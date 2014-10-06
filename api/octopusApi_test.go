package api

import (
	"./../clients"
	"github.com/gorilla/mux"
	//"github.com/tidepool-org/go-common"
	commonClients "github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/clients/disc"
	"github.com/tidepool-org/go-common/clients/mongo"
	"github.com/tidepool-org/go-common/clients/shoreline"
	//"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	THE_SECRET   = "shhh! don't tell"
	MAKE_IT_FAIL = true
)

type TestConfig struct {
	commonClients.Config
	Service disc.ServiceListing `json:"service"`
	Mongo   mongo.Config        `json:"mongo"`
	Api     Config              `json:"octopus"`
}

type MockShorelineClient struct{}

func (slc MockShorelineClient) CheckToken(token string) *shoreline.TokenData {
	return &shoreline.TokenData{"yolo", true}
}

func (slc MockShorelineClient) TokenProvide() string {
	return "yolo"
}

type MockSeagullClient struct{}

func (sgc MockSeagullClient) GetPrivatePair(userID, hashName, token string) *commonClients.PrivatePair {
	return &commonClients.PrivatePair{"yolo", "yolo"}
}

type MockGateKeeperClient struct{}

func (gkc MockGateKeeperClient) UserInGroup(userID, groupID string) (map[string]commonClients.Permissions, error) {
	return nil, nil
}

var (
	config           TestConfig
	shorelineClient  = MockShorelineClient{}
	seagullClient    = MockSeagullClient{}
	gatekeeperClient = MockGateKeeperClient{}
	store            = clients.NewMockStoreClient("salty", false, false)
	storeFail        = clients.NewMockStoreClient("salty", false, true)
	rtr              = mux.NewRouter()
	octopus          = InitApi(config.Api, shorelineClient, seagullClient, gatekeeperClient, store)
	octopuFail       = InitApi(config.Api, shorelineClient, seagullClient, gatekeeperClient, storeFail)
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

func TestGetLastEntry_Success(t *testing.T) {

	// The request isn't actually used
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	vars := make(map[string]string)
	vars["userID"] = "old greg"
	vars["deviceID"] = "johndeer"

	octopus.GetLastEntry(response, request, vars)

	if response.Code != http.StatusOK {
		t.Fatalf("Resp given [%s] expected [%s] ", response.Code, http.StatusOK)
	}
}
