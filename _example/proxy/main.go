package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/bradfitz/tcpproxy"
)

func main() {
	var lAddr, rAddr string
	flag.StringVar(&lAddr, "laddr", ":1935", `proxy local address`)
	flag.StringVar(&rAddr, "raddr", "", `proxy remote address`)
	flag.Parse()
	if rAddr == "" {
		log.Fatal("Please specify remote address via raddr option")
	}

	log.Println("Listening: " + lAddr)
	log.Println("Proxying: " + rAddr)

	var p tcpproxy.Proxy
	p.AddRoute(lAddr, To(rAddr))
	log.Fatal(p.Run())
}

func To(addr string) *DialProxy {
	return &DialProxy{Addr: addr}
}

// DialProxy implements Target by dialing a new connection to Addr
// and then proxying data back and forth with logging to Stderr.
//
// The To func is a shorthand way of creating a DialProxy.
type DialProxy struct {
	Addr            string
	KeepAlivePeriod time.Duration
	DialTimeout     time.Duration
	DialContext     func(ctx context.Context, network, address string) (net.Conn, error)
	OnDialError     func(src net.Conn, dstDialErr error)
}

func (dp *DialProxy) HandleConn(src net.Conn) {
	ctx := context.Background()
	var cancel context.CancelFunc
	if dp.DialTimeout >= 0 {
		ctx, cancel = context.WithTimeout(ctx, dp.dialTimeout())
	}
	dst, err := dp.dialContext()(ctx, "tcp", dp.Addr)
	if cancel != nil {
		cancel()
	}
	if err != nil {
		dp.onDialError()(src, err)
		return
	}
	defer src.Close()
	defer dst.Close()
	if ka := dp.keepAlivePeriod(); ka > 0 {
		if c, ok := tcpproxy.UnderlyingConn(src).(*net.TCPConn); ok {
			c.SetKeepAlive(true)
			c.SetKeepAlivePeriod(ka)
		}
		if c, ok := dst.(*net.TCPConn); ok {
			c.SetKeepAlive(true)
			c.SetKeepAlivePeriod(ka)
		}
	}
	errc := make(chan error, 1)
	go dp.proxyCopy(errc, src, dst, false)
	go dp.proxyCopy(errc, dst, src, true)
	<-errc
}

func (dp *DialProxy) proxyCopy(errc chan<- error, dst io.Writer, src io.Reader, toDst bool) {
	tee := debugReader(src, os.Stderr, toDst)
	_, err := io.Copy(dst, tee)
	errc <- err
}

func (dp *DialProxy) keepAlivePeriod() time.Duration {
	if dp.KeepAlivePeriod != 0 {
		return dp.KeepAlivePeriod
	}
	return time.Minute
}

func (dp *DialProxy) dialTimeout() time.Duration {
	if dp.DialTimeout > 0 {
		return dp.DialTimeout
	}
	return 10 * time.Second
}

var defaultDialer = new(net.Dialer)

func (dp *DialProxy) dialContext() func(ctx context.Context, network, address string) (net.Conn, error) {
	if dp.DialContext != nil {
		return dp.DialContext
	}
	return defaultDialer.DialContext
}

func (dp *DialProxy) onDialError() func(src net.Conn, dstDialErr error) {
	if dp.OnDialError != nil {
		return dp.OnDialError
	}
	return func(src net.Conn, dstDialErr error) {
		log.Printf("tcpproxy: for incoming conn %v, error dialing %q: %v", src.RemoteAddr().String(), dp.Addr, dstDialErr)
		src.Close()
	}
}

func debugReader(r io.Reader, w io.Writer, toDst bool) io.Reader {
	return &debugTeeReader{r, w, toDst}
}

type debugTeeReader struct {
	r     io.Reader
	w     io.Writer
	toDst bool
}

func (t *debugTeeReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if t.toDst {
			fmt.Fprintln(t.w, ">>>>>>>>>>>>>>>>>>>> Source -> Destination >>>>>>>>>>>>>>>>>>>")
		} else {
			fmt.Fprintln(t.w, "<<<<<<<<<<<<<<<<<<<< Source <- Destination <<<<<<<<<<<<<<<<<<<")
		}
		fmt.Fprintln(t.w, hex.Dump(p[:n]))
	}
	return
}
