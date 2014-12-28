package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestQueryResponds(t *testing.T) {

	tests := []toTest{
		{
			// always returns a 200 if properly formed
			method:   "POST",
			url:      "/query",
			respCode: http.StatusNotImplemented,
			token:    FAKE_TOKEN,
			body:     "METAQUERY WHERE userid IS \"12d7bc90fa\" QUERY TYPE IN update SORT BY time AS Timestamp REVERSED",
		},
		{
			method:     "POST",
			url:        "/query",
			respCode:   http.StatusNotImplemented,
			returnNone: true,
			token:      FAKE_TOKEN,
			body:       "METAQUERY WHERE userid IS \"12d7bc90fa\" QUERY TYPE IN cbg, smbg SORT BY time AS Timestamp REVERSED",
		},
		{
			//no data given
			method:   "POST",
			url:      "/query",
			respCode: http.StatusNotImplemented,
			token:    FAKE_TOKEN,
			body:     "METAQUERY WHERE userid IS \"12d7bc90fa\" QUERY TYPE IN foo SORT BY time AS Timestamp REVERSED",
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

func TestQuery(t *testing.T) {
	var jsonData = []byte(`"METAQUERY WHERE userid IS 12d7bc90fa QUERY TYPE IN update SORT BY time AS Timestamp REVERSED"`)

	req, _ := http.NewRequest("POST", "/query", bytes.NewBuffer(jsonData))
	req.Header.Add("content-type", "application/json")

	res := httptest.NewRecorder()

	octopus.Query(res, req)
	if res.Code != http.StatusNotImplemented {
		t.Fatalf("Resp given [%s] expected [%s] ", res.Code, http.StatusNotImplemented)
	}
}