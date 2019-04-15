package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-redis/redis"
)

func parseMatchers(query string) ([]*regexp.Regexp, error) {
	// TODO(gmagnusson: Make split aware of nesting dots? Or is error anyway?
	paths := strings.Split(query, ".")
	matchers := make([]*regexp.Regexp, 0, len(paths))

	for _, path := range paths {
		// TODO(gmagnusson): Deal with escaped *
		path = strings.Replace(path, "*", ".*", -1)
		if !strings.HasPrefix(path, "^") {
			path = "^" + path
		}
		if !strings.HasSuffix(path, "$") {
			path = path + "$"
		}

		re, err := regexp.Compile(path)
		if err != nil {
			return nil, err
		}

		matchers = append(matchers, re)
	}

	return matchers, nil
}

func matches(matchers []*regexp.Regexp, name string) bool {
	if strings.HasPrefix(name, METRIC_PREFIX) {
		name = strings.TrimPrefix(name, METRIC_PREFIX)
	}

	parts := strings.Split(name, ".")
	if len(parts) < len(matchers) {
		return false
	}

	for i := range matchers {
		if !matchers[i].MatchString(parts[i]) {
			return false
		}
	}

	return true
}

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
