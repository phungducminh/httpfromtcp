package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	rd := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(">")
		line, err := rd.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		_, werr := conn.Write([]byte(line))
		if werr != nil {
			log.Fatal(werr)
		}
	}
}
