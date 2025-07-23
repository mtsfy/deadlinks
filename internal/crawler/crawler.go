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

	return deadLinks
}

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
