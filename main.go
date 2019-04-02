package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

type Point struct {
	Name  string
	Value float32
}

func (p Point) StreamName() string {
	return fmt.Sprintf("metric:%s", p.Name)
}

func (p Point) ValueBytes() ([]byte, error) {
	b := make([]byte, 4)
	buf := bytes.NewBuffer(b)
	err := binary.Write(buf, binary.LittleEndian, p.Value)

	return buf.Bytes(), err
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

func handlePoints(rdb *redis.Client, ch chan Point) {
	for pt := range ch {
		v, err := pt.ValueBytes()
		if err != nil {
			log.Println(err)
			continue
		}

		args := redis.XAddArgs{
			Stream: pt.StreamName(),
			Values: map[string]interface{}{"v": v},
		}

		cmd := rdb.XAdd(&args)
		if _, err := cmd.Result(); err != nil {
			log.Println(err)
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":2003")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ch := make(chan Point, 1<<10)
	go handlePoints(rdb, ch)

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
