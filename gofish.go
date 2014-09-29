package main

import (
	//"./api"
	//sc "./clients"
	//"github.com/gorilla/mux"
	"crypto/tls"
	"github.com/gorilla/pat"
	"github.com/tidepool-org/go-common"
	"github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/clients/disc"
	"github.com/tidepool-org/go-common/clients/hakken"
	"github.com/tidepool-org/go-common/clients/mongo"
	"github.com/tidepool-org/go-common/clients/shoreline"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type (
	Config struct {
		clients.Config
		Service disc.ServiceListing `json:"service"`
		Mongo   mongo.Config        `json:"mongo"`
		//Api     api.Config          `json:"shoreline"`
	}
)

func main() {
	var config Config

	if err := common.LoadConfig([]string{"./config/env.json", "./config/server.json"}, &config); err != nil {
		log.Panic("Problem loading config", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	/*
	 * Hakken setup
	 */
	hakkenClient := hakken.NewHakkenBuilder().
		WithConfig(&config.HakkenConfig).
		Build()

	if err := hakkenClient.Start(); err != nil {
		log.Fatal(err)
	}
	defer hakkenClient.Close()

	/*
	 * Shoreline setup
	 */
	//store := sc.NewMongoStoreClient(&config.Mongo)

	//rtr := mux.NewRouter()
	//api := api.InitApi(config.Api, store)
	//api.SetHandlers("", rtr)

	/*
	 * Serve it up and publish
	 */

	shorelineClient := shoreline.NewShorelineClientBuilder().
		WithHostGetter(config.ShorelineConfig.ToHostGetter(hakkenClient)).
		WithHttpClient(httpClient).
		WithConfig(&config.ShorelineConfig.ShorelineClientConfig).
		Build()

	gatekeeperClient := clients.NewGatekeeperClientBuilder().
		WithHostGetter(config.GatekeeperConfig.ToHostGetter(hakkenClient)).
		WithHttpClient(httpClient).
		WithTokenProvider(shorelineClient).
		Build()

	userCanViewData := func(userID, groupID string) bool {
		if userID == groupID {
			return true
		}

		perms, err := gatekeeperClient.UserInGroup(userID, groupID)
		if err != nil {
			log.Println("Error looking up user in group", err)
			return false
		}

		log.Println(perms)
		return !(perms["root"] == nil && perms["view"] == nil)
	}

	if err := shorelineClient.Start(); err != nil {
		log.Fatal(err)
	}

	session, err := mongo.Connect(&config.Mongo)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	router := pat.New()

	router.Add("GET", "/status", http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		if err := session.Ping(); err != nil {
			res.WriteHeader(500)
			res.Write([]byte(err.Error()))
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("OK"))
	}))

	router.Add("GET", "/upload/lastentry/{userID}/{deviceID}",
		http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			userID := req.URL.Query().Get(":userID")
			//deviceID := req.URL.Query().Get(":deviceID")
			token := req.Header.Get("x-tidepool-session-token")
			td := shorelineClient.CheckToken(token)

			if td == nil || !(td.IsServer || td.UserID == userID || userCanViewData(td.UserID, userID)) {
				res.WriteHeader(403)
				return
			}

			res.WriteHeader(http.StatusOK)
			res.Write([]byte("Hi there"))

		}))

	done := make(chan bool)
	server := common.NewServer(&http.Server{
		Addr:    config.Service.GetPort(),
		Handler: router,
	})

	var start func() error
	if config.Service.Scheme == "https" {
		sslSpec := config.Service.GetSSLSpec()
		start = func() error { return server.ListenAndServeTLS(sslSpec.CertFile, sslSpec.KeyFile) }
	} else {
		start = func() error { return server.ListenAndServe() }
	}
	if err := start(); err != nil {
		log.Fatal(err)
	}

	//hakkenClient.Publish(&config.Service)

	signals := make(chan os.Signal, 40)
	signal.Notify(signals)
	go func() {
		for {
			sig := <-signals
			log.Printf("Got signal [%s]", sig)

			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				server.Close()
				done <- true
			}
		}
	}()

	<-done

}
