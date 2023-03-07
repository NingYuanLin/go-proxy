package socks5

import (
	"bytes"
	"reflect"
	"testing"
)

func TestNewClientAuthMessage(t *testing.T) {
	t.Run("should generate a message", func(t *testing.T) {
		b := []byte{Socks5Version, 2, MethodNoAuth, MethodGssApi}
		r := bytes.NewReader(b)
		message, err := NewClientAuthMessage(r)
		if err != nil {
			t.Fatalf("want err = nil but got %s", err)
		}
		if message.NMethods != 2 {
			t.Fatalf("want NMethods = 2 but got %d", message.NMethods)
		}
		methods := []byte{MethodNoAuth, MethodGssApi}
		if !reflect.DeepEqual(message.Methods, methods) {
			t.Fatalf("want methods: %v, but got %v", methods, message.Methods)
		}
	})

	t.Run("methods length is shorter than NMethods", func(t *testing.T) {
		b := []byte{Socks5Version, 2, MethodNoAuth}
		r := bytes.NewReader(b)
		_, err := NewClientAuthMessage(r)
		if err == nil {
			t.Fatalf("should get error != nil but got nil")
		}
	})
}

func TestNewServerAuthMessage(t *testing.T) {
	t.Run("should pass", func(t *testing.T) {
		var buf bytes.Buffer
		err := WriteServerAuthMessage(&buf, MethodNoAuth)
		if err != nil {
			t.Fatalf("err should be nil but got %s", err)
		}
		b := buf.Bytes()
		shouldWrite := []byte{Socks5Version, MethodNoAuth}
		if !reflect.DeepEqual(b, shouldWrite) {
			t.Fatalf("should write %v, but write %v", shouldWrite, b)
		}
	})
}

func TestNewClientPasswordAuthMessage(t *testing.T) {
	tests := []struct {
		Version        byte
		UsernameLength byte
		Username       []byte
		PasswordLength byte
		Password       []byte

		Err error
	}{
		{
			PasswordAuthVersion,
			0x03,
			[]byte{1, 2, 3},
			0x06,
			[]byte{1, 2, 3, 4, 5, 6},
			nil,
		},
		{
			PasswordAuthVersion,
			0x04,
			[]byte{1, 2, 3, 4},
			0x07,
			[]byte{1, 2, 3, 4, 5, 6, 7},
			nil,
		},
		{
			0x00,
			0x04,
			[]byte{1, 2, 3, 4},
			0x07,
			[]byte{1, 2, 3, 4, 5, 6, 7},
			ErrMethodVersionNotSupported,
		},
	}

	for _, test := range tests {
		conn := bytes.Buffer{}
		conn.Write([]byte{test.Version, test.UsernameLength})
		conn.Write(test.Username)
		conn.Write([]byte{test.PasswordLength})
		conn.Write(test.Password)

		message, err := NewClientPasswordAuthMessage(&conn)
		if err != test.Err {
			t.Fatalf("error should be %v, but got %v", test.Err, err)
		}
		if err != nil {
			continue
		}

		wantMessage := ClientPasswordAuthMessage{string(test.Username), string(test.Password)}
		if !reflect.DeepEqual(wantMessage, *message) {
			t.Fatalf("message should be %v, but got %v", wantMessage, message)
		}
	}
}
