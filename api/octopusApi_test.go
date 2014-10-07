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
	FAKE_USER_ID  = "oldgreg"
)

type TestConfig struct {
	commonClients.Config
	Service disc.ServiceListing `json:"service"`
	Mongo   mongo.Config        `json:"mongo"`
	Api     Config              `json:"octopus"`
}

type MockShorelineClient struct {
	validToken bool
	token      shoreline.TokenData
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

type MockSeagullClient struct{}

func (sgc MockSeagullClient) GetPrivatePair(userID, hashName, token string) *commonClients.PrivatePair {
	return &commonClients.PrivatePair{FAKE_GROUP_ID, FAKE_VALUE}
}

type MockGateKeeperClient struct{}

func (gkc MockGateKeeperClient) UserInGroup(userID, groupID string) (map[string]commonClients.Permissions, error) {
	return nil, nil
}

var (
	config           TestConfig
	vars             = httpVars{"userId": FAKE_USER_ID}
	tokenIsServer    = shoreline.TokenData{FAKE_TOKEN, true}
	tokenIsNotServer = shoreline.TokenData{FAKE_TOKEN, false}
	shorelineClient  = MockShorelineClient{true, tokenIsServer}
	slcNilToken      = MockShorelineClient{false, tokenIsServer}
	seagullClient    = MockSeagullClient{}
	gatekeeperClient = MockGateKeeperClient{}
	store            = clients.NewMockStoreClient("salty", false, false)
	storeFail        = clients.NewMockStoreClient("salty", false, true)
	rtr              = mux.NewRouter()
	octopus          = InitApi(config.Api, shorelineClient, seagullClient, gatekeeperClient, store)
	octopusNilToken  = InitApi(config.Api, slcNilToken, seagullClient, gatekeeperClient, store)
	octopusFail      = InitApi(config.Api, shorelineClient, seagullClient, gatekeeperClient, storeFail)
)

func genReqRes() (request *http.Request, response *httptest.ResponseRecorder) {
	request, _ = http.NewRequest("GET", "/", nil)
	response = httptest.NewRecorder()
	return
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
