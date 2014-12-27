package api

/*import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"../clients"
	commonClients "github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/clients/shoreline"
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
)

const (
	TOKEN_FOR_UID1 = "123-my-token"
	SOME_SALT      = "no-thanks"
	MAKE_IT_FAIL   = true
	RETURN_NOTHING = true
)

var (
	NO_PARAMS = map[string]string{}

	FAKE_CONFIG = Config{
		ServerSecret: "shhh! don't tell",
	}
	// Mocks
	mockShoreline  = shoreline.NewMock(FAKE_TOKEN)
	mockGatekeeper = commonClients.NewGatekeeperMock(nil, nil)
	mockSeagull    = commonClients.NewSeagullMock(`{}`, nil)
	// Stores
	mockStore      = clients.NewMockStoreClient(SOME_SALT, false, false)
	mockStoreEmpty = clients.NewMockStoreClient(SOME_SALT, RETURN_NOTHING, false)
	mockStoreFails = clients.NewMockStoreClient(SOME_SALT, false, MAKE_IT_FAIL)
)

func TestOctopusResponds(t *testing.T) {

	tests := []toTest{
		{
			// always returns a 200 if properly formed
			method:   "POST",
			url:      "/query",
			respCode: http.StatusNotImplemented,
			token:    TOKEN_FOR_UID1,
			body:     "METAQUERY WHERE userid IS \"12d7bc90fa\" QUERY TYPE IN update SORT BY time AS Timestamp REVERSED",
		},
		{
			method:     "POST",
			url:        "/query",
			respCode:   http.StatusNotImplemented,
			returnNone: true,
			token:      TOKEN_FOR_UID1,
			body:       "METAQUERY WHERE userid IS \"12d7bc90fa\" QUERY TYPE IN cbg, smbg SORT BY time AS Timestamp REVERSED",
		},
		{
			//no data given
			method:   "POST",
			url:      "/query",
			respCode: http.StatusNotImplemented,
			token:    TOKEN_FOR_UID1,
			body:     "METAQUERY WHERE userid IS \"12d7bc90fa\" QUERY TYPE IN foo SORT BY time AS Timestamp REVERSED",
		},
	}

	for idx, test := range tests {

		//fresh each time
		var testRtr = mux.NewRouter()

		if test.returnNone {
			octopusFindsNothing := InitApi(FAKE_CONFIG, shorelineClient, seagullClient, gatekeeperClient, mockStoreEmpty)
			octopusFindsNothing.SetHandlers("", testRtr)
		} else {
			octopus := InitApi(FAKE_CONFIG, shorelineClient, seagullClient, gatekeeperClient, mockStore)
			octopus.SetHandlers("", testRtr)
		}

		var body = &bytes.Buffer{}
		// build the body only if there is one defined in the test
		if len(test.body) != 0 {
			json.NewEncoder(body).Encode(test.body)
		}
		request, _ := http.NewRequest(test.method, test.url, body)
		if test.token != "" {
			request.Header.Set(TP_SESSION_TOKEN, FAKE_TOKEN)
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
}*/
