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

	"../model"
)

const (
	error_building_query = "There was an issue trying to build the query to run"
	error_no_userid      = "userid not found"
	error_no_permissons  = "permissons not found"
)

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

// http.StatusOK
// http.StatusBadRequest
// http.StatusUnauthorized
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
			http.Error(res, err.Error(), http.StatusInternalServerError)
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
