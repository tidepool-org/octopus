package model

import (
	"errors"
	"log"
	"strings"
)

const (
	ERROR_METAQUERY_REQUIRED = "Missing required METAQUERY e.g. METAQUERY WHERE userid IS 12d7bc90"
	ERROR_SORT_REQUIRED      = "Missing required SORT BY e.g. SORT BY time"
	ERROR_SORT_AS_REQUIRED   = "Missing required AS e.g. SORT BY time AS Timestamp"
	ERROR_TYPES_REQUIRED     = "Missing required TYPE IN e.g. TYPE IN cbg, smbg"

	KEYWORD_WHERE   = "where"
	KEYWORD_AND     = "and"
	KEYWORD_IS      = "is"
	KEYWORD_TYPE_IN = " type in "
	KEYWORD_SORT    = " sort by "
	KEYWORD_AS      = " as "
	KEYWORD_REVERSE = "reversed"
)

type (
	QueryData struct {
		MetaQuery      map[string]string
		WhereConditons []WhereCondition
		Types          []string
		Sort           map[string]string
		Reverse        bool
	}
	WhereCondition struct {
		Name      string
		Value     string
		Condition string
	}
)

//iterate over items and look for the key word
func keywordIndex(keyword string, list []string) int {
	for i, b := range list {
		if b == keyword {
			return i
		}
	}
	return 0
}

//e.g. METAQUERY WHERE userid IS 12d7bc90fa QUERY TYPE
func (qd *QueryData) buildMetaQuery(raw string) error {

	metaQuery := strings.Split(strings.ToLower(raw), " query")[0]

	metaQueryParts := strings.Fields(metaQuery)

	if len(metaQueryParts) != 5 ||
		strings.TrimSpace(metaQueryParts[1]) != KEYWORD_WHERE ||
		strings.TrimSpace(metaQueryParts[3]) != KEYWORD_IS {

		log.Printf("buildMetaQuery from %v gives error [%s]", metaQueryParts, ERROR_METAQUERY_REQUIRED)
		return errors.New(ERROR_METAQUERY_REQUIRED)
	}

	qd.MetaQuery = map[string]string{strings.TrimSpace(metaQueryParts[2]): strings.TrimSpace(metaQueryParts[4])}

	return nil
}

//e.g. WHERE time > starttime AND time < endtime
func (qd *QueryData) buildWhere(raw string) error {

	if strings.Contains(strings.ToLower(raw), KEYWORD_TYPE_IN) {
		//just get after the query ignoring the rest
		query := strings.Split(strings.ToLower(raw), "query type in")[1]
		//do we have a where statement?
		if strings.Contains(strings.ToLower(query), KEYWORD_WHERE) {

			queryParts := strings.Fields(query)
			whereIndex := keywordIndex(KEYWORD_WHERE, queryParts)
			log.Print("buildWhere adding initial where condition")

			qd.WhereConditons = append(qd.WhereConditons,
				WhereCondition{
					Name:      queryParts[whereIndex+1],
					Condition: queryParts[whereIndex+2],
					Value:     strings.ToUpper(queryParts[whereIndex+3]),
				})

			//do we also have an and?
			if andIndex := keywordIndex(KEYWORD_AND, queryParts); andIndex != 0 {
				log.Print("buildWhere appending where condition")
				qd.WhereConditons = append(qd.WhereConditons,
					WhereCondition{
						Name:      queryParts[andIndex+1],
						Condition: queryParts[andIndex+2],
						Value:     strings.ToUpper(queryParts[andIndex+3]),
					})
			}
		}

	}
	return nil
}

//e.g. TYPE IN update, cbg, smbg
func (qd *QueryData) buildTypes(raw string) error {

	if strings.Index(strings.ToLower(raw), KEYWORD_TYPE_IN) != -1 {
		typesIn := strings.Split(strings.ToLower(raw), KEYWORD_TYPE_IN)[1]

		if typesIn != "" {
			//now just get each individual type
			endsAt := KEYWORD_WHERE
			if strings.Index(typesIn, KEYWORD_WHERE) == -1 {
				endsAt = KEYWORD_SORT
			}

			typesIn = strings.Split(typesIn, endsAt)[0]

			typeParts := strings.Split(typesIn, ",")

			for i := range typeParts {
				qd.Types = append(qd.Types, strings.TrimSpace(typeParts[i]))
			}
			return nil
		}
	}
	log.Printf("buildTypes [%s] gives error [%s]", raw, ERROR_TYPES_REQUIRED)
	return errors.New(ERROR_TYPES_REQUIRED)
}

func (qd *QueryData) buildSort(raw string) error {
	if containsSort := strings.Index(strings.ToLower(raw), KEYWORD_SORT); containsSort != -1 {

		if containsSortAs := strings.Index(strings.ToLower(raw), KEYWORD_AS); containsSortAs != -1 {

			sortFieldName := strings.TrimSpace(raw[containsSort+len(KEYWORD_SORT) : containsSortAs])
			if sortEnd := strings.Index(strings.ToLower(raw), KEYWORD_REVERSE); sortEnd != -1 {

				sortAsValue := strings.TrimSpace(raw[containsSortAs+len(KEYWORD_AS) : sortEnd])

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

	if strings.Index(strings.ToLower(raw), KEYWORD_REVERSE) != -1 {
		qd.Reverse = true
	} else {
		qd.Reverse = false
	}
	return
}

//e.g. "METAQUERY WHERE userid IS 12d7bc90fa QUERY TYPE IN update WHERE time > starttime AND time < endtime SORT BY time AS Timestamp REVERSED",
func ExtractQuery(raw string) (parseErrs []error, qd *QueryData) {

	qd = &QueryData{}

	if metaErr := qd.buildMetaQuery(raw); metaErr != nil {
		parseErrs = append(parseErrs, metaErr)
	}
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
