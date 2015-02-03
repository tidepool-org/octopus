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
	"log"
	"strings"

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
		if len(details.InList) > 0 {
			log.Println("constructQuery: adding inlist")
			first := details.WhereConditions[0]
			switch strings.ToLower(first.Condition) {
			case "in":
				queryThis[first.Name] = bson.M{"$in": details.InList}
			case "not in":
				queryThis[first.Name] = bson.M{"$nin": details.InList}
			}
			queryThat[first.Name] = queryThis[first.Name]
		} else {

			//add where but only if there wasn't an InList
			if len(details.WhereConditions) == 1 {
				log.Println("constructQuery: where statement with just one condition")
				first := details.WhereConditions[0]
				op := getMongoOperator(first.Condition)
				queryThis[first.Name] = bson.M{op: first.Value}
				queryThat[first.Name] = bson.M{op: first.Value}
			} else if len(details.WhereConditions) == 2 {
				log.Println("constructQuery: where statement with two conditions")
				first := details.WhereConditions[0]
				op1 := getMongoOperator(first.Condition)
				second := details.WhereConditions[1]
				op2 := getMongoOperator(second.Condition)
				queryThis[first.Name] = bson.M{op1: first.Value, op2: second.Value}
				queryThat[first.Name] = bson.M{op1: first.Value, op2: second.Value}
			}
		}

		query = bson.M{"$or": []bson.M{queryThis, queryThat}}
		log.Printf("constructQuery: full query is %v", query)
	}
	//sort field and order
	for k := range details.Sort {
		sort = k
		if details.Reverse {
			sort = "-" + sort
		}
	}

	return query, sort
}

func (d MongoStoreClient) ExecuteQuery(details *model.QueryData) []byte {

	query, sort := constructQuery(details)

	log.Printf("ExecuteQuery query[%v] sort[%v]", query, sort)

	// Request a socket connection from the session to process our query.
	// Close the session when the goroutine exits and put the connection back
	// into the pool.
	sessionCopy := d.session.Copy()
	defer sessionCopy.Close()

	var results []interface{}
	//we don't want to return the _id
	filter := bson.M{"_id": 0}

	sessionCopy.DB("").C(DEVICE_DATA_COLLECTION).
		Find(query).
		Sort(sort).
		Select(filter).
		All(&results)

	log.Printf("ExecuteQuery found [%d] results", len(results))

	if len(results) == 0 {
		return []byte("[]")
	} else {
		bytes, _ := json.Marshal(results)
		return bytes
	}
}
