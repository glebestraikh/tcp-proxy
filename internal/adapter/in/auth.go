package in

import (
	"fmt"
	"io"
	"net"
	"tcp-proxy/internal/model"
)

func (s *ServerAdapter) handleAuth(conn net.Conn) error {
	// Read version
	versionBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, versionBuf); err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}
	version := versionBuf[0]
	if version != model.SOCKS5Version {
		return fmt.Errorf("unsupported SOCKS version: %d", version)
	}

	// Read nMethods
	nMethodsBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, nMethodsBuf); err != nil {
		return fmt.Errorf("failed to read number of methods: %w", err)
	}
	nMethods := nMethodsBuf[0]
	if nMethods == 0 {
		return fmt.Errorf("no authentication methods provided")
	}

	// Read methods
	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return fmt.Errorf("failed to read auth methods: %w", err)
	}

	// Check if NO AUTHENTICATION is supported
	noAuthSupported := false
	for _, method := range methods {
		if method == model.AuthNone {
			noAuthSupported = true
			break
		}
	}

	// Send response
	response := []byte{model.SOCKS5Version, model.AuthNone}
	if !noAuthSupported {
		response[1] = model.AuthNoAccept
	}

	if _, err := conn.Write(response); err != nil {
		return fmt.Errorf("failed to send auth response: %w", err)
	}

	if !noAuthSupported {
		return fmt.Errorf("no acceptable auth method")
	}

	return nil
}
