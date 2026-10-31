package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/kuroishi/golnk/internal/application"
	"github.com/kuroishi/golnk/internal/infrastructure/fetcher"
	"github.com/kuroishi/golnk/internal/infrastructure/idgen"
	"github.com/kuroishi/golnk/internal/infrastructure/repo"
	httpapi "github.com/kuroishi/golnk/internal/interface/http"
	"github.com/spf13/pflag"
)

func main() {
	addr := pflag.StringP("addr", "a", envOr("GOLNK_ADDR", ":8080"), "listen address")
	file := pflag.StringP("file", "f", envOr("GOLNK_FILE", "data.json"), "data file path")
	pflag.Parse()

	r, err := repo.NewFileRepo(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load data: %v\n", err)
		os.Exit(1)
	}

	shortener := application.NewShortener(r, fetcher.NewHTTPFetcher(), idgen.NewCryptoGen())
	handler := httpapi.NewHandler(shortener)

	mux := http.NewServeMux()
	handler.Register(mux)

	fmt.Printf("golnk listening on %s\n", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}