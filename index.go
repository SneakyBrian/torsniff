package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

var (
	db *sql.DB
)

func startIndex() {
	var err error

	db, err = sql.Open("sqlite", "torsniff.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS torrents (
		infohashHex TEXT PRIMARY KEY,
		name TEXT,
		length INTEGER,
		files TEXT,
		meta BLOB
	)`)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("SQLite database initialized")
}

func insertTorrent(t *torrent, meta []byte) error {
	_, err := db.Exec(`INSERT INTO torrents (infohashHex, name, length, files, meta) VALUES (?, ?, ?, ?, ?)`,
		t.InfohashHex, t.Name, t.Length, serializeFiles(t.Files), meta)
	return err
}

func getAllTorrents(from, size int) ([]*torrent, error) {
	query := `SELECT infohashHex, name, length, files FROM torrents LIMIT ? OFFSET ?`
	rows, err := db.Query(query, size, from)
	if err != nil {
		return nil, err
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
	return torrents, nil
}

func searchTorrents(searchText string, from, size int) ([]*torrent, error) {
	query := `SELECT infohashHex, name, length, files FROM torrents WHERE name LIKE ? LIMIT ? OFFSET ?`
	rows, err := db.Query(query, "%"+searchText+"%", size, from)
	if err != nil {
		return nil, err
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
	return torrents, nil
}

func getTorrentsByHashes(hashes []string) ([]*torrent, error) {
	query := `SELECT infohashHex, name, length, files FROM torrents WHERE infohashHex IN (?` + strings.Repeat(",?", len(hashes)-1) + `)`
	args := make([]interface{}, len(hashes))
	for i, hash := range hashes {
		args[i] = hash
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
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
	return torrents, nil
}

func deleteTorrents(hashes []string) error {
	query := `DELETE FROM torrents WHERE infohashHex IN (?` + strings.Repeat(",?", len(hashes)-1) + `)`
	args := make([]interface{}, len(hashes))
	for i, hash := range hashes {
		args[i] = hash
	}

	_, err := db.Exec(query, args...)
	return err
}

func getTorrentMeta(hash string) ([]byte, error) {
	var meta []byte
	err := db.QueryRow(`SELECT meta FROM torrents WHERE infohashHex = ?`, hash).Scan(&meta)
	return meta, err
}

func countTorrents() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM torrents`).Scan(&count)
	return count, err
}

func isTorrentExist(infohashHex string) bool {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM torrents WHERE infohashHex = ?)`, infohashHex).Scan(&exists)
	if err != nil {
		log.Println(err)
		return false
	}
	return exists
}

func serializeFiles(files []*tfile) string {
	data, err := json.Marshal(files)
	if err != nil {
		log.Printf("Error serializing files: %v", err)
		return ""
	}
	return string(data)
}

func deserializeFiles(files string) []*tfile {
	var tfiles []*tfile
	err := json.Unmarshal([]byte(files), &tfiles)
	if err != nil {
		log.Printf("Error deserializing files: %v", err)
		return nil
	}
	return tfiles
}
