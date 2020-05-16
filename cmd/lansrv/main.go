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
	var scanDir string
	flag.StringVar(&scanDir, "scan-dir", "/etc/systemd/system",
		"Specifies the directory to walk to find LanSrv service configurations.")
	var port int
	flag.IntVar(&port, "port", 42424, "Port to run the server on.")
	var discover bool
	flag.BoolVar(&discover, "discover", false,
		"Scan the local network for services published by other LanSrv nodes.  If not set, the server will start.")
	var seconds int
	flag.IntVar(&seconds, "time", 5, "Number of seconds to scan the local network for services.")
	var service string
	flag.StringVar(&service, "service", "", "Only print results matching the service name.")
	var delimiter string
	flag.StringVar(&delimiter, "delim", ",", "Delimiter to use when only printing specific service endpoints.")
	flag.Parse()

	switch {
	case discover:
		runDiscovery(seconds, service, delimiter)
	default:
		runServer(scanDir, port)
	}
}

func runServer(scanDir string, port int) {
	files := lansrv.GatherServiceConfigs(scanDir)
	ads := lansrv.ParseServiceFiles(files)
	if len(files) == 0 {
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
