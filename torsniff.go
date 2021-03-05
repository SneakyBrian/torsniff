package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/marksamman/bencode"
	"github.com/spf13/cobra"
)

type tfile struct {
	Name   string `json:"name"`
	Length int64  `json:"length"`
}

func (t *tfile) String() string {
	return fmt.Sprintf("name: %s\n, size: %d\n", t.Name, t.Length)
}

type torrent struct {
	InfohashHex string   `json:"infohashHex"`
	Name        string   `json:"name"`
	Length      int64    `json:"length"`
	Files       []*tfile `json:"files"`
	IndexType   string   `json:"indexType"`
}

func (t *torrent) String() string {
	return fmt.Sprintf(
		"link: %s\nname: %s\nsize: %d\nfile: %d\n",
		fmt.Sprintf("magnet:?xt=urn:btih:%s", t.InfohashHex),
		t.Name,
		t.Length,
		len(t.Files),
	)
}

func parseTorrent(meta []byte, infohashHex string) (*torrent, error) {
	dict, err := bencode.Decode(bytes.NewBuffer(meta))
	if err != nil {
		return nil, err
	}

	t := &torrent{InfohashHex: infohashHex}
	if name, ok := dict["name.utf-8"].(string); ok {
		t.Name = name
	} else if name, ok := dict["name"].(string); ok {
		t.Name = name
	}
	if length, ok := dict["length"].(int64); ok {
		t.Length = length
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
		t.Files = append(t.Files, &tfile{Name: filename, Length: filelength})
	}

	if files, ok := dict["files"].([]interface{}); ok {
		for _, file := range files {
			if f, ok := file.(map[string]interface{}); ok {
				extractFiles(f)
			}
		}
	}

	if t.Length == 0 {
		t.Length = totalSize
	}
	if len(t.Files) == 0 {
		t.Files = append(t.Files, &tfile{Name: t.Name, Length: t.Length})
	}

	t.IndexType = "torrent"

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

	ticker := time.NewTicker(5 * time.Second)

	go func() {
		var lastCount int
		for {
			<-ticker.C
			count := dht.peerCount()
			if count > lastCount {
				log.Printf("got %d peers (+%d)", count, count-lastCount)
				lastCount = count
			}
		}
	}()

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

	if t.isTorrentExist(ac.infohashHex) {
		return
	}

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

	log.Printf("indexing torrent %s...", torrent.InfohashHex)

	index.SetInternal([]byte(ac.infohashHex), meta)

	// index the torrent
	index.Index(torrent.InfohashHex, torrent)

	log.Println(torrent)
}

func (t *torsniff) isTorrentExist(infohashHex string) bool {
	_, err := index.GetInternal([]byte(infohashHex))
	return err == nil
}

func main() {
	log.SetFlags(0)

	var addr string
	var port uint16
	var peers int
	var timeout time.Duration
	var verbose bool
	var friends int

	fmt.Println("starting...")

	root := &cobra.Command{
		Use:          "torsniff-search",
		Short:        "torsniff-search - A sniffer that sniffs torrents from BitTorrent network.",
		SilenceUsage: true,
	}
	root.RunE = func(cmd *cobra.Command, args []string) error {

		log.SetOutput(ioutil.Discard)
		if verbose {
			log.SetOutput(os.Stdout)
		}

		startIndex()

		p := &torsniff{
			laddr:      net.JoinHostPort(addr, strconv.Itoa(int(port))),
			timeout:    timeout,
			maxFriends: friends,
			maxPeers:   peers,
			secret:     string(randBytes(20)),
			blacklist:  newBlackList(5*time.Minute, 50000),
		}
		go p.run()

		startHTTP()

		return nil
	}

	root.Flags().StringVarP(&addr, "addr", "a", "", "listen on given address (default all, ipv4 and ipv6)")
	root.Flags().Uint16VarP(&port, "port", "p", 6881, "listen on given port")
	root.Flags().IntVarP(&friends, "friends", "f", 500, "max fiends to make with per second")
	root.Flags().IntVarP(&peers, "peers", "e", 400, "max peers to connect to download torrents")
	root.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "max time allowed for downloading torrents")
	root.Flags().BoolVarP(&verbose, "verbose", "v", true, "run in verbose mode")

	if err := root.Execute(); err != nil {
		log.Fatal(fmt.Errorf("could not start: %s", err))
	}

	// wait for signal to shut down
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	log.Printf("we get signal! %s", sig)

	log.Println("closing index...")
	index.Close()
	fmt.Println("exiting...")
}
