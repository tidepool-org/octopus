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
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/tidepool-org/go-common/clients/mongo"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"../model"
)

var (
	theTime  = "2014-10-23T10:00:00.000Z"
	basalsQd = &model.QueryData{
		MetaQuery:       map[string]string{"userid": "1234"},
		WhereConditions: []model.WhereCondition{model.WhereCondition{Name: "time", Value: theTime, Condition: "<"}},
		Types:           []string{"basal"},
		Sort:            map[string]string{"time": "myTime"},
		Reverse:         false,
	}
)

func TestMongoStore(t *testing.T) {

	//we are setting to something other than the default so we can isolate the test data
	testingConfig := &mongo.Config{ConnectionString: "mongodb://localhost/streams_test"}

	mc := NewMongoStoreClient(testingConfig)

	/*
	 * INIT THE TEST - we use a clean copy of the collection before we start
	 */

	mc.deviceDataC.DropCollection()

	if err := mc.deviceDataC.Create(&mgo.CollectionInfo{}); err != nil {
		t.Fatalf("We couldn't create the device data collection for these tests ", err)
	}

	/*
	 * Load test data
	 */
	if testData, err := ioutil.ReadFile("./test_data.json"); err == nil {

		var toLoad []interface{}

		if err := json.Unmarshal(testData, &toLoad); err != nil {
			t.Fatalf("We could not load the test data ", err)
		}

		for i := range toLoad {
			//insert each test data item
			if insertErr := mc.deviceDataC.Insert(toLoad[i]); insertErr != nil {
				t.Fatalf("We could not load the test data ", insertErr)
			}
		}
	}
	/*
	 * Load test data
	 */
	if results := mc.ExecuteQuery(basalsQd); results == nil {
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

func TestWhereQueryConstruction(t *testing.T) {

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
