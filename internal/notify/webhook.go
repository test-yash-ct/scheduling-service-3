package notify

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func ProbeCallback(target string, allowlist []string) ([]byte, int, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, 0, err
	}
	if u.Scheme != "https" || u.Host == "" {
		return nil, 0, errors.New("https allowlisted url required")
	}
	host := strings.ToLower(u.Hostname())
	if !hostAllowed(host, allowlist) {
		return nil, 0, errors.New("host not allowlisted")
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, 0, err
	}
	var safeIPs []net.IP
	for _, ip := range ips {
		if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || isMetadataIP(ip) {
			return nil, 0, errors.New("blocked destination")
		}
		safeIPs = append(safeIPs, ip)
	}
	if len(safeIPs) == 0 {
		return nil, 0, errors.New("blocked destination")
	}

	port := u.Port()
	if port == "" {
		port = "443"
	}

	dialer := &net.Dialer{Timeout: 8 * time.Second}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			for _, ip := range safeIPs {
				conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
				if err == nil {
					return conn, nil
				}
			}
			return nil, errors.New("connection failed")
		},
		TLSClientConfig: &tls.Config{
			ServerName:         host,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		},
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Timeout:   8 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("redirects disabled")
		},
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return b, resp.StatusCode, err
}

func hostAllowed(host string, allowlist []string) bool {
	for _, a := range allowlist {
		a = strings.ToLower(strings.TrimSpace(a))
		if a != "" && host == a {
			return true
		}
	}
	return false
}

func isMetadataIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4.Equal(net.ParseIP("169.254.169.254"))
	}
	return false
}

func ValidateAllowlistConfigured(allowlist []string) error {
	if len(allowlist) == 0 {
		return fmt.Errorf("WEBHOOK_ALLOWLIST is required")
	}
	return nil
}
