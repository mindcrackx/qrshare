package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"net/http"
)

//go:embed index.html
var indexHtml []byte

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {

	addr := flag.String("addr", ":8001", "addr to listen on")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(indexHtml)
	})

	err := http.ListenAndServe(*addr, nil)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
