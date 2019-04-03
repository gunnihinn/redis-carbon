package main

import (
	"net/http"
)

type RenderHandler Carbon

func (h RenderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/render"))
}

type MetricsFindHandler Carbon

func (h MetricsFindHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/metrics/find"))
}

type MetricsIndexHandler Carbon

func (h MetricsIndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/metrics/index"))
}
