package log

import (
	"log"
	"unsafe"
)

//go:wasm-module env
//export log
func _log(ptr uint32, size uint32)

func NewLogger(l *log.Logger) *log.Logger {
	if l == nil {
		l = log.Default()
	}
	w := &logWriter{}
	l.SetOutput(w)
	return l
}

type logWriter struct {
}

func (w *logWriter) Write(p []byte) (int, error) {
	ptr, size := bytesToPrt(p)
	_log(ptr, size)
	return len(p), nil
}

func bytesToPrt(b []byte) (uint32, uint32) {
	ptr := &b[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return uint32(unsafePtr), uint32(len(b))
}
