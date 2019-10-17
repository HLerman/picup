package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
)

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong\n")
}

func upload(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		file, err := FileUpload(req)
		if err != nil {
			log.Println(err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
		} else {
			fmt.Fprintf(w, "%s\n", file)
		}
	}
}

func FileUpload(r *http.Request) (string, error) {
	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		return "", err
	}
	defer file.Close()

	f, err := os.OpenFile(createDirectory()+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		return "", err
	}

	io.Copy(f, file)

	return handler.Filename, nil
}

func main() {
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/upload", upload)

	http.ListenAndServe(":8090", nil)
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
	folderPath := "./download/" + randomName(7)
	newpath := filepath.Join(folderPath)
	os.MkdirAll(newpath, os.ModePerm)

	return folderPath + "/"
}
