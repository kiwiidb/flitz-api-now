package withdraw

import (
	"encoding/json"
	"net/http"

	"github.com/kiwiidb/bliksem-library/tokendb"
	"github.com/koding/multiconfig"
	"github.com/sirupsen/logrus"
)

var tdb tokendb.TokenDataBaseInterface

func init() {
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

//Handler main handler for this lambda
//for gathering info about who comes to our site
func Handler(w http.ResponseWriter, r *http.Request) {
	tdb.AddEntryToCollection(r, "metrics")	
	w.WriteHeader(http.StatusOK)
	return
}