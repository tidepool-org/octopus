/*
== BSD2 LICENSE ==
Copyright (c) 2015, Tidepool Project

This program is free software; you can redistribute it and/or modify it under
the terms of the associated License, which is identical to the BSD 2-Clause
License as published by the Open Source Initiative at opensource.org.

This program is distributed in the hope that it will be useful, but WITHOUT
ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
FOR A PARTICULAR PURPOSE. See the License for more details.

You should have received a copy of the License along with this program; if
not, you can obtain one from Tidepool Project at tidepool.org.
== BSD2 LICENSE ==
*/

package model

import (
	"testing"
)

const (
	QUERY_WHERE_AND  = "METAQUERY WHERE userid IS 12d7bc90fa QUERY TYPE IN update, cbg, smbg WHERE time > 2015-01-01T00:00:00.000Z AND time < 2015-01-01T01:00:00.000Z"
	QUERY_WHERE      = "METAQUERY WHERE userid IS 12d7bc90fa QUERY TYPE IN update, cbg, smbg WHERE time >= 2015-01-01T00:00:00.000Z"
	METAQUERY_EMAILS = "METAQUERY WHERE emails CONTAINS foo@bar.com QUERY TYPE IN update, cbg, smbg WHERE time >= 2015-01-01T00:00:00.000Z"
	QUERY_WHERE_IN   = "METAQUERY WHERE userid IS 12d7bc90fa QUERY TYPE IN cbg WHERE updateId NOT IN abcd, efgh, ijkl"
	IS_REVERSE       = "blah blah reversed"
	IS_REVERSE_CASE  = "blah blah REVERSED"
	NOT_REVERSE      = "blah blah"
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

	const givenId = "12d7bc90fa"

	qd := &QueryData{}

	qd.buildMetaQuery(QUERY_WHERE)

	if qd.GetMetaQueryId() == "" {
		t.Fatalf("should be a userid set on [%v]", qd.MetaQuery)
	}

	if qd.GetMetaQueryId() != givenId {
		t.Fatalf("the user id should have been %s, on [%v]", givenId, qd.MetaQuery)
	}

}

func TestMetaQueryWhere_Emails(t *testing.T) {

	const givenEmail = "foo@bar.com"

	qd := &QueryData{}

	qd.buildMetaQuery(METAQUERY_EMAILS)

	if qd.GetMetaQueryId() == "" {
		t.Fatalf("should be a emails set on [%v]", qd.MetaQuery)
	}

	if qd.GetMetaQueryId() != givenEmail {
		t.Fatalf("the email should have been %s, on [%v]", givenEmail, qd.MetaQuery)
	}

}

func TestMetaQueryWhere_Bad(t *testing.T) {

	const METAQUERY_BAD = "METAQUERY WHERE bad IS wrong QUERY TYPE IN update, cbg, smbg WHERE time >= 2015-01-01T00:00:00.000Z"

	qd := &QueryData{}

	if err := qd.buildMetaQuery(METAQUERY_BAD); err == nil {
		t.Fatalf("the meta query [%s] was badly formed and should have given an error", qd.MetaQuery)
	}

}

func TestQueryWhereAnd(t *testing.T) {

	//WHERE time > starttime AND time < endtime

	qd := &QueryData{}

	qd.buildTimeWhere(QUERY_WHERE_AND)

	t.Logf("%v", qd)

	if len(qd.WhereConditions) != 2 {
		t.Fatalf("there should be two where conditions got %v", qd.WhereConditions)
	}

	first := qd.WhereConditions[0]
	second := qd.WhereConditions[1]

	if first.Name != "time" || first.Condition != ">" || first.Value != "2015-01-01T00:00:00.000Z" {
		t.Fatalf("first where  %v doesn't match ", first)
	}

	if second.Name != "time" || second.Condition != "<" || second.Value != "2015-01-01T01:00:00.000Z" {
		t.Fatalf("second where  %v doesn't match ", second)
	}
}

func TestQueryWhere_WithGte(t *testing.T) {

	qd := &QueryData{}

	qd.buildTimeWhere(QUERY_WHERE)

	if len(qd.WhereConditions) != 1 {
		t.Fatalf("there should be two where conditions got %v", qd.WhereConditions)
	}

	first := qd.WhereConditions[0]

	if first.Name != "time" || first.Condition != ">=" || first.Value != "2015-01-01T00:00:00.000Z" {
		t.Fatalf("first where  %v doesn't match ", first)
	}

}

func TestQueryWhereIn(t *testing.T) {

	qd := &QueryData{}

	qd.buildInWhere(QUERY_WHERE_IN)

	if len(qd.WhereConditions) != 1 {
		t.Fatalf("there should be two where conditions got %v", qd.WhereConditions)
	}

	first := qd.WhereConditions[0]

	if first.Name != "updateId" {
		t.Fatalf("name [%s] doesn't match [updateId]", first.Name)
	}
	if first.Condition != "NOT IN" {
		t.Fatalf("condition [%s] doesn't match [NOT IN]", first.Condition)
	}
	if first.Value != "NOT USED" {
		t.Fatalf("value [%s] doesn't match [NOT USED] ", first.Value)
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

	qd.buildTypes(QUERY_WHERE_AND)

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

func TestBuildQuery_WithWhereAnd(t *testing.T) {

	//lets test it all

	errs, qd := BuildQuery(QUERY_WHERE_AND)

	if len(errs) != 0 {
		t.Fatalf("there should be no errors but got %v", errs)
	}

	if qd.MetaQuery[ANYID] != "12d7bc90fa" {
		t.Fatalf("userid should be 12d7bc90fa but %v", qd.MetaQuery)
	}

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

	if len(qd.WhereConditions) != 2 {
		t.Fatalf("there should be 2 conditions but got [%d]", len(qd.WhereConditions))
	}

	if qd.WhereConditions[0].Name != "time" || qd.WhereConditions[0].Condition != ">" || qd.WhereConditions[0].Value != "2015-01-01T00:00:00.000Z" {
		t.Fatalf("first where  %v doesn't match ", qd.WhereConditions[0])
	}

	if qd.WhereConditions[1].Name != "time" || qd.WhereConditions[1].Condition != "<" || qd.WhereConditions[1].Value != "2015-01-01T01:00:00.000Z" {
		t.Fatalf("second where  %v doesn't match ", qd.WhereConditions[1])
	}

}

func TestBuildQuery_WithWhere(t *testing.T) {

	//lets test it all

	errs, qd := BuildQuery(QUERY_WHERE)

	if len(errs) != 0 {
		t.Fatalf("there should be no errors but got %v", errs)
	}

	if qd.GetMetaQueryId() != "12d7bc90fa" {
		t.Fatalf("userid should be 12d7bc90fa but %v", qd.MetaQuery)
	}

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

	if len(qd.WhereConditions) != 1 {
		t.Fatalf("there should be 1 conditions but got [%d]", len(qd.WhereConditions))
	}

	first := qd.WhereConditions[0]

	if first.Name != "time" || first.Condition != ">=" || first.Value != "2015-01-01T00:00:00.000Z" {
		t.Fatalf("first where  %v doesn't match ", first)
	}

}

func TestBuildQuery_WithWhereIn(t *testing.T) {

	errs, qd := BuildQuery(QUERY_WHERE_IN)

	if len(errs) != 0 {
		t.Fatalf("there should be no errors but got %v", errs)
	}

	if qd.GetMetaQueryId() != "12d7bc90fa" {
		t.Fatalf("userid should be 12d7bc90fa but %v", qd.MetaQuery)
	}

	if len(qd.Types) != 1 {
		t.Fatalf("should listed the one types from query got [%v]", qd.Types)
	}

	if qd.Types[0] != "cbg" {
		t.Fatalf("type should have been cbg but [%s]", qd.Types[0])
	}

	if len(qd.WhereConditions) != 1 {
		t.Fatalf("there should be 1 conditions but got [%d]", len(qd.WhereConditions))
	}

	first := qd.WhereConditions[0]

	if first.Name != "updateId" {
		t.Fatalf("name [%s] doesn't match [updateId]", first.Name)
	}
	if first.Condition != "NOT IN" {
		t.Fatalf("condition [%s] doesn't match [NOT IN]", first.Condition)
	}
	if first.Value != "NOT USED" {
		t.Fatalf("value [%s] doesn't match [NOT USED] ", first.Value)
	}

}

func TestExtractQueryData_AccumulatesErrors(t *testing.T) {

	errs, _ := BuildQuery("blah blah")

	if len(errs) != 2 {
		t.Fatalf("there should be TWO errors but got [%d]", len(errs))
	}

}
