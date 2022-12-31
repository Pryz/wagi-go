package http

import (
	"fmt"
	"log"
	"net/http"
	"unsafe"
)

//go:wasm-module env
//export httpRoundTrip
func httpRoundTrip(reqPtr, reqSize uint32, respBufPtr **byte, respBufSize *int) uint32

type Transport struct {
	Logger *log.Logger
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	msg := &HttpRequest{
		Method: req.Method,
		URL:    req.URL.String(),
		Header: req.Header,
	}

	b, err := msg.MarshalMsg(nil)
	if err != nil {
		return &http.Response{StatusCode: http.StatusInternalServerError},
			fmt.Errorf("failed to marshal msgpack: %s", err)
	}
	reqPtr, reqSize := bytesPtr(b)

	var respBufPtr *byte
	var respBufSize int

	_ = httpRoundTrip(reqPtr, reqSize, &respBufPtr, &respBufSize)

	t.Logger.Println("size of response:", respBufSize)

	resp := &HttpResponse{}
	if respBufSize > 0 {
		result := unsafe.Slice(respBufPtr, respBufSize)

		t.Logger.Println("size of result:", len(result))

		_, err := resp.UnmarshalMsg(result)
		if err != nil {
			return &http.Response{StatusCode: http.StatusInternalServerError},
				fmt.Errorf("failed to unmarshal msgpack: %s", err)
		}
	}

	t.Logger.Printf("response from HTTP request: %+v", resp)

	return &http.Response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
	}, nil
}

func stringPtr(s string) (pos uint32, size uint32) {
	return bytesPtr([]byte(s))
}

func bytesPtr(b []byte) (pos uint32, size uint32) {
	ptr := &b[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return uint32(unsafePtr), uint32(len(b))
}

//export malloc_buf
//go:linkname malloc_buf
func malloc(size uint32) *byte {
	buf := make([]byte, size)
	return &buf[0]
}
