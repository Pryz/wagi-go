target ?= $(target)
wasmbin ?= $(patsubst %.rs,%.wasm,$(target))

rust:
	rustc -o $(target).wasm --target wasm32-wasi $(target).rs
