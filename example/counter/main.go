//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate --world counter --out gen ./wit
package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bytecodealliance/wasm-tools-go/cm"
	"github.com/jamesstocktonj1/ticker-provider/example/counter/gen/jamesstocktonj1/ticker/ticker"
	"github.com/jamesstocktonj1/ticker-provider/example/counter/gen/wasi/keyvalue/atomics"
	"github.com/jamesstocktonj1/ticker-provider/example/counter/gen/wasi/keyvalue/store"
	"go.wasmcloud.dev/component/log/wasilog"
	"go.wasmcloud.dev/component/net/wasihttp"
)

var logger = wasilog.ContextLogger("counter")

func init() {
	wasihttp.HandleFunc(handleRequest)
	ticker.Exports.Task = handleTask
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	bucket, _, isErr := store.Open("").Result()
	if isErr {
		logger.Error("error: unable to open keyvalue store")
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
		logger.Error("error: enable to increment counter")
		mar := json.NewEncoder(w)
		mar.SetIndent("", "  ")
		mar.Encode(map[string]string{
			"message": "error: unable to increment counter",
			"key":     path,
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Info("handleRequest", "count", count, "key", path)
	mar := json.NewEncoder(w)
	mar.SetIndent("", "  ")
	mar.Encode(map[string]any{
		"count": int(count),
		"key":   path,
	})
	w.WriteHeader(http.StatusOK)
	return
}

func handleTask() ticker.TaskError {
	logger.Info("handleTask")

	bucket, _, isErr := store.Open("").Result()
	if isErr {
		logger.Error("error: unable to open keyvalue store")
		return ticker.TaskErrorError("error: unable to open keyvalue store")
	}

	cursor := uint64(0)
	keysResult, _, isErr := bucket.ListKeys(cm.Some(cursor)).Result()
	if isErr {
		logger.Error("error: unable to fetch keys")
		return ticker.TaskErrorError("error: unable to fetch keys")
	}

	keys := keysResult.Keys.Slice()
	for _, k := range keys {
		_, _, isErr := bucket.Delete(k).Result()
		if isErr {
			logger.Error("error: unable to delete key", "key", k)
		}
	}
	return ticker.TaskErrorNone()
}

func main() {}
