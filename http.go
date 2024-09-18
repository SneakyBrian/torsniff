package main

import (
	"bytes"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
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

	from := getQSInt(r.URL.Query(), "f", 0)
	size := getQSInt(r.URL.Query(), "s", 10)

	torrents, err := getAllTorrents(from, size)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	response := searchResponse{Torrents: torrents}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	searchText := r.URL.Query().Get("q")

	from := getQSInt(r.URL.Query(), "f", 0)
	size := getQSInt(r.URL.Query(), "s", 10)

	torrents, err := searchTorrents(searchText, from, size)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	response := searchResponse{Torrents: torrents}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
}

func torrentHandler(w http.ResponseWriter, r *http.Request) {

	hashes := r.URL.Query()["h"]

	torrents, err := getTorrentsByHashes(hashes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	response := searchResponse{Torrents: torrents}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	hashes := r.URL.Query()["h"]

	err := deleteTorrents(hashes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func torrentFileHandler(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("h")
	if hash == "" {
		http.Error(w, "Missing hash parameter", http.StatusBadRequest)
		return
	}

	meta, err := getTorrentMeta(hash)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Torrent not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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

	// Decode the files field
	d, err := bencode.Decode(bytes.NewBuffer(meta))
	if err != nil {
		http.Error(w, "Error decoding torrent", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Add trackers to the torrent metadata
	trackersMutex.RLock()
	defer trackersMutex.RUnlock()

	// Re-encode with correct format
	ed := bencode.Encode(map[string]interface{}{
		"info":          d,
		"announce-list": trackersList,
	})

	// Use the torrent name for the filename, replacing any invalid characters
	filename := fmt.Sprintf("%s.torrent", sanitizeFilename(torrent.Name))

	w.Header().Set("Content-Type", "application/x-bittorrent")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	w.Write(ed)
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

func trackersHandler(w http.ResponseWriter, r *http.Request) {
	trackersMutex.RLock()
	response := map[string][]string{"trackers": trackersList}
	trackersMutex.RUnlock()

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan int)
	mutex     sync.Mutex
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		return
	}
	defer conn.Close()

	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	// Fetch the current torrent count
	currentCount, err := countTorrents()
	if err != nil {
		log.Println("Error fetching torrent count:", err)
	} else {
		// Use notifyClients to send the current count to the newly connected client
		notifyClients(currentCount)
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			break
		}
	}
}

func notifyClients(count int) {
	mutex.Lock()
	defer mutex.Unlock()
	for client := range clients {
		err := client.WriteJSON(map[string]int{"totalCount": count})
		if err != nil {
			log.Println("Error writing to websocket:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func startWebSocketServer() {
	http.HandleFunc("/ws", wsHandler)
	go func() {
		for {
			count := <-broadcast
			notifyClients(count)
		}
	}()
}

func startHTTP(port int) {
	http.HandleFunc("/query", Gzip(searchHandler))
	http.HandleFunc("/torrent", Gzip(torrentHandler))
	http.HandleFunc("/all", Gzip(allHandler))
	http.HandleFunc("/delete", Gzip(deleteHandler)) // Register the delete handler
	http.HandleFunc("/torrentfile", Gzip(torrentFileHandler))
	http.HandleFunc("/trackers", Gzip(trackersHandler)) // Register the trackers handler

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
