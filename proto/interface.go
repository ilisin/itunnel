package proto

import (
	"github.com/ilisin/itunnel/conn"
)

type Protocol interface {
	GetName() string
	WrapConn(conn.Conn, interface{}) conn.Conn
}
