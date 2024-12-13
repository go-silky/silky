package main

import (
	"log"
	"net/http"

	"github.com/go-silky/silky/example/basic/xhttp"
)

func main() {
	r := xhttp.SetupRouter()
	if err := http.ListenAndServe(":3333", r); err != nil {
		log.Fatal(err)
	}
}
