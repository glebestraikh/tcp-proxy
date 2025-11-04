package service

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	"tcp-proxy/internal/protocol"
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

func (p *ProxyService) Connect(addr string, port uint16, addrType byte) (net.Conn, byte) {
	var targetAddr string

	switch addrType {
	case protocol.AddrTypeIPv4:
		targetAddr = fmt.Sprintf("%s:%d", addr, port)
	case protocol.AddrTypeDomain:
		ips, err := p.resolver.Resolve(addr)
		if err != nil {
			return nil, protocol.RepHostUnreachable
		}
		if len(ips) == 0 {
			return nil, protocol.RepHostUnreachable
		}
		targetAddr = fmt.Sprintf("%s:%d", ips[0].String(), port)
	default:
		return nil, protocol.RepAddrTypeNotSupported
	}

	conn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			if opErr.Temporary() {
				return nil, protocol.RepTTLExpired
			}
			switch {
			case errors.Is(opErr.Err, syscall.ECONNREFUSED):
				return nil, protocol.RepConnectionRefused
			case errors.Is(opErr.Err, syscall.ENETUNREACH):
				return nil, protocol.RepNetworkUnreachable
			case errors.Is(opErr.Err, syscall.EHOSTUNREACH):
				return nil, protocol.RepHostUnreachable
			default:
				return nil, protocol.RepGeneralFailure
			}
		}
		return nil, protocol.RepGeneralFailure
	}

	return conn, protocol.RepSuccess
}
