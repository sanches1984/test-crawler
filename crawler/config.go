package crawler

var htmlTags = map[string]string{
	"a":      "href",
	"area":   "href",
	"link":   "href",
	"img":    "src",
	"iframe": "src",
	"frame":  "src",
	"embed":  "src",
	"script": "src",
	"input":  "src",
	"form":   "action",
	"object": "data",
	"body":   "background",
}
