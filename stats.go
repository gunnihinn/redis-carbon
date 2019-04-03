package main

import (
	"sync/atomic"
)

type Stats struct {
	pointTotal  int64
	pointErrors int64
}

func (s *Stats) PointTotal() int64  { return atomic.LoadInt64(&s.pointTotal) }
func (s *Stats) PointErrors() int64 { return atomic.LoadInt64(&s.pointErrors) }
func (s *Stats) IncPointTotal()     { atomic.AddInt64(&s.pointTotal, 1) }
func (s *Stats) IncPointErrors()    { atomic.AddInt64(&s.pointErrors, 1) }
