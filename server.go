package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	metrics := make([]string, 0)
	var cursor uint64
	for {
		keys, cursor, err := h.rdb.Scan(cursor, METRIC_PREFIX+"*", 1<<10).Result()
		if err != nil {
			http.Error(w, fmt.Sprintf("Couldn't list metrics: %s", err), http.StatusInternalServerError)
			return
		}

		for _, key := range keys {
			metrics = append(metrics, strings.TrimPrefix(key, METRIC_PREFIX))
		}

		if cursor == 0 {
			break
		}
	}

	blob, err := json.Marshal(metrics)
	if err != nil {
		http.Error(w, fmt.Sprintf("Couldn't serialize response: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(blob)
}
