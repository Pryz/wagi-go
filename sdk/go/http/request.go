package http

import (
	"net/http"
	"unsafe"
)

//go:wasm-module env
//export httpRoundTrip
func httpRoundTrip(mPos, mSize uint32) uint32

type Transport struct{}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	methodePos, methodeSize := stringPtr(req.Method)
	_ = httpRoundTrip(methodePos, methodeSize)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
	}
	return resp, nil
}

func stringPtr(s string) (pos uint32, size uint32) {
	return bytesPtr([]byte(s))
}

func bytesPtr(b []byte) (pos uint32, size uint32) {
	ptr := &b[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return uint32(unsafePtr), uint32(len(b))
}