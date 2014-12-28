package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

const (
	QUERY_NOT_PARSED = "Query could not be parsed"

	ERROR_WHERE_REQUIRED    = "Missing required WHERE e.g. WHERE userid IS 12d7bc90"
	ERROR_WHERE_IS_REQUIRED = "Missing required IS e.g. WHERE userid IS 12d7bc90"
	ERROR_SORT_REQUIRED     = "Missing required SORT BY e.g. SORT BY time"
	ERROR_SORT_AS_REQUIRED  = "Missing required AS e.g. SORT BY time AS Timestamp"
	ERROR_TYPES_REQUIRED    = "Missing required TYPE IN e.g. TYPE IN cbg, smbg"

	CLAUSE_WHERE    = "WHERE"
	CLAUSE_WHERE_IS = "IS"
	CLAUSE_TYPE_IN  = "TYPE IN"
	CLAUSE_SORT     = "SORT BY"
	CLAUSE_SORT_AS  = "AS"
	CLAUSE_REVERSE  = "REVERSED"
)

type (
	QueryData struct {
		Where   map[string]string
		Types   []string
		Sort    map[string]string
		Reverse bool
	}
)

func (qd *QueryData) buildWhere(raw string) error {
	if containsWhere := strings.Index(strings.ToUpper(raw), CLAUSE_WHERE); containsWhere != -1 {

		if containsWhereIs := strings.Index(strings.ToUpper(raw), CLAUSE_WHERE_IS); containsWhereIs != -1 {

			whereFieldName := strings.TrimSpace(raw[containsWhere+len(CLAUSE_WHERE) : containsWhereIs])
			queryStart := strings.Index(strings.ToUpper(raw), " QUERY")
			whereFieldValue := strings.TrimSpace(raw[containsWhereIs+len(CLAUSE_WHERE_IS) : queryStart])

			qd.Where = map[string]string{whereFieldName: whereFieldValue}

			return nil

		} else {
			log.Printf("buildWhere [%s] gives error [%s]", raw, ERROR_WHERE_IS_REQUIRED)
			return errors.New(ERROR_WHERE_IS_REQUIRED)
		}
	} else {
		log.Printf("buildWhere [%s] gives error [%s]", raw, ERROR_WHERE_IS_REQUIRED)
		return errors.New(ERROR_WHERE_REQUIRED)
	}
}

func (qd *QueryData) buildTypes(raw string) error {
	if containsTypes := strings.Index(strings.ToUpper(raw), CLAUSE_TYPE_IN); containsTypes != -1 {

		typesString := strings.TrimSpace(raw[containsTypes+len(CLAUSE_TYPE_IN) : strings.Index(strings.ToUpper(raw), CLAUSE_SORT)])

		log.Printf("buildTypes %v", strings.Split(typesString, ", "))

		qd.Types = strings.Split(typesString, ", ")

		return nil
	} else {
		log.Printf("buildTypes [%s] gives error [%s]", raw, ERROR_TYPES_REQUIRED)
		return errors.New(ERROR_TYPES_REQUIRED)
	}
}

func (qd *QueryData) buildSort(raw string) error {
	if containsSort := strings.Index(strings.ToUpper(raw), CLAUSE_SORT); containsSort != -1 {

		if containsSortAs := strings.Index(strings.ToUpper(raw), CLAUSE_SORT_AS); containsSortAs != -1 {

			sortFieldName := strings.TrimSpace(raw[containsSort+len(CLAUSE_SORT) : containsSortAs])
			if sortEnd := strings.Index(strings.ToUpper(raw), CLAUSE_REVERSE); sortEnd != -1 {

				sortAsValue := strings.TrimSpace(raw[containsSortAs+len(CLAUSE_SORT_AS) : sortEnd])

				qd.Sort = map[string]string{sortFieldName: sortAsValue}

				return nil
			} else {
				log.Printf("buildSort [%s] gives error [%s]", raw, "no end of the sort")
				return errors.New("no end of the sort")
			}

		} else {
			log.Printf("buildSort [%s] gives error [%s]", raw, ERROR_SORT_AS_REQUIRED)
			return errors.New(ERROR_SORT_AS_REQUIRED)
		}
	} else {
		log.Printf("buildSort [%s] gives error [%s]", raw, ERROR_SORT_REQUIRED)
		return errors.New(ERROR_SORT_REQUIRED)
	}
}

func (qd *QueryData) buildOrder(raw string) {

	if contains := strings.Index(strings.ToUpper(raw), CLAUSE_REVERSE); contains != -1 {
		qd.Reverse = true
	} else {
		qd.Reverse = false
	}
	return
}

//e.g. "METAQUERY WHERE userid IS \"12d7bc90fa\" QUERY TYPE IN update SORT BY time AS Timestamp REVERSED",
func extractQuery(raw string) (parseErrs []error, qd *QueryData) {

	qd = &QueryData{}

	if whereErr := qd.buildWhere(raw); whereErr != nil {
		parseErrs = append(parseErrs, whereErr)
	}
	if typeErr := qd.buildTypes(raw); typeErr != nil {
		parseErrs = append(parseErrs, typeErr)
	}
	if sortErr := qd.buildSort(raw); sortErr != nil {
		parseErrs = append(parseErrs, sortErr)
	}
	qd.buildOrder(raw)

	return parseErrs, qd
}

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
