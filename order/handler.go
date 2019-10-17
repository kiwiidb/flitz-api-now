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
	Value    int    //What the customer is buying, NOT the real value of what the voucher will be redeemed for
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

//Handler main handler for this lambda
//handle order, add charge to OpenNode
func Handler(w http.ResponseWriter, r *http.Request) {
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
	voucherValue, fiatPrice, err := calculateFiatPriceAndVoucherValue(req)
	if err != nil {
		logrus.Error(err.Error())
		http.Error(w, "Something wrong", http.StatusInternalServerError)
		return
	}
	cbURL := fmt.Sprintf(conf.CallBackURLTemplate, voucherValue, req.Amt, req.Currency, req.Email, req.Value) //req.Value is "price" in the cbURL template
	logrus.Info(cbURL)
	ch := opennode.Charge{
		CallbackURL: cbURL,
		Amount:      fiatPrice * float64(req.Amt),
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

//logic for how much profit we are making
//1 dollar/euro per voucher
//bulk vouchers have a value = face value - 2 split between reseller and us, so price for reseller is order value -1
//single vouchers have value = face value, so buyer needs to add 1 dollar/euro
//you can't buy 0 or a negative amount of vouchers
//TODO: move to seperate service
func calculateFiatPriceAndVoucherValue(order OrderRequest) (voucherValue int, fiatPrice float64, err error) {
	if order.Amt == 1 {
		if order.Value < 10 {
			//this is free of charge
			return order.Value, float64(order.Value), nil
		}
		return order.Value, float64(order.Value + 1), nil
	}
	if order.Amt > 1 {
		return order.Value - 2, float64(order.Value - 1), nil
	}
	return 0, 0, fmt.Errorf("Wrong amt of vouchers")
}
