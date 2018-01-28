package rtmp

import (
	"bufio"
	"log"
	"net"
	"time"
)

func ListenAndServe(addr string) error {
	server := &Server{Addr: addr}
	return server.ListenAndServe()
}

type Server struct {
	Addr     string      // If empty, use ":1935".
	ErrorLog *log.Logger // If nil, logging goes to os.Stderr.
}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":1935"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()
	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		rw, e := l.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				srv.logf("rtmp: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		c := srv.newConn(rw)
		go c.serve()
	}
}

func (srv *Server) newConn(nc net.Conn) *conn {
	return &conn{
		netconn:  nc,
		server:   srv,
		bufr:     bufio.NewReaderSize(nc, 1024*64),
		bufw:     bufio.NewWriterSize(nc, 1024*64),
		readbuf:  make([]byte, 4096),
		writebuf: make([]byte, 4096),
	}
}

func (srv *Server) logf(format string, args ...interface{}) {
	if srv.ErrorLog != nil {
		srv.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}
