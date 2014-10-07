package clients

type StoreClient interface {
	Close()
	Ping() error
	GetTimeLastEntryUser(groupId string) []byte
	GetTimeLastEntryUserAndDevice(groupId, deviceId string) []byte
}
