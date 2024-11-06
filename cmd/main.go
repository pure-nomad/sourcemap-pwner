package main

import (
	"bufio"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"os"
	"sourcemap-pwner/utils"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

func checkUrl(url string, ch chan utils.SourceMap, wg *sync.WaitGroup) {

	defer wg.Done()

	tr := &http.Transport{
		MaxIdleConns:        100,              // The maximum number of idle (keep-alive) connections
		MaxIdleConnsPerHost: 20,               // The maximum number of idle connections per host
		IdleConnTimeout:     90 * time.Second, // Idle connection timeout
		// Set connection timeout parameters
		DisableCompression:    true,            // Disable compression to avoid overhead for large amounts of requests
		TLSHandshakeTimeout:   5 * time.Second, // TLS handshake timeout
		ExpectContinueTimeout: 1 * time.Second, // Timeout for HTTP 100-Continue response
	}

	client := &http.Client{
		Transport: tr,
	}

	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var valid utils.SourceMap
	var once sync.Once

	// so to find js files we will just parse the html and look for javascript file references. ex: <script src="/main.js"/>

	r := strings.NewReader(string(body))

	doc, err := html.Parse(r)
	if err != nil {
		panic(err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, a := range n.Attr {
				if a.Key == "src" {

					// this works great! but what if the javascript files are loaded externally?
					// were going to add some methods to our sourcemap struct to do some checks on this :D
					// we should probably implement a manual mode that will show the user the urls so they can decide if they want to dump sourcemap from that js file.

					valid.Url = url
					valid.JsFiles = append(valid.JsFiles, a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	valid.CheckJsFileHostname()
	/* this call makes sure we only get one sourcemap struct back on the channel that has ALL the js files found :)
	AND NOW THEY ARE CLEANED READY FOR REQUESTS TO BE MADE! (CheckJsFileHostname) ... but I know I will have to parse the body of this as well to find sourceMappingURL=
	*/
	once.Do(func() { ch <- valid })
}

func main() {
	fmt.Println("Let's pwn some websites ~")

	if len(os.Args) <= 1 {
		fmt.Println("Usage: ./sourcemap-pwner urls.txt")
	}

	file, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	channel := make(chan utils.SourceMap)

	s := bufio.NewScanner(file)

	for s.Scan() {

		wg.Add(1)

		go checkUrl(s.Text(), channel, &wg)

	}

	go func() {
		defer close(channel)
		wg.Wait()
	}()

	for rec := range channel {
		fmt.Println(rec.JsFiles)
	}

}
