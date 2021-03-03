package main

import (
	"log"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/simple"
)

var (
	index bleve.Index
)

func startIndex() {

	var err error

	const indexFile = "torsniff.index"

	index, err = bleve.Open(indexFile)
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Println("creating new index...")

		indexMapping := bleve.NewIndexMapping()

		torrentMapping := bleve.NewDocumentMapping()

		// // a generic reusable mapping for english text
		// englishTextFieldMapping := bleve.NewTextFieldMapping()
		// englishTextFieldMapping.Analyzer = en.AnalyzerName

		// simpleTextFieldMapping := bleve.NewTextFieldMapping()
		// simpleTextFieldMapping.Analyzer = simple.Name

		// a generic reusable mapping for keyword text
		keywordFieldMapping := bleve.NewTextFieldMapping()
		keywordFieldMapping.Analyzer = keyword.Name

		// torrentMapping.AddFieldMappingsAt("Name", simpleTextFieldMapping)
		torrentMapping.AddFieldMappingsAt("InfohashHex", keywordFieldMapping)

		indexMapping.AddDocumentMapping("torrent", torrentMapping)

		indexMapping.TypeField = "IndexType"
		indexMapping.DefaultAnalyzer = simple.Name

		index, err = bleve.New(indexFile, indexMapping)
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	} else {
		log.Println("opening existing index...")
	}

	docCount, _ := index.DocCount()

	log.Printf("index contains %d documents", docCount)

}
