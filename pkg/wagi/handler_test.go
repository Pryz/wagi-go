package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		scenario string
		path     string
		function string
		header   http.Header
		query    map[string]string
		in       []byte
		out      []byte
	}{
		{
			scenario: "simple tinygo",
			path:     "./testdata/simple/go.wasm",
			function: "handle",
		},
		{
			scenario: "simple rust",
			path:     "./testdata/simple/rust.wasm",
		},
		{
			scenario: "query",
			path:     "./testdata/query/query.wasm",
			query:    map[string]string{"who": "bob"},
			out:      []byte("Hello bob"),
		},
	}

	ctx := context.Background()
	client := http.Client{}

	runtime := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			h := NewHandler(ctx, runtime, test.path, test.function)
			srv := httptest.NewServer(h)
			defer srv.Close()

			req, err := http.NewRequest("GET", srv.URL, nil)
			if err != nil {
				t.Fatal(err)
			}
			for key, vals := range test.header {
				for _, val := range vals {
					req.Header.Add(key, val)
				}
			}

			q := req.URL.Query()
			for key, val := range test.query {
				q.Add(key, val)
			}
			req.URL.RawQuery = q.Encode()

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Error("unexpected client request result")
				t.Log("got:", resp.Status)
			}
			defer resp.Body.Close()

			out, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("failed to read resp body: %s", err)
			}

			if len(test.out) > 0 {
				if !reflect.DeepEqual(test.out, out) {
					t.Error("unexpected response body")
					t.Logf("expected: %s", string(test.out))
					t.Logf("got: %s", string(out))
				}
			}
		})
	}
}
