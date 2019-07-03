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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const tmpFileName = "./tmp/tmp"

var (
	canMultiDownload            = false
	maxGoruntineNum             = 1000
	maxLengthPerGoruntine int64 = 1000000
	waitGroup                   = sync.WaitGroup{}
	tmpFile                     = make(map[string][]string)
	urlMd5String          string
)

func main() {
	startTime := time.Now().Unix()
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
		response.Body.Close()
		if acceptRange == "bytes" {
			canMultiDownload = true
		}
		if canMultiDownload == false {
			goruntine = 1
			perContent = contentLength
		} else {
			if contentLength <= maxLengthPerGoruntine {
				goruntine = maxGoruntineNum
				perContent = contentLength/int64(goruntine) + 1
			} else {
				perContent = maxLengthPerGoruntine
				hadMod := contentLength % maxLengthPerGoruntine
				goruntine = int(contentLength / maxLengthPerGoruntine)
				if hadMod != 0 {
					goruntine++
				}
			}
		}
	}
	if goruntine > maxGoruntineNum {
		leftGoruntine := goruntine
		downloadGoruntine := maxGoruntineNum
		j := 0
	download:
		for i := 0; i < downloadGoruntine; i++ {
			waitGroup.Add(1)
			startBytes := int64(j) * perContent
			endBytes := int64(j+1) * perContent
			tmpFile[urlMd5String] = append(tmpFile[urlMd5String], tmpFileName+strconv.FormatInt(startBytes, 10)+"-"+strconv.FormatInt(endBytes, 10))
			go sliceDownload(startBytes, endBytes, url, j)
			j++
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
			endBytes := int64(i+1) * perContent
			tmpFile[urlMd5String] = append(tmpFile[urlMd5String], tmpFileName+strconv.FormatInt(startBytes, 10)+"-"+strconv.FormatInt(endBytes, 10))
			go sliceDownload(startBytes, endBytes, url, i)
		}
		waitGroup.Wait()
	}
	arr := strings.Split(url, "/")
	requestFile := arr[len(arr)-1]
	finalFile, err := os.Create("./" + requestFile)
	if err != nil {
		log.Fatalln("Create File " + requestFile + "Error :" + err.Error())
	}
	var i int64 = 0
	for _, fileName := range tmpFile[urlMd5String] {
		file, _ := os.OpenFile(fileName, os.O_RDONLY, 0600)
		//content, err := ioutil.ReadAll(file)
		//if err != nil {
		//	log.Fatalln("Read from " + fileName + "Error: " + err.Error())
		//}
		//n, _ := finalFile.Seek(0, os.SEEK_END)
		//_, err = finalFile.WriteAt(content, n
		stat, err := file.Stat()
		if err != nil {
			panic(err)
		}
		num := stat.Size()
		buf := make([]byte, maxLengthPerGoruntine)
		for j := 0; int64(j) < num; {
			length, err := file.Read(buf)
			if err != nil {
				fmt.Println("读取文件错误")
			}
			size, err := finalFile.WriteAt(buf[:length], int64(i))
			if size <= 0 {
				log.Println("write error")
			}
			i += int64(length)
			j += length
			if err != nil {
				log.Fatalln("Write file Error: " + err.Error())
			}
		}
	}
	endTime := time.Now().Unix()
	log.Println("Download complete in " + strconv.FormatInt(endTime-startTime, 10) + "s")
	for _, mvFile := range tmpFile[urlMd5String] {
		err = os.Remove(mvFile)
		if err != nil {
			log.Println("Remove tmp file error " + err.Error())
		}
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
	contentRange := strconv.FormatInt(startBytes, 10) + "-" + strconv.FormatInt(endBytes-1, 10)
	request.Header.Add("Range", "bytes="+contentRange)
	response, err := client.Do(request)
	defer response.Body.Close()
	if err != nil || response == nil {
		log.Fatalln("Download error : " + err.Error())
		return
	}
	//httpStatus := response.StatusCode
	//log.Println("download status: " + response.Status)
	tmpFileName := tmpFile[urlMd5String][index]
	tmpFile, err := os.OpenFile(tmpFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	defer tmpFile.Close()
	if err != nil {
		log.Fatalln("open file error : " + err.Error())
		return
	}
	content, _ := ioutil.ReadAll(response.Body)
	size, err := tmpFile.Write(content)
	if err != nil {
		log.Fatalln("copy file to tmp file error :" + err.Error())
		return
	}
	if size <= 0 {
		log.Fatalln("download error")
		return
	}
	return
}
