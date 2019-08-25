package withdraw

import (
	"encoding/json"
	"net/http"

	"github.com/kiwiidb/bliksem-library/opennode"
	"github.com/kiwiidb/bliksem-library/tokendb"
	"github.com/koding/multiconfig"
	"github.com/sirupsen/logrus"
)

var on *opennode.OpenNode
var tdb tokendb.TokenDataBaseInterface

//Config for both tokens database and opennode api
type Config struct {
	OpenNodeURL    string
	OpenNodeAPIKey string
}

func init() {
	//init opennode
	conf := Config{}
	m := multiconfig.EnvironmentLoader{}
	err := m.Load(&conf)
	if err != nil {
		logrus.Fatal(err)
	}
	m.PrintEnvs(conf)
	logrus.Info(conf)
	on = &opennode.OpenNode{}
	on.BaseURL = conf.OpenNodeURL
	on.APIKey = conf.OpenNodeAPIKey
	logrus.Info(on)
	//init tokendb
	tdb = &tokendb.TokenDB{}
	tdbconf := tokendb.Config{}
	m = multiconfig.EnvironmentLoader{}
	err = m.Load(&tdbconf)
	if err != nil {
		logrus.Fatal(err)
	}
	m.PrintEnvs(conf)
	logrus.Info(conf)
	err = tdb.Initialize(tdbconf)
	if err != nil {
		logrus.Fatal(err)
	}
}

//Request with invoice and token to be verified
type Request struct {
	Token          string
	CollectionName string
}

//Handler main handler for this lambda
//redeem a token and withdraw your sats!
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	req := Request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logrus.Error(err.Error())
		http.Error(w, "Something wrong", http.StatusInternalServerError)
		return
	}
	if req.CollectionName == "" {
		http.Error(w, "Request or value missing", http.StatusBadRequest)
		return

	}
	authorized, euroValue, err := tdb.GetIfTokenAuthorized(req.Token, req.CollectionName)
	if err != nil {
		//not a real error possibly just wrong token
		logrus.Error(err.Error())
		logrus.WithField("token", req.Token).Info("Request coming in for bad token")
		http.Error(w, "Bad token", http.StatusUnauthorized)
		return
	}
	if !authorized {
		logrus.WithField("token", req.Token).Info("Request coming in for unauthorized token")
		http.Error(w, "Token already claimed", http.StatusUnauthorized)
		return
	}
	btcPrice, err := on.GetEuroRate()
	if err != nil {
		logrus.Error(err.Error())
		http.Error(w, "something wrong", http.StatusInternalServerError)
		return
	}
	satoshiValue := int(float64(euroValue) / btcPrice * 1e8)
	type rateresp struct {
		Authorized   bool
		SatoshiValue int
	}
	resp := rateresp{Authorized: true, SatoshiValue: satoshiValue}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		logrus.Error(err.Error())
		http.Error(w, "something wrong", http.StatusInternalServerError)
		return
	}
	w.Write(respBytes)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	return
}
