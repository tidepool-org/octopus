package api

import (
	"io/ioutil"
	"log"
	"net/http"

	"../model"
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
			res.WriteHeader(http.StatusBadRequest)
			return
		} else {
			query := string(rawQuery)

			log.Printf("Query: to execute [%s] ", query)

			if errs, qd := model.ExtractQuery(query); len(errs) != 0 {

				log.Printf("Query: errors [%v] found parsing raw query [%s]", errs, query)
				res.WriteHeader(http.StatusBadRequest)
				return

			} else {

				log.Printf("Query: data used [%v]", qd)

				result := a.Store.ExecuteQuery(qd)

				log.Printf("Query: results [%s]", string(result))

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
