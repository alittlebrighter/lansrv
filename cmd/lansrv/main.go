package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alittlebrighter/lansrv"
)

func main() {
	walkDir := ""
	flag.StringVar(&walkDir, "dir", walkDir,
		"Specifies the directory to walk to find LanSrv service configurations.  For systemd this would normally be: /etc/systemd/system")
	port := 42424
	flag.IntVar(&port, "port", port, "Port to run the server on.")
	publishServices := ""
	flag.StringVar(&publishServices, "publish", publishServices, "Comma delimited list of services to advertise in the format `<protocol>://<service-name>:<port>` e.g. http://files:9999")
	scan := false
	flag.BoolVar(&scan, "scan", scan,
		"Scan the local network for services published by other LanSrv nodes.  If not set, the server will start.")
	seconds := 5
	flag.IntVar(&seconds, "time", seconds, "Number of seconds to scan the local network for services.")
	service := ""
	flag.StringVar(&service, "service", service, "Only print results matching the service name.")
	format := lansrv.Protocol + "://" + lansrv.Address + ":" + lansrv.Port
	flag.StringVar(&format, "format", format, `Print results in a custom format delimited by the delim flag.  Keys start and end with %.
Valid keys are pro=protocol, addr=IP address, port=port, svc=service.`)
	var delimiter string
	flag.StringVar(&delimiter, "delim", ",", "Delimiter to use when only printing specific service endpoints.")
	var localhost bool
	flag.BoolVar(&localhost, "localhost", false,
		"Include services hosted on this computer.")
	flag.Parse()

	switch {
	case scan:
		runDiscovery(seconds, service, format, delimiter, localhost)
	default:
		runServer(walkDir, strings.Split(publishServices, ","), port)
	}
}

func runServer(scanDir string, services []string, port int) {
	ads := make([]lansrv.LanAd, 0)

	if len(scanDir) > 0 {
		files := lansrv.GatherServiceConfigs(scanDir)
		ads = append(ads, lansrv.ParseServiceFiles(files)...)
	}

svcs_loop:
	for _, svc := range services {
		ad := new(lansrv.LanAd)
		ad.FromString(svc)

		for _, listed := range ads {
			if ad.EqualTo(&listed) {
				continue svcs_loop
			}
		}

		ads = append(ads, *ad)
	}

	if len(ads) == 0 {
		fmt.Println("No LanSrv configurations found.  Exiting now.")
		return
	}

	server, err := lansrv.StartMdnsServer(ads, port)
	if err != nil {
		fmt.Println("Failed to start server:", err)
		return
	}
	defer server.Shutdown()

	fmt.Printf("mDNS server started on %d.\n", port)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	<-sigc
}

func runDiscovery(seconds int, service, format, delimiter string, localhost bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(seconds))
	defer cancel()

	networkAds, err := lansrv.ServicesLookup(ctx, localhost)
	if err != nil {
		fmt.Println("Failed to lookup services:", err)
		return
	}

	if len(service) > 0 {
		svcEndpoints := make(map[string]interface{})
		for _, svcs := range networkAds {
			for _, svc := range svcs {
				if svc.Service != service {
					continue
				}

				svcEndpoints[svc.ToFormattedString(format)] = struct{}{}
			}
		}

		i := 0
		svcList := make([]string, len(svcEndpoints))
		for k := range svcEndpoints {
			svcList[i] = k
			i++
		}

		fmt.Print(strings.Join(svcList, delimiter))
		return
	}

	data, _ := json.Marshal(networkAds)
	fmt.Println(string(data))
}
