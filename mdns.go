package lansrv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/grandcat/zeroconf"
)

const (
	service = "LanSrv"
	domain  = "local"
)

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
func ServicesLookup(ctx context.Context) (map[string][]LanAd, error) {
	// Discover all services on the network (e.g. _workstation._tcp)
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Failed to initialize resolver:", err.Error()))
	}

	entries := make(chan *zeroconf.ServiceEntry)
	ads := make(map[string][]LanAd)
	go func(results <-chan *zeroconf.ServiceEntry, store map[string][]LanAd) {
		for entry := range results {
			if len(entry.AddrIPv4) == 0 {
				continue
			}
			// will there ever be more than one?
			host := entry.AddrIPv4[0].String()

			if _, ok := store[host]; !ok {
				store[host] = make([]LanAd, 0)
			}

			for _, adData := range entry.Text {
				ad := LanAd{}
				adData = strings.ReplaceAll(adData, "\\", "")
				if err := json.Unmarshal([]byte(adData), &ad); err != nil {
					fmt.Println("could not parse ad:", err)
					fmt.Println("adData:", adData)
					continue
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
