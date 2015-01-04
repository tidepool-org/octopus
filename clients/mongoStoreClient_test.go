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
		MetaQuery:      map[string]string{"userid": "1234"},
		WhereConditons: []model.WhereCondition{model.WhereCondition{Name: "time", Value: theTime, Condition: "<"}},
		Types:          []string{"basal"},
		Sort:           map[string]string{"time": "myTime"},
		Reverse:        false,
	}
)

func TestMongoStore(t *testing.T) {

	type Config struct {
		Mongo *mongo.Config `json:"mongo"`
	}

	var (
		config Config
	)

	if jsonConfig, err := ioutil.ReadFile("../config/server.json"); err == nil {

		if err := json.Unmarshal(jsonConfig, &config); err != nil {
			t.Fatalf("We could not load the config ", err)
		}

		mc := NewMongoStoreClient(config.Mongo)

		/*
		 * INIT THE TEST - we use a clean copy of the collection before we start
		 */

		mc.deviceDataC.DropCollection()

		if err := mc.deviceDataC.Create(&mgo.CollectionInfo{}); err != nil {
			t.Fatalf("We couldn't created the device data collection for these tests ", err)
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
			//all we care about for this test
			type found struct {
				Time string
				Type string
			}
			const ISO_8601 = "2006-01-02T15:04:05Z"
			timeClause, _ := time.Parse(ISO_8601, theTime)

			records := []found{}
			json.Unmarshal(results, &records)

			//number of recoords
			if len(records) != 2 {
				t.Fatalf("we should have been given two results but got [%d]", len(records))
			}

			for rec := range records {
				//check the type
				if records[rec].Type != "basal" {
					t.Fatalf("they should be of type basal but where [%s] [%s]", records[rec].Type)
				}
				//check time in range
				timeIs, _ := time.Parse(ISO_8601, records[rec].Time)
				if timeIs.After(timeClause) {
					t.Fatalf("time [%v] should be before than [%v] ", timeIs, timeClause)
				}
			}
		}

	} else {
		t.Fatalf("wtf - failed parsing the config %v", err)
	}
}

func TestQueryConstruction(t *testing.T) {

	ourData := &model.QueryData{
		MetaQuery:      map[string]string{"userid": "1234"},
		WhereConditons: []model.WhereCondition{model.WhereCondition{Name: "Stuff", Value: "123", Condition: ">"}},
		Types:          []string{"cbg", "smbg"},
		Sort:           map[string]string{"time": "myTime"},
		Reverse:        false,
	}

	query, sort := constructQuery(ourData)

	if sort != "time" {
		t.Fatalf("sort returned [%s] but should be time", sort)
	}

	//check the meta query
	meta := query["$or"].([]bson.M)[0]

	if meta["groupId"] != "1234" {
		t.Fatalf("groupId [%v] should have been set to given 1234", meta)
	}

	//check the types
	types := meta["type"]
	expectedTypes := bson.M{"$in": []string{"cbg", "smbg"}}

	if reflect.DeepEqual(types, expectedTypes) != true {
		t.Fatalf("given %v but expected %v", types, expectedTypes)
	}

	//check the where condition
	where := meta["Stuff"]
	expectedWhere := bson.M{"$gt": "123"}

	if reflect.DeepEqual(where, expectedWhere) != true {
		t.Fatalf("given %v but expected %v", where, expectedWhere)
	}

	//check the other $or component of the query
	_meta := query["$or"].([]bson.M)[1]

	if _meta["_groupId"] != "1234" {
		t.Fatalf("_groupId [%v] should have been set to given 1234", _meta)
	}

	if _meta["type"] == nil {
		t.Fatalf("type should have two items [%v]", _meta["type"])
	}

}
