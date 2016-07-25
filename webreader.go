package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf"
	"github.com/k4s/phantomgo"
)

// WebReader defines an interface which just needs to be able to return
// the body of a page
type WebReader interface {
	GetBody(url string) (string, error)
}

// PhantomReader uses gophantom to create a headless browser
type PhantomReader struct {
	phantom phantomgo.Phantomer
}

// NewPhantomReader returns a pointer to a PhantomReader, no params required
func NewPhantomReader() *PhantomReader {
	phantom := phantomgo.NewPhantom()
	// @TODO(mark): Needs to be stored in config or flag if this were to be used
	phantom.SetPhantomjsPath("phantomjs", "/usr/local/bin/phantomjs")
	return &PhantomReader{
		phantom: phantom,
	}
}

// GetBody will construct a download to get the body
func (p *PhantomReader) GetBody(url string) (string, error) {
	resp, err := p.phantom.Download(&phantomgo.Param{
		Method:       "GET",
		Url:          url,
		Header:       http.Header{},
		UsePhantomJS: true,
		PostBody:     "",
	})
	if err != nil {
		return "", fmt.Errorf("could not open url: %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response body: %s", err)
	}
	return string(body), nil
}

// SurfReader uses surf, a browser which supports some DOM operations if
// we wanted to scrape using CSS selectors
type SurfReader struct{}

// NewSurfReader returns a pointer to a SurfReader, no params required
func NewSurfReader() *SurfReader {
	return &SurfReader{}
}

// GetBody will construct a browser and grab the contents from the URL
func (SurfReader) GetBody(url string) (string, error) {
	// Create the browser on each request for a body, it would require
	// locking otherwise
	bow := surf.NewBrowser()
	err := bow.Open(url)
	if err != nil {
		return "", fmt.Errorf("could not open url: %s", err)
	}
	body := ""
	bow.Find("body").Each(func(_ int, s *goquery.Selection) {
		body = s.Text()
	})
	return body, nil
}

// HttpReader will just use the built in http.Get
type HttpReader struct{}

// Returns a http reader
func NewHttpReader() *HttpReader {
	return &HttpReader{}
}

// GetBody will just execute http.Get and return the body or an error
func (HttpReader) GetBody(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("could not get from url: %s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response body: %s", err)
	}
	return string(body), nil
}

// GoogleCacheReader will just use the built in http.Get, but will grab
// it from the Google cache so that all of the JS has been rendered
// (i.e. SEO friendly version)
type GoogleCacheReader struct {
	HttpReader
}

// Returns a GoogleCache reader
func NewGoogleCacheReader() *GoogleCacheReader {
	return &GoogleCacheReader{}
}

// GetBody will just execute http.Get and return the body or an error
func (g GoogleCacheReader) GetBody(url string) (string, error) {

	if len(url) == 0 {
		return "", errors.New("url length cannot be 0")
	}

	// e.g. http://webcache.googleusercontent.com/search?q=cache:vbGcdXhWHFsJ:www.sainsburys.co.uk/shop/gb/groceries/fruit-veg/ripe---ready+&amp;cd=1&amp;hl=en&amp;ct=clnk&amp;gl=uk
	googleCacheUrl := "http://webcache.googleusercontent.com/search?q=cache:vbGcdXhWHFsJ:"
	newUrl := fmt.Sprintf("%s%s", googleCacheUrl, url)
	return g.HttpReader.GetBody(newUrl)
}
