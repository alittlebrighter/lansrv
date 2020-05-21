# LanSrv
This is a simple wrapper for [zeroconf](https://github.com/grandcat/zeroconf).  There are two major functions: 
- running an mDNS server that advertises all local services
  - Example: `$ lansrv -dir /etc/systemd/system # scans systemd service files`
- a scanning tool to find all services on the local network.
  - Example: `$ lansrv -scan`

## Why?
I wanted a way to automatically cluster [NATS](https://nats.io/) so each service in my home automation system just has to communicate with the local NATS instance.  With LanSrv I just need to add the following section to the systemd service file that defines the NATS service:
```
[LanSrv]
Name=nats-node
Port=4222
Protocol=nats
```
Other nodes that want to cluster can then be run with the following command:
```bash
lansrv -scan -service nats-node | xargs nats-server -routes
```
