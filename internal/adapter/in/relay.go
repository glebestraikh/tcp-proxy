package in

import (
	"io"
	"log/slog"
	"net"
	"sync"
)

func (s *ServerAdapter) relay(client, target net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go s.copyData(target, client, &wg)
	go s.copyData(client, target, &wg)

	wg.Wait()
}

func (s *ServerAdapter) copyData(dest, src net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	_, err := io.Copy(dest, src)
	if err != nil {
		slog.Debug("Connection interrupted",
			slog.Any("src", src.RemoteAddr()),
			slog.Any("dest", dest.RemoteAddr()),
			slog.Any("err", err),
		)
	}
}
