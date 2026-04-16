package certbootstrap

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

func ResolveTargetURL(args []string) (string, error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "--url=") {
			return normalizeURL(strings.TrimPrefix(a, "--url="))
		}
		if a == "--url" && i+1 < len(args) {
			return normalizeURL(args[i+1])
		}
	}

	if u := os.Getenv("JFROG_URL"); u != "" {
		return normalizeURL(u)
	}

	return "", nil
}

func normalizeURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func GetVerifiedChainIfWindowsTrusts(rawURL string) ([]*x509.Certificate, bool, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, false, err
	}
	if u.Scheme != "https" {
		return nil, false, nil
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "443"
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 5 * time.Second},
		"tcp",
		net.JoinHostPort(host, port),
		&tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
		},
	)
	if err != nil {
		return nil, false, nil
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.VerifiedChains) == 0 || len(state.VerifiedChains[0]) == 0 {
		return nil, false, nil
	}

	return state.VerifiedChains[0], true, nil
}