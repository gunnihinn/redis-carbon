package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"expvar"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type Point struct {
	Name  string
	Value float32
}

func (p Point) StreamName() string {
	return fmt.Sprintf("metric:%s", p.Name)
}

func (p Point) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p.Value)

	return buf.Bytes(), err
}

func handleStream(c net.Conn, ch chan Point) {
	r := bufio.NewReader(c)

	for {
		c.SetReadDeadline(time.Now().Add(time.Second))
		line, err := r.ReadString('\n')
		if err == io.EOF {
			return
		}

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Couldn't read from socket")
			return
		}

		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			log.WithFields(log.Fields{
				"line": line,
			}).Error("Invalid input")
			continue
		}

		val, err := strconv.ParseFloat(parts[1], 32)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"line":  line,
			}).Error("Invalid input")
			continue
		}

		ch <- Point{parts[0], float32(val)}
	}
}

func (s *Stats) handlePoints(rdb *redis.Client, ch chan Point) {
	for pt := range ch {
		atomic.AddInt64(&s.pointTotal, 1)

		v, err := pt.Encode()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"value": pt.Value,
			}).Error("Couldn't convert float to bytes")
			atomic.AddInt64(&s.pointErrors, 1)
			continue
		}

		args := redis.XAddArgs{
			Stream: pt.StreamName(),
			Values: map[string]interface{}{"v": v},
		}

		cmd := rdb.XAdd(&args)
		if _, err := cmd.Result(); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Redis error")
			atomic.AddInt64(&s.pointErrors, 1)
		}
	}
}

type Stats struct {
	pointTotal  int64
	pointErrors int64
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
	stats := new(Stats)
	go stats.handlePoints(rdb, ch)

	go func() {
		expvar.Publish("point_queue_length", expvar.Func((func() interface{} {
			return len(ch)
		})))
		expvar.Publish("point_total", expvar.Func((func() interface{} {
			return atomic.LoadInt64(&stats.pointTotal)
		})))
		expvar.Publish("point_errors", expvar.Func((func() interface{} {
			return atomic.LoadInt64(&stats.pointErrors)
		})))

		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("HTTP server died")
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Couldn't accept connection")
			continue
		}

		go func(c net.Conn, ch chan Point) {
			defer c.Close()
			handleStream(c, ch)
		}(conn, ch)
	}
}
