package http

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
)

// TODO: find a better name. Also is it the right API ?
// TODO: use the http.ResponseWriter interface ?
func Respond(resp *http.Response) {
	writer := bufio.NewWriter(os.Stdout)

	s := fmt.Sprintln("Content-Type: text/plain")
	s += fmt.Sprintf("Content-Length: %d", resp.ContentLength)
	if resp == nil {
		s += fmt.Sprintf("Status: %d\n", http.StatusInternalServerError)
		return
	}

	s += fmt.Sprintf("Status: %d\n", resp.StatusCode)

	for key, val := range resp.Header {
		s += fmt.Sprintf("%s: %s\n", key, val)
	}

	// TODO: catch writer errors
	writer.WriteString(s)

	if resp.Body != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			// FIXME
			return
		}
		resp.Body.Close()

		// TODO: catch writer errors
		writer.Write(data)
	}
}