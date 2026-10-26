package main

import (
	"fmt"
	"net"
	"strconv"
	"testing"
)

func TestListenIncrementsBusyPort(t *testing.T) {
	first, addr, err := listen("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = first.Close() })

	_, portText, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}

	second, nextAddr, err := listen(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Close() })

	_, nextPortText, err := net.SplitHostPort(nextAddr)
	if err != nil {
		t.Fatal(err)
	}
	nextPort, err := strconv.Atoi(nextPortText)
	if err != nil {
		t.Fatal(err)
	}
	if nextPort <= port {
		t.Fatalf("busy port %d was not incremented: got %d", port, nextPort)
	}
}
