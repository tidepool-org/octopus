package clients

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"labix.org/v2/mgo"

	"github.com/tidepool-org/go-common/clients/mongo"
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
