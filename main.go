package main

import (
	"net"
	"net/http"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

func main() {
	listener, err := net.Listen("tcp", ":2003")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Couldn't listen on TCP port")
	}
	defer listener.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	carbon := NewCarbon()
	go carbon.handlePoints(rdb)
	go carbon.handleSockets(listener)

	http.Handle("/render", RenderHandler(carbon))
	http.Handle("/metrics/find", MetricsFindHandler(carbon))
	http.Handle("/metrics/index", MetricsIndexHandler(carbon))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("HTTP server died")
	}
}
