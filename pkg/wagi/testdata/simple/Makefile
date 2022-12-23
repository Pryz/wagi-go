all: go rust

go:
	tinygo build -o go.wasm -target=wasi simple.go

rust:
	rustc -o rust.wasm --target wasm32-wasi simple.rs
