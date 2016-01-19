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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	httpgzip "github.com/daaku/go.httpgzip"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	commonClients "github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/clients/shoreline"

	"./../clients"
	"./../model"
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
	}

	ShorelineInterface interface {
		CheckToken(token string) *shoreline.TokenData
		GetUser(userID, token string) (*shoreline.UserData, error)
		TokenProvide() string
	}

	GatekeeperInterface interface {
		UserInGroup(userID, groupID string) (map[string]commonClients.Permissions, error)
	}

	SeagullInterface interface {
		GetPrivatePair(userID, hashName, token string) *commonClients.PrivatePair
	}

	// so we can wrap and marshal the detailed error
	detailedError struct {
		Status          int    `json:"status"`
		Id              string `json:"id"`
		Code            string `json:"code"`
		Message         string `json:"message"`
		InternalMessage string `json:"-"` //used only for logging so we don't want to serialize it out
	}

	httpVars    map[string]string
	varsHandler func(http.ResponseWriter, *http.Request, httpVars)
	gzipHandler func(http.ResponseWriter, *http.Request)
)

var (
	error_no_userid          = &detailedError{Status: http.StatusBadRequest, Code: "query_userid_notfound", Message: "userid not found"}
	error_getting_permissons = &detailedError{Status: http.StatusBadRequest, Code: "query_permissons_notfound", Message: "user does not have any permissons"}
	error_no_view_permisson  = &detailedError{Status: http.StatusForbidden, Code: "query_cant_view", Message: "user does not have permisson to view data"}
	error_not_authorized     = &detailedError{Status: http.StatusUnauthorized, Code: "query_not_authorized", Message: "user is not authorized"}
	error_building_query     = &detailedError{Status: http.StatusBadRequest, Code: "query_invalid_data", Message: "error building your query"}

	//generic server errors
	error_internal_server = &detailedError{Status: http.StatusInternalServerError, Code: "query_intenal_error", Message: "internal server error"}
	error_running_query   = &detailedError{Status: http.StatusInternalServerError, Code: "query_store_error", Message: "internal server error"}
	error_status_check    = &detailedError{Status: http.StatusInternalServerError, Code: "query_status_check", Message: "internal server error"}
)

//set this from the actual error if applicable
func (d *detailedError) setInternalMessage(internal error) *detailedError {
	d.InternalMessage = internal.Error()
	return d
}

//log error detail and write as application/json
func jsonError(res http.ResponseWriter, err *detailedError, startedAt time.Time) {

	err.Id = uuid.NewV4().String()

	log.Println(QUERY_API_PREFIX, fmt.Sprintf("[%s][%s] failed after [%.5f]secs with error [%s][%s] ", err.Id, err.Code, time.Now().Sub(startedAt).Seconds(), err.Message, err.InternalMessage))

	jsonErr, _ := json.Marshal(err)

	res.WriteHeader(err.Status)
	res.Header().Add("content-type", "application/json")
	res.Write(jsonErr)
	return
}

//find and validate the token
func (a *Api) authorized(req *http.Request) *shoreline.TokenData {

	if token := a.getToken(req); token != "" {
		if td := a.ShorelineClient.CheckToken(token); td != nil {
			log.Println(QUERY_API_PREFIX, "token check succeeded")
			return td
		}
		log.Println(QUERY_API_PREFIX, "token check failed")
		return nil
	}
	log.Println(QUERY_API_PREFIX, "no token to check")
	return nil
}

func (a *Api) userCanViewData(userID, groupID string) bool {

	if userID == groupID {
		return true
	}

	log.Println("checking if user", userID, " can view group", groupID)

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

func InitApi(
	slc ShorelineInterface,
	sgc SeagullInterface,
	gkc GatekeeperInterface,
	store clients.StoreClient) *Api {

	return &Api{
		Store:            store,
		ShorelineClient:  slc,
		SeagullClient:    sgc,
		GatekeeperClient: gkc,
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
		jsonError(res, error_status_check.setInternalMessage(err), start)
		return
	}
	log.Println(QUERY_API_PREFIX, fmt.Sprintf("GetStatus: completed in [%.5f] secs", time.Now().Sub(start).Seconds()))
	res.Write([]byte("OK"))
	return
}

// http.StatusOK, time of last entry
// http.StatusBadRequest - something was wrong with the request data
// http.StatusUnauthorized - you don't have a valid token
// http.StatusForbidden - you have a valid token but don't have permisson to look at the data
func (a *Api) TimeLastEntryUser(res http.ResponseWriter, req *http.Request, vars httpVars) {

	start := time.Now()

	if td := a.authorized(req); td != nil {

		userId := vars["userID"]

		if a.userCanViewData(td.UserID, userId) {

			group := a.SeagullClient.GetPrivatePair(userId, "uploads", a.ShorelineClient.TokenProvide())

			if group == nil {
				jsonError(res, error_getting_permissons, start)
				return
			}

			timeLastEntry, err := a.Store.GetTimeLastEntryUser(group.ID)
			if err != nil {
				jsonError(res, error_running_query.setInternalMessage(err), start)
				return
			}
			log.Println(QUERY_API_PREFIX, fmt.Sprintf("TimeLastEntryUser: completed in [%.5f] secs", time.Now().Sub(start).Seconds()))
			res.Header().Set("content-type", "application/json")
			res.Write(timeLastEntry)
			return
		}
		jsonError(res, error_no_view_permisson, start)
		return
	}
	jsonError(res, error_not_authorized, start)
	return
}

// http.StatusOK, time of last entry and device
// http.StatusBadRequest - something was wrong with the request data
// http.StatusUnauthorized - you don't have a valid token
// http.StatusForbidden - you have a valid token but don't have permisson to look at the data
func (a *Api) TimeLastEntryUserAndDevice(res http.ResponseWriter, req *http.Request, vars httpVars) {

	start := time.Now()

	if td := a.authorized(req); td != nil {
		userId := vars["userID"]

		if a.userCanViewData(td.UserID, userId) {
			group := a.SeagullClient.GetPrivatePair(userId, "uploads", a.ShorelineClient.TokenProvide())

			if group == nil {
				jsonError(res, error_getting_permissons, start)
				return
			}
			timeLastEntry, err := a.Store.GetTimeLastEntryUserAndDevice(group.ID, vars["deviceID"])
			if err != nil {
				jsonError(res, error_running_query.setInternalMessage(err), start)
				return
			}
			log.Println(QUERY_API_PREFIX, fmt.Sprintf("TimeLastEntryUserAndDevice: completed in [%.5f] secs", time.Now().Sub(start).Seconds()))
			res.Header().Set("content-type", "application/json")
			res.Write(timeLastEntry)
			return
		}
		jsonError(res, error_no_view_permisson, start)
		return
	}
	jsonError(res, error_not_authorized, start)
	return
}

//build up valid QueryData from the request or return any detailedError that happens while trying to build it
func buildQueryFrom(req *http.Request) (*model.QueryData, *detailedError) {
	defer req.Body.Close()
	rawQuery, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return nil, error_internal_server.setInternalMessage(err)
	}
	query := string(rawQuery)
	if query == "" {
		return nil, error_building_query
	}

	log.Println(QUERY_API_PREFIX, "Query: raw ", query)

	errs, qd := model.BuildQuery(query)

	if len(errs) != 0 {
		buildError := error_building_query
		buildError.Message = fmt.Sprintf("[%s] %v", buildError.Message, errs)
		return nil, buildError
	}
	return qd, nil
}

//as `userid` from our query could infact be an email we need to resolve that and then get the associated groupId or return any detailedError
func (a *Api) getGroupForQueriedUser(req *http.Request, givenId string) (string, *detailedError) {
	resolvedUser, err := a.ShorelineClient.GetUser(givenId, a.ShorelineClient.TokenProvide())
	if err != nil {
		return "", error_no_userid.setInternalMessage(err)
	}
	group := a.SeagullClient.GetPrivatePair(resolvedUser.UserID, "uploads", a.ShorelineClient.TokenProvide())
	if group == nil {
		return "", error_getting_permissons
	}
	return group.ID, nil
}

// http.StatusOK - the requested data
// http.StatusBadRequest - something was wrong with the request data
// http.StatusUnauthorized - you don't have a valid token
func (a *Api) Query(res http.ResponseWriter, req *http.Request) {

	start := time.Now()

	if td := a.authorized(req); td != nil {

		log.Println(QUERY_API_PREFIX, "Query: starting ... ")

		//build the query
		qd, detailedErr := buildQueryFrom(req)

		if detailedErr != nil {
			jsonError(res, detailedErr, start)
			return
		}

		//find the groupId
		groupId, detailedErr := a.getGroupForQueriedUser(req, qd.GetMetaQueryId())
		if detailedErr != nil {
			jsonError(res, detailedErr, start)
			return
		}

		qd.SetMetaQueryId(groupId)

		//run the query
		result, err := a.Store.ExecuteQuery(qd)

		if err != nil {
			jsonError(res, error_running_query.setInternalMessage(err), start)
			return
		}
		// yay we made it! lets give them what they asked for
		log.Println(QUERY_API_PREFIX, fmt.Sprintf("Query: completed in [%.5f] secs", time.Now().Sub(start).Seconds()))
		res.Header().Set("content-type", "application/json")
		res.Write(result)
		return

	}
	jsonError(res, error_not_authorized, start)
	return
}

func (h varsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	h(res, req, vars)
}

func (h gzipHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h(res, req)
}
