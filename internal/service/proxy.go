package service

import (
	"fmt"
	"net"
	"tcp-proxy/internal/model"
	"time"
)

type ProxyService struct {
	resolver *Resolver
}

func NewProxyService(resolver *Resolver) *ProxyService {
	return &ProxyService{
		resolver: resolver,
	}
}

func (p *ProxyService) Connect(addr string, port uint16, addrType byte) (net.Conn, error) {
	var targetAddr string

	switch addrType {
	case model.AddrTypeIPv4:
		// Direct IP connection
		targetAddr = fmt.Sprintf("%s:%d", addr, port)

	case model.AddrTypeDomain:
		// Resolve domain name
		ips, err := p.resolver.Resolve(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve domain %s: %w", addr, err)
		}
		if len(ips) == 0 {
			return nil, fmt.Errorf("no IP addresses found for domain %s", addr)
		}
		// Use first resolved IP
		targetAddr = fmt.Sprintf("%s:%d", ips[0].String(), port)

	default:
		return nil, fmt.Errorf("unsupported address type: %d", addrType)
	}

	// Connect to target with timeout
	conn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", targetAddr, err)
	}

	return conn, nil
}
