//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate --world counter --out gen ./wit
package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jamesstocktonj1/ticker-provider/example/counter/gen/wasi/keyvalue/atomics"
	"github.com/jamesstocktonj1/ticker-provider/example/counter/gen/wasi/keyvalue/store"
	"go.wasmcloud.dev/component/net/wasihttp"
)

func init() {
	wasihttp.HandleFunc(handleRequest)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	bucket, _, isErr := store.Open("counter").Result()
	if isErr {
		mar := json.NewEncoder(w)
		mar.SetIndent("", "  ")
		mar.Encode(map[string]string{
			"message": "error: unable to open keyvalue store",
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	path := r.URL.Path
	path, _ = strings.CutPrefix(path, "/")
	path = strings.ReplaceAll(path, "/", ".")

	count, _, isErr := atomics.Increment(bucket, path, 1).Result()
	if isErr {
		mar := json.NewEncoder(w)
		mar.SetIndent("", "  ")
		mar.Encode(map[string]string{
			"message": "error: unable to increment counter",
			"key":     path,
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	mar := json.NewEncoder(w)
	mar.SetIndent("", "  ")
	mar.Encode(map[string]any{
		"counter": int(count),
		"key":     path,
	})
	w.WriteHeader(http.StatusOK)
	return
}

func main() {}
