package api

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	QUERY_NOT_PARSED = "Query could not be parsed"
)

// http.StatusOK
// http.StatusNotAcceptable
func (a *Api) Query(res http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()
	var raw interface{}
	if err := json.NewDecoder(req.Body).Decode(&raw); err != nil {
		log.Printf("Query: error decoding query to parse %v\n", err)
	}

	log.Printf("req ", raw)

	res.WriteHeader(http.StatusNotImplemented)
	return
}
