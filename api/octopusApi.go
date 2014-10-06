package api

import (
	"./../clients"
	//"fmt"
	"github.com/gorilla/mux"
	commonClients "github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/clients/shoreline"
	"log"
	"net/http"
)

type ShorelineInterface interface {
	CheckToken(token string) *shoreline.TokenData
	TokenProvide() string
}

type SeagullInterface interface {
	GetPrivatePair(userID, hashName, token string) *commonClients.PrivatePair
}

type GatekeeperInterface interface {
	UserInGroup(userID, groupID string) (map[string]commonClients.Permissions, error)
}

type MongoInterface interface {
}

type (
	Api struct {
		Store            clients.StoreClient
		ShorelineClient  ShorelineInterface
		SeagullClient    SeagullInterface
		GatekeeperClient GatekeeperInterface
		Config           Config
	}
	Config struct {
		ServerSecret string `json:"serverSecret"` //used for services
		LongTermKey  string `json:"longTermKey"`
		Salt         string `json:"salt"`      //used for pw
		Secret       string `json:"apiSecret"` //used for token
	}
	varsHandler func(http.ResponseWriter, *http.Request, map[string]string)
)

func (a *Api) userCanViewData(userID, groupID string) bool {
	if userID == groupID {
		return true
	}

	perms, err := a.GatekeeperClient.UserInGroup(userID, groupID)
	if err != nil {
		log.Println("Error looking up user in group", err)
		return false
	}

	log.Println(perms)
	return !(perms["root"] == nil && perms["view"] == nil)
}

func InitApi(cfg Config, slc ShorelineInterface,
	sgc SeagullInterface, gkc GatekeeperInterface,
	store clients.StoreClient) *Api {

	return &Api{
		Store:            store,
		ShorelineClient:  slc,
		SeagullClient:    sgc,
		GatekeeperClient: gkc,
		Config:           cfg,
	}
}

func (a *Api) SetHandlers(prefix string, rtr *mux.Router) {
	rtr.HandleFunc("/status", a.GetStatus).Methods("GET")
	rtr.Handle("/upload/lastentry/{userID}/{deviceID}", varsHandler(a.GetLastEntry)).Methods("GET")
}

func (a *Api) GetStatus(res http.ResponseWriter, req *http.Request) {
	if err := a.Store.Ping(); err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}

func (a *Api) GetLastEntry(res http.ResponseWriter, req *http.Request, vars map[string]string) {
	userID := vars["userID"]
	deviceId := vars["deviceID"]

	token := req.Header.Get("x-tidepool-session-token")
	td := a.ShorelineClient.CheckToken(token)

	if td == nil || !(td.IsServer || td.UserID == userID || a.userCanViewData(td.UserID, userID)) {
		res.WriteHeader(http.StatusForbidden)
		return
	}

	pair := a.SeagullClient.GetPrivatePair(userID, "uploads", a.ShorelineClient.TokenProvide())
	if pair == nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	groupId := pair.ID

	timeLastEntry := a.Store.GetTimeLastEntry(groupId, deviceId)

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(timeLastEntry))
}

func (h varsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	h(res, req, vars)
}
