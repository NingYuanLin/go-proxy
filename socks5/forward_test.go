package socks5

import (
	"reflect"
	"testing"
)

func TestGetHostByteFromString(t *testing.T) {
	bytes, err := GetHostByteFromString("1.1.1.1:80")
	if err != nil {
		t.Fatal(err)
	}
	wantBytes := []byte{AddressTypeIpv4, 1, 1, 1, 1, 0x00, 0x50}
	if reflect.DeepEqual(wantBytes, bytes) != true {
		t.Fatalf("want %v, got %v", wantBytes, bytes)
	}
}
