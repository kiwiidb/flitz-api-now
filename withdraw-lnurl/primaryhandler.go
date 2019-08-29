package withdrawlnurl

import (
	"net/http"

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
	Tag             string //Always "withDrawRequest"
}

//PrimaryHandler main handler for this lambda
//redeem a token and withdraw your sats!
func PrimaryHandler(w http.ResponseWriter, r *http.Request) {

	//TODO
	//Extract token, collection from request route parameters
	//
	//look up if voucher is valid and active.
	//Get fiat amt from voucher => convert to sat amt
	//construct response and reply

}
