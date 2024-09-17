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

	// Create tables if they don't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS torrents (
		infohashHex TEXT PRIMARY KEY,
		name TEXT,
		length INTEGER,
		meta BLOB,
		seeds INTEGER DEFAULT 0,
		leechers INTEGER DEFAULT 0,
		added DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		torrentInfohashHex TEXT,
		name TEXT,
		length INTEGER,
		FOREIGN KEY (torrentInfohashHex) REFERENCES torrents(infohashHex)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("SQLite database initialized")
}

func insertTorrent(t *torrent, meta []byte) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO torrents (infohashHex, name, length, meta, seeds, leechers) VALUES (?, ?, ?, ?, ?, ?)`,
		t.InfohashHex, t.Name, t.Length, meta, t.Seeds, t.Leechers)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, file := range t.Files {
		_, err = tx.Exec(`INSERT INTO files (torrentInfohashHex, name, length) VALUES (?, ?, ?)`,
			t.InfohashHex, file.Name, file.Length)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func getAllTorrents(from, size int) ([]*torrent, error) {
	query := `SELECT t.infohashHex, t.name, t.length, t.seeds, t.leechers, t.added, f.name, f.length
			  FROM torrents t
			  LEFT JOIN files f ON t.infohashHex = f.torrentInfohashHex
			  LIMIT ? OFFSET ?`
	rows, err := db.Query(query, size, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	torrentsMap := make(map[string]*torrent)
	for rows.Next() {
		var infohashHex, name, fileName string
		var length, fileLength int64
		var added string
		var seeds, leechers int
		if err := rows.Scan(&infohashHex, &name, &length, &seeds, &leechers, &added, &fileName, &fileLength); err != nil {
			log.Println(err)
			continue
		}

		t, exists := torrentsMap[infohashHex]
		if !exists {
			t = &torrent{InfohashHex: infohashHex, Name: name, Length: length}
			torrentsMap[infohashHex] = t
		}
		t.Files = append(t.Files, &tfile{Name: fileName, Length: fileLength})
	}

	var torrents []*torrent
	for _, t := range torrentsMap {
		torrents = append(torrents, t)
	}
	return torrents, nil
}

func searchTorrents(searchText string, from, size int) ([]*torrent, error) {
	query := `SELECT infohashHex, name, length, files, added FROM torrents WHERE name LIKE ? LIMIT ? OFFSET ?`
	rows, err := db.Query(query, "%"+searchText+"%", size, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var torrents []*torrent
	for rows.Next() {
		var t torrent
		var files string
		var added string
		if err := rows.Scan(&t.InfohashHex, &t.Name, &t.Length, &files, &added); err != nil {
			log.Println(err)
			continue
		}
		t.Files = deserializeFiles(files)
		torrents = append(torrents, &t)
	}
	return torrents, nil
}

func getTorrentsByHashes(hashes []string) ([]*torrent, error) {
	query := `SELECT infohashHex, name, length, files, added FROM torrents WHERE infohashHex IN (?` + strings.Repeat(",?", len(hashes)-1) + `)`
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
		var added string
		if err := rows.Scan(&t.InfohashHex, &t.Name, &t.Length, &files, &added); err != nil {
			log.Println(err)
			continue
		}
		t.Files = deserializeFiles(files)
		torrents = append(torrents, &t)
	}
	return torrents, nil
}

func deleteTorrents(hashes []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	query := `DELETE FROM files WHERE torrentInfohashHex IN (?` + strings.Repeat(",?", len(hashes)-1) + `)`
	_, err = tx.Exec(query, toInterfaceSlice(hashes)...)
	if err != nil {
		tx.Rollback()
		return err
	}

	query = `DELETE FROM torrents WHERE infohashHex IN (?` + strings.Repeat(",?", len(hashes)-1) + `)`
	_, err = tx.Exec(query, toInterfaceSlice(hashes)...)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func toInterfaceSlice(strings []string) []interface{} {
	result := make([]interface{}, len(strings))
	for i, s := range strings {
		result[i] = s
	}
	return result
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
