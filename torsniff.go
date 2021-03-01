package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"

	"github.com/marksamman/bencode"
	"github.com/spf13/cobra"
)

var (
	index bleve.Index
)

type tfile struct {
	name   string
	length int64
}

func (t *tfile) String() string {
	return fmt.Sprintf("name: %s\n, size: %d\n", t.name, t.length)
}

type torrent struct {
	infohashHex string
	name        string
	length      int64
	files       []*tfile
}

func (t *torrent) String() string {
	return fmt.Sprintf(
		"link: %s\nname: %s\nsize: %d\nfile: %d\n",
		fmt.Sprintf("magnet:?xt=urn:btih:%s", t.infohashHex),
		t.name,
		t.length,
		len(t.files),
	)
}

func parseTorrent(meta []byte, infohashHex string) (*torrent, error) {
	dict, err := bencode.Decode(bytes.NewBuffer(meta))
	if err != nil {
		return nil, err
	}

	t := &torrent{infohashHex: infohashHex}
	if name, ok := dict["name.utf-8"].(string); ok {
		t.name = name
	} else if name, ok := dict["name"].(string); ok {
		t.name = name
	}
	if length, ok := dict["length"].(int64); ok {
		t.length = length
	}

	var totalSize int64
	var extractFiles = func(file map[string]interface{}) {
		var filename string
		var filelength int64
		if inter, ok := file["path.utf-8"].([]interface{}); ok {
			name := make([]string, len(inter))
			for i, v := range inter {
				name[i] = fmt.Sprint(v)
			}
			filename = strings.Join(name, "/")
		} else if inter, ok := file["path"].([]interface{}); ok {
			name := make([]string, len(inter))
			for i, v := range inter {
				name[i] = fmt.Sprint(v)
			}
			filename = strings.Join(name, "/")
		}
		if length, ok := file["length"].(int64); ok {
			filelength = length
			totalSize += filelength
		}
		t.files = append(t.files, &tfile{name: filename, length: filelength})
	}

	if files, ok := dict["files"].([]interface{}); ok {
		for _, file := range files {
			if f, ok := file.(map[string]interface{}); ok {
				extractFiles(f)
			}
		}
	}

	if t.length == 0 {
		t.length = totalSize
	}
	if len(t.files) == 0 {
		t.files = append(t.files, &tfile{name: t.name, length: t.length})
	}

	return t, nil
}

type torsniff struct {
	laddr      string
	maxFriends int
	maxPeers   int
	secret     string
	timeout    time.Duration
	blacklist  *blackList
}

func (t *torsniff) run() error {
	tokens := make(chan struct{}, t.maxPeers)

	dht, err := newDHT(t.laddr, t.maxFriends)
	if err != nil {
		return err
	}

	dht.run()

	log.Println("running, it may take a few minutes...")

	for {
		select {
		case <-dht.announcements.wait():
			for {
				if ac := dht.announcements.get(); ac != nil {
					tokens <- struct{}{}
					go t.work(ac, tokens)
					continue
				}
				break
			}
		case <-dht.die:
			return dht.errDie
		}
	}

}

func (t *torsniff) work(ac *announcement, tokens chan struct{}) {
	defer func() {
		<-tokens
	}()

	peerAddr := ac.peer.String()
	if t.blacklist.has(peerAddr) {
		return
	}

	wire := newMetaWire(string(ac.infohash), peerAddr, t.timeout)
	defer wire.free()

	meta, err := wire.fetch()
	if err != nil {
		t.blacklist.add(peerAddr)
		return
	}

	torrent, err := parseTorrent(meta, ac.infohashHex)
	if err != nil {
		return
	}

	log.Printf("indexing torrent %s...", torrent.infohashHex)

	// index the torrent
	index.Index(torrent.infohashHex, torrent)

	log.Println(torrent)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	searchText := r.URL.Query().Get("s")

	// search for some text
	query := bleve.NewMatchQuery(searchText)
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = json.NewEncoder(w).Encode(searchResults)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
}

func main() {
	log.SetFlags(0)

	var addr string
	var port uint16
	var peers int
	var timeout time.Duration
	var verbose bool
	var friends int

	root := &cobra.Command{
		Use:          "torsniff",
		Short:        "torsniff - A sniffer that sniffs torrents from BitTorrent network.",
		SilenceUsage: true,
	}
	root.RunE = func(cmd *cobra.Command, args []string) error {

		var err error

		log.SetOutput(ioutil.Discard)
		if verbose {
			log.SetOutput(os.Stdout)
		}

		const indexFile = "torsniff.index"

		index, err = bleve.Open(indexFile)
		if err == bleve.ErrorIndexPathDoesNotExist {
			log.Println("Creating new index...")

			indexMapping := bleve.NewIndexMapping()
			index, err = bleve.New(indexFile, indexMapping)
			if err != nil {
				log.Fatal(err)
			}
		} else if err != nil {
			log.Fatal(err)
		} else {
			log.Println("Opening existing index...")
		}

		p := &torsniff{
			laddr:      net.JoinHostPort(addr, strconv.Itoa(int(port))),
			timeout:    timeout,
			maxFriends: friends,
			maxPeers:   peers,
			secret:     string(randBytes(20)),
			blacklist:  newBlackList(5*time.Minute, 50000),
		}
		go p.run()

		http.HandleFunc("/query", searchHandler)

		return http.ListenAndServe(":8090", nil)
	}

	root.Flags().StringVarP(&addr, "addr", "a", "", "listen on given address (default all, ipv4 and ipv6)")
	root.Flags().Uint16VarP(&port, "port", "p", 6881, "listen on given port")
	root.Flags().IntVarP(&friends, "friends", "f", 500, "max fiends to make with per second")
	root.Flags().IntVarP(&peers, "peers", "e", 400, "max peers to connect to download torrents")
	root.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "max time allowed for downloading torrents")
	root.Flags().BoolVarP(&verbose, "verbose", "v", true, "run in verbose mode")

	if err := root.Execute(); err != nil {
		fmt.Println(fmt.Errorf("could not start: %s", err))
	}
}
