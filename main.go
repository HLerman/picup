package main

import (
	"errors"
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
	"github.com/spf13/viper"
)

const MB = 1 << 20

type FileSystem struct {
	fs http.FileSystem
}

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	initialize()
}

func main() {

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Config file not found")
		} else {
			log.Fatal("Config file seems not well formated")
		}
	}

	fileServer := http.FileServer(FileSystem{http.Dir(viper.GetString("directory"))})

	http.HandleFunc("/ping", ping)
	http.HandleFunc("/upload", upload)
	http.Handle("/download/", http.StripPrefix(strings.TrimRight("/download/", "/"), fileServer))

	http.ListenAndServe(":"+viper.GetString("port"), nil)
}

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong")
}

func upload(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, viper.GetInt64("maxSizeInMB")*MB)

	if req.Method == "POST" {
		file, err := fileUpload(req)
		if err != nil {
			sendLog("strerr", err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
		} else {
			sendLog("stdout", "file uploaded to "+file)
			fmt.Fprintf(w, "%s%s", viper.GetString("baseUrl"), file)
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

	directory := createAndReturnDirectory()
	f, err := os.OpenFile(directory+handler.Filename, os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return "", err
	}

	io.Copy(f, file)
	f.Close()

	contentType, err := getFileContentType(directory + handler.Filename)
	if err != nil {
		// remove file because it's not a valid type
		os.Remove(directory + handler.Filename)

		return "", err
	}

	acceptedType := viper.GetStringSlice("acceptedFileType")
	accepted := false
	for i := 0; i < len(acceptedType); i++ {
		if contentType == acceptedType[i] {
			accepted = true
		}
	}

	if accepted != true {
		// remove file because it's not a valid type
		os.Remove(directory + handler.Filename)

		return "", errors.New("Unaccepted file format " + contentType)
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
		letter := rand.Intn(len(letters))
		name = name + letters[letter]
	}

	return name
}

func createAndReturnDirectory() string {
	newpath := filepath.Join(viper.GetString("directory"), randomName(7))
	os.MkdirAll(newpath, os.ModePerm)

	return folderPath + "/"
}

func getFileContentType(filePath string) (string, error) {
	buf, err := ioutil.ReadFile(filePath)

	if err != nil {
		return "", err
	}

	kind, _ := filetype.Match(buf)
	if kind == filetype.Unknown {
		return "", errors.New("Unknown file type")
	}

	return kind.MIME.Value, nil
}

func initialize() {
	viper.SetDefault("port", ":8090")
	viper.SetDefault("baseUrl", "http://127.0.0.1:8090/")
	viper.SetDefault("directory", "./download")
	viper.SetDefault("maxSizeInMB", 10)
	viper.SetDefault("acceptedFileType", []string{})
}
