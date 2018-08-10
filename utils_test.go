package main

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestGet3(t *testing.T) {
	b := []byte{0x00, 0x01, 0x01}
	spew.Dump(get3bytes(b))

}
