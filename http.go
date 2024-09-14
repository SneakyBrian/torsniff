package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/blevesearch/bleve/v2"
)

//go:embed static/*
var staticFiles embed.FS

type searchResponse struct {
	SearchResults *bleve.SearchResult `json:"search"`
	Torrents      []*torrent          `json:"torrents"`
}

func getTorrentsFromSearch(searchResults *bleve.SearchResult) []*torrent {

	var torrents []*torrent

	for _, hit := range searchResults.Hits {

		meta, err := index.GetInternal([]byte(hit.ID))
		if err != nil {
			log.Println(err)
			continue
		}

		torrent, err := parseTorrent(meta, hit.ID)
		if err != nil {
			log.Println(err)
			continue
		}

		torrents = append(torrents, torrent)
	}

	return torrents
}

func getQSInt(qs url.Values, key string, defaultValue int) int {

	val := defaultValue

	valstr := qs.Get(key)
	if valstr != "" {
		i, err := strconv.Atoi(valstr)
		if err == nil {
			val = i
		}
	}

	return val

}

func allHandler(w http.ResponseWriter, r *http.Request) {

	query := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)

	searchRequest.From = getQSInt(r.URL.Query(), "f", searchRequest.From)
	searchRequest.Size = getQSInt(r.URL.Query(), "s", searchRequest.Size)

	searchResults, err := index.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	response := searchResponse{
		SearchResults: searchResults,
		Torrents:      getTorrentsFromSearch(searchResults),
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	searchText := r.URL.Query().Get("q")

	// search for some text
	query := bleve.NewQueryStringQuery(searchText)
	searchRequest := bleve.NewSearchRequest(query)

	searchRequest.From = getQSInt(r.URL.Query(), "f", searchRequest.From)
	searchRequest.Size = getQSInt(r.URL.Query(), "s", searchRequest.Size)

	searchResults, err := index.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	response := searchResponse{
		SearchResults: searchResults,
		Torrents:      getTorrentsFromSearch(searchResults),
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
}

func torrentHandler(w http.ResponseWriter, r *http.Request) {

	hashes := r.URL.Query()["h"]

	query := bleve.NewDocIDQuery(hashes)
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	response := searchResponse{
		SearchResults: searchResults,
		Torrents:      getTorrentsFromSearch(searchResults),
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	hashes := r.URL.Query()["h"]

	for _, hash := range hashes {
		err := index.Delete(hash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	docCount, err := index.DocCount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	response := map[string]uint64{"totalCount": docCount}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
}

func startHTTP(port int) {

	http.HandleFunc("/query", Gzip(searchHandler))
	http.HandleFunc("/torrent", Gzip(torrentHandler))
	http.HandleFunc("/all", Gzip(allHandler))
	http.HandleFunc("/delete", Gzip(deleteHandler)) // Register the delete handler
	http.HandleFunc("/count", Gzip(countHandler))   // Register the count handler

	// Create a file system from the embedded files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}

	// Serve embedded static files
	http.Handle("/", http.FileServer(http.FS(staticFS)))

	address := fmt.Sprintf(":%d", port) // Use the provided port
	go http.ListenAndServe(address, nil)
}
