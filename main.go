package main

import (
	"github.com/53jk1/jpeg-resizer/resizer"
	lru "github.com/hashicorp/golang-lru"
	"log"
	"net/http"
	"time"
)

type service struct {
	cache *lru.Cache
}

func main() {
	cache, err := lru.New(1024)
	if err != nil {
		log.Panicf("Failed to create cache: %v", err)
	}

	svc := &service{cache: cache}

	mux := http.NewServeMux()

	mux.Handle("/v1/resize", http.TimeoutHandler(&Resize{Resizer: &resizer.Resize{Cache: svc.cache}}, 10*time.Second, "Request has timed out"))

	mux.Handle("/v1/image/", svc.getImageHandler())

	address := resizer.HOSTPORT

	log.Print("Listening on ", resizer.HOSTPORT)
	// When running on docker mac, can't listen only on localhost
	panic(http.ListenAndServe(address, mux))
}
