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

package clients

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/tidepool-org/go-common/clients/mongo"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"../model"
)

const (
	valid_userid = "1234"

	valid_groupid  = "1234"
	valid_deviceid = "Paradigm Revel - 723-=-53571999" // there is only one match

	no_match_groupid  = "no_match"
	no_match_userid   = "no_match"
	no_match_deviceid = "Paradigm Revel - 723-=-77777777"
)

var (
	theTime  = "2014-10-23T10:00:00.000Z"
	basalsQd = &model.QueryData{
		MetaQuery:       map[string]string{"userid": valid_userid},
		WhereConditions: []model.WhereCondition{model.WhereCondition{Name: "time", Value: theTime, Condition: "<"}},
		Types:           []string{"basal"},
		Sort:            map[string]string{"time": "myTime"},
		Reverse:         false,
	}
	noDataQd = &model.QueryData{
		MetaQuery: map[string]string{"userid": no_match_userid},
		Types:     []string{"no_data"},
		Sort:      map[string]string{"time": "myTime"},
		Reverse:   false,
	}
	testingConfig = &mongo.Config{ConnectionString: "mongodb://localhost/streams_test"}
)

func initTestData() *MongoStoreClient {

	mongoSession, err := mongo.Connect(testingConfig)
	if err != nil {
		log.Fatal(err)
	}

	setupCopy := mongoSession.Copy()
	defer setupCopy.Close()

	//remove existing and start fresh
	setupCopy.DB("").C(DEVICE_DATA_COLLECTION).DropCollection()

	if err := setupCopy.DB("").C(DEVICE_DATA_COLLECTION).Create(&mgo.CollectionInfo{}); err != nil {
		log.Panic("We could not load the test data ", err.Error())
	}

	//initialize the test data
	if testData, err := ioutil.ReadFile("./test_data.json"); err == nil {

		var toLoad []interface{}

		if err := json.Unmarshal(testData, &toLoad); err != nil {
			log.Panic("We could not load the test data ", err.Error())
		}

		for i := range toLoad {
			//insert each test data item
			if insertErr := setupCopy.DB("").C(DEVICE_DATA_COLLECTION).Insert(toLoad[i]); insertErr != nil {
				log.Panic("We could not load the test data ", err.Error())
			}
		}
	}

	//return an instance of the store
	return NewMongoStoreClient(testingConfig)

}

func TestIndexes(t *testing.T) {

	const (
		//index names based on feilds used
		std_query_idx      = "_groupId_1__active_1_type_1_time_-1"
		uploadid_query_idx = "_groupId_1__active_1_type_1_uploadId_1_time_-1"
	)
	mc := initTestData()

	sCopy := mc.session
	defer sCopy.Close()

	if idxs, err := sCopy.DB("").C(DEVICE_DATA_COLLECTION).Indexes(); err != nil {
		t.Fatal("TestIndexes unexpected error ", err.Error())
	} else {
		// there are the two we have added and also the standard index
		if len(idxs) != 3 {
			t.Fatalf("TestIndexes should be 3 but found [%d] ", len(idxs))
		}

		if idxs[0].Name != std_query_idx {
			t.Fatalf("TestIndexes expected [%s] got [%s] ", std_query_idx, idxs[0].Name)
		}

		if idxs[1].Name != uploadid_query_idx {
			t.Fatalf("TestIndexes expected [%s] got [%s] ", uploadid_query_idx, idxs[1].Name)
		}
	}
}

func TestExecuteQuery(t *testing.T) {

	mc := initTestData()

	if results, err := mc.ExecuteQuery(basalsQd); err != nil {
		t.Fatalf("an error was thrown for query [%v] w error [%s]", basalsQd, err.Error())
	} else if results == nil {
		t.Fatalf("no results were found for the query [%v]", basalsQd)
	} else {

		type found map[string]interface{}

		const ISO_8601 = "2006-01-02T15:04:05Z"
		timeClause, _ := time.Parse(ISO_8601, theTime)

		records := []found{}
		json.Unmarshal(results, &records)

		//number of recoords
		if len(records) != 2 {
			t.Fatalf("we should have been given two results but got [%d]", len(records))
		}

		// test first results
		first := records[0]

		if first["_id"] != nil {
			t.Fatalf("the _id should not be set but is [%s] ", first["_id"])
		}

		if first["type"] != "basal" {
			t.Fatalf("first should be of type basal but where [%s] ", first["type"])
		}

		firstTimeIs, _ := time.Parse(ISO_8601, first["time"].(string))
		if firstTimeIs.After(timeClause) {
			t.Fatalf("first time [%v] should be before than [%v] ", firstTimeIs, timeClause)
		}
		if first["time"] != "2014-10-23T07:00:00.000Z" {
			t.Fatalf("first time [%s] should be 2014-10-23T07:00:00.000Z", first["time"])
		}

		if first["rate"] != 0.6 {
			t.Fatalf("first rate [%d] should be 0.6", first["rate"])
		}

		// test sec results
		second := records[1]

		if second["_id"] != nil {
			t.Fatalf("the _id should not be set but is [%s] ", second["_id"])
		}

		if second["type"] != "basal" {
			t.Fatalf("second should be of type basal but where [%s] ", second["type"])
		}

		secondTimeIs, _ := time.Parse(ISO_8601, second["time"].(string))
		if secondTimeIs.After(timeClause) {
			t.Fatalf(" second time [%v] should be before than [%v] ", secondTimeIs, timeClause)
		}
		if second["time"] != "2014-10-23T08:00:00.000Z" {
			t.Fatalf("second time [%s] should be 2014-10-23T08:00:00.000Z", second["time"])
		}

		if second["rate"] != 0.4 {
			t.Fatalf("second rate [%d] should be 0.4", second["rate"])
		}
	}
}

func TestExecuteQuery_NoData(t *testing.T) {

	mc := initTestData()

	if results, err := mc.ExecuteQuery(noDataQd); err != nil {
		t.Fatalf("an error was thrown for query [%v] w error [%s]", basalsQd, err.Error())
	} else {
		expectedData := []byte("[]")

		if reflect.DeepEqual(expectedData, results) == false {
			t.Fatalf("ExecuteQuery expected [%s] got [%s] ", expectedData, results)
		}
	}

}

func TestGetTimeLastEntryUser(t *testing.T) {

	mc := initTestData()

	entry, err := mc.GetTimeLastEntryUser(valid_groupid)

	t.Logf("TestGetTimeLastEntryUser %s ", entry)

	if len(entry) <= 0 {
		t.Fatal("GetTimeLastEntryUserAndDevice time entry hsould be set")
	}

	expectedTime := []byte("2015-01-13T08:44:04.000Z")

	if bytes.Equal(expectedTime, entry) && reflect.DeepEqual(expectedTime, entry) {
		t.Fatalf("GetTimeLastEntryUser expected [%s] got [%s] ", expectedTime, entry)
	}

	if err != nil {
		t.Fatalf("GetTimeLastEntryUser unexpected error [%s]", err.Error())
	}

}

func TestGetTimeLastEntryUser_NoData(t *testing.T) {

	mc := initTestData()

	entry, err := mc.GetTimeLastEntryUser(no_match_groupid)

	if len(entry) != 0 {
		t.Fatalf("GetTimeLastEntryUser found data when there should be none [%s]", string(entry[:]))
	}

	if err != nil {
		t.Fatalf("GetTimeLastEntryUser unexpected error [%s]", err.Error())
	}

}

func TestGetTimeLastEntryUserAndDevice(t *testing.T) {

	mc := initTestData()

	entry, err := mc.GetTimeLastEntryUserAndDevice(valid_groupid, valid_deviceid)

	t.Logf("TestGetTimeLastEntryUserAndDevice %s ", entry)

	if len(entry) <= 0 {
		t.Fatal("GetTimeLastEntryUserAndDevice time entry hsould be set")
	}

	expectedTime := []byte("2014-10-28T10:00:00.000Z")

	if bytes.Equal(expectedTime, entry) && reflect.DeepEqual(expectedTime, entry) {
		t.Fatalf("GetTimeLastEntryUserAndDevice expected [%s] got [%s] ", expectedTime, entry)
	}

	if err != nil {
		t.Fatalf("GetTimeLastEntryUserAndDevice unexpected error [%s]", err.Error())
	}

}

func TestGetTimeLastEntryUserAndDevice_NoData(t *testing.T) {

	mc := initTestData()

	entry, err := mc.GetTimeLastEntryUserAndDevice(no_match_groupid, no_match_deviceid)

	if len(entry) != 0 {
		t.Fatalf("GetTimeLastEntryUserAndDevice found data when there should be none [%s]", string(entry[:]))
	}

	if err != nil {
		t.Fatalf("GetTimeLastEntryUserAndDevice unexpected error [%s]", err.Error())
	}

}

func Test_constructQuery_WhereQueryConstruction(t *testing.T) {

	ourData := &model.QueryData{
		MetaQuery:       map[string]string{"userid": "1234"},
		WhereConditions: []model.WhereCondition{model.WhereCondition{Name: "Stuff", Value: "123", Condition: ">"}},
		Types:           []string{"cbg", "smbg"},
		InList:          []string{},
		Sort:            map[string]string{"time": "myTime"},
		Reverse:         false,
	}

	query, sort := constructQuery(ourData)

	if sort != "time" {
		t.Fatalf("sort returned [%s] but should be time", sort)
	}

	if query["_groupId"] != "1234" {
		t.Fatalf("_groupId [%v] should have been set to given 1234", query)
	}

	if query["type"] == nil {
		t.Fatalf("type should have two items [%v]", query["type"])
	}

	types := query["type"]
	expectedTypes := bson.M{"$in": []string{"cbg", "smbg"}}

	if reflect.DeepEqual(types, expectedTypes) != true {
		t.Fatalf("given %v but expected %v", types, expectedTypes)
	}

	//check the where condition
	where := query["Stuff"]
	expectedWhere := bson.M{"$gt": "123"}

	if reflect.DeepEqual(where, expectedWhere) != true {
		t.Fatalf("given %v but expected %v", where, expectedWhere)
	}

}

func TestInQueryConstruction(t *testing.T) {

	ourData := &model.QueryData{
		MetaQuery:       map[string]string{"userid": "1234"},
		WhereConditions: []model.WhereCondition{model.WhereCondition{Name: "updateId", Value: "NOTHING", Condition: "IN"}},
		Types:           []string{"cbg"},
		InList:          []string{"firstId", "secondId"},
		Sort:            map[string]string{"time": "myTime"},
		Reverse:         false,
	}

	query, sort := constructQuery(ourData)

	if sort != "time" {
		t.Fatalf("sort returned [%s] but should be time", sort)
	}

	if query["_groupId"] != "1234" {
		t.Fatalf("_groupId [%v] should have been set to given 1234", query)
	}

	//check the types
	types := query["type"]
	expectedTypes := bson.M{"$in": []string{"cbg"}}

	if reflect.DeepEqual(types, expectedTypes) != true {
		t.Fatalf("given %v but expected %v", types, expectedTypes)
	}

	//check the where condition
	where := query["updateId"]
	expectedWhere := bson.M{"$in": []string{"firstId", "secondId"}}

	if reflect.DeepEqual(where, expectedWhere) != true {
		t.Fatalf("given %v but expected %v", where, expectedWhere)
	}

}
