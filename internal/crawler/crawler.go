package crawler

import (
	"container/list"
	"net/url"
	"strings"
)

var Q list.List
var Visited = make(map[string]bool)
var BaseDomain string

func Start() {
	for Q.Len() > 0 {
		url := Q.Remove(Q.Front()).(string)
		// scraper
		// parser
		// isInternal -> !visited -> checker() -> add to queue
		Visited[url] = true
	}
}

func Init(url string) {
	BaseDomain = getDomain(url)
	Q.PushBack(url)
}

func getDomain(input string) string {
	url, err := url.Parse(input)
	if err != nil {
		panic(err)
	}
	hostname := url.Hostname()
	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return hostname
	}
	return parts[len(parts)-2]
}
