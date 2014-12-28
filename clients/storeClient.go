package clients

import "../model"

type StoreClient interface {
	Close()
	ExecuteQuery(details *model.QueryData) []byte
	GetTimeLastEntryUser(groupId string) []byte
	GetTimeLastEntryUserAndDevice(groupId, deviceId string) []byte
	Ping() error
}
