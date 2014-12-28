package api

import (
	"testing"
)

const (
	VALID_QUERY     = "METAQUERY WHERE userid IS 12d7bc90fa QUERY TYPE IN update, cbg, smbg SORT BY time AS Timestamp REVERSED"
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

func TestWhere_GivesError_WhenNoWhere(t *testing.T) {
	qd := &QueryData{}

	noWhereErr := qd.buildWhere("not right")

	if noWhereErr.Error() != ERROR_WHERE_REQUIRED {
		t.Fatalf("should give error when no WHERE is specified")
	}

}

func TestWhere_GivesError_WhenNoWhereIs(t *testing.T) {
	qd := &QueryData{}

	noWhereErr := qd.buildWhere("METAQUERY WHERE userid QUERY TYPE IN update")

	if noWhereErr.Error() != ERROR_WHERE_IS_REQUIRED {
		t.Fatalf("should give error when no IS  specified for WHERE clause")
	}

}

func TestWhere(t *testing.T) {

	const userid, givenId = "userid", "12d7bc90fa"

	qd := &QueryData{}

	qd.buildWhere(VALID_QUERY)

	if qd.Where[userid] == "" {
		t.Fatalf("should be a userid set")
	}

	if qd.Where[userid] != givenId {
		t.Fatalf("the user id should have been %s", givenId)
	}

}

func TestTypes_GivesError_WhenNoTypes(t *testing.T) {
	qd := &QueryData{}

	noTypesErr := qd.buildTypes("METAQUERY WHERE userid QUERY")

	if noTypesErr.Error() != ERROR_TYPES_REQUIRED {
		t.Fatalf("should give error when no TYPE IN is specified")
	}

}

func TestTypes(t *testing.T) {
	qd := &QueryData{}

	qd.buildTypes(VALID_QUERY)

	if len(qd.Types) != 3 {
		t.Fatalf("should listed the three types from query")
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
		t.Fatalf("should give error when no TYPE IN is specified")
	}

}

func TestSort_GivesError_WhenNoSortByAs(t *testing.T) {
	qd := &QueryData{}

	noSortAsErr := qd.buildSort("QUERY TYPE IN update, cbg, smbg SORT BY time")

	if noSortAsErr.Error() != ERROR_SORT_AS_REQUIRED {
		t.Fatalf("should give error when no TYPE IN is specified")
	}

}

func TestExractQueryData(t *testing.T) {

	errs, qd := extractQuery(VALID_QUERY)

	if len(errs) != 0 {
		t.Fatalf("there should be no errors")
	}

	if qd.Reverse != true {
		t.Fatalf("should be reversed")
	}

}

func TestExractQueryData_AccumulatesErrors(t *testing.T) {

	errs, _ := extractQuery("blah blah")

	if len(errs) != 3 {
		t.Fatalf("there should be errors [%d]", len(errs))
	}

}
