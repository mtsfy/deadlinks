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
						linkSet[a.Val] = true
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			getLinks(c)
		}
	}
	getLinks(doc)

	links := make([]string, 0)
	for link := range linkSet {
		links = append(links, link)
	}

	return links, nil
}
