package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	if r.Method == http.MethodPut {
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
			http.Error(w, "error reading from uploaded file", http.StatusInternalServerError)
			return
		}
		tmp.Write(fb)
		in, err := tmp.Stat()
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, "error writing file", http.StatusInternalServerError)
			return
		}
		log.Printf("Size of the new file: %d\n", in.Size)
		/* Now we need to send the temp file to the Artifactory Client*/
		err = PublishToArtifactory(tmp, "cocoapods-local", "HelloWorldSDK", "0.0.1")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			io.WriteString(w, "Successfully uploaded file to Artifactory")
		}
	} else {
		http.Error(w, "Unsupported method type", http.StatusBadRequest)
	}

}

var BASE_URL = "http://localhost:8080/artifactory"

func PublishToArtifactory(data *os.File, repo string, framework string, version string) error {
	url := prepareArtifactoryUploadURL(repo, framework, version)
	r, err := http.NewRequest("PUT", url, data)
	if err != nil {
		log.Printf("Error preparing the new request %s\n", err)
		return err
	}
	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		log.Printf("Error publishing the file to artifactory %s\n", err)
	}
	code := res.StatusCode

	if code != 200 || code != 201 {
		log.Printf("Error uploading file to artifactory %s\n", err)
		return errors.New("Unable to upload file successfully to artifactory")
	}
	bs, _ := ioutil.ReadAll(res.Body)
	fmt.Printf("Artifactory Response Body: %s\n", string(bs))
	defer res.Body.Close()
	return nil
}

func prepareArtifactoryUploadURL(repo, framework, version string) string {
	url := fmt.Sprintf("%s/%s/%s/%s/", BASE_URL, repo, framework, version)
	fmt.Printf("URL: %s\n", url)
	return url
}
