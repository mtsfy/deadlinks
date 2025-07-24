package parser

import (
	"strings"

	"golang.org/x/net/html"
)

func Extract(data string) ([]string, error) {
	reader := strings.NewReader(data)

	doc, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}

	linkSet := make(map[string]bool, 0)

	var getLinks func(*html.Node)
	getLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if a.Key == "href" && a.Val != "" && !strings.HasPrefix(a.Val, "#") {
						link := strings.TrimSpace(a.Val)
						if skipLink(link) {
							continue
						}
						linkSet[link] = true
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			getLinks(c)
		}
	}
	getLinks(doc)

	links := make([]string, 0, len(linkSet))
	for link := range linkSet {
		links = append(links, link)
	}

	return links, nil
}

func skipLink(link string) bool {
	if link == "" {
		return true
	}

	if strings.HasPrefix(link, "#") ||
		strings.HasPrefix(link, "mailto:") ||
		strings.HasPrefix(link, "tel:") ||
		strings.HasPrefix(link, "javascript:") {
		return true
	}

	exts := []string{".pdf", ".doc", ".docx", ".xlsx", ".ppt", ".zip"}
	for _, ext := range exts {
		if strings.HasSuffix(strings.ToLower(link), ext) {
			return true
		}
	}

	return false
}
