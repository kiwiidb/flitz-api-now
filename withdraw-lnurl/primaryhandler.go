package withdrawlnurl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/kiwiidb/bliksem-library/opennode"
	"github.com/kiwiidb/bliksem-library/tokendb"
	"github.com/koding/multiconfig"
	"github.com/sirupsen/logrus"
)

//MODELS

//PrimaryResponse as per LNURL spec for withdrawing
type PrimaryResponse struct {
	Callback           string `json:"callback"`
	K1                 string `json:"k1"`
	MaxWithdrawable    int    `json:"maxWithdrawable"`
	MinWithdrawable    int    `json:"minWithdrawable"`
	Tag                string `json:"tag"`
	DefaultDescription string `json:"defaultDescription"`
}

//SecondaryResponse as per LNURL spec for withdrawing
type SecondaryResponse struct {
	Status string
	Reason string
}

//Config for both tokens database and opennode api
type Config struct {
	OpenNodeURL    string
	OpenNodeAPIKey string
}

//main state/config providers for this lambda
var on opennode.LightningProvider
var tdb tokendb.TokenDataBaseInterface

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

//PrimaryHandler main handler for this lambda
//redeem a token and withdraw your sats!
func PrimaryHandler(w http.ResponseWriter, r *http.Request) {
	collection, token, err := getCollectionAndToken(r.URL.Path)
	if err != nil {
		logrus.Error(err)
		writeErrorResponse(w, "Bad Request", http.StatusInternalServerError)
		return
	}
	t, err := tdb.GetIfTokenAuthorized(token, collection)
	if err != nil {
		logrus.WithField("collection", collection).WithField("Token", token).Error(err)
		writeErrorResponse(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	//TODO add this whole thing to on library
	satValue, err := on.GetSatoshiValue(t.Value, t.Currency)
	if err != nil {
		logrus.Error(err.Error())
		writeErrorResponse(w, "Error getting fiat rates", http.StatusInternalServerError)
		return
	}
	milliSatoshiValue := satValue * 1e3
	secondaryRoute := fmt.Sprintf("%s/%s/%s", "/lnurl-secondary", collection, token)
	resp := PrimaryResponse{
		Callback:           fmt.Sprintf("https://%s%s", r.Host, secondaryRoute),
		K1:                 "", //not needed
		MinWithdrawable:    milliSatoshiValue,
		MaxWithdrawable:    milliSatoshiValue,
		Tag:                "withdrawRequest",
		DefaultDescription: "Redeem Flitz Voucher",
	}
	writeResponse(w, resp, http.StatusOK)
	return
}

//HELPER FUNCTIONS
func getCollectionAndToken(path string) (collection string, token string, err error) {
	//path is /{lnurl-primary,lnurl-secondary}/collection/token, so 2 and 3
	splittedRoute := strings.Split(path, "/")
	if len(splittedRoute) < 3 {
		return "", "", fmt.Errorf("Wrong number of route parameters in url %s", path)
	}
	return strings.Split(path, "/")[2], strings.Split((path), "/")[3], nil
}
func writeErrorResponse(w http.ResponseWriter, message string, status int) {
	resp := SecondaryResponse{
		Status: "ERROR",
		Reason: message,
	}
	writeResponse(w, resp, status)
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

//first calculate the amount of sat, ask on for exchange rate, calculate value in euros
func getFiatAmt(invoice string, token tokendb.Token) (int, error) {
	regexpp, _ := regexp.Compile("[0-9]+[munp]")
	amt := regexpp.FindString(invoice)
	intAmt, err := strconv.Atoi(amt[:len(amt)-1])
	if err != nil {
		return 0, err
	}
	suffix := amt[len(amt)-1]

	helperdict := map[string]float64{"m": 1e5, "u": 1e2, "n": 1e-1, "p": 1e-4}
	satAmt := float64(intAmt) * helperdict[string(suffix)]
	logrus.WithField("satamt", satAmt).Info("decoding invoice")
	return on.GetFiatValue(int(satAmt), token.Currency)
}
