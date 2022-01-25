package main

import (
	"flag"
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

var isVerbose bool

func main() {

	serverRoot := flag.String("serverRoot", contentRoot, "server root dir")
	webRoot := flag.String("webRoot", "", "content of this server is subordinate to this HTTP request path prefix")
	scaleGroup := flag.String("group", "not specified", "a key that groups services that are scaled together")
	isDebug := flag.Bool("debug", false, "enable debugging")
	flag.Parse()

	isVerbose = *isDebug

	model := struct {
		Hostname   string
		Request    string
		ScaleGroup string
		Date       string
	}{
		Hostname:   mustHostname(),
		ScaleGroup: *scaleGroup,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var err error

		filePath := path.Join(*serverRoot, stripRoot(*webRoot, r.URL.Path))
		log.Printf("Serving '%s'", filePath)

		if strings.HasSuffix(filePath, ".tmpl") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("I refuse to serve raw template files.\n"))
			return
		}

		if tmplPath, exists := isTemplateRequest(filePath); exists {
			dump, _ := httputil.DumpRequest(r, false)
			model.Request = string(dump)
			model.Date = time.Now().Format(time.RFC3339)
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
	log.Printf("Listening on port %d, serving from directory '%s'", listenPort, *serverRoot)
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", listenPort), nil))
}

// responds to the request with a 500 and the content of the error
func writeErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

// mustGetIntorDefault answers the value of the requested env var as an integer, or the default value if none exists.
// If the value of the requested env var cannot be converted to an int, mustGetIntOrDefault panics.
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

// mustHostname answers the hostname of the server answering the request, or panics.
func mustHostname() string {
	host, err := os.Hostname()
	if err != nil {
		log.Fatalf("error getting hostname: %s", err)
	}
	return host
}

// isTemplateRequest answers `true` if the request ought to be handled by a Go template
func isTemplateRequest(path string) (string, bool) {
	tmplPath := strings.ReplaceAll(path, ".html", ".tmpl")

	// if a template exists for the requested file, it's a template request
	_, err := os.Stat(tmplPath)
	if err == nil {
		return tmplPath, true
	}

	return "", false
}

// processTemplate executes the template at the supplied path, and returns file path to the output
func processTemplate(tmplPath string, model interface{}) (string, error) {
	if isVerbose {
		log.Printf("Processing %s", tmplPath)
	}

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
	if isVerbose {
		log.Printf("Processed %s to %s", tmplPath, outPath)
	}
	return outPath, nil
}

// stripRoot removes the prefix `root` from `requestPath`
func stripRoot(root, requestPath string) string {
	if strings.HasPrefix(requestPath, root) {
		return requestPath[len(root):]
	}
	return requestPath
}
