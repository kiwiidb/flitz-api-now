package withdraw

import (
	"net/http"
	"time"

	"github.com/kiwiidb/bliksem-library/tokendb"
	"github.com/koding/multiconfig"
	"github.com/sirupsen/logrus"
)

var tdb tokendb.TokenDataBaseInterface

func init() {
	tdb = &tokendb.TokenDB{}
	tdbconf := tokendb.Config{}
	m := multiconfig.EnvironmentLoader{}
	err := m.Load(&tdbconf)
	if err != nil {
		logrus.Fatal(err)
	}
	err = tdb.Initialize(tdbconf)
	if err != nil {
		logrus.Fatal(err)
	}
}

//Handler main handler for this lambda
//for gathering info about who comes to our site
func Handler(w http.ResponseWriter, r *http.Request) {
	type Metric struct {
		Header http.Header
		TimeStamp time.Time
	}
	err := tdb.AddEntryToCollection(Metric{r.Header, time.Now()}, "metrics")
	if err != nil {
		logrus.WithError(err).Error("error adding metric to database")
	}
	w.WriteHeader(http.StatusOK)
	return
}
