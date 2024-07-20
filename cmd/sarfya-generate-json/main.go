package main

import (
	"context"
	"flag"
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya/adapters/fwewdictionary"
	"github.com/gissleh/sarfya/adapters/jsonstorage"
	"github.com/gissleh/sarfya/adapters/placeholderdictionary"
	"github.com/gissleh/sarfya/adapters/sourcestorage"
	"log"
)

var flagSourceDir = flag.String("source-dir", "./data", "Source directory")
var flagOutputFile = flag.String("output-file", "./data-compiled.json", "Output file name")

func main() {
	flag.Parse()

	dict := sarfya.CombinedDictionary{
		fwewdictionary.Global(),
		placeholderdictionary.New(),
	}

	sourceStorage, err := sourcestorage.Open(context.Background(), *flagSourceDir, dict)
	if err != nil {
		log.Fatal("Failed to open source storage:", err)
	}

	destStorage := jsonstorage.New(*flagOutputFile)

	examples, err := sourceStorage.FetchExamples(context.Background(), nil, nil)
	if err != nil {
		log.Fatal("Failed to list examples from source storage:", err)
	}

	for _, example := range examples {
		err := destStorage.SaveExample(context.Background(), example)
		if err != nil {
			log.Fatal("Failed to save examples from source storage:", err)
		}
	}

	err = destStorage.WriteToFile()
	if err != nil {
		log.Fatal("Failed to write examples to destination file:", err)
	}
}
