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

func constructQuery(details *model.QueryData) (query bson.M, typesIn bson.M) {
	//METAQUERY WHERE userid IS "12d7bc90fa"
	//QUERY TYPE IN cbg, smbg, bolus, wizard WHERE time > starttime AND time < endtime SORT BY time AS Timestamp REVERSED

	for _, v := range details.Where {
		log.Printf("constructQuery for [%s]", v)
		query = bson.M{"$or": []bson.M{bson.M{"groupId": v}, bson.M{"_groupId": v, "_active": true}}}
	}

	typesIn = bson.M{}

	for i := range details.Types {
		typesIn[details.Types[i]] = 1
	}

	return query, typesIn
}

func (d MongoStoreClient) ExecuteQuery(details *model.QueryData) []byte {
	mongoSession := d.session.Copy()
	var result map[string]interface{}
	var sortField = ""
	c := mongoSession.DB("").C(DEVICE_DATA_COLLECTION)

	query, _ := constructQuery(details)

	for k := range details.Sort {
		sortField = k
		if details.Reverse {
			sortField = "-" + sortField
		}
	}

	log.Printf("sort by [%s]", sortField)

	c.Find(query).Sort(sortField).One(&result)
	bytes, err := json.Marshal(result)
	if err != nil {
		log.Print("Failed to marshall event", result, err)
	}
	return bytes
}
