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
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/tidepool-org/go-common/clients/mongo"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"../model"
)

const (
	DEVICE_DATA_COLLECTION = "deviceData"
)

type MongoStoreClient struct {
	session *mgo.Session
	logger  *log.Logger
	config  *StoreConfig
}

type StoreConfig struct {
	Connection    *mongo.Config `json:"mongo"`
	SchemaVersion `json:"schemaVersion"`
}

type SchemaVersion struct {
	Minimum int
	Maximum int
}

//all queries will be built on top of this
func (d MongoStoreClient) getBaseQuery(groupId string) bson.M {
	d.logger.Printf("target schema version %v", d.config.SchemaVersion)
	return bson.M{"_groupId": groupId, "_active": true, "_schemaVersion": bson.M{"$gte": d.config.Minimum, "$lte": d.config.Maximum}}
}

func NewMongoStoreClient(config *StoreConfig) *MongoStoreClient {

	mongoSession, err := mongo.Connect(config.Connection)
	if err != nil {
		log.Fatal(err)
	}

	//Note 1:  the order of the fields is important and should match query order
	//Note 2:  '-time' is the field we are sorting on must be the last field in the index
	queryIndex := mgo.Index{
		Key:        []string{"_groupId", "_active", "_schemaVersion", "type", "-time"},
		Background: true,
	}
	mongoSession.DB("").C(DEVICE_DATA_COLLECTION).EnsureIndex(queryIndex)

	//As above but includes uploadId for restriction of data returned
	queryUploadIdIndex := mgo.Index{
		Key:        []string{"_groupId", "_active", "_schemaVersion", "type", "uploadId", "-time"},
		Background: true,
	}
	mongoSession.DB("").C(DEVICE_DATA_COLLECTION).EnsureIndex(queryUploadIdIndex)

	storeLogger := log.New(os.Stdout, "api/query:", log.Lshortfile)

	return &MongoStoreClient{session: mongoSession, logger: storeLogger, config: config}
}

func (d MongoStoreClient) Close() {
	d.logger.Println("Close the session")
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

func (d MongoStoreClient) interpretQueryError(err error, startedAt time.Time, returnOnNotFound []byte) ([]byte, error) {

	failedAfter := time.Now().Sub(startedAt).Seconds()

	if err == mgo.ErrNotFound {
		d.logger.Println(fmt.Sprintf("mongo query took [%.5f] secs and found no results", failedAfter))
		return returnOnNotFound, nil
	} else {
		d.logger.Println(fmt.Sprintf("mongo query took [%.5f] secs but failed with error [%s] ", failedAfter, err.Error()))
		return nil, err
	}

}

func (d MongoStoreClient) GetTimeLastEntryUser(groupId string) ([]byte, error) {
	var result map[string]interface{}
	startQueryTime := time.Now()
	sessionCopy := d.session.Copy()
	defer sessionCopy.Close()

	// Get the entry with the latest time by reverse sorting and taking the first value
	err := sessionCopy.DB("").C(DEVICE_DATA_COLLECTION).
		Find(d.getBaseQuery(groupId)).
		Sort("-time").
		One(&result)

	if err != nil {
		return d.interpretQueryError(err, startQueryTime, []byte(""))
	}

	d.logger.Println(fmt.Sprintf("mongo query took [%.5f] secs ", time.Now().Sub(startQueryTime).Seconds()))
	return json.Marshal(result["time"])
}

func (d MongoStoreClient) GetTimeLastEntryUserAndDevice(groupId, deviceId string) ([]byte, error) {

	var result map[string]interface{}

	startQueryTime := time.Now()
	sessionCopy := d.session.Copy()
	defer sessionCopy.Close()

	// Get the entry with the latest time by reverse sorting and taking the first value
	err := sessionCopy.DB("").C(DEVICE_DATA_COLLECTION).
		Find(bson.M{"$and": []bson.M{d.getBaseQuery(groupId), bson.M{"deviceId": deviceId}}}).
		Sort("-time").
		One(&result)

	if err != nil {
		return d.interpretQueryError(err, startQueryTime, []byte(""))
	}

	d.logger.Println(fmt.Sprintf("mongo query took [%.5f] secs ", time.Now().Sub(startQueryTime).Seconds()))
	return json.Marshal(result["time"])
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

func (d MongoStoreClient) constructQuery(details *model.QueryData) (query bson.M, sort string) {
	for _, v := range details.MetaQuery {
		//start with the base query

		query = d.getBaseQuery(v)
		//add types
		if len(details.Types) > 0 {
			query["type"] = bson.M{"$in": details.Types}
		}
		if len(details.InList) > 0 {
			first := details.WhereConditions[0]
			switch strings.ToLower(first.Condition) {
			case "in":
				query[first.Name] = bson.M{"$in": details.InList}
			case "not in":
				query[first.Name] = bson.M{"$nin": details.InList}
			}
		} else {
			//add where but only if there wasn't an InList
			if len(details.WhereConditions) == 1 {
				first := details.WhereConditions[0]
				op := getMongoOperator(first.Condition)
				query[first.Name] = bson.M{op: first.Value}
			} else if len(details.WhereConditions) == 2 {
				first := details.WhereConditions[0]
				op1 := getMongoOperator(first.Condition)
				second := details.WhereConditions[1]
				op2 := getMongoOperator(second.Condition)
				query[first.Name] = bson.M{op1: first.Value, op2: second.Value}
			}
		}
		d.logger.Printf("mongo query %#v", query)
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

func (d MongoStoreClient) ExecuteQuery(details *model.QueryData) ([]byte, error) {

	startTime := time.Now()

	query, sort := d.constructQuery(details)
	d.logger.Println(fmt.Sprintf("mongo query built in [%.5f] secs", time.Now().Sub(startTime).Seconds()))

	var results []interface{}
	//we don't want to return these
	filter := bson.M{"_id": 0, "_active": 0}

	startQueryTime := time.Now()
	sessionCopy := d.session.Copy()
	defer sessionCopy.Close()

	err := sessionCopy.DB("").C(DEVICE_DATA_COLLECTION).
		Find(query).
		Sort(sort).
		Select(filter).
		All(&results)

	if err != nil {
		return d.interpretQueryError(err, startQueryTime, []byte("[]"))
	}
	d.logger.Println(fmt.Sprintf("mongo query took [%.5f] secs and returned [%d] records", time.Now().Sub(startQueryTime).Seconds(), len(results)))

	if len(results) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(results)

}
