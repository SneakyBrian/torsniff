package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/blevesearch/bleve/v2"
)

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

func startHTTP() {

	http.HandleFunc("/query", Gzip(searchHandler))
	http.HandleFunc("/torrent", Gzip(torrentHandler))
	http.HandleFunc("/all", Gzip(allHandler))

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	go http.ListenAndServe(":8090", nil)
}
