package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
)

func connectCommand(client *net.TCPConn) (*net.TCPConn, byte, error) {
	// Check version
	version := make([]byte, 1)
	_, err := io.ReadFull(client, version)
	if err != nil {
		return nil, RepConnectionNotAllowed, NewErrCommandRequestParsing("No socks version")
	}
	if version[0] != SOCKS5Version {
		return nil, RepConnectionNotAllowed, NewErrCommandRequestParsing(fmt.Sprintf(
			"Socks version %v is expected, but not %v", SOCKS5Version, version[0]))
	}

	// Check command
	command := make([]byte, 1)
	_, err = io.ReadFull(client, command)
	if err != nil {
		return nil, RepConnectionNotAllowed, NewErrCommandRequestParsing("No command")
	}
	if command[0] != CmdConnect {
		return nil, RepCommandNotSupported, NewErrCommandRequestParsing(fmt.Sprintf(
			"Unsupported command %v, %v command supported", command[0], CmdConnect))
	}

	// Check reserved byte
	reservedByte := make([]byte, 1)
	_, err = io.ReadFull(client, reservedByte)
	if err != nil {
		return nil, RepConnectionNotAllowed, NewErrCommandRequestParsing("No reserved byte")
	}
	if reservedByte[0] != Rsv {
		return nil, RepConnectionNotAllowed, NewErrCommandRequestParsing(fmt.Sprintf(
			"Reserved byte must be set to %v, but not %v", Rsv, reservedByte[0]))
	}

	// Check address type
	addressType := make([]byte, 1)
	_, err = io.ReadFull(client, addressType)
	if err != nil {
		return nil, RepConnectionNotAllowed, NewErrCommandRequestParsing("No address type")
	}
	switch addressType[0] {
	case AddrTypeIPv4:
		ipv4, port, err := readIpv4AndPort(client)
		if err != nil {
			return nil, RepConnectionNotAllowed, err
		}
		return ipv4Connect(ipv4, port)
	case AddrTypeDomain:
		domainName, port, err := readDomainNameAndPort(client)
		if err != nil {
			return nil, RepConnectionNotAllowed, err
		}
		return domainNameConnect(domainName, port, client)
	default:
		return nil, RepAddrTypeNotSupported, NewErrCommandRequestParsing(fmt.Sprintf(
			"Unsupported address type %v, %v is supported", addressType[0],
			[]byte{AddrTypeIPv4, AddrTypeDomain}))
	}
}

func readIpv4AndPort(client *net.TCPConn) (net.IP, int, error) {
	// Check ipv4
	ip := make([]byte, 4)
	_, err := io.ReadFull(client, ip)
	if err != nil {
		return nil, -1, NewErrCommandRequestParsing("No ipv4 address")
	}

	// Check port
	portBytes := make([]byte, 2)
	_, err = io.ReadFull(client, portBytes)
	if err != nil {
		return nil, -1, NewErrCommandRequestParsing("No port")
	}
	port := int(binary.BigEndian.Uint16(portBytes))

	return ip, port, nil
}

func ipv4Connect(ipv4 net.IP, port int) (*net.TCPConn, byte, error) {
	// Connect to peer
	tcpAddr := &net.TCPAddr{
		IP:   ipv4,
		Port: port,
	}

	peer, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok {
			if opErr.Temporary() {
				return nil, RepTTLExpired, NewErrPeerConnectionCreating(err.Error())
			}
			if opErr.Err.Error() == "network is unreachable" {
				return nil, RepNetworkUnreachable, NewErrPeerConnectionCreating(err.Error())
			}
			if opErr.Err.Error() == "no route to host" {
				return nil, RepHostUnreachable, NewErrPeerConnectionCreating(err.Error())
			}
			if opErr.Err.Error() == "connection refused" {
				return nil, RepConnectionRefused, NewErrPeerConnectionCreating(err.Error())
			}
		}
		return nil, RepGeneralFailure, NewErrPeerConnectionCreating(err.Error())
	}

	return peer, RepSuccess, nil
}

func readDomainNameAndPort(client *net.TCPConn) (string, int, error) {
	// Check domain name size
	domainNameSize := make([]byte, 1)
	_, err := io.ReadFull(client, domainNameSize)
	if err != nil {
		return "", -1, NewErrCommandRequestParsing("No domain name size")
	}

	// Check domain name
	domainName := make([]byte, domainNameSize[0])
	_, err = io.ReadFull(client, domainName)
	if err != nil {
		return "", -1, NewErrCommandRequestParsing("No domain name")
	}

	// Check port
	portBytes := make([]byte, 2)
	_, err = io.ReadFull(client, portBytes)
	if err != nil {
		return "", -1, NewErrCommandRequestParsing("No port")
	}
	port := int(binary.BigEndian.Uint16(portBytes))

	return string(domainName), port, nil
}

func domainNameConnect(domainName string, port int, client *net.TCPConn) (*net.TCPConn, byte, error) {
	// Resolve domain name
	ips, err := net.LookupIP(domainName)
	if err != nil {
		return nil, RepHostUnreachable, NewErrDNSResolving(err.Error())
	}
	slog.Debug("Domain name resolved",
		slog.Any("client", client.RemoteAddr()),
		slog.String("domain", domainName),
		slog.Any("ips", ips),
	)

	// Try connecting to each ipv4 address
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			peer, reply, err := ipv4Connect(ipv4, port)
			if err == nil {
				// Found working ip address
				return peer, reply, nil
			}
		}
	}

	// Not found working ipv4 address
	return nil, RepHostUnreachable,
		NewErrDNSResolving("No hosts with IPv6 addresses or working IPv4 addresses")
}

func sendCommandReply(client *net.TCPConn, reply byte) error {
	// Create message
	replyMsg := []byte{
		SOCKS5Version, reply, Rsv, AddrTypeIPv4, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	// Send reply
	_, err := client.Write(replyMsg)
	if err != nil {
		return NewErrCommandReplySending(err.Error())
	}
	return nil
}
