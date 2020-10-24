package lansrv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/grandcat/zeroconf"
)

const (
	service = "LanSrv"
	domain  = "local"
)

type LanAd struct {
	Service  string
	Address  net.IP `json:"-"`
	Port     int
	Protocol string
}

func (ad *LanAd) FromMap(adMap map[string]string) {
	if name, ok := adMap["Service"]; ok {
		ad.Service = name
	}

	if port, ok := adMap["Port"]; ok {
		portNum, _ := strconv.Atoi(port)
		ad.Port = portNum
	}

	if protocol, ok := adMap["Protocol"]; ok {
		ad.Protocol = protocol
	}
}

// FromString takes a string formatted {{protocol}}://{{service name}}:{{port}}
func (ad *LanAd) FromString(adStr string) {
	if protoSplitI := strings.Index(adStr, "://"); protoSplitI > -1 {
		ad.Protocol = adStr[:protoSplitI]
		adStr = adStr[protoSplitI+3:]
	}
	fmt.Println("adStr", adStr)

	hostPort := strings.Split(adStr, ":")
	if len(hostPort) >= 1 {
		ad.Service = hostPort[0]
	}
	if len(hostPort) >= 2 {
		ad.Port, _ = strconv.Atoi(hostPort[1])
	}
}

type LanAdFormatPart = string

const (
	marker = "%"

	Protocol LanAdFormatPart = marker + "pro" + marker
	Address                  = marker + "addr" + marker
	Port                     = marker + "port" + marker
	Service                  = marker + "svc" + marker
)

func (ad *LanAd) ToFormattedString(format string) string {
	builder := &strings.Builder{}
	fmtI := 0

	for fmtI < len(format) {
		search := format[fmtI:]
		toAppend := ""
		jump := 0
		switch {
		case strings.HasPrefix(search, Protocol):
			toAppend = ad.Protocol
			jump = len(Protocol)
		case strings.HasPrefix(search, Address):
			toAppend = ad.Address.String()
			jump = len(Address)
		case strings.HasPrefix(search, Port):
			toAppend = strconv.Itoa(ad.Port)
			jump = len(Address)
		case strings.HasPrefix(search, Service):
			toAppend = ad.Service
			jump = len(Service)
		default:
			toAppend = string(search[0])
			jump = 1
		}

		builder.WriteString(toAppend)
		fmtI += jump
	}

	return builder.String()
}

func (ad *LanAd) EqualTo(other *LanAd) bool {
	return ad.Service == other.Service && ad.Port == other.Port && ad.Protocol == other.Protocol && ad.Address.String() == other.Address.String()
}

func StartMdnsServer(ads []LanAd, port int) (*zeroconf.Server, error) {
	host, _ := os.Hostname()

	records := make([]string, len(ads))
	for i, ad := range ads {
		data, _ := json.Marshal(ad)
		records[i] = string(data)
	}

	return zeroconf.Register(host, service, domain, port, records, nil)
}

// ServicesLookup returns a map containing hostnames along with a list of LanAds published
// on that host.
func ServicesLookup(ctx context.Context, localhost bool) (map[string][]LanAd, error) {
	// Discover all services on the network (e.g. _workstation._tcp)
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Failed to initialize resolver:", err.Error()))
	}

	entries := make(chan *zeroconf.ServiceEntry)
	ads := make(map[string][]LanAd)
	localIPs := make(map[string]interface{})
	if !localhost {
		localIPs = hostIPs()
	}
	go func(results <-chan *zeroconf.ServiceEntry, store map[string][]LanAd) {
		for entry := range results {
			if len(entry.AddrIPv4) == 0 {
				continue
			}
			// will there ever be more than one?
			host := entry.AddrIPv4[0].String()
			if _, exists := localIPs[host]; exists {
				continue
			}

			if _, ok := store[host]; !ok {
				store[host] = make([]LanAd, 0)
			}

		new_ad:
			for _, adData := range entry.Text {
				ad := LanAd{Address: entry.AddrIPv4[0]}
				adData = strings.ReplaceAll(adData, "\\", "")
				if err := json.Unmarshal([]byte(adData), &ad); err != nil {
					fmt.Println("could not parse ad:", err)
					fmt.Println("adData:", adData)
					continue
				}

				for _, listed := range store[host] {
					if ad.EqualTo(&listed) {
						continue new_ad
					}
				}

				store[host] = append(store[host], ad)
			}
		}
	}(entries, ads)

	if err := resolver.Browse(ctx, service, domain, entries); err != nil {
		return nil, err
	}

	<-ctx.Done()

	return ads, nil
}

// this is stupid but it will work
func hostIPs() map[string]interface{} {
	allIPs := make(map[string]interface{})
	ifaces, _ := net.Interfaces()

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		addressCount := 0
		for _, addr := range addrs {
			addressCount++
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			allIPs[ip.String()] = struct{}{}
		}
	}
	return allIPs
}
