package in

import (
	"log/slog"
	"net"
	"tcp-proxy/internal/model"
	"tcp-proxy/internal/service"
)

type ServerAdapter struct {
	port         int
	proxyService *service.ProxyService
}

func NewServerAdapter(port int, proxyService *service.ProxyService) *ServerAdapter {
	return &ServerAdapter{
		port:         port,
		proxyService: proxyService,
	}
}

func (s *ServerAdapter) HandleConnection(clientConn net.Conn) {
	defer func() {
		if err := clientConn.Close(); err != nil {
			slog.Error("Client connection closing error", slog.Any("err", err))
		}
		slog.Info("Client connection closed", slog.Any("remote_addr", clientConn.RemoteAddr()))
	}()

	// 1. Handle authentication negotiation
	if err := s.handleAuth(clientConn); err != nil {
		slog.Error("Auth failed", slog.Any("err", err))
		return
	}

	// 2. Parse request
	req, replyCode, err := s.parseRequest(clientConn)
	if err != nil {
		slog.Error("Failed to parse request", slog.Any("err", err))
		err := s.sendReply(clientConn, replyCode, nil)
		if err != nil {
			slog.Error("Failed to send reply", slog.Any("err", err))
			return
		}
		return
	}

	// 3. Handle command
	if req.Command != model.CmdConnect {
		err := s.sendReply(clientConn, model.RepCommandNotSupported, nil)
		if err != nil {
			slog.Error("Failed to send reply", slog.Any("err", err))
			return
		}
		slog.Error("Unsupported command", slog.Int("command", int(req.Command)))
		return
	}

	// 4. Resolve address if needed and connect
	targetConn, reply := s.proxyService.Connect(req.DstAddr, req.DstPort, req.AddrType)
	if targetConn == nil {
		slog.Error("Failed to connect to target", slog.Any("err", err))
		err := s.sendReply(clientConn, reply, nil)
		if err != nil {
			slog.Error("Failed to send reply", slog.Any("err", err))
			return
		}
		return
	}
	defer func(targetConn net.Conn) {
		err := targetConn.Close()
		if err != nil {
			slog.Error("Target connection closing error", slog.Any("err", err))
		}
		slog.Info("Target connection closed", slog.Any("remote_addr", targetConn.RemoteAddr()))
	}(targetConn)

	// 5. Send success reply
	if err := s.sendReply(clientConn, model.RepSuccess, targetConn); err != nil {
		slog.Error("Failed to send reply", slog.Any("err", err))
		return
	}

	// 6. Relay data bidirectionally
	s.relay(clientConn, targetConn)
}
