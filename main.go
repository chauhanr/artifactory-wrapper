package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var port = ":80"
var BASE_URL = "http://localhost:8081/artifactory"

func main() {
	log.Println("[Artifactory Wrapper service starting ...]")
	setRoutes()
}

func setRoutes() {
	//	http.HandleFunc("/upload/android", UploadAndroid)
	//	http.HandleFunc("/upload/ios", UploadiOs)
	http.HandleFunc("/", CheckServer)
	http.HandleFunc("/master-data", ServeFile)
	log.Fatal(http.ListenAndServe(port, nil))
}

func UploadiOs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		r.ParseMultipartForm(10000000)
		fw := r.Header.Get("framework")
		v := r.Header.Get("version")
		repo := r.Header.Get("repo")

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
		defer os.Remove(handler.Filename)

		out, err := PublishToCocoaPodsArtifactory(handler.Filename, repo, fw, v, handler.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			io.WriteString(w, out)
		}
	} else {
		http.Error(w, "Unsupported method type", http.StatusBadRequest)
	}

}

func CheckServer(w http.ResponseWriter, r *http.Request) {
	log.Println("Check successful.")
	io.WriteString(w, "The artifactory wrapper server is up")
}

func ServeFile(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("master-data.json")
	if err != nil {
		http.Error(w, "error getting master file", http.StatusInternalServerError)
	}

	md, err := ioutil.ReadAll(bufio.NewReader(f))
	if err != nil {
		http.Error(w, "error getting master file", http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(md))
}

func UploadAndroid(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		r.ParseMultipartForm(10000000)
		groupId := r.Header.Get("groupId")
		artifactId := r.Header.Get("artifactId")
		v := r.Header.Get("version")

		repo := r.Header.Get("repo")
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
		defer os.Remove(handler.Filename)

		out, err := PublishToArtifactory(handler.Filename, repo, groupId, artifactId, v, handler.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			io.WriteString(w, out)
		}
	} else {
		http.Error(w, "Unsupported method type", http.StatusBadRequest)
	}

}

func PublishToCocoaPodsArtifactory(dataFile, repo, framework, version, fname string) (string, error) {
	data, err := os.Open(dataFile)
	if err != nil {
		return "", err
	}

	url := prepareCocoaPodsArtifactoryUploadURL(repo, framework, version, fname)
	r, err := http.NewRequest("PUT", url, data)

	if err != nil {
		log.Printf("Error preparing the new request %s\n", err)
		return "", err
	}
	r.SetBasicAuth("admin", "password")
	client := &http.Client{}
	res, err := client.Do(r)

	if err != nil {
		log.Printf("Error publishing the file to artifactory %s\n", err)
	}
	defer res.Body.Close()

	code := res.StatusCode
	bs, _ := ioutil.ReadAll(res.Body)
	out := string(bs)
	//fmt.Printf("Artifactory Status Code %d, Response Body: %s\n", code,out)
	if !(code == 200 || code == 201) {
		log.Printf("Error uploading file to artifactory %d\n", code)
		return "", errors.New("Unable to upload file successfully to artifactory")
	}
	return out, nil
}

func prepareCocoaPodsArtifactoryUploadURL(repo, framework, version, filename string) string {
	url := fmt.Sprintf("%s/%s/%s/%s", BASE_URL, repo, framework, version)
	fmt.Printf("URL: %s\n", url)

	return url
}

func PublishToArtifactory(dataFile, repo, groupId, artifactId, version, fname string) (string, error) {
	data, err := os.Open(dataFile)
	if err != nil {
		return "", err
	}

	url := prepareArtifactoryUploadURL(repo, groupId, artifactId, version, fname)
	r, err := http.NewRequest("PUT", url, data)

	if err != nil {
		log.Printf("Error preparing the new request %s\n", err)
		return "", err
	}
	r.SetBasicAuth("admin", "password")
	client := &http.Client{}
	res, err := client.Do(r)

	if err != nil {
		log.Printf("Error publishing the file to artifactory %s\n", err)
	}
	defer res.Body.Close()

	code := res.StatusCode
	bs, _ := ioutil.ReadAll(res.Body)
	out := string(bs)
	//fmt.Printf("Artifactory Status Code %d, Response Body: %s\n", code,out)
	if !(code == 200 || code == 201) {
		log.Printf("Error uploading file to artifactory %d\n", code)
		return "", errors.New("Unable to upload file successfully to artifactory")
	}
	return out, nil
}

func prepareArtifactoryUploadURL(repo, groupId, artifactId, version, filename string) string {
	ngid := transformGroupId(groupId)
	url := fmt.Sprintf("%s/%s/%s/%s/%s/%s", BASE_URL, repo, ngid, artifactId, version, filename)
	fmt.Printf("URL: %s\n", url)

	return url
}

func transformGroupId(groupId string) string {
	gid := strings.ReplaceAll(groupId, ".", "/")
	return gid
}
