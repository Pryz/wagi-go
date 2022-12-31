package main

import (
	"net/http"

	wagihttp "github.com/Pryz/wagi-go/sdk/go/http"
	wagilog "github.com/Pryz/wagi-go/sdk/go/log"
)

var (
	log = wagilog.NewLogger(nil)
)

func main() {}

// TODO: add the ability to return and handle error

//export handle
func handle() {
	req, err := http.NewRequest("GET", "https://httpbin.org/get", nil)
	if err != nil {
		log.Fatalf("failed to create http request: %s", err)
	}

	log.Printf("url: %s", req.URL)

	var resp *http.Response

	client := http.Client{
		Transport: &wagihttp.Transport{
			Logger: log,
		},
	}
	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("failed http request: %s", err)
	}

	if resp == nil {
		log.Println("HTTP reponse is nil")
	} else {
		log.Printf("http reponse %v", resp.Status)
	}

	//wagihttp.Respond(&http.Response{
	//Status:     resp.Status,
	//StatusCode: resp.StatusCode,
	//})
}
