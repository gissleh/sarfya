package main

import (
	"context"
	"flag"
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya/adapters/fwewdictionary"
	"github.com/gissleh/sarfya/adapters/placeholderdictionary"
	"github.com/gissleh/sarfya/adapters/sourcestorage"
	"github.com/gissleh/sarfya/adapters/templfrontend"
	"github.com/gissleh/sarfya/adapters/webapi"
	"github.com/gissleh/sarfya/service"
	"log"
	"strings"
)

var flagSourceDir = flag.String("source-dir", "./data", "Source directory")
var flagListenAddr = flag.String("listen", ":8080", "Listen address")

func main() {
	dict := sarfya.CombinedDictionary{
		fwewdictionary.Global(),
		placeholderdictionary.New(),
	}

	storage, err := sourcestorage.Open(context.Background(), *flagSourceDir, dict)
	if err != nil {
		log.Fatal("Failed to open storage:", err)
	}
	log.Println("Examples loaded:", storage.ExampleCount())

	svc := &service.Service{Dictionary: dict, Storage: storage}

	api, errCh := webapi.Setup(*flagListenAddr)

	webapi.Utils(api.Group("/api/utils"), dict)
	webapi.Examples(api.Group("/api/examples"), svc)
	templfrontend.Endpoints(api.Group(""), svc)

	go func() {
		example, err := storage.ListExamples(context.Background())
		if err != nil {
			return
		}

		exists := make(map[string]bool)
		for _, example := range example {
			rt := strings.TrimSpace(example.Text.RawText())
			if exists[rt] {
				log.Println("Duplicate example:", example.ID, example.Text.String())
			}

			exists[rt] = true
		}
	}()

	log.Println("Listening on", *flagListenAddr)

	err = <-errCh
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}
}
