package api

import (
	"./../clients"
	"github.com/gorilla/mux"
	commonClients "github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/clients/disc"
	"github.com/tidepool-org/go-common/clients/mongo"
	"github.com/tidepool-org/go-common/clients/shoreline"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	FAKE_TOKEN    = "yolo"
	FAKE_GROUP_ID = "abcdefg"
	FAKE_VALUE    = "gfedcba"
)

type TestConfig struct {
	commonClients.Config
	Service disc.ServiceListing `json:"service"`
	Mongo   mongo.Config        `json:"mongo"`
	Api     Config              `json:"octopus"`
}

type MockShorelineClient struct {
	validToken bool
}

func (slc MockShorelineClient) CheckToken(token string) *shoreline.TokenData {
	if slc.validToken {
		return &shoreline.TokenData{FAKE_TOKEN, true}
	} else {
		return nil
	}
}

func (slc MockShorelineClient) TokenProvide() string {
	return FAKE_TOKEN
}

type MockSeagullClient struct{}

func (sgc MockSeagullClient) GetPrivatePair(userID, hashName, token string) *commonClients.PrivatePair {
	return &commonClients.PrivatePair{FAKE_GROUP_ID, FAKE_VALUE}
}

type MockGateKeeperClient struct{}

func (gkc MockGateKeeperClient) UserInGroup(userID, groupID string) (map[string]commonClients.Permissions, error) {
	return nil, nil
}

var (
	config                TestConfig
	shorelineClient       = MockShorelineClient{true}
	shorelineClientNoAuth = MockShorelineClient{false}
	seagullClient         = MockSeagullClient{}
	gatekeeperClient      = MockGateKeeperClient{}
	store                 = clients.NewMockStoreClient("salty", false, false)
	storeFail             = clients.NewMockStoreClient("salty", false, true)
	rtr                   = mux.NewRouter()
	octopus               = InitApi(config.Api, shorelineClient, seagullClient, gatekeeperClient, store)
	octopusNoAuth         = InitApi(config.Api, shorelineClientNoAuth, seagullClient, gatekeeperClient, store)
	octopusFail           = InitApi(config.Api, shorelineClient, seagullClient, gatekeeperClient, storeFail)
)

func TestGetStatus_StatusOk(t *testing.T) {

	// The request isn't actually used
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	octopus.GetStatus(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Resp given [%s] expected [%s] ", response.Code, http.StatusOK)
	}
}

func TestGetStatus_StatusInternalServerError(t *testing.T) {

	// The request isn't actually used
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	octopusFail.GetStatus(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("Resp given [%s] expected [%s] ", response.Code, http.StatusInternalServerError)
	}
}

func TestTimeLastEntryUser_Success(t *testing.T) {

	// The request isn't actually used
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	vars := make(httpVars)
	vars["userID"] = "old greg"

	octopus.TimeLastEntryUser(response, request, vars)

	if response.Code != http.StatusOK {
		t.Fatalf("Resp given [%s] expected [%s] ", response.Code, http.StatusOK)
	}
}

func TestTimeLastEntryUser_StatusForbidden(t *testing.T) {
	// The request isn't actually used
	req, _ := http.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	vars := make(httpVars)
	vars["userID"] = "oldgreg"

	octopusNoAuth.TimeLastEntryUser(res, req, vars)
	if res.Code != http.StatusForbidden {
		t.Fatalf("Resp given [%s] expected [%s] ", res.Code, http.StatusForbidden)
	}
}
