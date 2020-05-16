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

const helpText = `LanSrv
Usage: lansrv [commands] [options]
Commands:
  none     : runs the server
  help     : print this help message
  discover : scans the local network for services advertised via mDNS
  
Options:
  --scan-dir : Specifies the directory to walk to find LanSrv service configurations.
  --time     : Number of seconds to scan the local network for services.
  --service  : Only print results matching the service name.`

func main() {
	var scanDir string
	flag.StringVar(&scanDir, "scan-dir", "/etc/systemd/system",
		"Specifies the directory to walk to find LanSrv service configurations.")
	var port int
	flag.IntVar(&port, "port", 42424, "Port to run the server on.")
	var seconds int
	flag.IntVar(&seconds, "time", 5, "Number of seconds to scan the local network for services.")
	var service string
	flag.StringVar(&service, "service", "", "Only print results matching the service name.")
	flag.Parse()

	switch {
	case len(os.Args) > 1 && strings.Contains(os.Args[1], "help"):
		fmt.Println(helpText)
	case len(os.Args) > 1 && os.Args[len(os.Args)-1] == "discover":
		discover(seconds, service)
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

func discover(seconds int, service string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(seconds))
	defer cancel()

	networkAds, err := lansrv.ServicesLookup(ctx)
	if err != nil {
		fmt.Println("Failed to lookup services:", err)
	}

	if len(service) > 0 {
		svcEndpoints := make([]string, 0)
		for host, svcs := range networkAds {
			for _, svc := range svcs {
				if svc.Name() != service {
					continue
				}

				svcEndpoints = append(svcEndpoints, host+":"+svc.Port())
			}
		}

		fmt.Print(strings.Join(svcEndpoints, " "))
		return
	}

	data, _ := json.Marshal(networkAds)
	fmt.Println(string(data))
}
