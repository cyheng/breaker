package netio

import (
	"io"
	"net"
	"sync"
)

func StartTunnel(src, dest net.Conn) {
	var wait sync.WaitGroup
	pipe := func(to io.ReadWriteCloser, from io.ReadWriteCloser) {
		defer to.Close()
		defer from.Close()
		defer wait.Done()
		io.Copy(to, from)
	}

	wait.Add(2)
	go pipe(src, dest)
	go pipe(dest, src)
	wait.Wait()
}
