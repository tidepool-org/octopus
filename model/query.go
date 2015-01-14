package model

import (
	"errors"
	"log"
	"regexp"
	"strings"
)

const (
	ERROR_METAQUERY_REQUIRED = "Missing required METAQUERY e.g. METAQUERY WHERE userid IS 12d7bc90"
	ERROR_SORT_REQUIRED      = "Missing required SORT BY e.g. SORT BY time AS Timestamp"
	ERROR_TYPES_REQUIRED     = "Missing required TYPE IN e.g. TYPE IN cbg, smbg"
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

func (qd *QueryData) buildMetaQuery(raw string) error {
	meta := regexp.MustCompile(`(?i)\bMETAQUERY WHERE (.*) IS (.*) \bQUERY`)
	data := meta.FindStringSubmatch(raw)

	if len(data) == 3 {
		qd.MetaQuery = map[string]string{data[1]: data[2]}
		return nil
	}

	log.Printf("buildMetaQuery from %v gives error [%s]", data, ERROR_METAQUERY_REQUIRED)
	return errors.New(ERROR_METAQUERY_REQUIRED)
}

func (qd *QueryData) buildTypes(raw string) error {
	typesMatch := regexp.MustCompile(`(?i)\bQUERY TYPE IN (.*) \bWHERE`)
	typesData := typesMatch.FindStringSubmatch(raw)

	if len(typesData) != 2 {
		typesMatch = regexp.MustCompile(`(?i)\bQUERY TYPE IN (.*) \bSORT`)
		typesData = typesMatch.FindStringSubmatch(raw)
	}

	if len(typesData) == 2 {
		types := strings.Split(typesData[1], ",")

		for i := range types {
			qd.Types = append(qd.Types, strings.TrimSpace(types[i]))
		}
		return nil
	}

	log.Printf("buildTypes from %v gives error [%s]", raw, ERROR_TYPES_REQUIRED)
	return errors.New(ERROR_TYPES_REQUIRED)
}

func (qd *QueryData) buildSort(raw string) error {
	sortMatch := regexp.MustCompile(`(?i)\bSORT BY (.*) AS (.*) REVERSED\b`)
	sortData := sortMatch.FindStringSubmatch(raw)

	if len(sortData) != 3 {
		sortMatch = regexp.MustCompile(`(?i)\bSORT BY (.*) AS (.*)`)
		sortData = sortMatch.FindStringSubmatch(raw)
	}

	if len(sortData) == 3 {
		qd.Sort = map[string]string{strings.TrimSpace(sortData[1]): strings.TrimSpace(sortData[2])}
		return nil
	}

	log.Printf("buildSort from %v gives error [%s]", raw, ERROR_SORT_REQUIRED)
	return errors.New(ERROR_SORT_REQUIRED)
}

func (qd *QueryData) buildWhere(raw string) {
	where := regexp.MustCompile(`(?i)(?:^METAQUERY.+)?QUERY.+\bWHERE (.*) (.*) (.*) AND (.*) (.*) (.*) \bSORT`)
	whereData := where.FindStringSubmatch(raw)

	if len(whereData) == 7 {

		qd.WhereConditons = append(qd.WhereConditons,
			WhereCondition{
				Name:      whereData[1],
				Condition: whereData[2],
				Value:     whereData[3],
			}, WhereCondition{
				Name:      whereData[4],
				Condition: whereData[5],
				Value:     whereData[6],
			})

		return
	} else {

		where = regexp.MustCompile(`(?i)(?:^METAQUERY.+)?QUERY.+\bWHERE (.*) (.*) (.*) \bSORT`)
		whereData = where.FindStringSubmatch(raw)

		if len(whereData) == 4 {
			qd.WhereConditons = append(qd.WhereConditons,
				WhereCondition{
					Name:      whereData[1],
					Condition: whereData[2],
					Value:     whereData[3],
				})
			return
		}
	}

	log.Printf("buildWhere from [%s] shows no where clause", raw)
}

func (qd *QueryData) buildOrder(raw string) {

	qd.Reverse = false

	if strings.Index(strings.ToLower(raw), "reverse") != -1 {
		qd.Reverse = true
	}

	return
}

func BuildQuery(raw string) (parseErrs []error, qd *QueryData) {

	qd = &QueryData{}

	//METAQUERY WHERE userid IS 12d7bc90fa
	//QUERY TYPE IN smbg, cbg
	//WHERE time > starttime AND time < endtime
	//SORT BY time AS Timestamp
	//REVERSED

	if metaErr := qd.buildMetaQuery(raw); metaErr != nil {
		parseErrs = append(parseErrs, metaErr)
	}
	if typeErr := qd.buildTypes(raw); typeErr != nil {
		parseErrs = append(parseErrs, typeErr)
	}
	if sortErr := qd.buildSort(raw); sortErr != nil {
		parseErrs = append(parseErrs, sortErr)
	}
	qd.buildWhere(raw)
	qd.buildOrder(raw)

	return parseErrs, qd
}
