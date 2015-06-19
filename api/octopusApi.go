/*
== BSD2 LICENSE ==
Copyright (c) 2015, Tidepool Project

This program is free software; you can redistribute it and/or modify it under
the terms of the associated License, which is identical to the BSD 2-Clause
License as published by the Open Source Initiative at opensource.org.

This program is distributed in the hope that it will be useful, but WITHOUT
ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
FOR A PARTICULAR PURPOSE. See the License for more details.

You should have received a copy of the License along with this program; if
not, you can obtain one from Tidepool Project at tidepool.org.
== BSD2 LICENSE ==
*/

package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	httpgzip "github.com/daaku/go.httpgzip"
	"github.com/gorilla/mux"
	commonClients "github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/clients/shoreline"

	"./../clients"
	"./../model"
)

const (
	SESSION_TOKEN    = "x-tidepool-session-token"
	QUERY_API_PREFIX = "api/query"

	//error messages
	error_building_query           = "There was an issue trying to build the query to run"
	error_no_userid                = "userid not found"
	error_no_permissons            = "permissons not found"
	error_running_query            = "error running query"
	error_checking_service_prereqs = "error when checking the service pre-reqs"
)

type (
	Api struct {
		Store            clients.StoreClient
		ShorelineClient  ShorelineInterface
		SeagullClient    SeagullInterface
		GatekeeperClient GatekeeperInterface
		Config           Config
	}
	Config struct {
		ServerSecret string `json:"serverSecret"` //used for services
	}

	GatekeeperInterface interface {
		UserInGroup(userID, groupID string) (map[string]commonClients.Permissions, error)
	}

	SeagullInterface interface {
		GetPrivatePair(userID, hashName, token string) *commonClients.PrivatePair
	}

	ShorelineInterface interface {
		CheckToken(token string) *shoreline.TokenData
		GetUser(userID, token string) (*shoreline.UserData, error)
		TokenProvide() string
	}

	httpVars    map[string]string
	varsHandler func(http.ResponseWriter, *http.Request, httpVars)
	gzipHandler func(http.ResponseWriter, *http.Request)
)

//find and validate the token
func (a *Api) authorized(req *http.Request) bool {

	if token := a.getToken(req); token != "" {
		if td := a.ShorelineClient.CheckToken(token); td != nil {
			log.Println(QUERY_API_PREFIX, "token check succeeded")
			return true
		}
		log.Println(QUERY_API_PREFIX, "token check failed")
		return false
	}
	log.Println(QUERY_API_PREFIX, "no token to check")
	return false
}

func (a *Api) userCanViewData(userID, groupID string) bool {
	if userID == groupID {
		return true
	}

	perms, err := a.GatekeeperClient.UserInGroup(userID, groupID)
	if err != nil {
		log.Println(QUERY_API_PREFIX, "Error looking up user in group", err)
		return false
	}
	log.Println(QUERY_API_PREFIX, "found perms ", perms)
	return !(perms["root"] == nil && perms["view"] == nil)
}

//just return the token
func (a *Api) getToken(req *http.Request) string {
	return req.Header.Get(SESSION_TOKEN)
}

//givenId could be the actual id or the users email address which we also treat as an id
func (a *Api) getUserPairId(givenId, token string) (string, error) {

	usr, err := a.ShorelineClient.GetUser(givenId, token)
	if err != nil {
		log.Println(QUERY_API_PREFIX, fmt.Sprintf("getUserPairId: error [%s] getting user id [%s]", err.Error(), givenId))
		return "", errors.New(error_no_userid)
	}
	pair := a.SeagullClient.GetPrivatePair(usr.UserID, "uploads", a.ShorelineClient.TokenProvide())
	if pair == nil {
		log.Println(QUERY_API_PREFIX, fmt.Sprintf("getUserPairId: no permissons found for [%s]", usr.UserID))
		return "", errors.New(error_no_permissons)
	}
	return pair.ID, nil
}

func InitApi(cfg Config, slc ShorelineInterface,
	sgc SeagullInterface, gkc GatekeeperInterface,
	store clients.StoreClient) *Api {

	return &Api{
		Store:            store,
		ShorelineClient:  slc,
		SeagullClient:    sgc,
		GatekeeperClient: gkc,
		Config:           cfg,
	}
}

func (a *Api) SetHandlers(prefix string, rtr *mux.Router) {
	rtr.HandleFunc("/status", a.GetStatus).Methods("GET")
	rtr.Handle("/upload/lastentry/{userID}", varsHandler(a.TimeLastEntryUser)).Methods("GET")
	rtr.Handle("/upload/lastentry/{userID}/{deviceID}", varsHandler(a.TimeLastEntryUserAndDevice)).Methods("GET")

	rtr.Handle("/data", httpgzip.NewHandler(gzipHandler(a.Query))).Methods("POST")

}

// http.StatusOK
// http.StatusInternalServerError - something is wrong with a service pre-req
func (a *Api) GetStatus(res http.ResponseWriter, req *http.Request) {
	start := time.Now()
	if err := a.Store.Ping(); err != nil {
		log.Println(QUERY_API_PREFIX, fmt.Sprintf("GetStatus: failed after [%.5f] secs with error[%s]", time.Now().Sub(start).Seconds(), err.Error()))
		//don't want to leak the actual error - we have logged it above
		http.Error(res, error_checking_service_prereqs, http.StatusInternalServerError)
		return
	}
	log.Println(QUERY_API_PREFIX, fmt.Sprintf("GetStatus: completed in [%.5f] secs", time.Now().Sub(start).Seconds()))
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}

// http.StatusOK, time of last entry
// http.StatusBadRequest - something was wrong with the request data
// http.StatusUnauthorized - you don't have a valid token
// http.StatusForbidden - you have a valid token but don't have permisson to look at the data
func (a *Api) TimeLastEntryUser(res http.ResponseWriter, req *http.Request, vars httpVars) {

	start := time.Now()

	if a.authorized(req) {

		userId := vars["userID"]

		groupId, err := a.getUserPairId(userId, a.getToken(req))
		if err != nil {
			log.Println(QUERY_API_PREFIX, fmt.Sprintf("TimeLastEntryUser: failed after [%.5f] secs with error[%s]", time.Now().Sub(start).Seconds(), err.Error()))
			http.Error(res, err.Error(), http.StatusBadRequest)
		}
		if a.userCanViewData(userId, groupId) {
			timeLastEntry := a.Store.GetTimeLastEntryUser(groupId)
			log.Println(QUERY_API_PREFIX, fmt.Sprintf("TimeLastEntryUser: completed in [%.5f] secs", time.Now().Sub(start).Seconds()))
			res.Header().Set("content-type", "application/json")
			res.WriteHeader(http.StatusOK)
			res.Write(timeLastEntry)
			return
		}
		log.Println(QUERY_API_PREFIX, fmt.Sprintf("TimeLastEntryUser: failed after [%.5f] secs with error[%s]", time.Now().Sub(start).Seconds(), error_no_permissons))
		http.Error(res, error_no_permissons, http.StatusForbidden)
		return
	}
	log.Println(QUERY_API_PREFIX, fmt.Sprintf("TimeLastEntryUser: failed authorization after [%.5f] secs", time.Now().Sub(start).Seconds()))
	http.Error(res, "failed authorization", http.StatusUnauthorized)
	return
}

// http.StatusOK, time of last entry and device
// http.StatusBadRequest - something was wrong with the request data
// http.StatusUnauthorized - you don't have a valid token
// http.StatusForbidden - you have a valid token but don't have permisson to look at the data
func (a *Api) TimeLastEntryUserAndDevice(res http.ResponseWriter, req *http.Request, vars httpVars) {

	start := time.Now()

	if a.authorized(req) {
		userId := vars["userID"]
		groupId, err := a.getUserPairId(userId, a.getToken(req))
		if err != nil {
			log.Println(QUERY_API_PREFIX, fmt.Sprintf("TimeLastEntryUserAndDevice: failed after [%.5f] secs with error[%s]", time.Now().Sub(start).Seconds(), err.Error()))
			http.Error(res, err.Error(), http.StatusBadRequest)
		}
		if a.userCanViewData(userId, groupId) {
			timeLastEntry := a.Store.GetTimeLastEntryUserAndDevice(groupId, vars["deviceID"])
			log.Println(QUERY_API_PREFIX, fmt.Sprintf("TimeLastEntryUserAndDevice: completed in [%.5f] secs", time.Now().Sub(start).Seconds()))
			res.Header().Set("content-type", "application/json")
			res.WriteHeader(http.StatusOK)
			res.Write(timeLastEntry)
			return
		}
		log.Println(QUERY_API_PREFIX, fmt.Sprintf("TimeLastEntryUserAndDevice: failed after [%.5f] secs with error[%s]", time.Now().Sub(start).Seconds(), error_no_permissons))
		http.Error(res, error_no_permissons, http.StatusForbidden)
		return
	}
	log.Println(QUERY_API_PREFIX, fmt.Sprintf("TimeLastEntryUserAndDevice: failed authorization after [%.5f] secs", time.Now().Sub(start).Seconds()))
	http.Error(res, "failed authorization", http.StatusUnauthorized)
	return

}

// http.StatusOK - the requested data
// http.StatusBadRequest - something was wrong with the request data
// http.StatusUnauthorized - you don't have a valid token
func (a *Api) Query(res http.ResponseWriter, req *http.Request) {

	start := time.Now()

	if a.authorized(req) {

		log.Println(QUERY_API_PREFIX, "Query: starting ... ")

		defer req.Body.Close()
		rawQuery, err := ioutil.ReadAll(req.Body)

		if err != nil || string(rawQuery) == "" {
			log.Println(QUERY_API_PREFIX, fmt.Sprintf("Query: err decoding nonempty response body: [%v]\n [%v]\n", err, req.Body))
			http.Error(res, error_building_query, http.StatusBadRequest)
			return
		}
		query := string(rawQuery)

		log.Println(QUERY_API_PREFIX, "Query: raw ", query)

		errs, qd := model.BuildQuery(query)

		if len(errs) != 0 {
			log.Println(QUERY_API_PREFIX, fmt.Sprintf("Query: errors [%v] found parsing raw query [%s]", errs, query))
			log.Println(QUERY_API_PREFIX, fmt.Sprintf("Query: failed after [%.5f] secs", time.Now().Sub(start).Seconds()))
			http.Error(res, fmt.Sprintf("Errors building query: [%v]", errs), http.StatusBadRequest)
			return
		}

		pairId, err := a.getUserPairId(qd.GetMetaQueryId(), a.getToken(req))

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		qd.SetMetaQueryId(pairId)

		result, err := a.Store.ExecuteQuery(qd)

		if err != nil {
			log.Println(QUERY_API_PREFIX, fmt.Sprintf("Query: failed after [%.5f] secs", time.Now().Sub(start).Seconds()))
			log.Println(QUERY_API_PREFIX, "Query:", error_running_query, err.Error())
			http.Error(res, error_running_query, http.StatusInternalServerError)
			return
		}
		// yay we made it! lets give them what they asked for
		log.Println(QUERY_API_PREFIX, fmt.Sprintf("Query: completed in [%.5f] secs", time.Now().Sub(start).Seconds()))
		res.Header().Set("content-type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(result)
		return

	}
	log.Print(QUERY_API_PREFIX, "Query: failed authorization")
	http.Error(res, "failed authorization", http.StatusUnauthorized)
	return
}

func (h varsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	h(res, req, vars)
}

func (h gzipHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h(res, req)
}
