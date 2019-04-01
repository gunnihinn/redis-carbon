package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type Point struct {
	Name  string
	Value float32
}

func handleStream(c net.Conn, ch chan Point) {
	r := bufio.NewReader(c)

	for {
		c.SetReadDeadline(time.Now().Add(time.Second))
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Println(err)
			break
		}

		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			log.Println(fmt.Errorf("Invalid input '%s'", line))
			continue
		}

		val, err := strconv.ParseFloat(parts[1], 32)
		if err != nil {
			log.Println(fmt.Errorf("Invalid float in input '%s': %s", line, err))
			continue
		}

		ch <- Point{parts[0], float32(val)}
	}
}

func handlePoints(ch chan Point) {
	for pt := range ch {
		fmt.Printf("%+v\n", pt)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":2003")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	ch := make(chan Point, 1<<10)
	go handlePoints(ch)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go func(c net.Conn, ch chan Point) {
			defer c.Close()
			handleStream(c, ch)
		}(conn, ch)
	}
}
