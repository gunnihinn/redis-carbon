package main

import (
	"bufio"
	"expvar"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type Carbon struct {
	ch    chan Point
	stats *Stats
}

func NewCarbon() Carbon {
	c := Carbon{
		ch:    make(chan Point, 1<<10),
		stats: new(Stats),
	}

	expvar.Publish("point_queue_length", expvar.Func((func() interface{} {
		return len(c.ch)
	})))
	expvar.Publish("point_total", expvar.Func((func() interface{} {
		return atomic.LoadInt64(&c.stats.pointTotal)
	})))
	expvar.Publish("point_errors", expvar.Func((func() interface{} {
		return atomic.LoadInt64(&c.stats.pointErrors)
	})))

	return c
}

type Stats struct {
	pointTotal  int64
	pointErrors int64
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
		atomic.AddInt64(&c.stats.pointTotal, 1)

		v, err := pt.Encode()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"value": pt.Value,
			}).Error("Couldn't convert float to bytes")
			atomic.AddInt64(&c.stats.pointErrors, 1)
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
			atomic.AddInt64(&c.stats.pointErrors, 1)
		}
	}
}
