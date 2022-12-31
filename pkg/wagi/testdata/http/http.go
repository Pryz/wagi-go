package main

import (
	"net/http"

	wagihttp "github.com/Pryz/wagi-go/sdk/go/http"
	wagilog "github.com/Pryz/wagi-go/sdk/go/log"
)

var (
	log    = wagilog.NewLogger(nil)
	client = http.DefaultClient
)

// Main is required by wazero (tinygo ?) since it will be executed
// as __start.
func main() {
	client = &http.Client{
		Transport: &wagihttp.Transport{
			Logger: log,
		},
	}
}

// TODO: how to instrument errors ? via logs or events ?
// TODO: split host and module logging streams
// TODO: add tracer to http transport. should we use stderr ?
// TODO: finish body reader impl

// `handle` use the wagihttp.Transport to leverage the host for
// an HTTP request. The host only process the request and forward
// the HTTP response to the module.

//export handle
func handle() {
	req, err := http.NewRequest("GET", "https://httpbin.org/get", nil)
	if err != nil {
		log.Fatalf("failed to create http request: %s", err)
	}

	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("failed http request: %s", err)
	}

	wagihttp.Respond(&http.Response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
	})
}
