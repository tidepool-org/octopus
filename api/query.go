package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tidepool-org/go-common/clients/status"

	"../model"
)

const (
	ERROR_READING_QUERY     = "There was an issue trying to build the query to run"
	ERROR_GETTING_UPLOAD_ID = "userid not found"
)

// http.StatusOK
// http.StatusBadRequest
// http.StatusUnauthorized
func (a *Api) Query(res http.ResponseWriter, req *http.Request) {

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

				statusErr := &status.StatusError{status.NewStatus(http.StatusBadRequest, fmt.Sprintf("Errors building query: [%v]", errs))}
				a.sendModelAsResWithStatus(res, statusErr, http.StatusBadRequest)
				return

			} else {

				if pair := a.SeagullClient.GetPrivatePair(qd.MetaQuery["userid"], "uploads", a.ShorelineClient.TokenProvide()); pair == nil {
					statusErr := &status.StatusError{status.NewStatus(http.StatusBadRequest, ERROR_GETTING_UPLOAD_ID)}
					a.sendModelAsResWithStatus(res, statusErr, http.StatusBadRequest)
					return
				} else {
					qd.MetaQuery["userid"] = pair.ID
				}

				log.Printf("Query: data used [%v]", qd)

				result := a.Store.ExecuteQuery(qd)

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
