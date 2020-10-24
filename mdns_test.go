package lansrv

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanAdFromString(t *testing.T) {
	ad := new(LanAd)
	ad.FromString("proto://svc:42")

	assert.True(t, ad.EqualTo(&LanAd{Service: "svc", Port: 42, Protocol: "proto"}),
		fmt.Sprintf("service: %s, protocol: %s, port: %d", ad.Service, ad.Protocol, ad.Port))
}

func TestLanAdToFormattedString(t *testing.T) {
	ad := &LanAd{
		Service:  "svc",
		Address:  net.IPv4(192, 168, 1, 4),
		Port:     42,
		Protocol: "test",
	}

	format := Protocol + "://" + Address + ":" + Port
	assert.Equal(t, "test://192.168.1.4:42", ad.ToFormattedString(format), "Standard format failed.")

	format = "%" + Service + "%{{" + Address + "}}"
	assert.Equal(t, "%svc%{{192.168.1.4}}", ad.ToFormattedString(format), "Format with extra marker chars failed.")
}
