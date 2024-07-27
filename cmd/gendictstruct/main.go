package main

import (
	"context"
	"fmt"
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya/adapters/fwewdictionary"
	"log"
	"os"
	"strings"
)

func main() {
	dict := sarfya.WithDerivedPoS(fwewdictionary.Global())
	args := strings.Join(os.Args[1:], " ")

	res, err := dict.Lookup(context.Background(), args)
	if err != nil {
		log.Fatalln(err)
		return
	}

	for _, res := range res {
		res.Definitions = map[string]string{"en": res.Definitions["en"]}
		fmt.Printf("Struct: %#+v\n", res)
		fmt.Printf("Filter: %#+v\n", res.ToFilter().String())
	}
}
