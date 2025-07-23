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

var q list.List
var visited = make(map[string]bool)
var baseDomain string
var baseURL string
var deadLinks = make(map[string][]string, 0)
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

	baseDomain = getDomain(normalizedURL)
	baseURL = getBaseURL(normalizedURL)
	q.PushBack(normalizedURL)
}

func Start() map[string][]string {
	for q.Len() > 0 {
		parentUrl := q.Remove(q.Front()).(string)
		t := time.Now()

		fmt.Printf("\033[33m[%s]: %s\n", t.Format("2006-01-02 15:04:05"), parentUrl)

		html, err := scraper.Fetch(parentUrl)
		if err != nil {
			continue
		}

		links, err := parser.Extract(html)
		if err != nil {
			continue
		}

		linkCh := make(chan string, len(links))
		resultCh := make(chan struct {
			page       string
			link       string
			isDead     bool
			isInternal bool
		})

		workers := 10
		var wg sync.WaitGroup

		for i := range workers {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for link := range linkCh {
					fullURL, isInternal := checkDomain(link)

					mu.Lock()
					alreadyVisited := visited[fullURL]
					mu.Unlock()

					if !alreadyVisited {
						isDead := checker.IsDead(fullURL)
						resultCh <- struct {
							page       string
							link       string
							isDead     bool
							isInternal bool
						}{parentUrl, fullURL, isDead, isInternal}
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
			visited[result.link] = true

			if result.isDead {
				deadLinks[result.page] = append(deadLinks[result.page], result.link)
			} else if result.isInternal {
				q.PushBack(result.link)
			}
			mu.Unlock()
		}
		mu.Lock()
		visited[parentUrl] = true
		mu.Unlock()
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
		baseURL := strings.TrimSuffix(baseURL, "/")
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
	return normalizedURL, baseDomain == getDomain(normalizedURL)
}
