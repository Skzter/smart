package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck
		fmt.Fprintf(w, "Hello World!")
	})

	//nolint:gosec
	if err := http.ListenAndServe(":8081", mux); err != nil {
		panic(err)
	}
}
