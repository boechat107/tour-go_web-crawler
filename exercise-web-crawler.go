package main

import (
	"fmt"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type Cache map[string]string

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, c chan Cache) {
	if depth <= 0 {
		t := <-c
		c <- t
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		t := <-c
		c <- t
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	// Only one goroutine have access to this data at some moment.
	cache := <-c
	//fmt.Printf("%s: %v\n", url, cache)
	var newurls []string
	for _, u := range urls {
		if _, status := cache[u]; !status {
			cache[u] = "done"
			newurls = append(newurls, u)
		}
	}
	if len(newurls) > 0 {
		children := make(chan Cache)
		for _, u := range newurls {
			go Crawl(u, depth-1, fetcher, children)
		}
		// Now all goroutines are waiting for some value on "children".
		children <- cache
		// The parent goroutine will be the last to have access to the channel.
		<-children
	}
	// Now it's safe to signal that the job is done.
	c <- cache
	return
}

func main() {
	url := "http://golang.org/"
	c := make(chan Cache)
	go Crawl(url, 4, fetcher, c)
	// To send data through the channel, first we need to have somebody waiting to
	// read.
	c <- Cache{url: "done"}
	// Now we wait the goroutine finish its job and send something back
	// (synchronization).
	<-c
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
