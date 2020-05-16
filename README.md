# LanSrv
This is a simple wrapper for [zeroconf](https://github.com/grandcat/zeroconf).  There are two major functions: 
- running an mDNS server that advertises all local services
  - Example: `$ lansrv`
- a scanning tool to find all services on the local network.
  - Example: `$ lansrv -discover`

## Why?
I wanted a way to automatically cluster [NATS](https://nats.io/) so each service in my home automation system just has to communicate with the local NATS instance.  With LanSrv I just need to add the following section to the systemd service file that defines the NATS service:
```
[LanSrv]
Name=nats
Port=4222
```
Other nodes that want to cluster can then be run with the following command:
```bash
lansrv -discover -service nats | xargs nats-server -routes
```