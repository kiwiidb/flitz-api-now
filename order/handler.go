package balance

import (
	"encoding/json"
	"net/http"

	"github.com/kiwiidb/bliksem-library/opennode"
	"github.com/koding/multiconfig"

	"github.com/sirupsen/logrus"
)

var on *opennode.OpenNode

//Config for both tokens database and opennode api
type Config struct {
	OpenNodeURL        string
	OpenNodeReadAPIKey string
}

//OrderRequest what you want to order
type OrderRequest struct {
	Amt      int    //amt in currency
	Currency string // EUR or USD
	Email    string //where to send vouchers to
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
	on.APIKey = conf.OpenNodeReadAPIKey

}

//DepositHandler main handler for this lambda
//deposit funds into OpenNode, only with firebase auth
func DepositHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	req := OrderRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logrus.Error(err.Error())
		http.Error(w, "Something wrong", http.StatusInternalServerError)
		return
	}
	ch := opennode.Charge{
		CallbackURL: "https://flitz-order-processor.kwintendebacker.now.sh/webhook?this=that&test=random",
		Amount:      float64(req.Amt),
		Currency:    req.Currency,
		Email:       req.Email,
		Description: "testtest",
	}
	chargeResp, err := on.CreateChargeAdvanced(ch)
	if err != nil {
		logrus.WithError(err).Info("something wrong")
		http.Error(w, "something wrong", http.StatusInternalServerError)
		return
	}

	respBytes, err := json.Marshal(&chargeResp)
	if err != nil {
		logrus.WithError(err).Info("something wrong")
		http.Error(w, "something wrong", http.StatusInternalServerError)
		return
	}
	w.Write(respBytes)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	return
}
