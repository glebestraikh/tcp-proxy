package service

import (
	"errors"
	"fmt"
	"net"
	"syscall"
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

func (p *ProxyService) Connect(addr string, port uint16, addrType byte) (net.Conn, byte) {
	var targetAddr string

	switch addrType {
	case model.AddrTypeIPv4:
		targetAddr = fmt.Sprintf("%s:%d", addr, port)
	case model.AddrTypeDomain:
		ips, err := p.resolver.Resolve(addr)
		if err != nil {
			return nil, model.RepHostUnreachable
		}
		if len(ips) == 0 {
			return nil, model.RepHostUnreachable
		}
		targetAddr = fmt.Sprintf("%s:%d", ips[0].String(), port)
	default:
		return nil, model.RepAddrTypeNotSupported
	}

	conn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			if opErr.Temporary() {
				return nil, model.RepTTLExpired
			}
			switch {
			case errors.Is(opErr.Err, syscall.ECONNREFUSED):
				return nil, model.RepConnectionRefused
			case errors.Is(opErr.Err, syscall.ENETUNREACH):
				return nil, model.RepNetworkUnreachable
			case errors.Is(opErr.Err, syscall.EHOSTUNREACH):
				return nil, model.RepHostUnreachable
			default:
				return nil, model.RepGeneralFailure
			}
		}
		return nil, model.RepGeneralFailure
	}

	return conn, model.RepSuccess
}
