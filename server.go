package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"strconv"
	"time"
)

const (
	envMsPort   = "MICROSERVICE_PORT"
	defaultPort = 80
	contentRoot = "/www"
)

func main() {

	debugEnabled := flag.Bool("debug", false, "enable debugging")

	handler := func(w http.ResponseWriter, r *http.Request) {
		filePath := path.Join(contentRoot, r.URL.Path)
		log.Printf("Serving '%s'", filePath)
		if *debugEnabled {
			dump, _ := httputil.DumpRequest(r, false)
			log.Printf("Request dump\n%s", dump)
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
