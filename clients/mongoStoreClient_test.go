package clients

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/tidepool-org/go-common/clients/mongo"
	//"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"../model"
)

var (
	qd = &model.QueryData{
		MetaQuery:      map[string]string{"userid": "1234"},
		WhereConditons: []model.WhereCondition{model.WhereCondition{Name: "Stuff", Value: "123", Condition: ">"}},
		Types:          []string{"cbg", "smbg"},
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

		//NewMongoStoreClient(config.Mongo)

	} else {
		t.Fatalf("wtf - failed parsing the config %v", err)
	}
}

func TestQueryConstruction(t *testing.T) {

	query, sort := constructQuery(qd)

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
