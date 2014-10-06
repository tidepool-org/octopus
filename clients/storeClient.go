package clients

type StoreClient interface {
	Close()
	Ping() error
	GetTimeLastEntry(groupId, deviceId string) []byte
}
