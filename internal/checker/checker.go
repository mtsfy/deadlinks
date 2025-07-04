package checker

import (
	"net/http"
)

func IsDead(link string) bool {
	resp, err := http.Get(link)
	if err != nil {
		return true
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 400
}
