package clients

import (
	"errors"
)

type MockStoreClient struct {
	salt            string
	doBad           bool
	returnDifferent bool
}

func NewMockStoreClient(salt string, returnDifferent, doBad bool) *MockStoreClient {
	return &MockStoreClient{salt: salt, doBad: doBad, returnDifferent: returnDifferent}
}

func (d MockStoreClient) Close() {}

func (d MockStoreClient) Ping() error {
	if d.doBad {
		return errors.New("Session failure")
	}
	return nil
}

func (d MockStoreClient) GetTimeLastEntryUser(deviceId string) []byte {
	return []byte("GetTimeLastEntryUser")
}

func (d MockStoreClient) GetTimeLastEntryUserAndDevice(groupId, deviceId string) []byte {
	return []byte("GetTimeLastEntryUserDevice")
}
