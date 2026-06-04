//go:build !js
// +build !js

package webtorrent

import (
	"github.com/pion/webrtc/v4"
)

func GetPeerConnStats(pc *wrappedPeerConnection) (stats webrtc.StatsReport) {
	// Skip closed connections. pc.GetStats() on a closed PeerConnection makes pion
	// log "ice ERROR: Failed to get candidate pair stats: the agent is closed", and
	// the client polls these stats ~5x/sec via the file-status endpoint — which
	// floods the logs. wrappedPeerConnection embeds *webrtc.PeerConnection, so
	// ConnectionState() is available directly.
	if pc != nil && pc.ConnectionState() != webrtc.PeerConnectionStateClosed {
		stats = pc.GetStats()
	}
	return
}
