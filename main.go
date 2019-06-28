/*
 | ---------------------------------------------------------
 | Author: Zoueature
 | Email: zoueature@gmail.com
 | Date: 2019/6/28
 | Time: 15:25
 | Description:
 | ---------------------------------------------------------
*/

package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
)

var (
	canMultiDownload = false
	maxGoruntineNum = 100
	maxLengthPerGoruntine int64 = 10000000
	waitGroup = sync.WaitGroup{}
)

func main() {
	argv := os.Args
	url := argv[1]
	var goruntine int
	var perContent int64
	if url == "" {
		fmt.Println("下载链接为空， 请检查后重新下载")
		os.Exit(-1)
	}
	client := http.Client{}
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}
	if response, err := client.Do(request); err == nil {
		contentLength := response.ContentLength
		acceptRange := response.Header.Get("Accept-Ranges")
		if acceptRange == "bytes" {
			canMultiDownload = true
		}
		if contentLength <= maxLengthPerGoruntine || canMultiDownload == false {
			goruntine = 1
			perContent = contentLength
		} else {
			perContent = maxLengthPerGoruntine
			hadMod := contentLength % maxLengthPerGoruntine
			goruntine = int(contentLength / maxLengthPerGoruntine)
			if hadMod != 0 {
				goruntine ++
			}
		}
	}
	if goruntine > maxGoruntineNum {

	} else {
		var downloadTotal int64
		for i := 0; i < goruntine; i ++ {
			waitGroup.Add(1)
			go sliceDownload()
		}
	}
	waitGroup.Wait()
}

func sliceDownload(rangeBytes int64) {
	waitGroup.Done()
}