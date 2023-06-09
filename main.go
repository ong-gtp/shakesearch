package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		totalV, ok := r.URL.Query()["total"]
		total := -1
		if ok {
			t, err := strconv.Atoi(totalV[0])
			if err != nil {
				total = -1
			} else {
				total = t
			}
		}

		results := searcher.Search(query[0], total)
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New([]byte(strings.ToLower(string(dat))))
	return nil
}

func (s *Searcher) Search(query string, total int) []string {

	if total <= 0 {
		total = -1
	}

	idxs := s.SuffixArray.Lookup([]byte(strings.ToLower(query)), total)
	results := []string{}
	lastWordIndex := len(s.CompleteWorks)
	// lastWordIndex := strings.LastIndex(s.CompleteWorks, "eBooks.")
	for _, idx := range idxs {
		if (idx-250) < 0 || (idx+250) > lastWordIndex {
			continue
		}
		subset := s.CompleteWorks[idx-250 : idx+250]
		subset = strings.ReplaceAll(subset, query, fmt.Sprintf("<b>%s</b>", query))
		results = append(results, subset)
	}
	return results
}
