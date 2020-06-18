package test

import (
	"github.com/sanches1984/web-crawler/crawler"
	"testing"
	"time"
)

func Test(t *testing.T) {
	cr := crawler.NewCrawler("https://snami.ru", 500*time.Millisecond, 3, 300)
	cr.Run()
}
