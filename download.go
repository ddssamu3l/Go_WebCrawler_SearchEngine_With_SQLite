package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
)

type DownloadContent struct{
	url string
	body []byte
}

func Download(url string, ch chan<- DownloadContent, wg *sync.WaitGroup){
	defer wg.Done()
	rsp, err := http.Get(url)
	if err != nil{
		fmt.Println(err)
		ch <- DownloadContent{url, []byte(``)}
		return
	}
	defer rsp.Body.Close()
	bts, err := io.ReadAll(rsp.Body)
	if err != nil{
		fmt.Println(err)
		ch <- DownloadContent{url, []byte(``)}
		return
	}

	ch <- DownloadContent{url, bts}
	return
}
