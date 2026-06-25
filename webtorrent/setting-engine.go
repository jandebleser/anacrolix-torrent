// These build constraints are copied from webrtc's settingengine.go.
//go:build !js
// +build !js

package webtorrent

import (
	"io"
	"os"
	"strconv"

	"github.com/pion/logging"
	"github.com/pion/webrtc/v4"
)

// s is the SettingEngine shared by every PeerConnection this package creates. It
// is configured once from the environment so only processes that opt in are
// affected: the cloud uploader sets the vars below; desktop clients and the
// tracker leave them unset and keep pion's defaults (host candidates gathered
// from every local interface, random ephemeral UDP ports, no NAT1To1).
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

	return se
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
