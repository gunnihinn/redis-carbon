package main

import (
	"bufio"
	"expvar"
	"io"
	"net"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type Carbon struct {
	ch    chan Point
	stats *Stats
}

func NewCarbon(s *Stats) Carbon {
	c := Carbon{
		ch:    make(chan Point, 1<<10),
		stats: s,
	}

	expvar.Publish("point_queue_length", expvar.Func((func() interface{} {
		return len(c.ch)
	})))
	expvar.Publish("point_total", expvar.Func((func() interface{} {
		return c.stats.PointTotal()
	})))
	expvar.Publish("point_errors", expvar.Func((func() interface{} {
		return c.stats.PointErrors()
	})))

	return c
}

func (c Carbon) handleSockets(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Couldn't accept connection")
			continue
		}

		go func(co net.Conn) {
			defer co.Close()
			c.handleConnection(co)
		}(conn)
	}
}

func (c Carbon) handleConnection(conn net.Conn) {
	r := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(time.Second))
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

		pt, err := PointFromLine(line)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"line":  line,
			}).Error("Invalid input")
			continue
		}

		c.ch <- pt
	}
}

func (c Carbon) handlePoints(rdb *redis.Client) {
	for pt := range c.ch {
		c.stats.IncPointTotal()

		v, err := pt.Encode()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"value": pt.Value,
			}).Error("Couldn't convert float to bytes")
			c.stats.IncPointErrors()
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
			c.stats.IncPointErrors()
		}
	}
}
