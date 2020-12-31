package lansrv

import (
	"context"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/grandcat/zeroconf"
)

type ZeroConfActor struct {
	server *zeroconf.Server
	Ads    []LanAd
}

type Discover struct {
	Localhost bool
	Timeout   time.Duration
}

type LanAdsDiscovery struct {
	Ads map[string][]LanAd
	Err error
}

type Broadcast struct {
	Name string
}

func (state *ZeroConfActor) Receive(actorCtx actor.Context) {
	switch msg := actorCtx.Message().(type) {
	case Discover:
		ctx, cancel := context.WithTimeout(context.Background(), msg.Timeout)
		defer cancel()

		svcs, err := ServicesLookup(ctx, msg.Localhost)

		actorCtx.Respond(LanAdsDiscovery{svcs, err})
	case Broadcast:

	}
}
