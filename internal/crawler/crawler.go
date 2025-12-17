package crawler

import (
	"container/list"
	"deadlinks/internal/checker"
	"deadlinks/internal/parser"
	"deadlinks/internal/scraper"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type DeadLink struct {
	Link     string
	Internal bool
}

var q list.List
var visited = make(map[string]bool)

var baseDomain string
var baseURL string
var deadLinks = make(map[string][]DeadLink, 0)
var deadLinksSet = make(map[string]map[string]struct{}, 0)

var mu sync.Mutex

var pagesProcessed int64
var linksChecked int64
var activeWorkers int64
var pendingPages int64
var httpSemaphore = make(chan struct{}, 20)

func Init(link string) {
	q = list.List{}
	visited = make(map[string]bool)
	deadLinks = make(map[string][]DeadLink, 0)
	deadLinksSet = make(map[string]map[string]struct{}, 0)
	atomic.StoreInt64(&pagesProcessed, 0)
	atomic.StoreInt64(&linksChecked, 0)
	atomic.StoreInt64(&activeWorkers, 0)
	atomic.StoreInt64(&pendingPages, 0)

	parsedURL, err := url.Parse(link)
	if err != nil {
		panic(err)
	}

	if parsedURL.Path == "/" {
		parsedURL.Path = ""
	}

	normalizedURL := parsedURL.String()

	baseDomain = parsedURL.Hostname()
	baseURL = getBaseURL(normalizedURL)
	q.PushBack(normalizedURL)
	atomic.StoreInt64(&pendingPages, 1)
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

func Start() map[string][]DeadLink {
	pageCh := make(chan string, 100)
	errCh := make(chan error, 100)
	var wg sync.WaitGroup

	workers := 10
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pageURL := range pageCh {
				// increment counter of active workers currently processing
				atomic.AddInt64(&activeWorkers, 1)
				// process the page and extract links
				if err := processPage(pageURL); err != nil {
					errCh <- fmt.Errorf("failed to process %s: %w", pageURL, err)
				}
				// decrement active worker count
				atomic.AddInt64(&activeWorkers, -1)
				// decrement pending pages counter (this page is done)
				atomic.AddInt64(&pendingPages, -1)
			}
		}()
	}

	// prints errors as they occur
	go func() {
		for err := range errCh {
			fmt.Printf("⚠️  Error: %v\n", err)
		}
	}()

	// feeds pages from queue to workers (queue processor)
	go func() {
		defer close(pageCh) // close channel when done to signal workers
		for {
			mu.Lock()
			qLen := q.Len()
			mu.Unlock()

			if qLen == 0 {
				// empty queue so wait a bit to see if workers add more pages
				time.Sleep(50 * time.Millisecond)

				// check if there are any pending pages still being processed
				pending := atomic.LoadInt64(&pendingPages)
				if pending == 0 {
					// no pending pages and queue is empty, it's done!
					break
				}
				// workers are still processing, wait and check again
				continue
			}

			mu.Lock()
			if q.Len() > 0 {
				parentUrl := q.Remove(q.Front()).(string)
				if !visited[parentUrl] {
					visited[parentUrl] = true
				}
				mu.Unlock()
				// send the page to a worker to process
				pageCh <- parentUrl
			} else {
				mu.Unlock()
			}
		}
	}()

	// wait for all workers to finish processing
	wg.Wait()
	// close error channel
	close(errCh)

	return deadLinks
}

func processPage(parentUrl string) error {
	// increment counter of pages processed so far
	processed := atomic.AddInt64(&pagesProcessed, 1)

	html, err := scraper.Fetch(parentUrl)
	if err != nil {
		return err
	}

	links, err := parser.Extract(html)
	if err != nil {
		return err
	}

	type linkJob struct {
		fullURL    string
		isInternal bool
	}

	// deduplicate links to avoid checking the same link twice per page
	uniqueJobs := make(map[string]linkJob)
	for _, link := range links {
		fullURL, isInternal := checkDomain(link)
		uniqueJobs[fullURL] = linkJob{fullURL, isInternal}
	}

	// send link jobs to workers
	linkCh := make(chan linkJob, len(uniqueJobs))
	resultCh := make(chan struct {
		page       string
		link       string
		isDead     bool
		isInternal bool
	}, len(uniqueJobs))

	// workers for checking links
	workers := 10
	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range linkCh {
				fullURL := job.fullURL
				isInternal := job.isInternal

				// check if already visited this link to avoid duplicate checks
				mu.Lock()
				alreadyVisited := visited[fullURL]
				mu.Unlock()

				if !alreadyVisited {
					// acquire a slot from the HTTP semaphore
					httpSemaphore <- struct{}{}
					// check if the link returns an error status code
					isDead := checker.IsDead(fullURL)
					// release the HTTP semaphore slot
					<-httpSemaphore
					// send result back through the result channel
					resultCh <- struct {
						page       string
						link       string
						isDead     bool
						isInternal bool
					}{parentUrl, fullURL, isDead, isInternal}
				}
			}
		}()
	}

	// send all unique link jobs to workers
	for _, job := range uniqueJobs {
		linkCh <- job
	}

	close(linkCh)
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// process results as they come in from workers
	for result := range resultCh {
		// increment total links checked counter
		checked := atomic.AddInt64(&linksChecked, 1)
		fmt.Printf("🔍 Checked: %s -> Dead: %v, Internal: %v\n", result.link, result.isDead, result.isInternal)

		mu.Lock()
		if result.isDead {
			if _, ok := deadLinksSet[result.page]; !ok {
				deadLinksSet[result.page] = make(map[string]struct{})
			}
			if _, dup := deadLinksSet[result.page][result.link]; !dup {
				deadLinksSet[result.page][result.link] = struct{}{}
				deadLinks[result.page] = append(deadLinks[result.page], DeadLink{Link: result.link, Internal: result.isInternal})
				fmt.Printf("💀 Added dead link: %s\n", result.link)
			}
		} else if result.isInternal {
			q.PushBack(result.link)
			atomic.AddInt64(&pendingPages, 1)
		}
		// mark link as visited regardless of status
		visited[result.link] = true
		mu.Unlock()

		if checked%10 == 0 {
			fmt.Printf("📊 Progress: %d pages, %d links checked\n", processed, checked)
		}
	}

	// periodically run garbage collection
	if atomic.LoadInt64(&pagesProcessed)%100 == 0 {
		runtime.GC()
	}
	return nil
}

func getDomain(link string) string {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return ""
	}
	return parsedURL.Hostname()
}

func isExactDomain(link, targetDomain string) bool {
	linkDomain := getDomain(link)
	return linkDomain == targetDomain
}

func checkDomain(link string) (string, bool) {
	var fullURL string

	if strings.HasPrefix(link, "http") {
		fullURL = link
	} else {
		cleanBaseURL := strings.TrimSuffix(baseURL, "/")
		if !strings.HasPrefix(link, "/") {
			fullURL = cleanBaseURL + "/" + link
		} else {
			fullURL = cleanBaseURL + link
		}
	}

	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return fullURL, false
	}

	if parsedURL.Path == "/" {
		parsedURL.Path = ""
	}

	normalizedURL := parsedURL.String()
	return normalizedURL, isExactDomain(normalizedURL, baseDomain)
}
