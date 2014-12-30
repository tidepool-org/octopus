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

	var result map[string]interface{}
	groupIdQuery := bson.M{"$or": []bson.M{bson.M{"groupId": groupId},
		bson.M{"_groupId": groupId, "_active": true}}}
	// Get the entry with the latest time by reverse sorting and taking the first value
	d.deviceDataC.Find(groupIdQuery).Sort("-time").One(&result)
	bytes, err := json.Marshal(result["time"])
	if err != nil {
		log.Print("Failed to marshall event", result, err)
	}
	return bytes
}

func (d MongoStoreClient) GetTimeLastEntryUserAndDevice(groupId, deviceId string) []byte {

	var result map[string]interface{}

	groupIdQuery := bson.M{"$or": []bson.M{bson.M{"groupId": groupId},
		bson.M{"_groupId": groupId, "_active": true}}}
	deviceIdQuery := bson.M{"deviceId": deviceId}
	// Full query matches groupId and deviceId
	fullQuery := bson.M{"$and": []bson.M{groupIdQuery, deviceIdQuery}}
	// Get the entry with the latest time by reverse sorting and taking the first value
	d.deviceDataC.Find(fullQuery).Sort("-time").One(&result)
	bytes, err := json.Marshal(result["time"])
	if err != nil {
		log.Print("Failed to marshall event", result, err)
	}
	return bytes
}

//map to the mongo conditions
func getMongoOperator(op string) string {
	switch op {
	case "<":
		return "$lt"
	case "<=":
		return "$lte"
	case ">":
		return "$gt"
	case ">=":
		return "$gte"
	default:
		return ""
	}
}

func constructQuery(details *model.QueryData) (query bson.M, sort string) {
	//query
	for _, v := range details.MetaQuery {
		log.Println("constructQuery: create base queries")
		//base query
		queryThis := bson.M{"groupId": v}
		queryThat := bson.M{"_groupId": v, "_active": true}
		//add types
		if len(details.Types) > 0 {
			log.Println("constructQuery: adding types")
			queryThis["type"] = bson.M{"$in": details.Types}
			queryThat["type"] = bson.M{"$in": details.Types}
		}
		//add where
		if len(details.WhereConditons) > 0 {

			if len(details.WhereConditons) == 1 {
				log.Println("constructQuery: where statement with just the one condition")
				first := details.WhereConditons[0]
				op := getMongoOperator(first.Condition)
				queryThis[first.Name] = bson.M{op: first.Value}
				queryThat[first.Name] = bson.M{op: first.Value}
			} else if len(details.WhereConditons) == 2 {
				log.Println("constructQuery: where statement with with two conditions")
				first := details.WhereConditons[0]
				op1 := getMongoOperator(first.Condition)
				second := details.WhereConditons[1]
				op2 := getMongoOperator(second.Condition)
				queryThis[first.Name] = bson.M{op1: first.Value, op2: second.Value}
				queryThat[first.Name] = bson.M{op1: first.Value, op2: second.Value}

			}
		}
		query = bson.M{"$or": []bson.M{queryThis, queryThat}}
		log.Printf("constructQuery: full query is %v", query)
	}

	//sort
	for k := range details.Sort {
		sort = k
		if details.Reverse {
			sort = "-" + sort
		}
	}
	return query, sort
}

func (d MongoStoreClient) ExecuteQuery(details *model.QueryData) []byte {
	var results []interface{}

	query, sort := constructQuery(details)

	log.Printf("ExecuteQuery query[%v] sort[%s]", query, sort)

	d.deviceDataC.Find(query).Sort(sort).All(&results)
	bytes, err := json.Marshal(results)
	if err != nil {
		log.Print("Failed to marshall event", results, err)
	}
	return bytes
}
