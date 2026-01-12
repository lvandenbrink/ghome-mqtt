package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	log "log/slog"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"

	"github.com/gorilla/mux"
	auth2 "github.com/mrlauy/ghome-mqtt/auth"
	"github.com/mrlauy/ghome-mqtt/config"
	"github.com/mrlauy/ghome-mqtt/fulfillment"
	"github.com/mrlauy/ghome-mqtt/mqtt"
)

//go:embed templates/*
var files embed.FS

const requestFullDump = false

func main() {
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Error("failed to read config", "err", err)
		return
	}

	config.InitLogging(cfg.Log.Level)

	auth := auth2.NewAuth(cfg.Auth)
	messageHandler, err := mqtt.NewMqtt(cfg.Mqtt)
	if err != nil {
		log.Error("failed to start mqtt", "err", err)
		return
	}

	fulfillmentManager, err := fulfillment.NewFulfillment(messageHandler, cfg.Devices, cfg.ExecutionTemplates)
	if err != nil {
		log.Error("failed to start fulfillment handler", "err", err)
		return
	}

	staticResources, err := fs.Sub(files, "templates/static")
	if err != nil {
		log.Error("failed to load static resources", "err", err)
		return
	}

	staticFileHandler := http.StripPrefix("/static/", http.FileServer(http.FS(staticResources)))
	loginPage := template.Must(template.ParseFS(files, "templates/login.html"))
	authPage := template.Must(template.ParseFS(files, "templates/auth.html"))

	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	router.PathPrefix("/static/").Handler(staticFileHandler)
	router.HandleFunc("/login", auth.Login(loginPage))
	router.HandleFunc("/confirm", auth.Confirm(authPage))

	router.HandleFunc("/oauth/authorize", auth.Authorize)
	router.HandleFunc("/oauth/token", auth.Token)

	smarthomeRouter := router.PathPrefix("/smarthome").Subrouter()
	smarthomeRouter.Use(auth.ValidateToken)
	smarthomeRouter.HandleFunc("/fulfillment", fulfillmentManager.Handler).Methods("POST")

	http.Handle("/", router)

	port := cfg.Server.Port
	log.Info("started server", "port", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	log.Error("failure during execution", "err", err)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestFullDump {
			headers := r.Header

			// only print the following headers
			r.Header = map[string][]string{}
			r.Header.Add("Cookie", headers.Get("Cookie"))
			r.Header.Add("Referer", headers.Get("Referer"))
			r.Header.Add("Authorization", headers.Get("Authorization"))

			data, err := httputil.DumpRequest(r, true)
			if err != nil {
				log.Error("error dumping request", "err", err)
				return
			}
			r.Header = headers

			log.Info(fmt.Sprintf("\n> %s \n%v", r.URL, string(data)))

			recorder := httptest.NewRecorder()
			next.ServeHTTP(recorder, r)

			dump, err := httputil.DumpResponse(recorder.Result(), true)
			if err != nil {
				log.Error("error dumping response", "err", err)
				return
			}
			log.Info(fmt.Sprintf("\n< %s \n %v\n", r.URL, string(dump)))

			// we copy the captured response headers to our new response
			for k, v := range recorder.Header() {
				w.Header()[k] = v
			}

			// grab the captured response body
			response := recorder.Body.Bytes()

			w.WriteHeader(recorder.Code)
			_, _ = w.Write(response)
		} else {
			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r)
		}
	})
}
