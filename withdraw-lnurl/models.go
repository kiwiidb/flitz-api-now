package withdrawlnurl

import (
	"github.com/kiwiidb/bliksem-library/opennode"
	"github.com/kiwiidb/bliksem-library/tokendb"
	"github.com/koding/multiconfig"
	"github.com/sirupsen/logrus"
)

//PrimaryResponse as per LNURL spec for withdrawing
type PrimaryResponse struct {
	Callback        string
	K1              string
	MaxWithdrawable int
	MinWithdrawable int
	Tag             string
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
