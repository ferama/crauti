package utils

import (
	"net/http"
	"testing"
)

func TestExtractHost(t *testing.T) {
	list := []string{
		"host.dom:8080",
		"host.dom",
	}
	expected := []string{
		"host.dom",
		"host.dom",
	}

	for idx, s := range list {
		req := &http.Request{
			Host: s,
		}
		host, err := GetRequestHost(req)
		if err != nil {
			t.Fatal(err)
		}
		if host != expected[idx] {
			t.Fatalf("expected %s, got %s", expected[idx], host)
		}
	}
}

func TestConvertToBytes(t *testing.T) {

	list := []string{
		"1KB",
		"3Kb",
		"5mb",
		"0",
		"10b",
		"-10b",
		"-10Kb",
	}
	expected := []int64{
		1024,
		3 * 1024,
		5 * 1024 * 1024,
		0,
		10,
		-10,
		-10 * 1024,
	}

	for idx, s := range list {
		bytes, err := ConvertToBytes(s)
		if err != nil {
			t.Fatal(err)
		}
		if bytes != expected[idx] {
			t.Fatalf("expected %d, got %d", expected[idx], bytes)
		}
	}
}
