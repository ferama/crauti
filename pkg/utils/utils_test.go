package utils

import "testing"

func TestConvertToBytes(t *testing.T) {

	list := []string{
		"1KB",
		"3Kb",
		"5mb",
		"0",
		"10b",
	}
	expected := []int64{
		1024,
		3 * 1024,
		5 * 1024 * 1024,
		0,
		10,
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
