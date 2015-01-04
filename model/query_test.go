package model

import (
	"testing"
)

const (
	VALID_QUERY     = "METAQUERY WHERE userid IS 12d7bc90fa QUERY TYPE IN update, cbg, smbg WHERE time > 2015-01-01T00:00:00.000Z AND time < 2015-01-01T01:00:00.000Z SORT BY time AS Timestamp REVERSED"
	IS_REVERSE      = "blah blah reversed"
	IS_REVERSE_CASE = "blah blah REVERSED"
	NOT_REVERSE     = "blah blah"
)

func TestReverse_True(t *testing.T) {
	qd := &QueryData{}

	qd.buildOrder(IS_REVERSE)

	if qd.Reverse == false {
		t.Fatalf(" reverse should have been true")
	}

}

func TestReverse_False(t *testing.T) {
	qd := &QueryData{}

	qd.buildOrder(NOT_REVERSE)

	if qd.Reverse == true {
		t.Fatalf(" reverse should have been false")
	}

}

func TestReverse_IgnoresCase(t *testing.T) {
	qd1 := &QueryData{}
	qd2 := &QueryData{}

	qd1.buildOrder(IS_REVERSE)
	qd2.buildOrder(IS_REVERSE_CASE)

	if qd1.Reverse != qd2.Reverse {
		t.Fatalf("should have the same result as we ignore case")
	}

}

func TestMetaQuery_GivesError_WhenNoWhere(t *testing.T) {
	qd := &QueryData{}

	noWhereErr := qd.buildMetaQuery("not right")

	if noWhereErr.Error() != ERROR_METAQUERY_REQUIRED {
		t.Fatalf("got err [%s] expected err [%s]", noWhereErr.Error(), ERROR_METAQUERY_REQUIRED)
	}

}

func TestMetaQuery_GivesError_WhenNoWhereIs(t *testing.T) {
	qd := &QueryData{}

	noWhereErr := qd.buildMetaQuery("METAQUERY WHERE userid QUERY TYPE IN update")

	if noWhereErr.Error() != ERROR_METAQUERY_REQUIRED {
		t.Fatalf("got err [%s] expected err [%s]", noWhereErr.Error(), ERROR_METAQUERY_REQUIRED)
	}

}

func TestMetaQueryWhere(t *testing.T) {

	const userid, givenId = "userid", "12d7bc90fa"

	qd := &QueryData{}

	qd.buildMetaQuery(VALID_QUERY)

	if qd.MetaQuery[userid] == "" {
		t.Fatalf("should be a userid set on [%v]", qd.MetaQuery)
	}

	if qd.MetaQuery[userid] != givenId {
		t.Fatalf("the user id should have been %s, on [%v]", givenId, qd.MetaQuery)
	}

}

func TestQueryWhere_GivesNoError_WhenNoWhereStatement(t *testing.T) {
	qd := &QueryData{}

	noWhereErr := qd.buildWhere("METAQUERY WHERE userid QUERY TYPE IN update")

	if noWhereErr != nil {
		t.Fatalf("got err [%s] but expected none", noWhereErr.Error())
	}

}

func TestQueryWhere(t *testing.T) {

	//WHERE time > starttime AND time < endtime

	qd := &QueryData{}

	qd.buildWhere(VALID_QUERY)

	if len(qd.WhereConditons) != 2 {
		t.Fatalf("there should be two where conditions got %v", qd.WhereConditons)
	}

	first := qd.WhereConditons[0]
	second := qd.WhereConditons[1]

	if first.Name != "time" || first.Condition != ">" || first.Value != "2015-01-01T00:00:00.000Z" {
		t.Fatalf("first where  %v doesn't match ", first)
	}

	if second.Name != "time" || second.Condition != "<" || second.Value != "2015-01-01T01:00:00.000Z" {
		t.Fatalf("second where  %v doesn't match ", second)
	}
}

func TestQueryWhere_Gte(t *testing.T) {

	const GTE_QUERY = "QUERY TYPE IN update, cbg, smbg WHERE time >= 2015-01-01T00:00:00.000Z AND time <= 2015-01-01T01:00:00.000Z SORT BY time AS Timestamp REVERSED"

	qd := &QueryData{}

	qd.buildWhere(GTE_QUERY)

	if len(qd.WhereConditons) != 2 {
		t.Fatalf("there should be two where conditions got %v", qd.WhereConditons)
	}

	first := qd.WhereConditons[0]
	second := qd.WhereConditons[1]

	if first.Name != "time" || first.Condition != ">=" || first.Value != "2015-01-01T00:00:00.000Z" {
		t.Fatalf("first where  %v doesn't match ", first)
	}

	if second.Name != "time" || second.Condition != "<=" || second.Value != "2015-01-01T01:00:00.000Z" {
		t.Fatalf("second where  %v doesn't match ", second)
	}
}

func TestTypes_GivesError_WhenNoTypes(t *testing.T) {
	qd := &QueryData{}

	noTypesErr := qd.buildTypes("METAQUERY WHERE userid QUERY")

	if noTypesErr.Error() != ERROR_TYPES_REQUIRED {
		t.Fatalf("got err [%s] expected err [%s]", noTypesErr.Error(), ERROR_TYPES_REQUIRED)
	}

}

func TestTypes(t *testing.T) {
	qd := &QueryData{}

	qd.buildTypes(VALID_QUERY)

	if len(qd.Types) != 3 {
		t.Fatalf("should listed the three types from query got [%v]", qd.Types)
	}

	if qd.Types[0] != "update" {
		t.Fatalf("type should have been update but [%s]", qd.Types[0])
	}

	if qd.Types[1] != "cbg" {
		t.Fatalf("type should have been cbg but [%s]", qd.Types[1])
	}

	if qd.Types[2] != "smbg" {
		t.Fatalf("type should have been smbg but [%s]", qd.Types[2])
	}

}

func TestSort_GivesError_WhenNoSortBy(t *testing.T) {
	qd := &QueryData{}

	noSortErr := qd.buildSort("QUERY TYPE IN update, cbg, smbg AS Timestamp REVERSED")

	if noSortErr.Error() != ERROR_SORT_REQUIRED {
		t.Fatalf("got err [%s] expected err [%s]", noSortErr.Error(), ERROR_SORT_REQUIRED)
	}

}

func TestSort_GivesError_WhenNoSortByAs(t *testing.T) {
	qd := &QueryData{}

	noSortAsErr := qd.buildSort("QUERY TYPE IN update, cbg, smbg SORT BY time")

	if noSortAsErr.Error() != ERROR_SORT_AS_REQUIRED {
		t.Fatalf("got err [%s] expected err [%s]", noSortAsErr.Error(), ERROR_SORT_AS_REQUIRED)
	}

}

func TestExractQueryData(t *testing.T) {

	errs, qd := ExtractQuery(VALID_QUERY)

	if len(errs) != 0 {
		t.Fatalf("there should be no errors but got %v", errs)
	}

	if qd.Reverse != true {
		t.Fatalf("should be reversed")
	}

}

func TestExractQueryData_AccumulatesErrors(t *testing.T) {

	errs, _ := ExtractQuery("blah blah")

	if len(errs) != 3 {
		t.Fatalf("there should be errors [%d]", len(errs))
	}

}
