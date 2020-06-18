package main

import (
	"github.com/sanches1984/web-crawler/crawler"
	"log"
	"os"
	"time"
)

// количество одновременно обрабатываемых страниц
const queueLimit = 3

// максимальное количество страниц для вывода
const maxCount = 200

// период задержки
const delay = 500 * time.Millisecond

func main() {
	if len(os.Args) < 2 {
		log.Println("Params not set")
		return
	}

	// получаем адрес домена
	host, err := crawler.GetHost(os.Args[1])
	if err != nil {
		log.Println("Error: ", err)
	}

	// запускаем индексатор
	crawl := crawler.NewCrawler(host, delay, queueLimit, maxCount)
	log.Println("Started crawling for host: ", host)
	crawl.Run()
}
