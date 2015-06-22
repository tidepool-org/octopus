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

package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	commonClients "github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/clients/shoreline"

	"./../clients"
)

const (
	valid_token   = "yolo"
	invalid_token = "bad-token"

	valid_groupid = "abcdefg"

	valid_userid           = "oldgreg"
	userid_can_only_upload = "upload-only"
	userid_no_match_found  = "no-match"

	valid_deviceid = "some-supported-device"

	SOME_SALT = "salty"
)

type (
	MockShorelineClient struct{}

	MockSeagullClient struct{}

	MockGateKeeperClient struct{}
)

func (slc MockShorelineClient) CheckToken(token string) *shoreline.TokenData {
	if token == invalid_token {
		log.Print("MockShorelineClient.CheckToken ", "return nil as you gave the token", invalid_token)
		return nil
	}
	return &shoreline.TokenData{UserID: valid_userid, IsServer: false}
}

func (slc MockShorelineClient) TokenProvide() string {
	log.Print("MockShorelineClient.TokenProvide", "return a valid token")
	return valid_token
}

func (slc MockShorelineClient) GetUser(userID, token string) (*shoreline.UserData, error) {
	log.Print("MockShorelineClient.GetUser", "return the user asked for")
	return &shoreline.UserData{UserID: userID, UserName: userID, Emails: []string{userID}}, nil
}

func (sgc MockSeagullClient) GetPrivatePair(userID, hashName, token string) *commonClients.PrivatePair {
	if userID == userid_no_match_found {
		log.Println("MockSeagullClient.GetPrivatePair", "no private pair found")
		return nil
	}
	return &commonClients.PrivatePair{ID: hashName, Value: "value-to-use"}
}

func (gkc MockGateKeeperClient) UserInGroup(userID, groupID string) (map[string]commonClients.Permissions, error) {
	permissonsToReturn := make(map[string]commonClients.Permissions)
	p := make(commonClients.Permissions)

	if userID == userid_can_only_upload {
		log.Println("MockGateKeeperClient.UserInGroup", "Allow `upload` perms only")
		p["userid"] = userID
		permissonsToReturn["upload"] = p
		return permissonsToReturn, nil
	}

	log.Println("MockGateKeeperClient.UserInGroup", "Allow `view` perms only")
	p["userid"] = userID
	permissonsToReturn["view"] = p
	return permissonsToReturn, nil
}

// initialize the api in a working state:
// we may reset some clients depending on what we are trying to assert in our tests
func initApiForTest() *Api {
	return InitApi(
		Config{ServerSecret: "shhh! don't tell"},
		MockShorelineClient{},
		MockSeagullClient{},
		MockGateKeeperClient{},
		clients.NewMockStoreClient(SOME_SALT, false, false),
	)
}

func Test_GetStatus_OK(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	octo.GetStatus(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusOK)
	}

}

func Test_GetStatus_InternalServerError(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	//the store will throw an exception
	octo.Store = clients.NewMockStoreClient(SOME_SALT, false, true)

	octo.GetStatus(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusInternalServerError)
	}
}

func Test_TimeLastEntryUser_OK(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(SESSION_TOKEN, valid_token)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	octo.TimeLastEntryUser(res, req, httpVars{"userID": valid_userid})
	if res.Code != http.StatusOK {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusOK)
	}
}

func Test_TimeLastEntryUser_BadRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(SESSION_TOKEN, valid_token)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	octo.TimeLastEntryUser(res, req, httpVars{"userID": userid_no_match_found})
	if res.Code != http.StatusBadRequest {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusBadRequest)
	}
}

func Test_TimeLastEntryUser_Unauthorized(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(SESSION_TOKEN, invalid_token)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	octo.TimeLastEntryUser(res, req, httpVars{"userID": valid_userid})

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusUnauthorized)
	}
}

func Test_TimeLastEntryUser_Forbidden(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(SESSION_TOKEN, valid_token)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	octo.TimeLastEntryUser(res, req, httpVars{"userID": userid_can_only_upload})
	if res.Code != http.StatusForbidden {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusForbidden)
	}
}

func Test_TimeLastEntryUserAndDevice_OK(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(SESSION_TOKEN, valid_token)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	octo.TimeLastEntryUserAndDevice(res, req, httpVars{"userID": valid_userid, "deviceID": valid_deviceid})
	if res.Code != http.StatusOK {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusOK)
	}
}

func Test_TimeLastEntryUserAndDevice_BadRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(SESSION_TOKEN, valid_token)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	octo.TimeLastEntryUserAndDevice(res, req, httpVars{"userID": userid_no_match_found, "deviceID": valid_deviceid})
	if res.Code != http.StatusBadRequest {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusBadRequest)
	}
}

func Test_TimeLastEntryUserAndDevice_Unauthorized(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(SESSION_TOKEN, invalid_token)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	octo.TimeLastEntryUserAndDevice(res, req, httpVars{"userID": valid_userid, "deviceID": valid_deviceid})

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusUnauthorized)
	}
}

func Test_TimeLastEntryUserAndDevice_Forbidden(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(SESSION_TOKEN, valid_token)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	octo.TimeLastEntryUserAndDevice(res, req, httpVars{"userID": userid_can_only_upload, "deviceID": valid_deviceid})
	if res.Code != http.StatusForbidden {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusForbidden)
	}
}

func encodeQuery(queryString string) *bytes.Buffer {
	var body = &bytes.Buffer{}
	json.NewEncoder(body).Encode(queryString)
	return body
}

func Test_Query_Unauthorized(t *testing.T) {

	//query is valid
	body := encodeQuery("METAQUERY WHERE userid IS 12d7bc90fa QUERY TYPE IN cbg, smbg SORT BY time AS Timestamp REVERSED")

	req, _ := http.NewRequest("POST", "/", body)
	req.Header.Set(SESSION_TOKEN, invalid_token)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	octo.Query(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusUnauthorized)
	}
}

func Test_Query_BadRequest(t *testing.T) {

	//query will be invalid
	body := encodeQuery("METAQUERY WHERE REVERSED")

	req, _ := http.NewRequest("POST", "/", body)
	req.Header.Set(SESSION_TOKEN, valid_token)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	octo.Query(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusBadRequest)
	}
}

func Test_Query_InternalServerError(t *testing.T) {

	//query is valid
	body := encodeQuery("METAQUERY WHERE userid IS 12d7bc90fa QUERY TYPE IN cbg, smbg SORT BY time AS Timestamp REVERSED")

	req, _ := http.NewRequest("POST", "/", body)
	req.Header.Set(SESSION_TOKEN, valid_token)
	res := httptest.NewRecorder()

	octo := initApiForTest()
	//set the store so it will throw an error
	octo.Store = clients.NewMockStoreClient(SOME_SALT, false, true)

	octo.Query(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusInternalServerError)
	}
}
