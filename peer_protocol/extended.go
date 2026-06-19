package peer_protocol

import (
	"net"
)

// http://www.bittorrent.org/beps/bep_0010.html
type (
	ExtendedHandshakeMessage struct {
		M    map[ExtensionName]ExtensionNumber `bencode:"m"`
		V    string                            `bencode:"v,omitempty"`
		Reqq int                               `bencode:"reqq,omitempty"`
		// The only mention of this I can find is in https://www.bittorrent.org/beps/bep_0011.html
		// for bit 0x01.
		Encryption bool `bencode:"e"`
		// BEP 9
		MetadataSize int `bencode:"metadata_size,omitempty"`
		// The local client port. It would be redundant for the receiving side of
		// a connection to send this.
		Port   int       `bencode:"p,omitempty"`
		YourIp CompactIp `bencode:"yourip,omitempty"`
		Ipv4   CompactIp `bencode:"ipv4,omitempty"`
		Ipv6   net.IP    `bencode:"ipv6,omitempty"`
		// Terashare extension (non-standard, ignored by other clients): a paid
		// restore token (HMAC, bound to the info_hash). The seeder of a paywalled
		// (expired-grace) share verifies it at handshake time and drops peers that
		// don't present a valid one — enforcing the restore paywall at the
		// connection level, independent of how the peer was discovered.
		CloudToken string `bencode:"ts_token,omitempty"`
	}

	ExtensionName   string
	ExtensionNumber uint8
)

const (
	// http://www.bittorrent.org/beps/bep_0011.html
	ExtensionNamePex ExtensionName = "ut_pex"

	ExtensionDeleteNumber ExtensionNumber = 0
)

func (me *ExtensionNumber) UnmarshalBinary(b []byte) error {
	*me = ExtensionNumber(b[0])
	return nil
}
