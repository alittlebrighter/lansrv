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
	var walkDir string
	flag.StringVar(&walkDir, "dir", "",
		"Specifies the directory to walk to find LanSrv service configurations.  For systemd this would normally be: /etc/systemd/system")
	var port int
	flag.IntVar(&port, "port", 42424, "Port to run the server on.")
	var publishServices string
	flag.StringVar(&publishServices, "publish", "", "Comma delimited list of services to advertise in the format `<protocol>://<service-name>:<port>` e.g. http://files:9999")
	var scan bool
	flag.BoolVar(&scan, "scan", false,
		"Scan the local network for services published by other LanSrv nodes.  If not set, the server will start.")
	var seconds int
	flag.IntVar(&seconds, "time", 5, "Number of seconds to scan the local network for services.")
	var service string
	flag.StringVar(&service, "service", "", "Only print results matching the service name.")
	var delimiter string
	flag.StringVar(&delimiter, "delim", ",", "Delimiter to use when only printing specific service endpoints.")
	flag.Parse()

	switch {
	case scan:
		runDiscovery(seconds, service, delimiter)
	default:
		runServer(walkDir, strings.Split(publishServices, ","), port)
	}
}

func runServer(scanDir string, services []string, port int) {
	ads := make([]lansrv.LanAd, 0)

	if len(scanDir > 0) {
		files := lansrv.GatherServiceConfigs(scanDir)
		ads = append(ads, lansrv.ParseServiceFiles(files)...)
	}

	for _, svc := range services {
		ad := new(lansrv.LanAd)
		ad.FromString(svc)

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

func runDiscovery(seconds int, service, delimiter string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(seconds))
	defer cancel()

	networkAds, err := lansrv.ServicesLookup(ctx)
	if err != nil {
		fmt.Println("Failed to lookup services:", err)
		return
	}

	if len(service) > 0 {
		svcEndpoints := make([]string, 0)
		for host, svcs := range networkAds {
			for _, svc := range svcs {
				if svc.Name != service {
					continue
				}

				var endpoint strings.Builder
				if len(svc.Protocol) > 0 {
					endpoint.WriteString(svc.Protocol + "://")
				}
				endpoint.WriteString(host)
				if svc.Port > 0 {
					endpoint.WriteString(fmt.Sprintf(":%d", svc.Port))
				}
				svcEndpoints = append(svcEndpoints, endpoint.String())
			}
		}

		fmt.Print(strings.Join(svcEndpoints, delimiter))
		return
	}

	data, _ := json.Marshal(networkAds)
	fmt.Println(string(data))
}
