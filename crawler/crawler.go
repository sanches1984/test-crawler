package crawler

import (
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"sync"
	"time"
)

type Crawler struct {
	host       string
	queueLimit int
	maxCount   int
	delay      time.Duration
	urlMap     map[string]struct{}
	queue      chan string
	process    chan string
	finish     chan string
	exit       chan struct{}
	mutex      sync.Mutex
	index      int
}

func NewCrawler(host string, delay time.Duration, queueLimit, maxCount int) *Crawler {
	queue := make(chan string)
	process := make(chan string, queueLimit)
	finish := make(chan string, queueLimit)
	exit := make(chan struct{})
	return &Crawler{
		host:     host,
		urlMap:   map[string]struct{}{},
		delay:    delay,
		queue:    queue,
		process:  process,
		finish:   finish,
		exit:     exit,
		maxCount: maxCount,
	}
}

// запуск
func (c *Crawler) Run() {
	go c.processQueue()
	go c.processParser()
	go c.processFinish()
	c.queue <- c.host
	<-c.exit
}

// обработка очереди
func (c *Crawler) processQueue() {
	for {
		select {
		case address := <-c.queue:
			go c.scan(address)
		}
	}
}

// обработка парсера
func (c *Crawler) processParser() {
	for {
		select {
		case address := <-c.finish:
			<-c.process
			c.index++
			fmt.Printf("URL #%d: %s\n", c.index, address)
			if c.index >= c.maxCount {
				c.exit <- struct{}{}
			}
		}
	}
}

// обработка завершения
func (c *Crawler) processFinish() {
	oldIndex := c.index
	for {
		time.Sleep(c.delay * 2)
		if oldIndex == c.index {
			c.exit <- struct{}{}
		} else {
			oldIndex = c.index
		}
	}
}

// чтение станицы
func (c *Crawler) scan(address string) {
	defer func() { c.finish <- address }()
	c.process <- address
	time.Sleep(c.delay)

	// читаем контент html
	resp, err := http.Get(address)
	if err != nil {
		log.Println("Can't get page:", address, err)
		return
	}
	defer resp.Body.Close()

	// игнорируем редиректы и иной контент
	if !IsResponseValid(resp) {
		return
	}

	// парсим контент
	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Println("Can't read page:", address, err)
		return
	}
	c.parse(doc)
}

func (c *Crawler) parse(n *html.Node) {
	// спуск вглубь html
	for childNode := n.FirstChild; childNode != nil; childNode = childNode.NextSibling {
		c.parse(childNode)
	}

	if n.Type != html.ElementNode {
		return
	}

	// тэг содержит url
	if attr, ok := htmlTags[n.Data]; ok {
		for i := 0; i < len(n.Attr); i++ {
			// нашли атрибут
			address := n.Attr[i].Val
			if n.Attr[i].Key == attr {
				host, err := GetHost(address)
				if err != nil {
					//log.Println("Can't find host: ", err)
					continue
				}
				if host != c.host {
					continue
				}
				// проверяем, добавлен ли в список
				if ok := c.addToMap(address); !ok {
					continue
				}

				// добавляем в очередь
				//log.Println("Queued", address)
				c.queue <- address
			}
		}
	}
}

// проверка и добавление в список
func (c *Crawler) addToMap(address string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.urlMap[address]; !ok {
		c.urlMap[address] = struct{}{}
		return true
	}
	return false
}
