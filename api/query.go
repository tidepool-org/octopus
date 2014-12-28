package api

import (
	"errors"
	"log"
	"net/http"
	"strings"
)

const (
	QUERY_NOT_PARSED = ""

	ERROR_WHERE_REQUIRED    = "Missing required WHERE e.g. WHERE userid IS 12d7bc90"
	ERROR_WHERE_IS_REQUIRED = "Missing required IS e.g. WHERE userid IS 12d7bc90"
	ERROR_TYPES_REQUIRED    = "Missing required TYPE IN e.g. TYPE IN cbg, smbg"

	whereClause    = "WHERE"
	whereIsClause  = "IS"
	spaceInClause  = " "
	typeInClause   = "TYPE IN"
	sortByClause   = "SORT BY"
	sortByAsClause = "AS"
	reverseClause  = "REVERSED"
)

type (
	QueryData struct {
		Where   map[string]string
		Types   []string
		Sort    map[string]string
		Reverse bool
	}
)

func (qd *QueryData) buildWhere(raw string) (error, *QueryData) {
	if containsWhere := strings.Index(strings.ToUpper(raw), whereClause); containsWhere != -1 {

		if containsWhereIs := strings.Index(strings.ToUpper(raw), whereIsClause); containsWhereIs != -1 {

			whereFieldName := strings.TrimSpace(raw[containsWhere+len(whereClause) : containsWhereIs])
			queryStart := strings.Index(strings.ToUpper(raw), " QUERY")
			whereFieldValue := strings.TrimSpace(raw[containsWhereIs+len(whereIsClause) : queryStart])

			qd.Where = map[string]string{whereFieldName: whereFieldValue}

			return nil, qd

		} else {
			log.Printf("buildWhere [%s] gives error [%s]", raw, ERROR_WHERE_IS_REQUIRED)
			return errors.New(ERROR_WHERE_IS_REQUIRED), qd
		}
	} else {
		log.Printf("buildWhere [%s] gives error [%s]", raw, ERROR_WHERE_IS_REQUIRED)
		return errors.New(ERROR_WHERE_REQUIRED), qd
	}
}

func (qd *QueryData) buildTypes(raw string) (error, *QueryData) {
	if containsTypes := strings.Index(strings.ToUpper(raw), typeInClause); containsTypes != -1 {

		typesString := strings.TrimSpace(raw[containsTypes+len(typeInClause) : strings.Index(strings.ToUpper(raw), sortByClause)])

		log.Printf("buildTypes %v", strings.Split(typesString, ", "))

		qd.Types = strings.Split(typesString, ", ")

		return nil, qd
	} else {
		log.Printf("buildTypes [%s] gives error [%s]", raw, ERROR_TYPES_REQUIRED)
		return errors.New(ERROR_TYPES_REQUIRED), qd
	}
}

func (qd *QueryData) buildSort(raw string) *QueryData {
	return qd
}

func (qd *QueryData) buildOrder(raw string) *QueryData {

	if contains := strings.Index(strings.ToUpper(raw), reverseClause); contains != -1 {
		qd.Reverse = true
	} else {
		qd.Reverse = false
	}
	return qd
}

//e.g. "METAQUERY WHERE userid IS \"12d7bc90fa\" QUERY TYPE IN update SORT BY time AS Timestamp REVERSED",
func (a *Api) extractQuery(rawQuery string) (qd QueryData) {

	//qt := template.Must(template.New("query")).Parse(queryTemplate)

	//t := template.Must(template.New("letter").Parse(letter))

	//WHERE - what e.g. userid

	//IS - userid == '1234'

	//TYPE IN - e.g. update, cbg, smbg

	//SORT BY - e.g. time

	//AS - e.g. Timestamp

	return qd
}

// http.StatusOK
// http.StatusNotAcceptable
func (a *Api) Query(res http.ResponseWriter, req *http.Request) {

	res.WriteHeader(http.StatusNotImplemented)
	return
}
