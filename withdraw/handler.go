package withdraw

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"

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

//Request with invoice and token to be verified
type Request struct {
	Invoice    string
	Token      string
	Collection string
}

//Handler main handler for this lambda
//redeem a token and withdraw your sats!
func Handler(w http.ResponseWriter, r *http.Request) {
	//ugly check for options call
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
	if req.Invoice == "" || req.Collection == "" {
		http.Error(w, "Request or collection missing", http.StatusBadRequest)
		return

	}

	token, err := tdb.GetIfTokenAuthorized(req.Token, req.Collection)
	if err != nil {
		//not a real error possibly just wrong token
		logrus.Error(err.Error())
		logrus.WithField("token", req.Token).Info("Request coming in for bad token")
		http.Error(w, "Bad token", http.StatusUnauthorized)
		return
	}

	fiatAmt, err := getFiatAmt(req.Invoice, token)
	if err != nil {
		logrus.WithError(err).Error("error decoding invoice")
		http.Error(w, "something wrong", http.StatusInternalServerError)
		return
	}
	if fiatAmt != token.Value {
		logrus.WithField("token", req.Token).WithField("invoice amt", fiatAmt).WithField("token value", token.Value).Info("Request coming in for wrongly priced invoice")
		http.Error(w, "Value of invoice is wrong", http.StatusUnauthorized)
		return
	}
	count, err := tdb.SetTokenClaimed(req.Token, req.Invoice, fmt.Sprintf("%v", *r), req.Collection)
	if err != nil {
		logrus.Error(err.Error())
		http.Error(w, "Something wrong", http.StatusInternalServerError)
		return
	}
	if count != 1 {
		logrus.WithError(err).Error("SOMETHING FISHY GOING ON HERE")
		http.Error(w, "Something wrong", http.StatusInternalServerError)
		return

	}
	wd, err := on.Withdraw(req.Invoice)
	if err != nil {
		logrus.Error(err.Error())
		http.Error(w, "Something wrong", http.StatusInternalServerError)
		return
	}
	logrus.Info(wd)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
}

//first calculate the amount of sat, ask on for exchange rate, calculate value in euros
//round to nearest int
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
	floatFiatValue, err := on.GetFiatValue(int(satAmt), token.Currency)
	if err != nil {
		return 0, err
	}
	return int(math.Round(floatFiatValue)), nil
}
