package main

import (
	"fmt"
	"net"
	"time"

	"github.com/xtaci/kcp-go/v5"
)

var (
	k          *kcp.KCP
	udpconn    net.PacketConn
	remoteAddr net.Addr
	conv       = 1
)

func main() {
	go func() {
		var err error
		udpconn, err = net.ListenPacket("udp", "10.60.91.7:8081")
		if err != nil {
			panic(err)
		}

		k = kcp.NewKCP(uint32(conv), output_callback)

		go recv(udpconn)

		time.Sleep(3 * time.Second)
		send(udpconn)
	}()

	time.Sleep(time.Second)
	go client()

	select {}
}

func client() {
	conn, err := net.Dial("udp", "10.60.91.7:8081")
	if err != nil {
		panic(err)
	}

	k := kcp.NewKCP(uint32(conv), func(buf []byte, size int) {
		n, err := conn.Write(buf[:size])
		if err != nil {
			panic(err)
		}
		if n != size {
			panic(n)
		}
	})

	go func() {
		time.Sleep(time.Second)
		k.Send([]byte("hello"))
		ticker := time.NewTicker(100 * time.Millisecond)
		for range ticker.C {
			k.Update()
		}
	}()

	udpbuf := make([]byte, 15000)
	kcpbuf := make([]byte, 15000)
	for {
		n, err := conn.Read(udpbuf)
		if err != nil {
			panic(err)
		}

		code := k.Input(udpbuf[:n], true, false)
		if code != 0 {
			panic(code)
		}

		n = k.PeekSize()
		if n > 0 {
			n = k.Recv(kcpbuf)
			if n <= 0 {
				panic(n)
			}
			fmt.Println("client recv ", kcpbuf[:n])
			fmt.Println("client recv n", n)
		}
	}
}

func recv(conn net.PacketConn) {
	udpbuf := make([]byte, 1500)
	kcpbuf := make([]byte, 1500)
	for {
		n, addr, err := conn.ReadFrom(udpbuf)
		if err != nil {
			panic(err)
		}
		remoteAddr = addr

		code := k.Input(udpbuf[:n], true, false)
		if code != 0 {
			panic(code)
		}

		n = k.PeekSize()
		if n > 0 {
			n = k.Recv(kcpbuf)
			if n <= 0 {
				panic(n)
			}
			fmt.Println("server recv ", kcpbuf[:n])
		}
	}
}

var (
	buf = make([]byte, 1376*10+1)
)

func init() {
	for i := 0; i < 1376*10; i++ {
		buf[i] = 8
	}
	for i := 1376 * 10; i < len(buf); i++ {
		buf[i] = 6
	}
}

func send(conn net.PacketConn) {
	go func() {
		k.Send(buf)
		ticker := time.NewTicker(100 * time.Millisecond)
		for range ticker.C {
			k.Update()
		}
	}()
}

func output_callback(buf []byte, size int) {
	n, err := udpconn.WriteTo(buf[:size], remoteAddr)
	if err != nil {
		panic(err)
	}
	if n != size {
		panic(n)
	}
}
