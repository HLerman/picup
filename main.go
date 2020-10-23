package main

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

const MB = 1 << 20

// Initialize configuration
func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Fatal error config file: " + err.Error())
	}

	viper.SetDefault("port", "8090")
	viper.SetDefault("baseUrl", "http://127.0.0.1:8090")
	viper.SetDefault("directory", "download/")
	viper.SetDefault("virtualDirectory", "/download")
	viper.SetDefault("maxSizeInMB", 10)
	viper.SetDefault("acceptedFileType", []string{})

	// Rewrite properly the virtualDirectory
	viper.Set("virtualDirectory", "/"+strings.Trim(viper.GetString("virtualDirectory"), "/"))

	// Rewrite properly the baseUrl
	viper.Set("baseUrl", strings.TrimRight(viper.GetString("baseUrl"), "/"))
}

// Initialize logs
func init() {
	ljack := &lumberjack.Logger{
		Filename:   viper.GetString("log.path"),
		MaxSize:    viper.GetInt("log.maxSize"), // megabytes
		MaxBackups: viper.GetInt("log.maxBackups"),
		MaxAge:     viper.GetInt("log.maxAge"), //days
		Compress:   viper.GetBool("log.compress"),
	}

	gin.DefaultWriter = io.MultiWriter(os.Stdout, ljack)
	log.SetOutput(gin.DefaultWriter)
}

func main() {
	// Starting application
	log.WithFields(log.Fields{
		"Runtime Version": runtime.Version(),
		"Number of CPUs":  runtime.NumCPU(),
		"Arch":            runtime.GOARCH,
	}).Info("Application Initializing")

	// Release or Debug mode
	if viper.GetString("mode") == "release" {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
	}

	// Load Logger and Recovery feature
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Allow CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	router.Use(cors.New(config))

	// Serving pictures
	router.Static(viper.GetString("virtualDirectory"), viper.GetString("directory"))

	// Upload picture
	router.MaxMultipartMemory = viper.GetInt64("maxSizeInMB") << 20
	router.POST("/upload", fileUpload)

	srv := &http.Server{
		Addr:    ":" + viper.GetString("port"),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	/* Wait for interrupt signal to gracefully shutdown the server with
	a timeout of 5 seconds. */
	quit := make(chan os.Signal)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}

func fileUpload(c *gin.Context) {
	// Multipart form
	form, err := c.MultipartForm()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	files := form.File["upload"]

	var results []string

	for _, file := range files {
		// Upload the file to specific dst.
		directory := createAndReturnDirectory()

		filePath := filepath.Join(viper.GetString("directory"), directory, file.Filename)

		err := c.SaveUploadedFile(file, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		contentType, err := getFileContentType(filePath)
		if err != nil {
			// remove file because it's not a valid type
			os.Remove(filePath)

			c.JSON(http.StatusUnsupportedMediaType, gin.H{"message": err.Error()})
			return
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
			os.Remove(filePath)

			c.JSON(http.StatusUnsupportedMediaType, gin.H{"message": "Unaccepted file format " + contentType})
			return
		}

		url, err := url.Parse(viper.GetString("baseUrl"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		url.Path = path.Join(url.Path, viper.GetString("virtualDirectory"), directory, file.Filename)
		results = append(results, url.String())
	}

	c.JSON(http.StatusOK, results)
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
	randomPath := randomName(7)
	newpath := filepath.Join(viper.GetString("directory"), randomPath)
	os.MkdirAll(newpath, os.ModePerm)

	return randomPath + "/"
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
