package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
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
	return METRIC_PREFIX + p.Name
}

func (p Point) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p.Value)

	return buf.Bytes(), err
}
