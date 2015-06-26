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
	"errors"

	"../model"
)

type MockStoreClient struct {
	salt        string
	ThrowError  bool
	ReturnOther bool
}

func NewMockStoreClient(salt string, returnDifferent, doBad bool) *MockStoreClient {
	return &MockStoreClient{salt: salt, ThrowError: doBad, ReturnOther: returnDifferent}
}

func (d MockStoreClient) Close() {}

func (d MockStoreClient) Ping() error {
	if d.ThrowError {
		return errors.New("Session failure")
	}
	return nil
}

func (d MockStoreClient) GetTimeLastEntryUser(deviceId string) ([]byte, error) {
	if d.ThrowError {
		return nil, errors.New("GetTimeLastEntryUser mongo error")
	}
	return []byte("GetTimeLastEntryUser"), nil
}

func (d MockStoreClient) GetTimeLastEntryUserAndDevice(groupId, deviceId string) ([]byte, error) {
	if d.ThrowError {
		return nil, errors.New("GetTimeLastEntryUserAndDevice mongo error")
	}
	return []byte("GetTimeLastEntryUserDevice"), nil
}

func (d MockStoreClient) ExecuteQuery(details *model.QueryData) ([]byte, error) {
	if d.ThrowError {
		return nil, errors.New("ExecuteQuery mongo error")
	}
	return []byte("ExecuteQuery"), nil
}
