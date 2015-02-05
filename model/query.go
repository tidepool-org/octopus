package model

import (
	"errors"
	"log"
	"regexp"
	"strings"
)

const (
	ERROR_METAQUERY_REQUIRED = "Missing required METAQUERY e.g. METAQUERY WHERE userid IS 12d7bc90 or  METAQUERY WHERE emails CONTAINS foo@bar.org"
	ERROR_SORT_REQUIRED      = "Missing required SORT BY e.g. SORT BY time AS Timestamp"
	ERROR_TYPES_REQUIRED     = "Missing required TYPE IN e.g. TYPE IN cbg, smbg"
	INWHERE_PAT              = `(?i)\bQUERY.+\bWHERE +([^ ]*) +(?:(NOT IN|IN) +)(.*)\bSORT`
	ANYID                    = "anyid" // as an we can use either the userid or an email as an 'id' here
)

type (
	QueryData struct {
		MetaQuery       map[string]string
		WhereConditions []WhereCondition
		Types           []string
		InList          []string
		Sort            map[string]string
		Reverse         bool
	}
	WhereCondition struct {
		Name      string
		Value     string
		Condition string
	}
)

func (qd *QueryData) GetMetaQueryId() string {
	return qd.MetaQuery[ANYID]
}

func (qd *QueryData) SetMetaQueryId(anyid string) {
	qd.MetaQuery[ANYID] = anyid
}

func (qd *QueryData) buildMetaQuery(raw string) error {

	const USERID, EMAILS = "userid", "emails"

	useridMeta := regexp.MustCompile(`(?i)\bMETAQUERY WHERE (.*) IS (.*) \bQUERY`)
	useridData := useridMeta.FindStringSubmatch(raw)

	if len(useridData) == 3 && useridData[1] == USERID {
		qd.MetaQuery = map[string]string{ANYID: useridData[2]}
		return nil
	} else {
		//if it is not a userid see if its an email
		emailsMeta := regexp.MustCompile(`(?i)\bMETAQUERY WHERE (.*) CONTAINS (.*) \bQUERY`)
		emailsData := emailsMeta.FindStringSubmatch(raw)

		if len(emailsData) == 3 && emailsData[1] == EMAILS {
			qd.MetaQuery = map[string]string{ANYID: emailsData[2]}
			return nil
		}
		log.Printf("buildMetaQuery from %v gives error [%s]", emailsData, ERROR_METAQUERY_REQUIRED)
		return errors.New(ERROR_METAQUERY_REQUIRED)
	}

	log.Printf("buildMetaQuery from %v gives error [%s]", useridData, ERROR_METAQUERY_REQUIRED)
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

func (qd *QueryData) isTimeWhere(raw string) bool {
	where := regexp.MustCompile(`(?i)(?:^METAQUERY.+)?QUERY.+\bWHERE (.*) (.*) (.*) (AND (.*) (.*) (.*) )?\bSORT`)
	indices := where.FindStringIndex(raw)
	return indices != nil
}

func (qd *QueryData) isInWhere(raw string) bool {
	where := regexp.MustCompile(INWHERE_PAT)
	indices := where.FindStringIndex(raw)
	return indices != nil
}

func (qd *QueryData) buildTimeWhere(raw string) {
	where := regexp.MustCompile(`(?i)[^METAQUERY] \bWHERE (.*) (.*) (.*) AND (.*) (.*) (.*) \bSORT`)
	whereData := where.FindStringSubmatch(raw)

	if len(whereData) == 7 {

		qd.WhereConditions = append(qd.WhereConditions,
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
		where = regexp.MustCompile(`(?i)[^METAQUERY] \bWHERE (.*) (.*) (.*) \bSORT`)
		whereData = where.FindStringSubmatch(raw)

		if len(whereData) == 4 {
			qd.WhereConditions = append(qd.WhereConditions,
				WhereCondition{
					Name:      whereData[1],
					Condition: whereData[2],
					Value:     whereData[3],
				})
			return
		}
	}

	log.Printf("buildTimeWhere from [%s] shows incorrect or no where clause", raw)
}

func (qd *QueryData) buildInWhere(raw string) {
	where := regexp.MustCompile(INWHERE_PAT)
	whereData := where.FindStringSubmatch(raw)

	if len(whereData) == 4 {

		listre := regexp.MustCompile("[ ,]+")
		qd.InList = listre.Split(whereData[3], -1)

		qd.WhereConditions = append(qd.WhereConditions,
			WhereCondition{
				Name:      whereData[1],
				Condition: whereData[2],
				Value:     "NOT USED",
			})
		return
	}

	log.Printf("buildInWhere from [%s] shows incorrect or no where clause", raw)
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

	if metaErr := qd.buildMetaQuery(raw); metaErr != nil {
		parseErrs = append(parseErrs, metaErr)
	}
	if typeErr := qd.buildTypes(raw); typeErr != nil {
		parseErrs = append(parseErrs, typeErr)
	}
	if sortErr := qd.buildSort(raw); sortErr != nil {
		parseErrs = append(parseErrs, sortErr)
	}
	if qd.isInWhere(raw) {
		qd.buildInWhere(raw)
	} else if qd.isTimeWhere(raw) {
		qd.buildTimeWhere(raw)
	}
	qd.buildOrder(raw)

	return parseErrs, qd
}
