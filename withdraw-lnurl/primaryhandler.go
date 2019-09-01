package withdrawlnurl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kiwiidb/bliksem-library/opennode"
	"github.com/kiwiidb/bliksem-library/tokendb"
	"github.com/koding/multiconfig"
	"github.com/sirupsen/logrus"
)

var on opennode.LightningProvider
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
	on.Initialize(conf.OpenNodeAPIKey, conf.OpenNodeURL)
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

//PrimaryResponse as per LNURL spec for withdrawing
type PrimaryResponse struct {
	Callback        string
	K1              string
	MaxWithdrawable int
	MinWithdrawable int
	Tag             string
}

//PrimaryHandler main handler for this lambda
//redeem a token and withdraw your sats!
func PrimaryHandler(w http.ResponseWriter, r *http.Request) {
	collection, token, err := getCollectionAndToken(r.URL.Path)
	if err != nil {
		logrus.Error(err)
		resp := SecondaryResponse{
			Status: "ERROR",
			Reason: "Bad Request",
		}
		writeResponse(w, resp, http.StatusInternalServerError)
		return
	}
	authorized, euroValue, err := tdb.GetIfTokenAuthorized(token, collection)
	if err != nil {
		logrus.WithField("collection", collection).WithField("Token", token).Error(err)
		resp := SecondaryResponse{
			Status: "ERROR",
			Reason: "Internal Error",
		}
		writeResponse(w, resp, http.StatusInternalServerError)
		return
	}
	if !authorized {
		resp := SecondaryResponse{
			Status: "ERROR",
			Reason: "Token not valid or already claimed",
		}
		writeResponse(w, resp, http.StatusInternalServerError)
		return
	}
	//TODO add this whole thing to on library
	btcPrice, err := on.GetEuroRate()
	if err != nil {
		logrus.Error(err.Error())
		resp := SecondaryResponse{
			Status: "ERROR",
			Reason: "Error getting fiat rates",
		}
		writeResponse(w, resp, http.StatusInternalServerError)
		return
	}
	satoshiValue := int(float64(euroValue) / btcPrice * 1e8)
	secondaryRoute := fmt.Sprintf("%s/%s/%s", "/lnurl-secondary", collection, token)
	resp := PrimaryResponse{
		Callback:        fmt.Sprintf("https://%s%s", r.Host, secondaryRoute),
		K1:              "", //not needed
		MinWithdrawable: satoshiValue,
		MaxWithdrawable: satoshiValue,
		Tag:             "withdrawRequest",
	}
	writeResponse(w, resp, http.StatusOK)
	return
}

func writeResponse(w http.ResponseWriter, resp interface{}, status int) {
	respBytes, err := json.Marshal(resp)
	if err != nil {
		logrus.Error(err.Error())
		http.Error(w, "something wrong", http.StatusInternalServerError)
		return
	}
	w.Write(respBytes)
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
}
func getCollectionAndToken(path string) (collection string, token string, err error) {
	//path is /{lnurl-primary,lnurl-secondary}/collection/token, so 2 and 3
	splittedRoute := strings.Split(path, "/")
	if len(splittedRoute) < 3 {
		return "", "", fmt.Errorf("Wrong number of route parameters in url %s", path)
	}
	return strings.Split(path, "/")[2], strings.Split((path), "/")[3], nil
}
