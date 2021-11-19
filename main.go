package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/muesli/cache2go"
	"willnorris.com/go/imageproxy"
)

// LinkpreviewAPI endpoint
var LinkpreviewAPI = "https://api.linkpreview.net"

var linkpreviewKey = ""

// Cache engine for the app
var Cache *cache2go.CacheTable

// CachedResponse struct
type CachedResponse struct {
	Body   []byte
	Status int
}

type lpreq struct {
	Key    string `json:"key"`
	Q      string `json:"q"`
	Fields string `json:"fields"`
}

func init() {
	Cache = cache2go.Cache("myCache")
}

func main() {

	fmt.Println("LinkPreview Proxy Server v2.0.0")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	addr := os.Getenv("ADDR")
	linkpreviewKey = os.Getenv("LINK_PREVIEW_KEY")

	http.HandleFunc("/linkpreview/", MWrefererCheck(os.Getenv("REFERER"), LinkpreviewProxyHandler))
	http.HandleFunc("/imageproxy/", MWrefererCheck(os.Getenv("REFERER"), ImageProxyHandler))

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

// LinkpreviewProxyHandler is a Proxy for the API
func LinkpreviewProxyHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query().Get("q")

	// keep results in cache for one day (based on the key scheme)
	cacheKey := fmt.Sprintf("%s%d", query, time.Now().Day())

	cached, err := Cache.Value(cacheKey)
	if err == nil {
		fmt.Println("Serving from Cache: " + query)
		w.WriteHeader(cached.Data().(*CachedResponse).Status)
		w.Write(cached.Data().(*CachedResponse).Body)
		return
	}

	body, _ := json.Marshal(&lpreq{
		Key:    linkpreviewKey,
		Q:      query,
		Fields: "title,description,image,url", // see https://docs.linkpreview.net/#query-parameters
	})

	fmt.Println("Requesting from the API: " + query)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", LinkpreviewAPI, bytes.NewBuffer(body))
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	cr := CachedResponse{}
	cr.Body, err = ioutil.ReadAll(resp.Body)
	cr.Status = resp.StatusCode
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(cr.Status)
	w.Write(cr.Body)

	// add to cache and automatically expire unused after one day
	Cache.Add(cacheKey, 24*time.Hour, &cr)
}

// ImageProxyHandler is a Proxy for serving images
// for advanced usage see https://github.com/willnorris/imageproxy
func ImageProxyHandler(w http.ResponseWriter, r *http.Request) {
	p := imageproxy.NewProxy(nil, nil)

	// convert query input to root path, imageproxy will be confused otherwise
	r.URL.Path = "/" + r.URL.Query().Get("src")
	p.ServeHTTP(w, r)
}

// MWrefererCheck is a Middleware that can protect agains unknown referrers
func MWrefererCheck(referer string, next http.HandlerFunc) http.HandlerFunc {

	// skip if not configured
	if referer == "" {
		return next
	}

	return func(w http.ResponseWriter, r *http.Request) {

		if !strings.HasPrefix(r.Referer(), referer) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
