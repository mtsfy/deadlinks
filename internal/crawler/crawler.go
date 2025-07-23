package crawler

import (
	"container/list"
	"deadlinks/internal/checker"
	"deadlinks/internal/parser"
	"deadlinks/internal/scraper"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

var Q list.List
var Visited = make(map[string]bool)
var BaseDomain string
var BaseURL string
var DeadLinksSet = make(map[string]bool, 0)
var mu sync.Mutex

func Init(link string) {
	parsedURL, err := url.Parse(link)
	if err != nil {
		panic(err)
	}

	if parsedURL.Path == "/" {
		parsedURL.Path = ""
	}

	normalizedURL := parsedURL.String()

	BaseDomain = getDomain(normalizedURL)
	BaseURL = getBaseURL(normalizedURL)
	Q.PushBack(normalizedURL)
}

func Start() []string {
	for Q.Len() > 0 {
		url := Q.Remove(Q.Front()).(string)
		t := time.Now()

		fmt.Printf("[%s]: %s\n", t.Format("2006-01-02 15:04:05"), url)

		html, err := scraper.Fetch(url)
		if err != nil {
			continue
		}

		links, err := parser.Extract(html)
		if err != nil {
			continue
		}

		linkCh := make(chan string, len(links))
		resultCh := make(chan struct {
			url        string
			isDead     bool
			isInternal bool
		})

		workers := 10
		var wg sync.WaitGroup

		for i := range workers {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for url := range linkCh {
					fullURL, isInternal := checkDomain(url)

					mu.Lock()
					alreadyVisited := Visited[fullURL]
					mu.Unlock()

					if !alreadyVisited {
						isDead := checker.IsDead(fullURL)
						resultCh <- struct {
							url        string
							isDead     bool
							isInternal bool
						}{fullURL, isDead, isInternal}
					}
				}
			}(i)
		}

		for _, link := range links {
			linkCh <- link
		}
		close(linkCh)

		go func() {
			wg.Wait()
			close(resultCh)
		}()

		for result := range resultCh {
			mu.Lock()
			Visited[result.url] = true

			if result.isDead {
				DeadLinksSet[result.url] = true
			} else if result.isInternal {
				Q.PushBack(result.url)
			}
			mu.Unlock()
		}
		mu.Lock()
		Visited[url] = true
		mu.Unlock()
	}

	deadLinks := make([]string, 0)
	for dl := range DeadLinksSet {
		deadLinks = append(deadLinks, dl)
	}

	return deadLinks
}

func getDomain(link string) string {
	parsedURL, err := url.Parse(link)
	if err != nil {
		panic(err)
	}
	hostname := parsedURL.Hostname()
	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return hostname
	}
	return parts[len(parts)-2]
}

func getBaseURL(link string) string {
	parsedURL, err := url.Parse(link)
	if err != nil {
		panic(err)
	}
	parsedURL.Path = ""
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""
	return parsedURL.String()
}

func checkDomain(link string) (string, bool) {
	var fullURL string

	if strings.HasPrefix(link, "http") {
		fullURL = link
	} else {
		baseURL := strings.TrimSuffix(BaseURL, "/")
		fullURL = baseURL + link
	}

	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return fullURL, false
	}

	if parsedURL.Path == "/" {
		parsedURL.Path = ""
	}

	normalizedURL := parsedURL.String()
	return normalizedURL, BaseDomain == getDomain(normalizedURL)
}
