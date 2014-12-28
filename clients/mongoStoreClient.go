package clients

import (
	"encoding/json"
	"log"

	"github.com/tidepool-org/go-common/clients/mongo"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"../model"
)

const (
	DEVICE_DATA_COLLECTION = "deviceData"
)

type MongoStoreClient struct {
	session     *mgo.Session
	deviceDataC *mgo.Collection
}

func NewMongoStoreClient(config *mongo.Config) *MongoStoreClient {

	mongoSession, err := mongo.Connect(config)
	if err != nil {
		log.Fatal(err)
	}

	return &MongoStoreClient{
		session:     mongoSession,
		deviceDataC: mongoSession.DB("").C(DEVICE_DATA_COLLECTION),
	}
}

func (d MongoStoreClient) Close() {
	log.Println("Close the session")
	d.session.Close()
	return
}

func (d MongoStoreClient) Ping() error {
	// do we have a store session
	if err := d.session.Ping(); err != nil {
		return err
	}
	return nil
}

func (d MongoStoreClient) GetTimeLastEntryUser(groupId string) []byte {
	mongoSession := d.session.Copy()
	var result map[string]interface{}
	c := mongoSession.DB("").C(DEVICE_DATA_COLLECTION)
	groupIdQuery := bson.M{"$or": []bson.M{bson.M{"groupId": groupId},
		bson.M{"_groupId": groupId, "_active": true}}}
	// Get the entry with the latest time by reverse sorting and taking the first value
	c.Find(groupIdQuery).Sort("-time").One(&result)
	bytes, err := json.Marshal(result["time"])
	if err != nil {
		log.Print("Failed to marshall event", result, err)
	}
	return bytes
}

func (d MongoStoreClient) GetTimeLastEntryUserAndDevice(groupId, deviceId string) []byte {
	mongoSession := d.session.Copy()
	var result map[string]interface{}
	c := mongoSession.DB("").C(DEVICE_DATA_COLLECTION)
	groupIdQuery := bson.M{"$or": []bson.M{bson.M{"groupId": groupId},
		bson.M{"_groupId": groupId, "_active": true}}}
	deviceIdQuery := bson.M{"deviceId": deviceId}
	// Full query matches groupId and deviceId
	fullQuery := bson.M{"$and": []bson.M{groupIdQuery, deviceIdQuery}}
	// Get the entry with the latest time by reverse sorting and taking the first value
	c.Find(fullQuery).Sort("-time").One(&result)
	bytes, err := json.Marshal(result["time"])
	if err != nil {
		log.Print("Failed to marshall event", result, err)
	}
	return bytes
}

func (d MongoStoreClient) constructQuery(details model.QueryData) (query bson.M) {
	return query
}

func (d MongoStoreClient) ExecuteQuery(details model.QueryData) []byte {
	mongoSession := d.session.Copy()
	var result map[string]interface{}
	c := mongoSession.DB("").C(DEVICE_DATA_COLLECTION)

	query := bson.M{}
	// Get the entry with the latest time by reverse sorting and taking the first value
	c.Find(query).Sort("-time").One(&result)
	bytes, err := json.Marshal(result["time"])
	if err != nil {
		log.Print("Failed to marshall event", result, err)
	}
	return bytes
}
