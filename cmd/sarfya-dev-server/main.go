package main

import (
	"context"
	"flag"
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya/adapters/fwewdictionary"
	"github.com/gissleh/sarfya/adapters/placeholderdictionary"
	"github.com/gissleh/sarfya/adapters/sourcestorage"
	"github.com/gissleh/sarfya/adapters/webapi"
	"github.com/gissleh/sarfya/service"
	"log"
)

var flagSourceDir = flag.String("source-dir", "./data", "Source directory")

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

	api, errCh := webapi.Setup("localhost:8080")

	webapi.Utils(api.Group("/api/utils"), dict)
	webapi.Examples(api.Group("/api/examples"), svc)

	err = <-errCh
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}
}
