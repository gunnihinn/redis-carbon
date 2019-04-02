package main

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestPointBytes(t *testing.T) {
	p := Point{Value: 4}
	bs, err := p.Encode()
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewReader(bs)
	var got float32
	if err := binary.Read(buf, binary.LittleEndian, &got); err != nil {
		t.Fatal(err)
	}

	if got != p.Value {
		t.Fatal("Bad decoding")
	}
}
