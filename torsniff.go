package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
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
	// log.Printf("Parsing torrent for infohash: %s", infohashHex)
	dict, err := bencode.Decode(bytes.NewBuffer(meta))
	if err != nil {
		log.Printf("Error parsing torrent: %v", err)
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

	// log.Printf("Parsed torrent: %+v", t)
	return t, nil
}

type torsniff struct {
	laddr      string
	maxFriends int
	maxPeers   int
	secret     string
	timeout    time.Duration
	blacklist  *blackList
	maxRetries int // New field for max retries
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
	log.Printf("Processing announcement for infohash: %s", ac.infohashHex)
	defer func() {
		<-tokens
	}()

	if isTorrentExist(ac.infohashHex) {
		log.Printf("infohash %s already exists", ac.infohashHex)
		return
	}

	peerAddr := ac.peer.String()
	if t.blacklist.has(peerAddr) {
		log.Printf("peer %s already blacklisted", peerAddr)
		return
	}

	var meta []byte
	var err error

	for attempt := 1; attempt <= t.maxRetries; attempt++ { // Use the maxRetries field
		wire := newMetaWire(string(ac.infohash), peerAddr, t.timeout)
		defer wire.free()

		meta, err = wire.fetch()
		if err == nil {
			break
		}

		log.Printf("Attempt %d to fetch meta failed for peer %s: %v", attempt, peerAddr, err)

		// Exponential backoff delay
		backoffDuration := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		log.Printf("Waiting for %v before retrying...", backoffDuration)
		time.Sleep(backoffDuration)
	}

	if err != nil {
		log.Printf("adding peer %s to blacklist after %d failed attempts", peerAddr, t.maxRetries)
		t.blacklist.add(peerAddr)
		return
	}

	torrent, err := parseTorrent(meta, ac.infohashHex)
	if err != nil {
		log.Printf("error parsing torrent: %v", err)
		return
	}

	log.Printf("Indexing torrent: %s", torrent.InfohashHex)

	err = insertTorrent(torrent, meta)
	if err != nil {
		log.Printf("error inserting torrent into database: %v", err)
		return
	}

	log.Println(torrent)
}

var trackersList []string

const trackerURL = "https://raw.githubusercontent.com/ngosang/trackerslist/master/trackers_best.txt"

func downloadTrackers() error {
	resp, err := http.Get(trackerURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download trackers: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Split the body into lines and store them in the trackersList
	trackersList = strings.Split(string(body), "\n")
	return nil
}

func startTrackerDownloadScheduler() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			err := downloadTrackers()
			if err != nil {
				log.Printf("Error downloading trackers: %v", err)
			} else {
				log.Println("Successfully downloaded trackers")
			}
			<-ticker.C
		}
	}()
}

func main() {
	// log.SetFlags(0)

	var addr string
	var port int // Change to int to allow -1 as a default value
	var peers int
	var timeout time.Duration
	var verbose bool
	var friends int
	var httpPort int
	var maxRetries int
	var enableHTTPPortMapping bool // New variable for enabling HTTP port mapping

	fmt.Println("starting...")

	root := &cobra.Command{
		Use:          "torsniff",
		Short:        "torsniff - A sniffer that sniffs torrents from BitTorrent network.",
		SilenceUsage: true,
	}
	root.RunE = func(cmd *cobra.Command, args []string) error {

		log.SetOutput(io.Discard)
		if verbose {
			log.SetOutput(os.Stdout)
		}

		startIndex()

		startTrackerDownloadScheduler()

		// Create a new random generator
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		if port == -1 {
			port = rng.Intn(1000) + 6000 // Random port between 6000 and 6999
			log.Printf("No DHT port specified, using random port: %d", port)
		}
		portMappings := []PortMapping{
			{Port: int(port), Protocol: "UDP"},
		}

		// Conditionally add HTTP port mapping
		if enableHTTPPortMapping {
			portMappings = append(portMappings, PortMapping{Port: httpPort, Protocol: "TCP"})
		}

		err := SetupPortForwarding(portMappings)
		if err != nil {
			log.Printf("Warning: Failed to set up port forwarding: %v", err)
		}

		p := &torsniff{
			laddr:      net.JoinHostPort(addr, strconv.Itoa(int(port))),
			timeout:    timeout,
			maxFriends: friends,
			maxPeers:   peers,
			secret:     string(randBytes(20)),
			blacklist:  newBlackList(5*time.Minute, 50000),
			maxRetries: maxRetries,
		}
		go p.run()

		startHTTP(httpPort) // Pass the HTTP port to startHTTP

		return nil
	}

	root.Flags().StringVarP(&addr, "addr", "a", "", "listen on given address (default all, ipv4 and ipv6)")
	root.Flags().IntVarP(&port, "port", "p", -1, "listen on given port") // Default to -1
	root.Flags().IntVarP(&friends, "friends", "f", 500, "max fiends to make with per second")
	root.Flags().IntVarP(&peers, "peers", "e", 400, "max peers to connect to download torrents")
	root.Flags().DurationVarP(&timeout, "timeout", "t", 30*time.Second, "max time allowed for downloading torrents")
	root.Flags().BoolVarP(&verbose, "verbose", "v", true, "run in verbose mode")
	root.Flags().IntVarP(&httpPort, "http-port", "H", 8090, "HTTP server port")
	root.Flags().IntVarP(&maxRetries, "max-retries", "r", 3, "maximum number of retries to fetch metadata") // New flag for max retries

	root.Flags().BoolVarP(&enableHTTPPortMapping, "enable-http-port-mapping", "m", false, "enable HTTP port mapping for UPnP") // New flag with short option

	if err := root.Execute(); err != nil {
		log.Fatal(fmt.Errorf("could not start: %s", err))
	}

	// wait for signal to shut down
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	log.Printf("we get signal! %s", sig)

	log.Println("closing database...")
	db.Close()
	fmt.Println("exiting...")
}
