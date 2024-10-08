package protocol

import (
	"context"
	"net"
	"net/netip"

	"github.com/favonia/cloudflare-ddns/internal/ipnet"
	"github.com/favonia/cloudflare-ddns/internal/pp"
)

// Local detects the IP address by pretending to send out an UDP packet
// and using the source IP address assigned by the system. In most cases
// it will detect the IP address of the network interface toward the internet.
// (No actual UDP packets will be sent out.)
type Local struct {
	// Name of the detection protocol.
	ProviderName string

	// The target IP address of the UDP packet to be sent.
	RemoteUDPAddr map[ipnet.Type]string
}

// Name of the detection protocol.
func (p Local) Name() string {
	return p.ProviderName
}

// GetIP detects the IP address by pretending to send an UDP packet.
// (No actual UDP packets will be sent out.)
func (p Local) GetIP(_ context.Context, ppfmt pp.PP, ipNet ipnet.Type) (netip.Addr, Method, bool) {
	var invalidIP netip.Addr

	remoteUDPAddr, found := p.RemoteUDPAddr[ipNet]
	if !found {
		ppfmt.Noticef(pp.EmojiImpossible, "Unhandled IP network: %s", ipNet.Describe())
		return invalidIP, MethodUnspecified, false
	}

	conn, err := net.Dial(ipNet.UDPNetwork(), remoteUDPAddr)
	if err != nil {
		ppfmt.Noticef(pp.EmojiError, "Failed to detect a local %s address: %v", ipNet.Describe(), err)
		return invalidIP, MethodUnspecified, false
	}
	defer conn.Close()

	ip := conn.LocalAddr().(*net.UDPAddr).AddrPort().Addr() //nolint:forcetypeassert

	normalizedIP, ok := ipNet.NormalizeDetectedIP(ppfmt, ip)
	return normalizedIP, MethodPrimary, ok
}
