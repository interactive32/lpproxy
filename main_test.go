package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	proxy "linkpreview.net/proxy/v2"
)

func TestLinkpreviewProxy(t *testing.T) {
	apihit := false

	// mock linkpreview api
	apiresp := []byte(`{"title":"Wikipedia","description":"Wikipedia is a free online encyclopedia, created and edited by volunteers around the world and hosted by the Wikimedia Foundation.","image":"https://www.wikipedia.org/static/apple-touch/wikipedia.png","url":"https://www.wikipedia.org/"}`)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apihit = true
		w.WriteHeader(http.StatusOK)
		w.Write(apiresp)
	}))
	defer s.Close()
	proxy.LinkpreviewAPI = s.URL

	// hit proxy
	req := httptest.NewRequest("GET", "/linkpreview/?q=https://wikipedia.org", nil)
	res := httptest.NewRecorder()
	proxy.LinkpreviewProxyHandler(res, req)

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	// check the actual response
	assert.Equal(t, true, apihit)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, apiresp, body)

	// check if we have this cached now
	cacheKey := fmt.Sprintf("%s%d", "https://wikipedia.org", time.Now().Day())
	cached, err := proxy.Cache.Value(cacheKey)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, cached.Data().(*proxy.CachedResponse).Status)
	assert.Equal(t, apiresp, cached.Data().(*proxy.CachedResponse).Body)
}

func TestLinkpreviewProxyFromCached(t *testing.T) {

	// put something in the cache
	cr := proxy.CachedResponse{}
	cr.Body = []byte(`{"title":"Example site","description":"Example site","image":"","url":"https://example.com/"}`)
	cr.Status = http.StatusOK
	cacheKey := fmt.Sprintf("%s%d", "https://example.com/test", time.Now().Day())
	proxy.Cache.Add(cacheKey, 24*time.Hour, &cr)

	// hit proxy
	req := httptest.NewRequest("GET", "/linkpreview/?q=https://example.com/test", nil)
	res := httptest.NewRecorder()
	proxy.LinkpreviewProxyHandler(res, req)

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Contains(t, string(body), "Example site")
}

func TestGoodReferrer(t *testing.T) {

	passed := false
	s := proxy.MWrefererCheck("https://example.com/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passed = true
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Referer", "https://example.com/")
	res := httptest.NewRecorder()

	// hit
	s(res, req)
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	// check the actual response
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, true, passed)
	assert.Equal(t, "", string(body))
}

func TestNoReferrer(t *testing.T) {

	passed := false
	s := proxy.MWrefererCheck("https://example.com/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passed = true
	}))

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	// hit
	s(res, req)
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	// check the actual response
	assert.Equal(t, http.StatusForbidden, res.Code)
	assert.Equal(t, false, passed)
	assert.Equal(t, "Forbidden\n", string(body))
}

func TestBadReferrer(t *testing.T) {

	passed := false
	s := proxy.MWrefererCheck("https://example.com/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passed = true
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Referer", "https://strange.com/")
	res := httptest.NewRecorder()

	// hit
	s(res, req)
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	// check the actual response
	assert.Equal(t, http.StatusForbidden, res.Code)
	assert.Equal(t, false, passed)
	assert.Equal(t, "Forbidden\n", string(body))
}
