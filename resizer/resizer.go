package resizer

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/53jk1/Server-Coding/requests"
	"github.com/53jk1/Server-Coding/responses"
	lru "github.com/hashicorp/golang-lru"
	jpgresize "github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const (
	PROTO    = "http://"
	SUCCESS  = "success"
	FAILURE  = "failure"
	HOSTPORT = "localhost:8080"
)

type Resizer interface {
	ProcessResizes(request requests.Resize) ([]responses.Resize, error)
	ProcessResizesAsync(request requests.Resize) ([]responses.Resize, error)
}

type Resize struct {
	Cache *lru.Cache
}

// ProcessResizesAsync is a stub for the async version of ProcessResizes
func (s *Resize) ProcessResizesAsync(request requests.Resize) ([]responses.Resize, error) {
	results := make([]responses.Resize, 0, len(request.URLs))
	for _, url := range request.URLs {

		// generate a unique IDs for the images to be resized in go routines
		id := genID(strconv.Itoa(int(request.Width)), strconv.Itoa(int(request.Width)), url)

		key := "/v1/image/" + id + ".jpeg"

		// check if KEY already exist in the cache
		if s.Cache.Contains(key) {
			result := responses.Resize{}
			newURL := PROTO + HOSTPORT + key
			result.URL = newURL
			result.Result = SUCCESS
			result.Cached = true
			results = append(results, result)
			continue
		}

		newURL := PROTO + HOSTPORT + key

		// if the image is not in the cache, resize the image in a go routine
		go func(url string, id string, key string, newURL string) {
			data, err := fetchAndResize(url, request.Width, request.Height)
			if err != nil {
				log.Printf("failed to resize %s: %v", url, err)
				return
			}
			log.Print("caching ", key)
			s.Cache.Add(key, data)
		}(url, id, key, newURL)

		// return the new URL for the resized image
		result := responses.Resize{}
		result.URL = newURL
		result.Result = SUCCESS
		result.Cached = false
		results = append(results, result)
	}
	return results, nil
}

func (s *Resize) ProcessResizes(request requests.Resize) ([]responses.Resize, error) {
	results := make([]responses.Resize, 0, len(request.URLs))
	for _, url := range request.URLs {
		result := responses.Resize{}
		// request.Width and request.Height are uint, so we need to convert them to string
		id := genID(strconv.Itoa(int(request.Width)), strconv.Itoa(int(request.Width)), url)

		key := "/v1/image/" + id + ".jpeg"

		// check if KEY already exist in the cache
		if s.Cache.Contains(key) {
			result := responses.Resize{}
			newURL := PROTO + HOSTPORT + key
			result.URL = newURL
			result.Result = SUCCESS
			result.Cached = true
			results = append(results, result)
			continue
		}

		newURL := PROTO + HOSTPORT + key

		data, err := fetchAndResize(url, request.Width, request.Height)
		if err != nil {
			log.Printf("failed to resize %s: %v", url, err)
			result.Result = FAILURE
			results = append(results, result)
			continue
		}

		log.Print("caching ", key)
		s.Cache.Add(key, data)

		result.URL = newURL
		result.Result = SUCCESS
		result.Cached = false
		results = append(results, result)
	}

	return results, nil
}

func fetchAndResize(url string, width uint, height uint) ([]byte, error) {
	data, err := fetch(url)
	if err != nil {
		return nil, err
	}

	return resize(data, width, height)
}

func fetch(url string) ([]byte, error) {
	log.Print("fetching ", url)
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %v", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status: %d", r.StatusCode)
	}

	data, err := ioutil.ReadAll(io.LimitReader(r.Body, 15*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read fetch data: %v", err)
	}

	return data, nil
}

func resize(data []byte, width uint, height uint) ([]byte, error) {
	// decode jpeg into image.Image
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to jped decode: %v", err)
	}

	var newImage image.Image

	// if either width or height is 0, it will resize respecting the aspect ratio
	newImage = jpgresize.Resize(width, height, img, jpgresize.Lanczos3)

	newData := bytes.Buffer{}
	err = jpeg.Encode(bufio.NewWriter(&newData), newImage, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to jpeg encode resized image: %v", err)
	}

	return newData.Bytes(), nil
}

func genID(width, height, url string) string {
	hash := sha256.Sum256([]byte(width + height + url))
	return base64.StdEncoding.EncodeToString(hash[:])
}
