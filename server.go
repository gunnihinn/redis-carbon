package main

import (
	"net/http"

	"github.com/go-redis/redis"
)

// TODO(gmagnusson): Is *redis.Client thread safe?
type handler struct {
	rdb *redis.Client
}

type RenderHandler handler

func (h RenderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/render"))
}

type MetricsFindHandler handler

func (h MetricsFindHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/metrics/find"))
}

type MetricsIndexHandler handler

func (h MetricsIndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Ask Redis about all keys that start with 'metric:'
	// Put them into an array
	// Write that array out as JSON
}
