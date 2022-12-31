package http

//go:generate msgp
type HttpRequest struct {
	Method string              `msg:"method"`
	URL    string              `msg:"url"`
	Proto  string              `msg:"proto,omitempty"`
	Header map[string][]string `msg:"header,omitempty"`
	Body   []byte              `msg:"body,omitempty"`
}

type HttpResponse struct {
	Status     string `msgpack:"status"`
	StatusCode int    `msgpack:"status_code"`
}
