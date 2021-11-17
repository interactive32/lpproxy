package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/muesli/cache2go"
)

const linkpreviewAPI = "https://api.linkpreview.net"

var linkpreviewKey = ""

var cache *cache2go.CacheTable

type cachedResponse struct {
	body   []byte
	status int
}

type lpreq struct {
	Key    string `json:"key"`
	Q      string `json:"q"`
	Fields string `json:"fields"`
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	addr := os.Getenv("ADDR")
	linkpreviewKey = os.Getenv("LINK_PREVIEW_KEY")

	cache = cache2go.Cache("myCache")

	http.HandleFunc("/", proxyHandler)

	if os.Getenv("SSL_CERT") != "" && os.Getenv("SSL_KEY") != "" {
		fmt.Printf("Proxy secure server started on https://%s\n", addr)
		if err := http.ListenAndServeTLS(addr, os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"), nil); err != nil {
			panic(err)
		}
	}

	fmt.Printf("Proxy server started on http://%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}

}

func proxyHandler(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query().Get("q")

	// keep results in cache for one day (based on the key scheme)
	cacheKey := fmt.Sprintf("%s%d", query, time.Now().Day())

	cached, err := cache.Value(cacheKey)
	if err == nil {
		fmt.Println("Serving from Cache: " + query)
		w.WriteHeader(cached.Data().(*cachedResponse).status)
		w.Write(cached.Data().(*cachedResponse).body)
		return
	}

	body, _ := json.Marshal(&lpreq{
		Key:    linkpreviewKey,
		Q:      query,
		Fields: "title,description,image,url", // see https://docs.linkpreview.net/#query-parameters
	})

	fmt.Println("Requesting from the API: " + query)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", linkpreviewAPI, bytes.NewBuffer(body))
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	cr := cachedResponse{}
	cr.body, err = io.ReadAll(resp.Body)
	cr.status = resp.StatusCode
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(cr.status)
	w.Write(cr.body)

	// add to cache and automatically expire unused after one day
	cache.Add(cacheKey, 24*time.Hour, &cr)
}
