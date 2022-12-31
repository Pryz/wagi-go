package wagi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/vmihailenco/msgpack/v5"
)

func SetEnvironment(ctx context.Context, r wazero.Runtime, ns wazero.Namespace) {
	builder := r.NewHostModuleBuilder("env")

	builder.NewFunctionBuilder().
		WithFunc(logPrintf).
		Export("log").
		Instantiate(ctx, r)

	exportHttpRoundTrip(ctx, ns, builder.NewFunctionBuilder())
}

// TODO: add ability to split the different streams of log (runtime vs modules)
func logPrintf(ctx context.Context, mod api.Module, pos, size uint32) {
	buf, ok := mod.Memory().Read(pos, size)
	if !ok {
		log.Printf("ERROR - memory out of range: pos=%d size=%d", pos, size)
	}

	fmt.Printf(string(buf))
}

func exportHttpRoundTrip(ctx context.Context, ns wazero.Namespace, builder wazero.HostFunctionBuilder) {
	apiFunc := api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
		mPos, mSize := stack[0], stack[1]
		log.Printf("request at %d, size %d", mPos, mSize)
		buf, _ := mod.Memory().Read(uint32(mPos), uint32(mSize))

		msg := &httpRequest{}

		msgpack.Unmarshal(buf, msg)

		req, err := msg.request()
		if err != nil {
			log.Printf("ERROR - failed to parse HTTP request from module")
			return
		}

		log.Printf("DEBUG - receive request from module: %+v", msg)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("ERROR - failed http request: %s", err)
			return
		}

		msgResp, err := msgpack.Marshal(responseFrom(resp))
		if err != nil {
			log.Printf("ERROR - failed to marshal response for module: %s", err)
			return
		}

		mallocRet, err := mod.ExportedFunction("malloc_buf").Call(ctx, uint64(len(msgResp)))
		if err != nil {
			log.Printf("ERROR - failed to malloc buffer in module memory: %s", err)
			return
		}
		bufPtr := uint32(mallocRet[0])

		respBufPtr := uint32(stack[2])
		respBufSize := uint32(stack[3])

		if !mod.Memory().Write(uint32(bufPtr), msgResp) {
			log.Printf("ERROR - write out of range")
			return
		}

		mod.Memory().WriteUint32Le(respBufPtr, uint32(bufPtr))
		mod.Memory().WriteUint32Le(respBufSize, uint32(len(msgResp)))
		mod.Memory().Write(bufPtr, msgResp)

		// TODO: the following only works with tinygo. Need a standard ABI.
		//malloc := mod.ExportedFunction("malloc")
		//free := mod.ExportedFunction("free")

		//results, err := malloc.Call(ctx, uint64(len(msgResp)))
		//if err != nil {
		//log.Printf("ERROR - failed to malloc for http response: %s", err)
		//return
		//}
		//msgRespPtr := results[0]
		//// FIXME
		//defer free.Call(ctx, msgRespPtr)

		//if !mod.Memory().Write(uint32(msgRespPtr), msgResp) {
		//log.Println("ERROR - failed to write http response to module mem")
		//return
		//}

		//stack[0] = api.EncodeI32(int32(msgRespPtr))
		//stack[1] = api.EncodeI32(int32(len(msgResp)))
	})

	params := []api.ValueType{
		api.ValueTypeI32, // request position
		api.ValueTypeI32, // request size
		api.ValueTypeI32, // response position
		api.ValueTypeI32, // response size
	}

	results := []api.ValueType{
		api.ValueTypeI32, // result
	}

	builder.WithGoModuleFunction(apiFunc, params, results).
		Export("httpRoundTrip").
		Instantiate(ctx, ns)
}

type httpRequest struct {
	Method string              `msgpack:"method"`
	URL    string              `msgpack:"url"`
	Proto  string              `msgpack:"proto,omitempty"`
	Header map[string][]string `msgpack:"header,omitempty"`
	Body   []byte              `msgpack:"body,omitempty"`
}

func (h *httpRequest) request() (*http.Request, error) {
	u, err := url.Parse(h.URL)
	if err != nil {
		return nil, err
	}

	var reader io.ReadCloser
	if len(h.Body) > 0 {
		reader = io.NopCloser(bytes.NewReader(h.Body))
	}

	return &http.Request{
		Method: h.Method,
		URL:    u,
		Proto:  h.Proto,
		Header: h.Header,
		Body:   reader,
	}, nil
}

type httpResponse struct {
	Status     string `msgpack:"status"`
	StatusCode int    `msgpack:"status_code"`
	//Proto      string              `msgpack:"proto,omitempty"`
	//Header     map[string][]string `msgpack:"header,omitempty"`
	//Body          []byte              `msgpack:"body,omitempty"`
	//ContentLength int64               `msgpack:"content_length"`
}

func (h *httpResponse) response() *http.Response {
	return &http.Response{
		Status:     h.Status,
		StatusCode: h.StatusCode,
		//Proto:      h.Proto,
		//Header:     h.Header,
	}
}

func responseFrom(resp *http.Response) *httpResponse {
	r := &httpResponse{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
	}
	// TODO: deal with body
	return r
}
