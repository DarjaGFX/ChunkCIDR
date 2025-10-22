package chunker

import (
	"reflect"
	"testing"
)

func TestChunkCIDRIPv4(t *testing.T) {
	res, err := ChunkCIDR("192.168.0.0/24", 26)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{
		"192.168.0.0/26",
		"192.168.0.64/26",
		"192.168.0.128/26",
		"192.168.0.192/26",
	}
	if !reflect.DeepEqual(res, want) {
		t.Fatalf("got %v, want %v", res, want)
	}
}

func TestChunkCIDRSamePrefix(t *testing.T) {
	res, err := ChunkCIDR("10.0.0.0/8", 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 || res[0] != "10.0.0.0/8" {
		t.Fatalf("unexpected result: %v", res)
	}
}
