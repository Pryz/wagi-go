package http

import (
	"fmt"
	"net/http"
)

// TODO: find a better name. Also is it the right API ?
// TODO: use the http.ResponseWriter interface ?
// TODO: pass resp.Body to stdout
func Respond(resp *http.Response) {
	fmt.Println("Content-Type: text/plain")

	if resp == nil {
		fmt.Println("Status: ", http.StatusInternalServerError)
		return
	}

	fmt.Println("Status: ", resp.Status)
}