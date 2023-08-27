package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var directory string

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	directory = os.Getenv("FILE_DIRECTORY")
	if directory == "" {
		log.Fatal("FILE_DIRECTORY must be set in .env")
	}
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	basePath := filepath.Join(directory, r.URL.Path)

	entries, err := ioutil.ReadDir(basePath)
	if err != nil {
		http.Error(w, "Unable to read files", http.StatusInternalServerError)
		return
	}

	// Constructing breadcrumbs
	parts := strings.Split(r.URL.Path, "/")
	breadcrumbPath := "/"
	w.Write([]byte("<html><body><div><a href='/'>./</a>"))
	for _, part := range parts {
		if part != "" {
			breadcrumbPath = filepath.Join(breadcrumbPath, part)
			w.Write([]byte(fmt.Sprintf("<a href='%s/'>%s/</a>", breadcrumbPath, part)))
		}
	}
	w.Write([]byte("</div><ul>"))

	// List links to files and directories
	for _, entry := range entries {
		entryPath := filepath.Join(r.URL.Path, entry.Name())
		if entry.IsDir() {
			w.Write([]byte(fmt.Sprintf("<li><a href='%s/'>%s/</a></li>", entryPath, entry.Name())))
		} else {
			w.Write([]byte(fmt.Sprintf("<li><a href='%s'>%s</a></li>", entryPath, entry.Name())))
		}
	}
	w.Write([]byte("</ul></body></html>"))
}

func serveFileOrDirectory(w http.ResponseWriter, r *http.Request) {
	relativePath := strings.TrimPrefix(r.URL.Path, "/")
	absolutePath := filepath.Join(directory, relativePath)

	fi, err := os.Stat(absolutePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if fi.IsDir() {
		listFiles(w, r)
	} else {
		// Logging the download
		log.Printf("[%s] File downloaded: %s by IP %s", time.Now().Format(time.RFC1123), absolutePath, r.RemoteAddr)
		http.ServeFile(w, r, absolutePath)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set in .env")
	}

	http.HandleFunc("/", serveFileOrDirectory)

	log.Printf("Server started on :%s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
