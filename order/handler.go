package balance

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kiwiidb/bliksem-library/opennode"
	"github.com/koding/multiconfig"

	"github.com/sirupsen/logrus"
)

var on *opennode.OpenNode
var conf Config

//Config for both tokens database and opennode api
type Config struct {
	OpenNodeURL         string
	OpenNodeReadAPIKey  string
	CallBackURLTemplate string
}

//OrderRequest what you want to order
type OrderRequest struct {
	Amt      int    //amt of vouchers
	Value    int    //value in currency
	Currency string // EUR or USD
	Email    string //where to send vouchers to
}

func init() {
	//init opennode
	conf = Config{}
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
	cbURL := fmt.Sprintf(conf.CallBackURLTemplate, req.Value, req.Amt, req.Currency, req.Email)
	ch := opennode.Charge{
		CallbackURL: cbURL,
		Amount:      float64(req.Amt * req.Value),
		Currency:    req.Currency,
		Email:       req.Email,
		Description: "Flitz cards order",
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
