package main

import (
	"bufio"
	"log"
	"net"
	"time"

	"github.com/xtaci/kcp-go/v5"
)

type marsHeader struct {
	HeaderLength  int32
	ClientVersion int32
	Cmd           int32
	Sequence      int32
	BodyLength    int32
}

func main() {
	if listener, err := kcp.Listen("10.60.91.7:8081"); err == nil {
		// spin-up the client
		go client()
		for {
			s, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			s.(*kcp.UDPSession).SetNoDelay(0, 40, 0, 0)
			writer = bufio.NewWriterSize(s, 16384)
			go handleEcho(s)
		}
	} else {
		log.Fatal(err)
	}
}

// handleEcho send back everything it received
func handleEcho(conn net.Conn) {

	// pb := msg.GameMsg{
	// 	ModId: 1,
	// 	MsgId: 1,
	// 	Data:  buf,
	// }
	// data, err := proto.Marshal(&pb)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// header.BodyLength = int32(len(data))
	// err = binary.Write(writer, binary.BigEndian, header)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// n, err := writer.Write(buf)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println(n)
	// err = writer.Flush()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	conn.(*kcp.UDPSession).Write(buf)
}

var buf []byte
var header *marsHeader
var writer *bufio.Writer

// 1400-24 = 1376
func init() {
	buf = make([]byte, 1376*10+1)
	for i := 0; i < 1376*10; i++ {
		buf[i] = 8
	}
	for i := 1376 * 10; i < len(buf); i++ {
		buf[i] = 6
	}

	header = &marsHeader{
		HeaderLength:  20,
		ClientVersion: 200,
		Cmd:           0,
		Sequence:      0,
		BodyLength:    0,
	}

}

func client() {
	// wait for server to become ready
	time.Sleep(time.Second)

	// dial to the echo server
	if sess, err := kcp.Dial("10.60.91.7:8081"); err == nil {
		sess.Write([]byte("hello"))
		recvbuf := make([]byte, 15000)
		for {

			// read back the data

			//if n, err := io.ReadFull(sess, buf); err == nil {
			//if n, err := sess.Read(recvbuf); err == nil {
			if n, err := sess.(*kcp.UDPSession).Read(recvbuf); err == nil {
				log.Println("recv:", recvbuf[:n])
				log.Println(n)
				recvbuf = recvbuf[n:]
			} else {
				log.Fatal(err)
			}
		}
	} else {
		log.Fatal(err)
	}
}
