package service

import (
	"context"
	"net"
	"time"
)

type Resolver struct {
	resolver *net.Resolver
}

func NewResolver() *Resolver {
	return &Resolver{
		resolver: &net.Resolver{
			PreferGo: true,
		},
	}
}

func (r *Resolver) Resolve(domain string) ([]net.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ips, err := r.resolver.LookupIP(ctx, "ip4", domain)
	if err != nil {
		return nil, err
	}

	return ips, nil
}
