package main

import (
	"flag"
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya/adapters/fwewdictionary"
	"github.com/gissleh/sarfya/adapters/jsonstorage"
	"github.com/gissleh/sarfya/adapters/placeholderdictionary"
	"github.com/gissleh/sarfya/adapters/templfrontend"
	"github.com/gissleh/sarfya/adapters/webapi"
	"github.com/gissleh/sarfya/service"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var flagSourceFile = flag.String("source-file", "./data-compiled.json", "File containing data.")
var flagListenAddr = flag.String("listen", ":8080", "Listen address")

func main() {
	dict := sarfya.CombinedDictionary{
		fwewdictionary.Global(),
		placeholderdictionary.New(),
	}

	storage, err := jsonstorage.Open(*flagSourceFile, true)
	if err != nil {
		log.Fatalln("Failed to open json storage:", err)
		return
	}

	svc := &service.Service{Dictionary: dict, Storage: storage, ReadOnly: true}
	api, errCh := webapi.Setup(*flagListenAddr)

	webapi.Utils(api.Group("/api/utils"), dict)
	webapi.Examples(api.Group("/api/examples"), svc)
	templfrontend.Endpoints(api.Group(""), svc)

	log.Println("Listening on", *flagListenAddr)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-signalCh:
		log.Println("Shutting down due to signal:", sig)
	case err := <-errCh:
		log.Fatal("Failed to listen:", err)
	}
}
