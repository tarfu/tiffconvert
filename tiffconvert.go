package main

import (
	"os"
	"path/filepath"
	"log"
	"strings"
	"path"
	"io/ioutil"
	"golang.org/x/image/tiff"
	"bytes"
	"image/jpeg"
)

var queue chan convertJob
var scanPrefix string
var convertPrefix string

type convertJob struct{
	tiffPath, jpegPath string
}

func main() {
	queue = make(chan convertJob, 10)

	err := filepath.Walk(scanPrefix, walked)
	if err != nil {
		log.Printf("error walking")
	}



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
		err = ioutil.WriteFile(job.jpegPath, buf.Bytes(), 777)
		if err != nil {
			log.Printf("could not write file %v", err)
		}

	}
}





func walked(walkedpath string, info os.FileInfo, err error) error {
	jpegpath := path.Join(convertPrefix, string(walkedpath[len(scanPrefix)]))
	if strings.HasPrefix(walkedpath, convertPrefix){
		return filepath.SkipDir
	}
	if !strings.HasSuffix(walkedpath, ".tiff") || !strings.HasSuffix(walkedpath, ".tif"){
		return nil
	}
	err2 := checkConverted(walkedpath, jpegpath)
	if err2 != nil {
		log.Printf("%v check said %v", walkedpath, err2)
	}

	queue<- convertJob{tiffPath: walkedpath, jpegPath:jpegpath}


	return nil
}

func checkConverted(tiffPath, jpegPath string) error{

	return nil
}