package urlshort

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	yaml "gopkg.in/yaml.v2"
)

type WrongTypeError struct {
	Type string

	Err error
}

func (r *WrongTypeError) Error() string {
	return fmt.Sprintf("type %s: err %v", r.Type, r.Err)
}

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Ran map handler")
		// Check if the requested path exists in the map
		if dest, ok := pathsToUrls[r.URL.Path]; ok {
			// Redirect to the corresponding URL
			http.Redirect(w, r, dest, http.StatusFound)
			return
		}

		// If the path is not in the map, call the fallback handler
		fallback.ServeHTTP(w, r)
	}
}

type pathUrl struct {
	Path string `yaml:"path,omitempty" json:"path,omitempty" db:"path,omitempty"`
	URL  string `yaml:"url,omitempty" json:"url,omitempty" db:"url,omitempty"`
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//   - path: /some-path
//     url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yamlBytes []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var pathUrls []pathUrl
	pathMap := make(map[string]string)

	err := yaml.Unmarshal(yamlBytes, &pathUrls)
	if err != nil {
		return nil, err
	}

	pathMap = buildMap(pathUrls)
	return MapHandler(pathMap, fallback), nil
}

func JSONHandler(jsonBytes []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var pathUrls []pathUrl
	pathMap := make(map[string]string)

	err := json.Unmarshal(jsonBytes, &pathUrls)
	if err != nil {
		return nil, err
	}

	pathMap = buildMap(pathUrls)
	return MapHandler(pathMap, fallback), nil
}

func SQLiteHandler(fileName string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var res pathUrl
		db, err := sqlx.Open("sqlite3", fileName)
		if err != nil {
			fallback.ServeHTTP(w, r)
		}
		defer db.Close()
		err = db.Get(&res, "SELECT * FROM urls WHERE path = $1", r.URL.Path)
		fmt.Printf("result object %#v\n", res)
		if err != nil {
			fmt.Printf("%#v", err)
			fallback.ServeHTTP(w, r)

		}
		http.Redirect(w, r, res.URL, http.StatusFound)
	}
}

func FileHandler(fileName, fileType string, fallback http.Handler) (http.HandlerFunc, error) {
	pathMap := make(map[string]string)
	var byteValue []byte
	if fileType != "sqlite" {
		file, err := os.Open(fileName)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		byteValue, err = io.ReadAll(file)
		if err != nil {
			return nil, err
		}
	}

	switch fileType {
	case "json":
		return JSONHandler(byteValue, fallback)
	case "yaml":
		return YAMLHandler(byteValue, fallback)
	case "sqlite":
		return SQLiteHandler(fileName, fallback), nil
	default:
		return MapHandler(pathMap, fallback), &WrongTypeError{
			Type: fileType,
			Err:  errors.New("unavailable"),
		}
	}
}

func buildMap(urls []pathUrl) map[string]string {
	urlDict := make(map[string]string)

	for _, url := range urls {
		urlDict[url.Path] = url.URL
	}
	return urlDict
}
