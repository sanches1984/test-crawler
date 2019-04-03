package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

const waitSec = 1

var tags = make(map[string]string)
var exts = make(map[string]string)

// App app
type App struct {
	Host  string
	URLs  []string
	mutex sync.Mutex
}

func init() {
	// список можно расширить, но этого для начала будет достаточно
	tags["a"] = "href"
	tags["area"] = "href"
	tags["link"] = "href"
	tags["img"] = "src"
	tags["iframe"] = "src"
	tags["frame"] = "src"
	tags["embed"] = "src"
	tags["script"] = "src"
	tags["input"] = "src"
	tags["form"] = "action"
	tags["object"] = "data"
	tags["body"] = "background"

	exts["jpg"] = ""
	exts["jpeg"] = ""
	exts["png"] = ""
	exts["gif"] = ""
	exts["bmp"] = ""
}

func main() {
	app := App{}
	var err error
	if len(os.Args) < 2 {
		fmt.Println("No url")
		return
	}

	// получаем адрес домена
	sch, domain, err := getDomain(os.Args[1])
	if err != nil || domain == "" || (sch != "http" && sch != "https") {
		fmt.Println("Bad url")
		return
	}

	// получаем адрес хоста, сравниваем с ним
	app.Host = getHost(domain)
	fmt.Println("Current host:", app.Host)

	ch := make(chan string)
	exit := make(chan struct{})

	go app.Run(ch, exit, sch+"://"+domain)
	app.Printer(ch, exit)
}

// Run запуск
func (app *App) Run(ch chan<- string, exit chan<- struct{}, xurl string) {
	app.ReadPage(ch, xurl)
	exit <- struct{}{}
}

// ReadPage чтение станицы
func (app *App) ReadPage(ch chan<- string, xurl string) {
	if ok := app.addToList(xurl); !ok {
		return
	}

	// фильтруем на картинки
	if _, ok := exts[filepath.Ext(xurl)]; ok {
		return
	}

	ch <- xurl

	// читаем контент html
	resp, err := http.Get(xurl)
	if err != nil {
		log.Println("Can't get page:", xurl)
		return
	}
	defer resp.Body.Close()

	// парсим
	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Println("Can't read page:", xurl)
		return
	}

	var f func(*html.Node)
	// разбор html
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// тэг содержит url
			if val, ok := tags[n.Data]; ok {
				for i := 0; i < len(n.Attr); i++ {
					// нашли атрибут
					if n.Attr[i].Key == val && needWatch(app.Host, n.Attr[i].Val) {
						// рекурсия
						app.ReadPage(ch, n.Attr[i].Val)
					}
				}
			}
		}
		// спуск гвлубь html
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

// Printer вывод на печать
func (app *App) Printer(ch <-chan string, exit <-chan struct{}) {
	for {
		select {
		case xurl := <-ch:
			fmt.Println("URL:", xurl)
			time.Sleep(waitSec * time.Second)
		case <-exit:
			return
		}
	}
}

// проверка и добавление в список
func (app *App) addToList(xurl string) bool {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	for _, v := range app.URLs {
		if strings.ToLower(v) == strings.ToLower(xurl) {
			return false
		}
	}

	app.URLs = append(app.URLs, xurl)
	return true
}

func getDomain(xurl string) (scheme string, domain string, err error) {
	u, err := url.Parse(xurl)
	if err != nil {
		return
	}

	scheme = u.Scheme
	domain = u.Host
	return
}

func getHost(domain string) (host string) {
	sub := strings.Split(domain, ".")
	if len(sub) > 2 {
		host = sub[len(sub)-2] + "." + sub[len(sub)-1]
	} else {
		host = domain
	}

	return
}

// проверка на просмотр
func needWatch(host string, xurl string) bool {
	sch, dom, err := getDomain(xurl)
	if err != nil {
		return false
	}

	if sch != "http" && sch != "https" {
		return false
	}

	if strings.ToLower(getHost(dom)) != strings.ToLower(host) {
		return false
	}

	return true
}
