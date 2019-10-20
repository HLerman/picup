package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
)

const baseURL string = "http://127.0.0.1:8090/"
const MB = 1 << 20

var acceptedType = [...]string{"image/x-icon", "image/gif", "image/png", "image/jpeg", "image/vnd.adobe.photoshop", "image/tiff"}

type FileSystem struct {
	fs http.FileSystem
}

func main() {
	directory := flag.String("d", "./download", "directory of uploaded files")
	port := flag.String("p", "8090", "port to serve on")
	flag.Parse()

	fileServer := http.FileServer(FileSystem{http.Dir(*directory)})

	http.HandleFunc("/ping", ping)
	http.HandleFunc("/upload", upload)
	http.Handle("/download/", http.StripPrefix(strings.TrimRight("/download/", "/"), fileServer))

	http.ListenAndServe(":"+*port, nil)
}

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong")
}

func upload(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, 10*MB)

	if req.Method == "POST" {
		file, err := fileUpload(req)
		if err != nil {
			sendLog("strerr", err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
		} else {
			sendLog("stdout", "file uploaded to "+file)
			fmt.Fprintf(w, "%s%s", baseURL, file)
		}
	}
}

func (fs FileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := fs.fs.Open(index); err != nil {
			return nil, err
		}
	}

	return f, nil
}

func fileUpload(r *http.Request) (string, error) {
	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		return "", err
	}
	defer file.Close()

	directory := createDirectory()
	f, err := os.OpenFile(directory+handler.Filename, os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return "", err
	}

	io.Copy(f, file)

	_, err = getFileContentType(directory + handler.Filename)
	if err != nil {
		return "", err
	}

	return directory + handler.Filename, nil
}

func sendLog(std string, message ...interface{}) {
	log.SetOutput(os.Stdout)
	if std == "stderr" {
		log.SetOutput(os.Stderr)
	}

	log.Println(message)
}

func randomName(size uint8) string {
	if size == 0 {
		size = 1
	}

	letters := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

	name := ""
	for i := uint8(0); i < size; i++ {
		letter := rand.Intn(25)
		name = name + letters[letter]
	}

	return name
}

func createDirectory() string {
	folderPath := "download/" + randomName(7)
	newpath := filepath.Join(folderPath)
	os.MkdirAll(newpath, os.ModePerm)

	return folderPath + "/"
}

func getFileContentType(filePath string) (string, error) {
	buf, _ := ioutil.ReadFile(filePath)

	kind, _ := filetype.Match(buf)
	if kind == filetype.Unknown {
		return "", errors.New("Unknown file type")
	}

	contentType := kind.MIME.Value

	accepted := false
	for i := 0; i < len(acceptedType); i++ {
		if contentType == acceptedType[i] {
			accepted = true
		}
	}

	if accepted != true {
		return "", errors.New("Unaccepted file format " + kind.MIME.Value)
	}

	return contentType, nil
}
