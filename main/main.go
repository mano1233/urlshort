package main

import (
	"flag"
	"fmt"
	"net/http"

	urlshort "mano/urlshort/urlshort"
)

func main() {
	mux := defaultMux()
	Filename := flag.String("fileName", "targets.yaml", "a file to be parsed")
	fileType := flag.String("t", "yaml", "the type of file name thats needs to be parsed available types (json,yaml,sqlite) please specify fileName option as well")
	flag.Parse()
	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)
	var handler http.Handler
	var err error
	handler, err = urlshort.FileHandler(*Filename, *fileType, mapHandler)
	if err != nil {
		panic(err)
	}
	// yamlHandler, err := urlshort.YAMLHandler([]byte(yaml), mapHandler)
	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", handler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
