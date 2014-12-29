package api

import (
	"io/ioutil"
	"log"
	"net/http"

	"../model"
)

const (
	QUERY_NOT_PARSED = "Query could not be parsed"
)

// http.StatusOK
// http.StatusBadRequest
func (a *Api) Query(res http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()
	if rawQuery, err := ioutil.ReadAll(req.Body); err != nil {
		log.Printf("Query - err decoding nonempty response body: [%v]\n [%v]\n", err, req.Body)
		res.WriteHeader(http.StatusBadRequest)
		return
	} else {
		query := string(rawQuery)
		log.Printf("Query - raw [%s] ", query)

		if errs, qd := model.ExtractQuery(query); len(errs) != 0 {

			log.Printf("Query - errors [%v] found parsing raw query [%s]", errs, query)
			res.WriteHeader(http.StatusBadRequest)
			return

		} else {

			log.Printf("Query data used [%v]", qd)

			result := a.Store.ExecuteQuery(qd)

			log.Printf("Query results [%s]", string(result))

			res.WriteHeader(http.StatusOK)
			res.Write(result)
			return
		}
	}
}
