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
	"log"
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
		skip          bool
		returnNone    bool
		returnFailure bool
		method        string
		url           string
		body          string
		token         string
		respCode      int
		response      jo
	}
	// These two types make it easier to define blobs of json inline.
	// We don't use the types defined by the API because we want to
	// be able to test with partial data structures.
	// jo is a generic json object
	jo map[string]interface{}

	// and ja is a generic json array
	ja []interface{}

	MockShorelineClient struct {
		tokenSeenAsValid bool
		tokenData        shoreline.TokenData
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
	log.Print("MockShorelineClient.CheckToken")
	if slc.tokenSeenAsValid {
		return &slc.tokenData
	} else {
		return nil
	}
}

func (slc MockShorelineClient) TokenProvide() string {
	log.Print("MockShorelineClient.TokenProvide")
	return FAKE_TOKEN
}

func (slc MockShorelineClient) GetUser(userID, token string) (*shoreline.UserData, error) {
	log.Print("MockShorelineClient.GetUser")
	return &shoreline.UserData{UserID: userID, UserName: userID, Emails: []string{userID}}, nil
}

func (sgc MockSeagullClient) GetPrivatePair(userID, hashName, token string) *commonClients.PrivatePair {
	log.Print("MockSeagullClient.GetPrivatePair")
	return &commonClients.PrivatePair{ID: FAKE_GROUP_ID, Value: FAKE_VALUE}
}

func (gkc MockGateKeeperClient) UserInGroup(userID, groupID string) (map[string]commonClients.Permissions, error) {
	log.Print("MockGateKeeperClient.UserInGroup")
	//we give the user `view` permissons
	permissonsToReturn := make(map[string]commonClients.Permissions)
	p := make(commonClients.Permissions)
	p["userid"] = userID
	permissonsToReturn["view"] = p
	return permissonsToReturn, nil
}

var (
	mockConfig    = Config{ServerSecret: "shhh! don't tell"}
	vars          = httpVars{"userID": FAKE_USER_ID}
	varsDifferent = httpVars{"userID": FAKE_USER_ID_DIFFERENT}
	//shoreline mocks
	mockShoreline = MockShorelineClient{
		tokenSeenAsValid: true,
		tokenData:        shoreline.TokenData{UserID: FAKE_USER_ID, IsServer: true},
	}
	mockShorelineNilToken = MockShorelineClient{
		tokenSeenAsValid: false,
		tokenData:        shoreline.TokenData{},
	}
	mockShorelineForbiddenToken = MockShorelineClient{
		tokenSeenAsValid: false,
		tokenData:        shoreline.TokenData{UserID: FAKE_USER_ID, IsServer: false},
	}
	mockSeagullClient = MockSeagullClient{}
	//gatekeeper mocks
	mockeGatekeeperClient = MockGateKeeperClient{}
	//storage client mocks
	mockStore        = clients.NewMockStoreClient(SOME_SALT, false, false)
	mockStoreEmpty   = clients.NewMockStoreClient(SOME_SALT, true, false)
	mockStoreFailure = clients.NewMockStoreClient(SOME_SALT, false, true)
	rtr              = mux.NewRouter()
	//api's
	octopus               = InitApi(mockConfig, mockShoreline, mockSeagullClient, mockeGatekeeperClient, mockStore)
	octopusNoToken        = InitApi(mockConfig, mockShorelineNilToken, mockSeagullClient, mockeGatekeeperClient, mockStore)
	octopusFail           = InitApi(mockConfig, mockShoreline, mockSeagullClient, mockeGatekeeperClient, mockStoreFailure)
	octopusTokenForbidden = InitApi(mockConfig, mockShorelineForbiddenToken, mockSeagullClient, mockeGatekeeperClient, mockStore)
)

func genReqRes(tokenString string) (request *http.Request, response *httptest.ResponseRecorder) {
	request, _ = http.NewRequest("GET", "/", nil)
	request.Header.Set(SESSION_TOKEN, tokenString)
	response = httptest.NewRecorder()
	return
}

func TestOctopusResponds(t *testing.T) {

	tests := []toTest{

		//
		// service status response
		//
		{
			method: "GET", url: "/status",
			respCode: http.StatusOK,
		},
		{
			method: "GET", url: "/status",
			returnFailure: true,
			respCode:      http.StatusInternalServerError,
		},

		//
		// lastentry for user response
		//
		{
			method: "GET", url: "/upload/lastentry/" + FAKE_USER_ID,
			token:    FAKE_TOKEN,
			respCode: http.StatusOK,
		},
		{
			method: "GET", url: "/upload/lastentry/" + FAKE_USER_ID,
			token:    "",
			respCode: http.StatusUnauthorized,
		},
		{
			method: "GET", url: "/upload/lastentry/" + FAKE_USER_ID_DIFFERENT,
			token:    FAKE_TOKEN,
			respCode: http.StatusForbidden,
		},

		//
		// lastentry for user and device response
		//
		{
			method: "GET", url: "/upload/lastentry/" + FAKE_USER_ID + "/123-my-device-id",
			token:    FAKE_TOKEN,
			respCode: http.StatusOK,
		},

		// data query
		{method: "POST", url: "/data",
			token: FAKE_TOKEN, body: "METAQUERY WHERE userid IS 12d7bc90fa QUERY TYPE IN update SORT BY time AS Timestamp REVERSED",
			respCode: http.StatusOK},

		{method: "POST", url: "/data",

			returnNone: true, token: FAKE_TOKEN, body: "METAQUERY WHERE userid IS 12d7bc90fa QUERY TYPE IN cbg, smbg SORT BY time AS Timestamp REVERSED",
			respCode: http.StatusOK,
		},
		{method: "POST", url: "/data",

			token:    FAKE_TOKEN,
			respCode: http.StatusBadRequest,
		},
		{method: "POST", url: "/data",

			token: FAKE_TOKEN, body: "blah balh blah",
			respCode: http.StatusBadRequest,
		},
	}

	for idx, test := range tests {

		//fresh each time
		var testRtr = mux.NewRouter()

		if test.returnNone {
			octopus := InitApi(mockConfig, mockShoreline, mockSeagullClient, mockeGatekeeperClient, mockStoreEmpty)
			octopus.SetHandlers("", testRtr)
		} else if test.returnFailure {
			octopus := InitApi(mockConfig, mockShoreline, mockSeagullClient, mockeGatekeeperClient, mockStoreFailure)
			octopus.SetHandlers("", testRtr)
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

/*func TestGetStatus_StatusOk(t *testing.T) {
	req, res := genReqRes("")
	octopus.GetStatus(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusOK)
	}
}*/

/*func TestGetStatus_StatusInternalServerError(t *testing.T) {
	req, res := genReqRes("")
	octopusFail.GetStatus(res, req)

	request, _ := http.NewRequest(test.method, test.url, body)
	if test.token != "" {
		request.Header.Set(SESSION_TOKEN, FAKE_TOKEN)
	}
	response := httptest.NewRecorder()
	testRtr.ServeHTTP(response, request)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusInternalServerError)
	}
}

func TestTimeLastEntryUser_Success(t *testing.T) {
	req, res := genReqRes(FAKE_TOKEN)
	octopus.TimeLastEntryUser(res, req, vars)
	if res.Code != http.StatusOK {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusOK)
	}
}

func TestTimeLastEntryUser_NilToken_StatusUnauthorized(t *testing.T) {
	req, res := genReqRes("")
	octopusNoToken.TimeLastEntryUser(res, req, vars)
	//no token so this is not authorized
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusUnauthorized)
	}
}*/

func TestTimeLastEntryUser_NotAuthorized_StatusForbidden(t *testing.T) {
	req, res := genReqRes(FAKE_TOKEN)
	octopusTokenForbidden.TimeLastEntryUser(res, req, varsDifferent)
	if res.Code != http.StatusForbidden {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusForbidden)
	}
}

func TestTimeLastEntryUserAndDevice_Success(t *testing.T) {
	req, res := genReqRes(FAKE_TOKEN)
	octopus.TimeLastEntryUserAndDevice(res, req, vars)
	if res.Code != http.StatusOK {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusOK)
	}
}

func TestTimeLastEntryUserAndDevice_NilToken_StatusUnauthorized(t *testing.T) {
	req, res := genReqRes(FAKE_TOKEN)
	octopusNoToken.TimeLastEntryUserAndDevice(res, req, vars)
	if res.Code != http.StatusForbidden {
		t.Fatalf("Resp given [%d] expected [%d] ", res.Code, http.StatusUnauthorized)
	}
}
