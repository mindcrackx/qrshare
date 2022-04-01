package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/skip2/go-qrcode"
)

//go:embed index.html
var indexHtml []byte

var qrMaxContentSize int = 1024

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {

	addr := flag.String("addr", ":8000", "addr to listen on")
	flag.IntVar(&qrMaxContentSize, "max-part-size", 1024, "max content size per qr code part")
	file := flag.String("f", "", "path to file")
	flag.Parse()

	filePath := strings.TrimSpace(*file)
	if filePath == "" {
		return errors.New("no file path provided")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}

	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		if _, err := w.Write(indexHtml); err != nil {
			log.Println("unable to write indexHtml")
		}
	})

	r.Get("/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, fmt.Sprintf("no/wrong image part number provided: %s", err), http.StatusBadRequest)
			return
		}
		if id < 1 {
			http.Error(w, fmt.Sprintf("id has to be >= 1, but was: %d", id), http.StatusBadRequest)
			return
		}

		skip := qrMaxContentSize * (id - 1)
		take := qrMaxContentSize * id

		if len(data) < skip {
			http.Error(w, "no more data", http.StatusBadRequest)
			return
		}

		if len(data) < take {
			take = len(data)
		}

		png, err := qrcode.Encode(string(data[skip:take]), qrcode.Low, 1024)
		if err != nil {
			http.Error(w, fmt.Sprintf("error encoding qrcode: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", fmt.Sprint(len(png)))

		if _, err := w.Write(png); err != nil {
			log.Println("unable to write image")
		}
	})

	err = http.ListenAndServe(*addr, r)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
