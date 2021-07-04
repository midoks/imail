package smtpd

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// Envelope holds a message
type Envelope struct {
	Sender     string
	Recipients []string
	Data       []byte
}

// AddReceivedLine prepends a Received header to the Data
func (env *Envelope) AddReceivedLine(peer Peer) {

	tlsDetails := ""

	tlsVersions := map[uint16]string{
		tls.VersionSSL30: "SSL3.0",
		tls.VersionTLS10: "TLS1.0",
		tls.VersionTLS11: "TLS1.1",
		tls.VersionTLS12: "TLS1.2",
		tls.VersionTLS13: "TLS1.3",
	}

	if peer.TLS != nil {
		version := "unknown"

		if val, ok := tlsVersions[peer.TLS.Version]; ok {
			version = val
		}

		cipher := tls.CipherSuiteName(peer.TLS.CipherSuite)

		tlsDetails = fmt.Sprintf(
			"\r\n\t(version=%s cipher=%s);",
			version,
			cipher,
		)
	}

	peerIP := ""
	if addr, ok := peer.Addr.(*net.TCPAddr); ok {
		peerIP = addr.IP.String()
	}

	line := wrap([]byte(fmt.Sprintf(
		"Received: from %s ([%s]) by %s with %s;%s\r\n\t%s\r\n",
		peer.HeloName,
		peerIP,
		peer.ServerName,
		peer.Protocol,
		tlsDetails,
		time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700 (MST)"),
	)))

	env.Data = append(env.Data, line...)

	// Move the new Received line up front

	copy(env.Data[len(line):], env.Data[0:len(env.Data)-len(line)])
	copy(env.Data, line)

}
