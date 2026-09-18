// These build constraints are copied from webrtc's settingengine.go.
//go:build !js
// +build !js

package webtorrent

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/pion/logging"
	"github.com/pion/webrtc/v4"
)

// s is the SettingEngine shared by every PeerConnection this package creates.
// It is configured once from the environment: the cloud uploader sets the
// NAT1To1/port-range vars below; other processes keep pion's defaults except
// for the always-on virtual-interface filter (opt out with
// WEBRTC_GATHER_VIRTUAL_INTERFACES=1).
var s = newSettingEngine()

func newSettingEngine() webrtc.SettingEngine {
	se := webrtc.SettingEngine{
		// This could probably be done with better integration into anacrolix/log, but I'm not sure if
		// it's worth the effort.
		LoggerFactory: logging.NewDefaultLoggerFactory(),
	}

	// WEBRTC_NAT_1TO1_IP: advertise this single public IP as the host ICE
	// candidate, replacing whatever local-interface IPs pion would otherwise
	// enumerate (e.g. a cloud droplet's private anchor IP 10.x, which a browser
	// can't reach). With a reachable public host candidate, ICE can select the
	// direct browser<->uploader pair (host has higher priority than relay) instead
	// of falling back to the TURN relay. Only the cloud uploader sets this.
	if ip := os.Getenv("WEBRTC_NAT_1TO1_IP"); ip != "" {
		se.SetNAT1To1IPs([]string{ip}, webrtc.ICECandidateTypeHost)
	}

	// WEBRTC_UDP_PORT_MIN / _MAX: pin WebRTC to a fixed UDP port range instead of
	// random ephemeral ports. Not required when all UDP is open to the host; useful
	// only when the host-candidate port must be known in advance to be allowed
	// through a firewall. Both must be set and form a valid range, else pion's
	// default (the whole ephemeral range) is kept.
	if pmin, pmax := envPort("WEBRTC_UDP_PORT_MIN"), envPort("WEBRTC_UDP_PORT_MAX"); pmin > 0 && pmax >= pmin {
		// The only error is an invalid range, which the guard above already
		// excludes; swallow it so a misconfigured env var can't crash every
		// WebRTC-capable process at startup.
		_ = se.SetEphemeralUDPPortRange(pmin, pmax)
	}

	// Skip host-side container/VM plumbing when gathering host candidates. Every
	// gathered candidate is a live socket plus a read goroutine FOR EACH open
	// peer connection, and a long-running seeder keeps one open offer per
	// torrent — on a docker/libvirt-heavy host that was 42 candidates per offer,
	// none of them reachable by a remote peer (they are the host side of a
	// bridge; inside a container the interface is eth0, which stays eligible).
	// WEBRTC_GATHER_VIRTUAL_INTERFACES=1 restores pion's gather-everything
	// default for the exotic setups where a peer really sits behind one of these
	// (e.g. a browser inside a local NATed VM answering via virbr0).
	if os.Getenv("WEBRTC_GATHER_VIRTUAL_INTERFACES") != "1" {
		se.SetInterfaceFilter(func(name string) bool {
			return !isVirtualInterface(name)
		})
	}

	return se
}

// virtualInterfacePrefixes name interfaces that are the host's side of
// container/VM networking. Deliberately absent: VPN interfaces (wg*, tun*,
// tailscale*) — those carry reachable routes.
var virtualInterfacePrefixes = []string{
	"docker", "br-", "veth", "virbr", "vnet", "lxc", "lxd", "cni", "flannel", "kube",
}

func isVirtualInterface(name string) bool {
	for _, p := range virtualInterfacePrefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

// envPort parses a uint16 UDP port from the named env var, returning 0 when it is
// unset or unparseable so the caller falls back to pion's default behaviour.
func envPort(name string) uint16 {
	v := os.Getenv(name)
	if v == "" {
		return 0
	}
	n, err := strconv.ParseUint(v, 10, 16)
	if err != nil {
		return 0
	}
	return uint16(n)
}

type discardLoggerFactory struct{}

func (discardLoggerFactory) NewLogger(scope string) logging.LeveledLogger {
	return logging.NewDefaultLeveledLoggerForScope(scope, logging.LogLevelInfo, io.Discard)
}
