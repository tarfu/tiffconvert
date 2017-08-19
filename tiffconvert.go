package main

import (
	"os"
	"path/filepath"
	"log"
	"strings"
	"io/ioutil"
	"golang.org/x/image/tiff"
	"bytes"
	"image/jpeg"
	"flag"
	"sync"
	"runtime"
	"fmt"
)

var queue chan convertJob
var scanPrefix string
var convertPrefix string

var wg sync.WaitGroup

type convertJob struct{
	tiffPath, jpegPath string
}

func main() {

	scanPrefixL := flag.String("tiff", ".", "path to the folder with tiffs")
	convertPrefixL := flag.String("jpeg", "converted", "path to the folder with the converted jpegs")

	flag.Parse()

	scanPrefix = filepath.Clean(*scanPrefixL)
	convertPrefix = filepath.Clean(*convertPrefixL)
	queue = make(chan convertJob, 10)

	for i := 0; i<runtime.NumCPU() ; i++  {
		go converter(queue)
	}

	err := filepath.Walk(scanPrefix, walked)
	if err != nil {
		log.Printf("error walking")
	}


	wg.Wait()


}

func converter(receiver chan convertJob){
	for job := range receiver {
		imgData, err := ioutil.ReadFile(job.tiffPath)
		if err != nil {
			log.Printf("could not read %v", err)
		}
		reader := bytes.NewReader(imgData)
		image, err := tiff.Decode(reader)
		if err != nil {
			log.Printf("Could not decode tiff %v", err)
		}

		buf := bytes.NewBuffer(make([]byte, 0))
		err = jpeg.Encode(buf, image, &jpeg.Options{Quality:100})
		if err != nil {
			log.Printf("could not encode %v", err)
		}
		err = ioutil.WriteFile(job.jpegPath, buf.Bytes(), os.ModePerm)
		if err != nil {
			log.Printf("could not write file %v", err)
		}
		fmt.Println(job)
		wg.Done()

	}
}





func walked(walkedpath string, info os.FileInfo, err error) error {
	jpegpath := filepath.Join(convertPrefix, string(walkedpath[len(scanPrefix)-1:]))
	if strings.HasPrefix(walkedpath, convertPrefix){
		return filepath.SkipDir
	}
	if strings.HasSuffix(walkedpath, ".tiff") {
		jpegpath = jpegpath[:len(jpegpath)-4]+"jpg"
	}else if strings.HasSuffix(walkedpath, ".tif"){
		jpegpath = jpegpath[:len(jpegpath)-3]+"jpg"
	} else{
		return nil
	}


	converted, err2 := checkConvertedAndCreateFolder(walkedpath, jpegpath)
	if err2 != nil {
		log.Printf("%v check said %v", walkedpath, err2)
	}
	if !converted {
		wg.Add(1)
		queue <- convertJob{tiffPath: walkedpath, jpegPath: jpegpath}
	}

	return nil
}

func checkConvertedAndCreateFolder(tiffPath, jpegPath string) (bool, error){

	if _, err := os.Stat(jpegPath); !os.IsNotExist(err) {
		return true, nil

	}
	dir := filepath.Dir(jpegPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(jpegPath), os.ModePerm)
		if err != nil {
			return false, err
		}
	}
	return false, nil
}