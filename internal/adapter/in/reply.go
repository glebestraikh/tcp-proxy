package in

import (
	"encoding/binary"
	"net"
	"tcp-proxy/internal/model"
)

func (s *ServerAdapter) sendReply(conn net.Conn, replyCode byte, targetConn net.Conn) error {
	reply := []byte{
		model.SOCKS5Version,
		replyCode,
		model.Reserved,
		model.AddrTypeIPv4,
		0, 0, 0, 0, // Bind address (0.0.0.0)
		0, 0, // Bind port (0)
	}

	if targetConn != nil {
		if addr, ok := targetConn.LocalAddr().(*net.TCPAddr); ok {
			if ipv4 := addr.IP.To4(); ipv4 != nil {
				copy(reply[4:8], ipv4)
				binary.BigEndian.PutUint16(reply[8:10], uint16(addr.Port))
			}
		}
	}

	_, err := conn.Write(reply)
	return err
}
