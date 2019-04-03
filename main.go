package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type Point struct {
	Name  string
	Value float32
}

func PointFromLine(line string) (Point, error) {
	parts := strings.SplitN(line, " ", 3)
	if len(parts) != 3 {
		return Point{}, fmt.Errorf("Line didn't have 3 parts")
	}

	val, err := strconv.ParseFloat(parts[1], 32)
	if err != nil {
		return Point{}, err
	}

	return Point{parts[0], float32(val)}, nil
}

func (p Point) StreamName() string {
	return fmt.Sprintf("metric:%s", p.Name)
}

func (p Point) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p.Value)

	return buf.Bytes(), err
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

	carbon := NewCarbon()
	go carbon.handlePoints(rdb)
	go carbon.handleSockets(listener)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("HTTP server died")
	}
}
