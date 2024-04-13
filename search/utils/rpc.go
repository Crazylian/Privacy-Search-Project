package utils

import (
	"fmt"
	"net"
	"net/rpc"
	"strconv"
)

const (
	EmbServerPort = 1240
	UrlServerPort = 1450
)

func LocalAddr(port int) string {
	return localIP().String() + ":" + strconv.Itoa(port)
}

func RemoteAddr(ip string, port int) string {
	return ip + ":" + strconv.Itoa(port)
}

func localIP() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		panic("Error looking up own IP")
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP
			}
		}
	}
	panic("Own IP not found")
}

/*
 * callTLS and callTCP send an RPC to the server; then wait for the response.
 */

func DialTCP(addr string) *rpc.Client {
	c, err := rpc.Dial("tcp", addr)

	if err != nil {
		fmt.Printf("Tried to dial %s\n", addr)
		fmt.Printf("DialHTTP error: %s\n", err)
		panic("Dialing error")
	}

	return c
}

func CallTCP(c *rpc.Client, rpcname string, args interface{}, reply interface{}) {
	err := c.Call(rpcname, args, reply)
	if err == nil {
		return
	}

	fmt.Printf("Err: %s\n", err)
	panic("Call failed")
}

/*
 * serveTLS and serveTCP implement the server-side networking logic.
 */
func ListenAndServeTCP(server *rpc.Server, port int) {
	addr := LocalAddr(port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Listener error: %v\n", err)
		panic("Listener error")
	}
	defer l.Close()

	fmt.Printf("TCP server listening on %s\n", addr)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Listener error: %v\n", err)
			continue
		}

		defer conn.Close()
		go server.ServeConn(conn)
	}
}
