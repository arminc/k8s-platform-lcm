package internal

import (
	"context"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

// WebData is all the information web page needs to show the data
type WebData struct {
	Status          string
	LastTimeFetched string
	ContainerInfo   []ContainerInfo
	ChartInfo       []ChartInfo
	ToolInfo        []ToolInfo
}

var (
	// WebDataVar makes the WebData available
	WebDataVar = WebData{}
)

// StartServer starts the server
func StartServer() {
	r := mux.NewRouter()

	health := healthcheck.NewHandler()
	r.Handle("/live", health)
	r.Handle("/ready", health)

	r.HandleFunc("/", index)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	n := negroni.Classic()
	n.UseHandler(r)

	addr := ":7321"
	log.WithFields(log.Fields{"addr": addr}).Info("Started server")

	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      n,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.WithError(err).Error("Could not start server")
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to properly shutdown")
	}
	log.Info("Shutting down")
	os.Exit(0)
}

func index(w http.ResponseWriter, req *http.Request) {
	templates := template.Must(template.ParseGlob("templates/*"))
	err := templates.ExecuteTemplate(w, "index.gohtml", WebDataVar)
	if err != nil {
		log.WithError(err).Error("Could not server index template")
	}
}
