package fwebsocket

import (
	"errors"
	"net"
	"syscall"

	"github.com/wundergraph/graphql-go-tools/v2/pkg/netpoll"
)

func isReadTimeout(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}

func socketFd(conn net.Conn) int {
	if con, ok := conn.(syscall.Conn); ok {
		raw, err := con.SyscallConn()
		if err != nil {
			return 0
		}
		sfd := 0
		_ = raw.Control(func(fd uintptr) {
			sfd = int(fd)
		})
		return sfd
	}
	if con, ok := conn.(netpoll.ConnImpl); ok {
		return con.GetFD()
	}
	return 0
}
