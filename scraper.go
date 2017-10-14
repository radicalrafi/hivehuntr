package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

const MAGIC = "https://coinhive.com/lib/coinhive.min.js"

func fetchBody(url string) {
	resp, _ := http.Get(url)
	bytes, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("HTML : \n", string(bytes))
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
func getSrc(t html.Token) (ok bool, src string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "src" {
			src = a.Val
			ok = true
		}
	}

	// "bare" return will return the variables (ok, href) as defined in
	// the function definition
	return
}
func crawl(url string, ch chan string, chFinished chan bool) {
	resp, err := http.Get(url)

	defer func() {
		// Notify that we're done after this function
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \""+url+"\""+"%s", err)
		return
	}

	b := resp.Body
	defer b.Close() // close Body when the function returns

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			// Check if the token is an <a> tag
			isScript := t.Data == "script"
			if !isScript {
				continue
			}

			// Extract the href value, if there is one
			ok, url := getSrc(t)
			if !ok {
				continue
			}

			// Make sure the url begines in http**
			hasProto := strings.Compare(url, MAGIC) == 0
			if hasProto {
				ch <- url
			}
		}
	}
}

func main() {
	foundUrls := make(map[string]bool)
	seedUrls, _ := readLines("topsites.txt")
	fmt.Println(seedUrls[0])
	// Channels
	chUrls := make(chan string)
	chFinished := make(chan bool)

	// Kick off the crawl process (concurrently)
	for _, url := range seedUrls {
		go crawl("http://"+url, chUrls, chFinished)
	}

	// Subscribe to both channels
	for c := 0; c < len(seedUrls); {
		select {
		case url := <-chUrls:
			foundUrls[url] = true
		case <-chFinished:
			c++
		}
	}

	// We're done! Print the results...

	fmt.Println("\nFound", len(foundUrls), "unique urls:\n")

	for url := range foundUrls {
		fmt.Println(" - " + url)
	}

	close(chUrls)
}
