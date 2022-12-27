package wagi

import (
	"context"
	"fmt"
	"log"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

func SetEnvironment(ctx context.Context, r wazero.Runtime, ns wazero.Namespace) {
	builder := r.NewHostModuleBuilder("env")

	builder.NewFunctionBuilder().
		WithFunc(logPrintf).
		Export("log").
		Instantiate(ctx, r)

	apiFunc := api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
		mPos, mSize := stack[0], stack[1]
		buf, _ := mod.Memory().Read(uint32(mPos), uint32(mSize))
		log.Printf("DEBUG - receive request with method %s", string(buf))
	})

	params := []api.ValueType{
		api.ValueTypeI32, // method position
		api.ValueTypeI32, // method size
	}

	results := []api.ValueType{
		api.ValueTypeI32,
	}

	builder.NewFunctionBuilder().
		WithGoModuleFunction(apiFunc, params, results).
		Export("httpRoundTrip").
		Instantiate(ctx, ns)
}

func logPrintf(ctx context.Context, mod api.Module, pos, size uint32) {
	buf, ok := mod.Memory().Read(pos, size)
	if !ok {
		log.Printf("ERROR - memory out of range: pos=%d size=%d", pos, size)
	}
	fmt.Printf(string(buf))
	//log.Printf("INFO - message from module: %s", string(buf))
}
