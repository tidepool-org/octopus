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
	"encoding/json"
	"errors"
	"log"
	"net/http"

	httpgzip "github.com/daaku/go.httpgzip"
	"github.com/gorilla/mux"
	commonClients "github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/clients/shoreline"

	"./../clients"
)

const (
	SESSION_TOKEN    = "x-tidepool-session-token"
	QUERY_API_PREFIX = "api/query"
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

func (a *Api) sendModelAsResWithStatus(res http.ResponseWriter, model interface{}, statusCode int) {
	if jsonDetails, err := json.Marshal(model); err != nil {
		log.Printf(QUERY_API_PREFIX, "Error trying to send [%v]", model)
		http.Error(res, "Error marshaling data for response", http.StatusInternalServerError)
	} else {
		res.Header().Set("content-type", "application/json")
		res.WriteHeader(statusCode)
		res.Write(jsonDetails)
	}
	return
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

//find and validate the token
func (a *Api) authorized(req *http.Request) bool {

	if token := a.getToken(req); token != "" {
		if td := a.ShorelineClient.CheckToken(token); td == nil {
			return false
		}
		//all good!
		return true
	}
	return false
}

//just return the token
func (a *Api) getToken(req *http.Request) string {
	return req.Header.Get(SESSION_TOKEN)
}

func (a *Api) authorizeAndGetGroupId(res http.ResponseWriter, req *http.Request, vars httpVars) (string, error) {
	userID := vars["userID"]

	if td := a.ShorelineClient.CheckToken(req.Header.Get(SESSION_TOKEN)); td == nil || !(td.IsServer || a.userCanViewData(td.UserID, userID)) {
		res.WriteHeader(http.StatusForbidden)
		return "fail", errors.New("Forbidden")
	}

	if pair := a.SeagullClient.GetPrivatePair(userID, "uploads", a.ShorelineClient.TokenProvide()); pair == nil {
		res.WriteHeader(http.StatusInternalServerError)
		return "fail", errors.New("Internal server error")
	} else {
		return pair.ID, nil
	}

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

func (a *Api) GetStatus(res http.ResponseWriter, req *http.Request) {
	if err := a.Store.Ping(); err != nil {
		log.Println(QUERY_API_PREFIX, "GetStatus ", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}

// http.StatusOK,  time of last entry
func (a *Api) TimeLastEntryUser(res http.ResponseWriter, req *http.Request, vars httpVars) {
	if groupId, err := a.authorizeAndGetGroupId(res, req, vars); err != nil {
		res.WriteHeader(http.StatusOK)
		return
	} else {
		timeLastEntry := a.Store.GetTimeLastEntryUser(groupId)
		res.WriteHeader(http.StatusOK)
		res.Write(timeLastEntry)
	}
}

// http.StatusOK, time of last entry and device
func (a *Api) TimeLastEntryUserAndDevice(res http.ResponseWriter, req *http.Request, vars httpVars) {
	if groupId, err := a.authorizeAndGetGroupId(res, req, vars); err != nil {
		return
	} else {

		deviceId := vars["deviceID"]

		timeLastEntry := a.Store.GetTimeLastEntryUserAndDevice(groupId, deviceId)

		res.WriteHeader(http.StatusOK)
		res.Write(timeLastEntry)
	}
}

func (h varsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	h(res, req, vars)
}

func (h gzipHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h(res, req)
}
