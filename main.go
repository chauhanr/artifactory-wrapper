package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var port = ":8080"

func main() {
	log.Println("[Artifactory Wrapper service starting ...]")
	setRoutes()
}

func setRoutes() {
	http.HandleFunc("/upload", Upload)
	log.Fatal(http.ListenAndServe(port, nil))
}

func Upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	/* not getting the file from the form.*/
	file, handler, err := r.FormFile("sdk-binary")
	if err != nil {
		log.Printf("Error retriving sdk binary %s", err)
		return
	}
	defer file.Close()
	log.Printf("Upload file: %+v size: %+v header: %+v\n", handler.Filename, handler.Size, handler.Header)
	tmp, err := ioutil.TempFile("", "ios-sdk.framework")
	if err != nil {
		log.Printf("Error in temp file creation %s\n", err)
	}
	defer tmp.Close()
	fb, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintf(w, "error reading form file", http.StatusInternalServerError)
		return
	}
	tmp.Write(fb)

	fmt.Fprintf(w, "Successfully uploaded file", http.StatusOK)
}
