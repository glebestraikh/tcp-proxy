package in

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"tcp-proxy/internal/model"
)

func (s *ServerAdapter) parseRequest(conn net.Conn) (*model.Request, error) {
	version := make([]byte, 1)
	if _, err := io.ReadFull(conn, version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}
	if version[0] != model.SOCKS5Version {
		return nil, fmt.Errorf("unsupported SOCKS version: %d", version[0])
	}

	command := make([]byte, 1)
	if _, err := io.ReadFull(conn, command); err != nil {
		return nil, fmt.Errorf("failed to read command: %w", err)
	}

	reserved := make([]byte, 1)
	if _, err := io.ReadFull(conn, reserved); err != nil {
		return nil, fmt.Errorf("failed to read reserved byte: %w", err)
	}

	addrType := make([]byte, 1)
	if _, err := io.ReadFull(conn, addrType); err != nil {
		return nil, fmt.Errorf("failed to read address type: %w", err)
	}

	req := &model.Request{
		Version:  version[0],
		Command:  command[0],
		Reserved: reserved[0],
		AddrType: addrType[0],
	}

	switch req.AddrType {
	case model.AddrTypeIPv4:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return nil, fmt.Errorf("failed to read IPv4 address: %w", err)
		}
		req.DstAddr = net.IP(addr).String()

	case model.AddrTypeDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return nil, fmt.Errorf("failed to read domain length: %w", err)
		}
		domain := make([]byte, lenBuf[0])
		if _, err := io.ReadFull(conn, domain); err != nil {
			return nil, fmt.Errorf("failed to read domain: %w", err)
		}
		req.DstAddr = string(domain)

	default:
		return nil, fmt.Errorf("unsupported address type: %d", req.AddrType)
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return nil, fmt.Errorf("failed to read port: %w", err)
	}
	req.DstPort = binary.BigEndian.Uint16(portBuf)

	return req, nil
}
