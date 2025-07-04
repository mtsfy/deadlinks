package crawler

import (
	"container/list"
	"deadlinks/internal/checker"
	"deadlinks/internal/parser"
	"deadlinks/internal/scraper"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var Q list.List
var Visited = make(map[string]bool)
var BaseDomain string
var BaseURL string
var DeadLinksSet = make(map[string]bool, 0)

func Start() {
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

		for _, link := range links {
			fullURL, isInternal := checkDomain(link)
			if isInternal && !Visited[fullURL] {
				isDeadLink := checker.IsDead(fullURL)
				if isDeadLink {
					DeadLinksSet[fullURL] = true
					continue
				}
				Q.PushBack(fullURL)
			}
		}
		Visited[url] = true
	}

	deadLinks := make([]string, 0)
	for dl := range DeadLinksSet {
		deadLinks = append(deadLinks, dl)
	}

	fmt.Println(deadLinks)
}

func Init(url string) {
	BaseDomain = getDomain(url)
	BaseURL = getBaseURL(url)
	Q.PushBack(url)
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

func checkDomain(url string) (string, bool) {
	var fullURL string
	if strings.HasPrefix(url, "http") {
		fullURL = url
	} else {
		baseURL := strings.TrimSuffix(BaseURL, "/")
		fullURL = baseURL + url
	}
	return fullURL, BaseDomain == getDomain(fullURL)
}
