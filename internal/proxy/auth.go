package proxy

import (
	"fmt"
	"io"
	"net"
)

/*
+----+----------+----------+
|VER | NMETHODS | METHODS  |
+----+----------+----------+
| 1  |    1     | 1 to 255 |
+----+----------+----------+
*/
func authenticate(client *net.TCPConn) (byte, error) {
	// Read version
	version := make([]byte, 1)
	_, err := io.ReadFull(client, version)
	if err != nil {
		return AuthNoAccept, NewErrAuthRequestParsing("No socks version")
	}
	if version[0] != SOCKS5Version {
		return AuthNoAccept, NewErrAuthRequestParsing(fmt.Sprintf(
			"Socks version %v is expected, but not %v", SOCKS5Version, version[0]))
	}

	// Read method count
	methodCount := make([]byte, 1)
	_, err = io.ReadFull(client, methodCount)
	if err != nil {
		return AuthNoAccept, NewErrAuthRequestParsing("No authentication method count")
	}

	// Read auth methods
	methods := make([]byte, methodCount[0])
	actualMethodCount, err := io.ReadFull(client, methods)
	if err != nil {
		return AuthNoAccept, NewErrAuthRequestParsing(fmt.Sprintf(
			"Not enough authentication methods: Expected %v, received %v", methodCount, actualMethodCount))
	}

	for _, methods := range methods {
		if methods == AuthNone {
			// Found supported method
			return AuthNone, nil
		}
	}

	// Not found supported method
	return AuthNoAccept, NewErrAuthRequestParsing(fmt.Sprintf(
		"Unsupported auth methods, %v method supported", []byte{AuthNone}))
}

/*
+----+--------+
|VER | METHOD |
+----+--------+
| 1  |   1    |
+----+--------+
*/
func sendAuthReply(client *net.TCPConn, method byte) error {
	// Prepare reply message
	replyMsg := []byte{SOCKS5Version, method}

	// Send reply message
	_, err := client.Write(replyMsg)
	if err != nil {
		return NewErrAuthReplySending(err.Error())
	}
	return nil
}
