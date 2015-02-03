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
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	commonClients "github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/clients/shoreline"

	"./../clients"
)

const (
	FAKE_TOKEN             = "yolo"
	FAKE_GROUP_ID          = "abcdefg"
	FAKE_VALUE             = "gfedcba"
	FAKE_USER_ID           = "oldgreg"
	FAKE_USER_ID_DIFFERENT = "different"
	SOME_SALT              = "salty"
)

type (

	//common test structure
	toTest struct {
		skip       bool
		returnNone bool
		method     string
		url        string
		body       string
		token      string
		respCode   int
		response   jo
	}
	// These two types make it easier to define blobs of json inline.
	// We don't use the types defined by the API because we want to
	// be able to test with partial data structures.
	// jo is a generic json object
	jo map[string]interface{}

	// and ja is a generic json array
	ja []interface{}

	MockShorelineClient struct {
		validToken bool
		token      shoreline.TokenData
	}

	MockSeagullClient struct{}

	MockGateKeeperClient struct{}
)

func (i *jo) deepCompare(j *jo) string {
	for k, _ := range *i {
		if reflect.DeepEqual((*i)[k], (*j)[k]) == false {
			return fmt.Sprintf("for [%s] was [%v] expected [%v] ", k, (*i)[k], (*j)[k])
		}
	}
	return ""
}

func (slc MockShorelineClient) CheckToken(token string) *shoreline.TokenData {
	if slc.validToken {
		return &slc.token
	} else {
		return nil
	}
}

func (slc MockShorelineClient) TokenProvide() string {
	return FAKE_TOKEN
}

func (sgc MockSeagullClient) GetPrivatePair(userID, hashName, token string) *commonClients.PrivatePair {
	return &commonClients.PrivatePair{FAKE_GROUP_ID, FAKE_VALUE}
}

func (gkc MockGateKeeperClient) UserInGroup(userID, groupID string) (map[string]commonClients.Permissions, error) {
	return nil, nil
}

var (
	mockConfig = Config{
		ServerSecret: "shhh! don't tell",
	}
	vars                        = httpVars{"userID": FAKE_USER_ID}
	varsDifferent               = httpVars{"userID": FAKE_USER_ID_DIFFERENT}
	tokenIsServer               = shoreline.TokenData{FAKE_USER_ID, true}
	tokenIsNotServer            = shoreline.TokenData{FAKE_USER_ID, false}
	mockShoreline               = MockShorelineClient{true, tokenIsServer}
	mockShorelineNilToken       = MockShorelineClient{false, tokenIsServer}
	mockShorelineTokenNotServer = MockShorelineClient{true, tokenIsNotServer}
	mockSeagullClient           = MockSeagullClient{}
	mockeGatekeeperClient       = MockGateKeeperClient{}
	mockStore                   = clients.NewMockStoreClient(SOME_SALT, false, false)
	mockStoreEmpty              = clients.NewMockStoreClient(SOME_SALT, true, false)
	mockStoreFailure            = clients.NewMockStoreClient(SOME_SALT, false, true)
	rtr                         = mux.NewRouter()
	octopus                     = InitApi(mockConfig, mockShoreline, mockSeagullClient, mockeGatekeeperClient, mockStore)
	octopusNilToken             = InitApi(mockConfig, mockShorelineNilToken, mockSeagullClient, mockeGatekeeperClient, mockStore)
	octopusFail                 = InitApi(mockConfig, mockShoreline, mockSeagullClient, mockeGatekeeperClient, mockStoreFailure)
	octopusTokenNotServer       = InitApi(mockConfig, mockShorelineTokenNotServer, mockSeagullClient, mockeGatekeeperClient, mockStore)
)

func genReqRes() (request *http.Request, response *httptest.ResponseRecorder) {
	request, _ = http.NewRequest("GET", "/", nil)
	response = httptest.NewRecorder()
	return
}

func TestOctopusResponds(t *testing.T) {

	tests := []toTest{
		{
			// always returns a 200 if properly formed
			method:   "GET",
			url:      "/status",
			respCode: http.StatusOK,
		},
		{
			// always returns a 200 if properly formed
			method:   "GET",
			url:      "/upload/lastentry/" + FAKE_USER_ID,
			token:    FAKE_TOKEN,
			respCode: http.StatusOK,
		},

		{
			// always returns a 200 if properly formed
			method:   "GET",
			url:      "/upload/lastentry/" + FAKE_USER_ID + "/123-my-device-id",
			respCode: http.StatusOK,
			token:    FAKE_TOKEN,
		},
	}

	for idx, test := range tests {

		//fresh each time
		var testRtr = mux.NewRouter()

		if test.returnNone {
			octopusFindsNothing := InitApi(mockConfig, mockShoreline, mockSeagullClient, mockeGatekeeperClient, mockStoreEmpty)
			octopusFindsNothing.SetHandlers("", testRtr)
		} else {
			octopus := InitApi(mockConfig, mockShoreline, mockSeagullClient, mockeGatekeeperClient, mockStore)
			octopus.SetHandlers("", testRtr)
		}

		var body = &bytes.Buffer{}
		// build the body only if there is one defined in the test
		if len(test.body) != 0 {
			json.NewEncoder(body).Encode(test.body)
		}
		request, _ := http.NewRequest(test.method, test.url, body)
		if test.token != "" {
			request.Header.Set(SESSION_TOKEN, FAKE_TOKEN)
		}
		response := httptest.NewRecorder()
		testRtr.ServeHTTP(response, request)

		if response.Code != test.respCode {
			t.Fatalf("Test %d url: '%s'\nNon-expected status code %d (expected %d):\n\tbody: %v",
				idx, test.url, response.Code, test.respCode, response.Body)
		}

		if response.Body.Len() != 0 && len(test.response) != 0 {
			// compare bodies by comparing the unmarshalled JSON results
			var result = &jo{}

			if err := json.NewDecoder(response.Body).Decode(result); err != nil {
				t.Logf("Err decoding nonempty response body: [%v]\n [%v]\n", err, response.Body)
				return
			}

			if cmp := result.deepCompare(&test.response); cmp != "" {
				t.Fatalf("Test %d url: '%s'\n\t%s\n", idx, test.url, cmp)
			}
		}
	}
}

func TestGetStatus_StatusOk(t *testing.T) {
	req, res := genReqRes()
	octopus.GetStatus(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("Resp given [%s] expected [%s] ", res.Code, http.StatusOK)
	}
}

func TestGetStatus_StatusInternalServerError(t *testing.T) {
	req, res := genReqRes()
	octopusFail.GetStatus(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("Resp given [%s] expected [%s] ", res.Code, http.StatusInternalServerError)
	}
}

func TestTimeLastEntryUser_Success(t *testing.T) {
	req, res := genReqRes()
	octopus.TimeLastEntryUser(res, req, vars)
	if res.Code != http.StatusOK {
		t.Fatalf("Resp given [%s] expected [%s] ", res.Code, http.StatusOK)
	}
}

func TestTimeLastEntryUser_NilToken_StatusForbidden(t *testing.T) {
	req, res := genReqRes()
	octopusNilToken.TimeLastEntryUser(res, req, vars)
	if res.Code != http.StatusForbidden {
		t.Fatalf("Resp given [%s] expected [%s] ", res.Code, http.StatusForbidden)
	}
}

func TestTimeLastEntryUser_NotAuthorized_StatusForbidden(t *testing.T) {
	req, res := genReqRes()
	octopusTokenNotServer.TimeLastEntryUser(res, req, varsDifferent)
	if res.Code != http.StatusForbidden {
		t.Fatalf("Resp given [%s] expected [%s] ", res.Code, http.StatusForbidden)
	}
}

func TestTimeLastEntryUserAndDevice_Success(t *testing.T) {
	req, res := genReqRes()
	octopus.TimeLastEntryUserAndDevice(res, req, vars)
	if res.Code != http.StatusOK {
		t.Fatalf("Resp given [%s] expected [%s] ", res.Code, http.StatusOK)
	}
}

func TestTimeLastEntryUserAndDevice_NilToken_StatusForbidden(t *testing.T) {
	req, res := genReqRes()
	octopusNilToken.TimeLastEntryUserAndDevice(res, req, vars)
	if res.Code != http.StatusForbidden {
		t.Fatalf("Resp given [%s] expected [%s] ", res.Code, http.StatusForbidden)
	}
}
