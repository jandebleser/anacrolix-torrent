//go:build !js
// +build !js

package webtorrent

import "testing"

func TestIsVirtualInterface(t *testing.T) {
	virtual := []string{
		"docker0", "br-8f2a1c", "veth1a2b3c", "virbr0", "vnet3",
		"lxcbr0", "lxdbr0", "cni0", "flannel.1", "kube-bridge",
	}
	for _, name := range virtual {
		if !isVirtualInterface(name) {
			t.Errorf("%s should be filtered from ICE candidate gathering", name)
		}
	}
	// Real NICs and VPNs carry reachable routes and must stay eligible; a
	// container's own eth0 must never match (only the host side is bridge-named).
	real := []string{"eth0", "enp3s0", "wlp2s0", "lo", "wg0", "tun0", "tailscale0", "en0"}
	for _, name := range real {
		if isVirtualInterface(name) {
			t.Errorf("%s must not be filtered from ICE candidate gathering", name)
		}
	}
}
