package clients

import (
	//"./../models"
	//"fmt"
	"github.com/tidepool-org/go-common/clients/mongo"
	"labix.org/v2/mgo"
	//"labix.org/v2/mgo/bson"
	"log"
)

const (
	USERS_COLLECTION  = "users"
	TOKENS_COLLECTION = "tokens"
)

type MongoStoreClient struct {
	session *mgo.Session
	usersC  *mgo.Collection
	tokensC *mgo.Collection
}

func NewMongoStoreClient(config *mongo.Config) *MongoStoreClient {

	mongoSession, err := mongo.Connect(config)
	if err != nil {
		log.Fatal(err)
	}

	return &MongoStoreClient{
		session: mongoSession,
		usersC:  mongoSession.DB("").C(USERS_COLLECTION),
		tokensC: mongoSession.DB("").C(TOKENS_COLLECTION),
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

func (d MongoStoreClient) GetTimeLastEntry(groupId, deviceId string) string {
	return "always the same"
}
