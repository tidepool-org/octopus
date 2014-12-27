package api

import (
	"testing"
)

const (
	VALID_QUERY     = "METAQUERY WHERE userid IS \"12d7bc90fa\" QUERY TYPE IN update SORT BY time AS Timestamp REVERSED"
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

	noWhereErr, qd := qd.buildWhere("not right")

	if noWhereErr.Error() != ERROR_WHERE_REQUIRED {
		t.Fatalf("should give error when no WHERE is specified")
	}

}

func TestWhere_GivesError_WhenNoWhereIs(t *testing.T) {
	qd := &QueryData{}

	noWhereErr, qd := qd.buildWhere("METAQUERY WHERE userid QUERY TYPE IN update")

	if noWhereErr.Error() != ERROR_WHERE_IS_REQUIRED {
		t.Fatalf("should give error when no IS  specified for WHERE clause")
	}

}
