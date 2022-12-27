package wagi

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type Handler struct {
	context  context.Context
	path     string
	function string
	runtime  wazero.Runtime
}

func NewHandler(ctx context.Context, runtime wazero.Runtime, path, function string) *Handler {
	return &Handler{
		context:  ctx,
		path:     path,
		function: function,
		runtime:  runtime,
	}
}

func (h *Handler) Close(ctx context.Context) {
	h.runtime.Close(ctx)
}

func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	data, err := loadModule(h.context, h.path)
	if err != nil {
		log.Printf("ERROR - failed to load module: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	stdout := new(bytes.Buffer)
	stdin := new(bytes.Buffer)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("ERROR - failed to read request body: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)

	}
	stdin.Write(body)

	config := prepareConfig(stdin, stdout, req.Header, parseQuery(req.URL.Query()))

	// TODO: define default headers.
	if err := h.run(h.function, data, config); err != nil {
		log.Printf("ERROR - failed to run module: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := processResponse(resp, stdout); err != nil {
		log.Printf("ERROR - failed to process response: %s", err)
		resp.WriteHeader(http.StatusPartialContent)
	}
}

func parseQuery(values url.Values) []string {
	args := []string{}
	for key, vals := range values {
		for _, val := range vals {
			args = append(args, fmt.Sprintf("%s=%s", key, val))
		}
	}
	return args
}

func processResponse(resp http.ResponseWriter, buf *bytes.Buffer) error {
	scanner := bufio.NewScanner(buf)
	// FIXME
	bodyBuf := new(bytes.Buffer)

	var body bool
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			body = true
			continue
		}
		if !body {
			_, err := processHeaderLine(resp, scanner.Text())
			if err != nil {
				log.Printf("ERROR - failed to process stdout line: %s", err)
			}
		} else {
			bodyBuf.Write(scanner.Bytes())
		}
	}

	_, err := resp.Write(bodyBuf.Bytes())
	return err
}

func processHeaderLine(resp http.ResponseWriter, line string) (bool, error) {
	if len(line) == 0 {
		return true, nil
	}
	split := strings.Split(line, ":")
	if len(split) > 1 {
		header := strings.TrimSpace(split[0])
		// TODO: do something with status text.
		// TODO: handle the location header.
		switch header {
		case "content-type", "Content-Type":
			resp.Header().Add("Content-Type", split[1])
		case "status", "Status":
			status := strings.TrimSpace(split[1])
			code, _ := stringToStatus(status)
			if code != -1 {
				resp.WriteHeader(code)
			}
		}
	}
	return false, nil
}

func stringToStatus(str string) (code int, text string) {
	status, err := strconv.Atoi(str)
	if err != nil {
		return -1, ""
	}
	return status, http.StatusText(status)
}

// TODO: load modules once and reuse it
func loadModule(ctx context.Context, path string) ([]byte, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	return io.ReadAll(fh)
}

func prepareConfig(stdin io.Reader, stdout io.Writer, env map[string][]string, args []string) wazero.ModuleConfig {
	config := wazero.NewModuleConfig()
	config = config.WithStdin(stdin)
	config = config.WithStdout(stdout)

	if env != nil {
		for key, vals := range env {
			for _, val := range vals {
				config = config.WithEnv(key, val)
			}
		}
	}

	if len(args) > 0 {
		config = config.WithArgs(args...)
	}

	return config
}

func (h *Handler) run(function string, data []byte, config wazero.ModuleConfig) error {
	compiled, err := h.runtime.CompileModule(h.context, data)
	if err != nil {
		return fmt.Errorf("compile module: %s", err)
	}

	namespace := h.runtime.NewNamespace(h.context)
	defer namespace.Close(h.context)

	builder := wasi_snapshot_preview1.NewBuilder(h.runtime)
	builder.Instantiate(h.context, namespace)

	SetEnvironment(h.context, h.runtime, namespace)

	// NOTE: wazero will automatically run a function called _start if exists.
	mod, err := namespace.InstantiateModule(h.context, compiled, config)
	if err != nil {
		return fmt.Errorf("instantiate module: %s", err)
	}
	defer mod.Close(h.context)

	// When a function is not defined, we rely on executing the `_start` function.
	if function != "" {
		fn := mod.ExportedFunction(function)
		if fn == nil {
			return fmt.Errorf("function not exporter")
		}
		_, err = fn.Call(h.context)
		if err != nil {
			return fmt.Errorf("call main: %s", err)
		}
	}

	return nil
}