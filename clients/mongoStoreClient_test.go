package clients

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/tidepool-org/go-common/clients/mongo"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"../model"
)

var (
	qd = &model.QueryData{
		Where:   map[string]string{"userid": "1234"},
		Types:   []string{"cbg", "smbg"},
		Sort:    map[string]string{"time": "myTime"},
		Reverse: false,
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

		//drop it like its hot
		mc.deviceDataC.DropCollection()

		if err := mc.deviceDataC.Create(&mgo.CollectionInfo{}); err != nil {
			t.Fatalf("We couldn't created the users collection for these tests ", err)
		}

	} else {
		t.Fatalf("wtf - failed parsing the config %v", err)
	}
}

func TestQueryConstruction(t *testing.T) {

	q, in := constructQuery(qd)

	if in["cbg"] == nil {
		t.Fatalf("should include cbg [%v]", in)
	}
	if in["smbg"] == nil {
		t.Fatalf("should include smbg [%v]", in)
	}

	groupWhere := q["$or"].([]bson.M)[0]

	if groupWhere["groupId"] != "1234" {
		t.Fatalf("groupId [%v] should have been set to given 1234", groupWhere)
	}

	_groupWhere := q["$or"].([]bson.M)[1]

	if _groupWhere["_groupId"] != "1234" {
		t.Fatalf("_groupId [%v] should have been set to given 1234", _groupWhere)
	}

}
