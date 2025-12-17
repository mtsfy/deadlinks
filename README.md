<h1 align="center">deadlinks</h1>

> [!NOTE]
> Work in progress.

## :page_facing_up: Description

Fast CLI to crawl a domain and report dead links. It parses pages within the same domain, checks all discovered links in parallel, and prints a color-coded report.

## :hammer_and_wrench: Development

- Go 1.24+
- Docker

## :sparkles: Features

- Parallel page crawling and link checking
- Domain-scoped crawl (internal links are followed)
- Distinguishes internal vs external links
- URL normalization and per-page deduplication
- HTTP HEAD with timeout, limited concurrency via semaphore

##### :crystal_ball: Future Features

- **Headless fallback**
  - Using Playwright or Selenium for JS-rendered pages
- **Respect robots.txt**
  - Crawl-delay
  - Disallow

## :rocket: Quick Start

### Using Docker

```bash
# Build locally
docker build -t deadlinks .

# Run (replace URL)
docker run --rm deadlinks --url https://scrape-me.dreamsofcode.io
```

```bash
# Output
Found 12 dead links:

PAGE                                          LINK                                          STATUS
----                                          ---------                                     ------
https://scrape-me.dreamsofcode.io/about       http://10.255.255.1                           Maybe dead
https://scrape-me.dreamsofcode.io/believe     http://10.255.255.1                           Maybe dead
https://scrape-me.dreamsofcode.io/nirvana     https://scrape-me.dreamsofcode.io/nevermind   Dead
https://scrape-me.dreamsofcode.io/nirvana     https://scrape-me.dreamsofcode.io/in-utero    Dead
https://scrape-me.dreamsofcode.io/locations   https://twitter.com                           Maybe dead
https://scrape-me.dreamsofcode.io/nav         https://scrape-me.dreamsofcode.io/teapot      Dead
https://scrape-me.dreamsofcode.io/moon        https://scrape-me.dreamsofcode.io/busted      Dead
https://scrape-me.dreamsofcode.io/sun         https://scrape-me.dreamsofcode.io/earth       Dead
https://scrape-me.dreamsofcode.io/sun         https://scrape-me.dreamsofcode.io/mars        Dead
https://scrape-me.dreamsofcode.io/sun         https://scrape-me.dreamsofcode.io/venus       Dead
https://scrape-me.dreamsofcode.io/galaxy      https://scrape-me.dreamsofcode.io/recursion   Dead
https://scrape-me.dreamsofcode.io/galaxy      https://scrape-me.dreamsofcode.io/mountain    Dead
```
