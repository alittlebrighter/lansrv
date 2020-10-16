package lansrv

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanAdFromString(t *testing.T) {
	ad := new(LanAd)
	ad.FromString("proto://svc:42")

	assert.True(t, ad.EqualTo(&LanAd{Name: "svc", Port: 42, Protocol: "proto"}),
		fmt.Sprintf("name: %s, proto: %s, port: %d", ad.Name, ad.Protocol, ad.Port))
}
