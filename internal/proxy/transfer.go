package proxy

import (
	"io"
	"log/slog"
	"net"
	"sync"
)

func transferData(client *net.TCPConn, peer *net.TCPConn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go copyData(client, peer, &wg)
	go copyData(peer, client, &wg)

	wg.Wait()
}

func copyData(dest *net.TCPConn, src *net.TCPConn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func(dest *net.TCPConn) {
		err := dest.Close()
		if err != nil {
			slog.Error("Destination connection closing error",
				slog.Any("remote", dest.RemoteAddr()),
				slog.Any("err", err),
			)
		}
	}(dest)

	_, err := io.Copy(dest, src)
	if err != nil {
		slog.Error("Reading error",
			slog.Any("remote", dest.RemoteAddr()),
			slog.Any("err", err),
		)
	}
}
