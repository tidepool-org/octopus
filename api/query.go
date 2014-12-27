package api

import (
	"net/http"
)

const (
	QUERY_NOT_PARSED = ""
)

// http.StatusOK
// http.StatusNotAcceptable
func (a *Api) Query(res http.ResponseWriter, req *http.Request) {

	res.WriteHeader(http.StatusNotImplemented)
	return
}
