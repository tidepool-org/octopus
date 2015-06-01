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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/tidepool-org/go-common/clients/status"

	"../model"
)

const (
	ERROR_READING_QUERY     = "There was an issue trying to build the query to run"
	ERROR_GETTING_UPLOAD_ID = "userid not found"
)

//givenId could be the actual id or the users email address which we also treat as an id
func (a *Api) getUserPairId(givenId, token string) (string, error) {

	if usr, err := a.ShorelineClient.GetUser(givenId, token); err != nil {
		log.Printf("getUserPairId: error [%s] getting user id [%s]", err.Error(), givenId)
		return "", &status.StatusError{status.NewStatus(http.StatusBadRequest, ERROR_GETTING_UPLOAD_ID)}
	} else {
		if pair := a.SeagullClient.GetPrivatePair(usr.UserID, "uploads", a.ShorelineClient.TokenProvide()); pair == nil {
			log.Printf("getUserPairId: error GetPrivatePair")
			return "", &status.StatusError{status.NewStatus(http.StatusBadRequest, ERROR_GETTING_UPLOAD_ID)}
		} else {
			return pair.ID, nil
		}
	}
}

// http.StatusOK
// http.StatusBadRequest
// http.StatusUnauthorized
func (a *Api) Query(res http.ResponseWriter, req *http.Request) {

	start := time.Now()

	if a.authorized(req) {

		log.Print("Query: starting ... ")

		defer req.Body.Close()
		if rawQuery, err := ioutil.ReadAll(req.Body); err != nil || string(rawQuery) == "" {
			log.Printf("Query: err decoding nonempty response body: [%v]\n [%v]\n", err, req.Body)
			statusErr := &status.StatusError{status.NewStatus(http.StatusBadRequest, ERROR_READING_QUERY)}
			a.sendModelAsResWithStatus(res, statusErr, http.StatusBadRequest)
			return
		} else {
			query := string(rawQuery)

			log.Printf("Query: to execute [%s] ", query)

			if errs, qd := model.BuildQuery(query); len(errs) != 0 {

				log.Printf("Query: errors [%v] found parsing raw query [%s]", errs, query)
				log.Println("Query: failed after [", time.Now().Sub(start).Seconds(), "] secs")

				statusErr := &status.StatusError{status.NewStatus(http.StatusBadRequest, fmt.Sprintf("Errors building query: [%v]", errs))}
				a.sendModelAsResWithStatus(res, statusErr, http.StatusBadRequest)
				return

			} else {

				if pairId, err := a.getUserPairId(qd.GetMetaQueryId(), a.getToken(req)); err != nil {
					a.sendModelAsResWithStatus(res, err, http.StatusBadRequest)
					return
				} else {
					qd.SetMetaQueryId(pairId)
				}

				log.Printf("Query: data used [%v]", qd)

				result := a.Store.ExecuteQuery(qd)

				log.Println("Query: completed in [", time.Now().Sub(start).Seconds(), "] secs")

				res.WriteHeader(http.StatusOK)
				res.Write(result)
				return
			}
		}
	}
	log.Print("Query: failed authorization")
	res.WriteHeader(http.StatusUnauthorized)
	return
}
