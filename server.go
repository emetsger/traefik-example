package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	envMsPort   = "MICROSERVICE_PORT"
	defaultPort = 80
	contentRoot = "/www"
)

func main() {

	model := struct {
		Hostname string
		Request  string
	}{
		Hostname: mustHostname(),
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var err error

		filePath := path.Join(contentRoot, r.URL.Path)
		log.Printf("Serving '%s'", filePath)

		if strings.HasSuffix(filePath, ".tmpl") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("I refuse to serve raw template files.\n"))
			return
		}

		if tmplPath, exists := isTemplateRequest(filePath); exists {
			dump, _ := httputil.DumpRequest(r, false)
			model.Request = string(dump)
			if filePath, err = processTemplate(tmplPath, model); err != nil {
				writeErr(w, err)
			}
		}

		f, err := os.Open(filePath)
		var lastMod time.Time
		var name string
		if err != nil {
			writeErr(w, err)
			return
		} else {
			if info, err := f.Stat(); err == nil {
				if info.IsDir() {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("I refuse to serve directory contents.\n"))
					return
				}
				lastMod = info.ModTime()
				name = info.Name()
			} else {
				writeErr(w, err)
				return
			}
		}

		http.ServeContent(w, r, name, lastMod, f)
	}
	http.HandleFunc("/", handler)

	listenPort := mustGetIntOrDefault(envMsPort, defaultPort)
	log.Printf("Listening on port %d, serving from directory '%s'", listenPort, contentRoot)
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", listenPort), nil))
}

func writeErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func mustGetIntOrDefault(env string, defaultV int) int {
	if value, exists := os.LookupEnv(env); exists {
		if intVal, err := strconv.Atoi(value); err != nil {
			log.Fatalf("cannot convert value of %s to int: %s", env, err)
		} else {
			return intVal
		}
	}

	return defaultV
}

func mustHostname() string {
	host, err := os.Hostname()
	if err != nil {
		log.Fatalf("error getting hostname: %s", err)
	}
	return host
}

func isTemplateRequest(path string) (string, bool) {
	_, err := os.Stat(path)
	if err == nil {
		return "", false
	}

	if errors.Is(err, os.ErrNotExist) {
		tmplPath := strings.ReplaceAll(path, ".html", ".tmpl")
		_, err := os.Stat(tmplPath)
		return tmplPath, err == nil
	}

	return "", false
}

func processTemplate(tmplPath string, model interface{}) (string, error) {

	var (
		tmpl    *template.Template
		outFile *os.File
		err     error
	)

	defer func() {
		if outFile != nil {
			outFile.Close()
		}
	}()

	if tmpl, err = template.ParseFiles(tmplPath); err != nil {
		return "", err
	}

	outPath := strings.ReplaceAll(tmplPath, ".tmpl", ".html")
	outFile, err = os.OpenFile(outPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(outFile, model)
	if err != nil {
		return "", err
	}
	return outPath, nil
}
