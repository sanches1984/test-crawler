package crawler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func GetHost(address string) (domain string, err error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	if u.Host == "" {
		return "", errors.New("Domain is empty")
	}
	if !strings.Contains(u.Scheme, "http") {
		return "", errors.New("Bad scheme")
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}

func IsResponseValid(resp *http.Response) bool {
	return resp.StatusCode == 200 && strings.Contains(resp.Header.Get("Content-Type"), "text/html")
}
