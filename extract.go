package main

import (
	//"fmt"
	//"log"
	"bytes"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/net/html"
)

type ExtractResult struct{
    words []string
    hrefs []string
    title string
    url string
    err error
}

func Extract(input []byte, currentURL string, extractWg *sync.WaitGroup, extractChannel chan ExtractResult) {
    defer extractWg.Done()
    result := ExtractResult{}
    var words, hrefs []string
    title := "untitled"
    doc, err := html.Parse(bytes.NewReader(input))
    if err != nil {
        result.err = err
        extractChannel <- result
        return
    }

    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.ElementNode && n.Data == "style" {
            return
        }
        if n.Type == html.ElementNode && n.Data == "title" && n.Parent.Data == "head" && n.FirstChild != nil {
            title = n.FirstChild.Data
        }
        if n.Type == html.TextNode {
            checkWords := func(c rune) bool {
                return !unicode.IsLetter(c) && !unicode.IsNumber(c)
            }
            words = append(words, strings.FieldsFunc(n.Data, checkWords)...)
        } else if n.Type == html.ElementNode && n.Data == "a" {
            for _, a := range n.Attr {
                if a.Key == "href" {
                    hrefs = append(hrefs, a.Val)
                    break
                }
            }
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(doc)

    result.words = words
    result.hrefs = hrefs
    result.title = title
    result.url = currentURL
    extractChannel <- result
}
