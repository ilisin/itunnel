package server

import (
    "fmt"

    "github.com/ilisin/itunnel/conn"
    "github.com/ilisin/itunnel/log"
    "github.com/ilisin/itunnel/msg"
    "github.com/ilisin/itunnel/util"
)

func startTcpListener(addr string) (listener *conn.Listener) {
    // bind/listen for incoming connections
    var err error
    if listener, err = conn.Listen(addr, "tcp"); err != nil {
        panic(err)
    }

    proto := "tcp"

    log.Info("Listening for public %s connections on %v", proto, listener.Addr.String())
    go func() {
        for conn := range listener.Conns {
            go tcpHandler(conn)
        }
    }()

    return
}

func tcpHandler(c conn.Conn) {
    defer c.Close()
    defer func() {
        if r := recover(); r != nil {
            c.Warn("HandlePublicConnection failed with error %v", r)
        }
    }()

    // multiplex to find the right backend host
    c.Debug("Found hostname %s in request", c.RemoteAddr())

    //regKey := fmt.Sprintf("%s://%v", "tcp",c.LocalAddr().String())
    //fmt.Println(regKey)
    regKey := tcpRegKey

    tunnel := tunnelRegistry.Get(regKey) // 固定
    if tunnel == nil {
        c.Info("No tunnel found for hostname tcp")
        c.Write([]byte("Not F"))
        return
    }

    //startTime := time.Now()

    var proxyConn conn.Conn
    var err error
    for i := 0; i < (2 * proxyMaxPoolSize); i++ {
        // get a proxy connection
        if proxyConn, err = tunnel.ctl.GetProxy(); err != nil {
            tunnel.Warn("Failed to get proxy connection: %v", err)
            return
        }
        defer proxyConn.Close()
        tunnel.Info("Got proxy connection %s", proxyConn.Id())
        proxyConn.AddLogPrefix(tunnel.Id())

        // tell the client we're going to start using this proxy connection
        startPxyMsg := &msg.StartProxy{
            Url:        tunnel.url,
            ClientAddr: c.RemoteAddr().String(),
        }

        if err = msg.WriteMsg(proxyConn, startPxyMsg); err != nil {
            proxyConn.Warn("Failed to write StartProxyMessage: %v, attempt %d", err, i)
            proxyConn.Close()
        } else {
            // success
            break
        }
    }

    if err != nil {
        // give up
        c.Error("Too many failures starting proxy connection")
        return
    }

    util.PanicToError(func() { tunnel.ctl.out <- &msg.ReqProxy{} })

    for{
        bytesIn, bytesOut := conn.Join(c, proxyConn)
        if bytesIn == 0 {
            break
        }
        fmt.Println("bytesIn:",bytesIn,"bytesOut:",bytesOut)
    }
}
