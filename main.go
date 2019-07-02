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
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const tmpFileName = "tmp."

var (
	canMultiDownload = false
	maxGoruntineNum = 100
	maxLengthPerGoruntine int64 = 10000000
	waitGroup = sync.WaitGroup{}
	tmpFile = make(map[string][]string)
	urlMd5String string
)

func main() {
	argv := os.Args
	if len(argv) < 2 {
		log.Fatalln("no url found")
	}
	url := argv[1]
	urlMd5 := md5.Sum([]byte(url))
	urlMd5String = string(urlMd5[:])
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
		if canMultiDownload == false {
			goruntine = 1
			perContent = contentLength
		} else  {
			if contentLength <= maxLengthPerGoruntine {
				goruntine = maxGoruntineNum
				perContent = contentLength / int64(goruntine) + 1
			} else {
				perContent = maxLengthPerGoruntine
				hadMod := contentLength % maxLengthPerGoruntine
				goruntine = int(contentLength / maxLengthPerGoruntine)
				if hadMod != 0 {
					goruntine ++
				}
			}
		}
	}
	if goruntine > maxGoruntineNum {
		leftGoruntine := goruntine
		downloadGoruntine := maxGoruntineNum
		j := 0;
		download:
			for i := 0 ; i < downloadGoruntine ; i ++ {
				waitGroup.Add(1)
				startBytes := int64(j) * perContent
				endBytes := int64(j + 1) * perContent
				tmpFile[urlMd5String] = append(tmpFile[urlMd5String], tmpFileName + string(startBytes) + "-" + string(endBytes))
				go sliceDownload(startBytes, endBytes, url, j)
				j ++
			}
			waitGroup.Wait()
			leftGoruntine -= downloadGoruntine
			if leftGoruntine > 0 {
				if leftGoruntine >= maxGoruntineNum {
					downloadGoruntine = maxGoruntineNum
				} else {
					downloadGoruntine = leftGoruntine
				}
				goto download
			}
	} else {
		for i := 0; i < goruntine; i++ {
			waitGroup.Add(1)
			startBytes := int64(i) * perContent
			endBytes := int64(i + 1) * perContent
			tmpFile[urlMd5String] = append(tmpFile[urlMd5String], tmpFileName + string(startBytes) + "-" + string(endBytes))
			go sliceDownload(startBytes, endBytes, url, i)
		}
		waitGroup.Wait()
	}
}

/**
 * 分块下载函数
 */
func sliceDownload(startBytes, endBytes int64, url string, index int) {
	defer waitGroup.Done()
	client := http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln("init request error : " + err.Error())
		return
	}

	contentRange := strconv.FormatInt(startBytes, 10) + "-" + strconv.FormatInt(endBytes, 10)
	request.Header.Add("Range", "bytes=" + contentRange)
	response, err := client.Do(request)
	if err != nil {
		log.Fatalln("Download error : " + err.Error())
		return
	}
	httpStatus := response.StatusCode
	if httpStatus != http.StatusOK {
		log.Fatalln("download error : " + err.Error())
		return
	}
	tmpFileName := tmpFile[urlMd5String][index]
	tmpFile, err := os.Create(tmpFileName)
	if err != nil {
		log.Fatalln("open file error : " + err.Error())
		return
	}
	size, err := io.Copy(tmpFile, response.Body)
	if err != nil {
		log.Fatalln("copy file to tmp file error :" + err.Error())
		return
	}
	if size <= 0{
		log.Fatalln("download error")
		return
	}
	return
}