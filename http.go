package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/marksamman/bencode"
)

//go:embed static/*
var staticFiles embed.FS

type searchResponse struct {
	Torrents []*torrent `json:"torrents"`
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

	// Implement logic to fetch all torrents from the SQLite database
	rows, err := db.Query(`SELECT infohashHex, name, length, files FROM torrents`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer rows.Close()

	var torrents []*torrent
	for rows.Next() {
		var t torrent
		var files string
		if err := rows.Scan(&t.InfohashHex, &t.Name, &t.Length, &files); err != nil {
			log.Println(err)
			continue
		}
		t.Files = deserializeFiles(files)
		torrents = append(torrents, &t)
	}

	response := searchResponse{
		Torrents: torrents,
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
	rows, err := db.Query(`SELECT infohashHex, name, length, files FROM torrents WHERE name LIKE ?`, "%"+searchText+"%")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer rows.Close()

	var torrents []*torrent
	for rows.Next() {
		var t torrent
		var files string
		if err := rows.Scan(&t.InfohashHex, &t.Name, &t.Length, &files); err != nil {
			log.Println(err)
			continue
		}
		t.Files = deserializeFiles(files)
		torrents = append(torrents, &t)
	}

	response := searchResponse{
		Torrents: torrents,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
}

func torrentHandler(w http.ResponseWriter, r *http.Request) {

	hashes := r.URL.Query()["h"]

	// Implement logic to fetch torrents by hashes from the SQLite database
	query := `SELECT infohashHex, name, length, files FROM torrents WHERE infohashHex IN (?` + strings.Repeat(",?", len(hashes)-1) + `)`
	args := make([]interface{}, len(hashes))
	for i, hash := range hashes {
		args[i] = hash
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer rows.Close()

	var torrents []*torrent
	for rows.Next() {
		var t torrent
		var files string
		if err := rows.Scan(&t.InfohashHex, &t.Name, &t.Length, &files); err != nil {
			log.Println(err)
			continue
		}
		t.Files = deserializeFiles(files)
		torrents = append(torrents, &t)
	}

	response := searchResponse{
		Torrents: torrents,
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

func torrentFileHandler(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("h")
	if hash == "" {
		http.Error(w, "Missing hash parameter", http.StatusBadRequest)
		return
	}

	meta, err := index.GetInternal([]byte(hash))
	if err != nil {
		http.Error(w, "Torrent not found", http.StatusNotFound)
		log.Println(err)
		return
	}

	// Parse the torrent to get the name
	torrent, err := parseTorrent(meta, hash)
	if err != nil {
		http.Error(w, "Error parsing torrent", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// decode data
	d, err := bencode.Decode(bytes.NewBuffer(meta))
	if err != nil {
		http.Error(w, "Torrent not decoded", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// re-encode with correct format
	ed := bencode.Encode(map[string]interface{}{
		"info": d,
	})

	// Use the torrent name for the filename, replacing any invalid characters
	filename := fmt.Sprintf("%s.torrent", sanitizeFilename(torrent.Name))

	w.Header().Set("Content-Type", "application/x-bittorrent")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	w.Write(ed)
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	// Implement logic to count torrents in the SQLite database
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM torrents`).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	response := map[string]int{"totalCount": count}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
}

func sanitizeFilename(name string) string {
	// Replace any characters that are not allowed in filenames
	return strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, name)
}

func startHTTP(port int) {

	http.HandleFunc("/query", Gzip(searchHandler))
	http.HandleFunc("/torrent", Gzip(torrentHandler))
	http.HandleFunc("/all", Gzip(allHandler))
	http.HandleFunc("/delete", Gzip(deleteHandler))           // Register the delete handler
	http.HandleFunc("/count", Gzip(countHandler))             // Register the count handler
	http.HandleFunc("/torrentfile", Gzip(torrentFileHandler)) // Register the new handler

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
