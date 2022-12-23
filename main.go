package main

import "fmt"

// curl -vvv -H "HOST:foo.example.com" localhost:3000/env/foo?greet=matt\&foo=bar
// - query parameters are passed via command line arguments. here: /env greet=matt foo=bar
// - HTTP headers become env variable. As well as default env vars: https://github.com/deislabs/wagi/blob/main/docs/environment_variables.md
//
// - STDOUT: WAGI reads STDOUT fh and reformats the data into an HTTP response.
// - STDIN: HTTP POST data is passed through the module via STDIN
//
// Advanced:
// - sub routes: https://github.com/deislabs/wagi/blob/main/docs/writing_modules.md#advanced-declaring-sub-routes-in-the-module
// - mapped volumes: ability to mount FS volumes to module.
// - outbound http requests: https://github.com/deislabs/wagi/blob/main/docs/writing_modules.md#advanced-declaring-sub-routes-in-the-module
// 		-> WASI-sockets is still in progress.
//		-> Here the idea is to provide a `request` function that modules can call. The execution of the HTTP request will be done by the runtime.
//		-> to keep the thing "secure", WAGI has a `allowed_hosts` configuration. If not set or there is no match, the request is denied.

func main() {
	fmt.Println("hello")
}
