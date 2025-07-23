package checker

import (
	"net/http"
	"time"
)

func IsDead(link string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Head(link)
	if err != nil {
		return true
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 400
}
