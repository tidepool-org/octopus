package api

import (
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

func (qd *QueryData) buildWhere(raw string) error {
	if containsWhere := strings.Index(strings.ToUpper(raw), whereClause); containsWhere != -1 {

		if containsWhereIs := strings.Index(strings.ToUpper(raw), whereIsClause); containsWhereIs != -1 {

			whereFieldName := strings.TrimSpace(raw[containsWhere+len(whereClause) : containsWhereIs])
			queryStart := strings.Index(strings.ToUpper(raw), " QUERY")
			whereFieldValue := strings.TrimSpace(raw[containsWhereIs+len(whereIsClause) : queryStart])

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
	if containsTypes := strings.Index(strings.ToUpper(raw), typeInClause); containsTypes != -1 {

		typesString := strings.TrimSpace(raw[containsTypes+len(typeInClause) : strings.Index(strings.ToUpper(raw), sortByClause)])

		log.Printf("buildTypes %v", strings.Split(typesString, ", "))

		qd.Types = strings.Split(typesString, ", ")

		return nil
	} else {
		log.Printf("buildTypes [%s] gives error [%s]", raw, ERROR_TYPES_REQUIRED)
		return errors.New(ERROR_TYPES_REQUIRED)
	}
}

func (qd *QueryData) buildSort(raw string) error {
	if containsSort := strings.Index(strings.ToUpper(raw), sortByClause); containsSort != -1 {

		if containsSortAs := strings.Index(strings.ToUpper(raw), sortByAsClause); containsSortAs != -1 {

			sortFieldName := strings.TrimSpace(raw[containsSort+len(sortByClause) : containsSortAs])
			if sortEnd := strings.Index(strings.ToUpper(raw), reverseClause); sortEnd != -1 {

				sortAsValue := strings.TrimSpace(raw[containsSortAs+len(sortByAsClause) : sortEnd])

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

	if contains := strings.Index(strings.ToUpper(raw), reverseClause); contains != -1 {
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

	res.WriteHeader(http.StatusNotImplemented)
	return
}
