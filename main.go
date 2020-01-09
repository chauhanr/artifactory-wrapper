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

		fb, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, "error reading from uploaded file", http.StatusInternalServerError)
			return
		}

		err = ioutil.WriteFile(handler.Filename, fb, 0644)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, "error writing file", http.StatusInternalServerError)
			return
		}
		/* Now we need to send the temp file to the Artifactory Client*/
		err = PublishToArtifactory(handler.Filename, "mvn-local", "HelloWorldSDK", "0.0.1", handler.Filename)
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

func PublishToArtifactory(dataFile string, repo string, framework string, version string, fname string) error {
	data, err := os.Open(dataFile)
	if err != nil {
		return err
	}
	url := prepareArtifactoryUploadURL(repo, framework, version, fname)
	r, err := http.NewRequest("PUT", url, data)
	if err != nil {
		log.Printf("Error preparing the new request %s\n", err)
		return err
	}
	r.SetBasicAuth("admin", "password")
	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		log.Printf("Error publishing the file to artifactory %s\n", err)
	}
	code := res.StatusCode

	bs, _ := ioutil.ReadAll(res.Body)
	fmt.Printf("Artifactory Status Code %d, Response Body: %s\n", code, string(bs))

	if code != 200 || code != 201 {
		log.Printf("Error uploading file to artifactory %d\n", code)
		return errors.New("Unable to upload file successfully to artifactory")
	}

	defer res.Body.Close()
	return nil
}

func prepareArtifactoryUploadURL(repo, framework, version, filename string) string {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", BASE_URL, repo, framework, version, filename)
	fmt.Printf("URL: %s\n", url)
	return url
}
