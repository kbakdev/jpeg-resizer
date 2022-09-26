package main

import (
	"container/list"
	"encoding/json"
	"github.com/53jk1/Server-Coding/requests"
	"github.com/53jk1/Server-Coding/resizer"
	"github.com/53jk1/Server-Coding/responses"
	"io"
	"log"
	"net/http"
)

type Resize struct {
	Resizer      resizer.Resizer
	handlerQueue *list.List
}

func (s *Resize) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Expecting POST request"))
		return
	}

	request := requests.Resize{}
	err := json.NewDecoder(io.LimitReader(r.Body, 8*1024)).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse request"))
		return
	}

	var results []responses.Resize

	query := r.URL.Query()

	switch query.Get("async") {
	case "true":
		results, err = s.Resizer.ProcessResizesAsync(request)
	default:
		results, err = s.Resizer.ProcessResizes(request)
	}

	data, err := json.Marshal(results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to marshal response"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Add("content-type", "application/json")
	w.Write(data)

}

func (s *service) getImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("fetching ", r.URL.String())
		data, ok := s.cache.Get(r.URL.String())
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("content-type", "image/jpeg")
		w.Write(data.([]byte))
	}
}
