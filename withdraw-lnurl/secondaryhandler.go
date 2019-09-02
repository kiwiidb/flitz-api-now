package withdrawlnurl

import (
	"math"
	"net/http"

	"github.com/sirupsen/logrus"
)

//SecondaryHandler as per LNURL specs
func SecondaryHandler(w http.ResponseWriter, r *http.Request) {
	collection, token, err := GetCollectionAndToken(r.URL.Path)
	if err != nil {
		logrus.Error(err)
		writeErrorResponse(w, "Bad Request", http.StatusBadRequest)
		return
	}
	authorized, euroValue, err := tdb.GetIfTokenAuthorized(token, collection)
	if err != nil {
		logrus.WithField("collection", collection).WithField("Token", token).Error(err)
		writeErrorResponse(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	if !authorized {
		writeErrorResponse(w, "Token not valid or already claimed", http.StatusBadRequest)
		return
	}
	invoices := r.URL.Query()["pr"]
	//LNURL specifies multiple invoices possible
	totalFiatAmt := 0.0
	for _, inv := range invoices {
		value, err := getFiatAmt(inv)
		if err != nil {
			logrus.Error(err)
			writeErrorResponse(w, "Internal Error decoding payreq", http.StatusInternalServerError)
			return
		}
		totalFiatAmt += value
		//check value with value in db
	}
	if int(math.Round(totalFiatAmt)) != euroValue {
		logrus.WithField("token", token).WithField("invoice amt", math.Round(totalFiatAmt)).WithField("token value", euroValue).Info("Request coming in for wrongly priced invoice")
		writeErrorResponse(w, "Bad request", http.StatusBadRequest)
		return
	}
	for _, inv := range invoices {
		wd, err := on.Withdraw(inv)
		if err != nil {
			writeErrorResponse(w, "Bad request", http.StatusInternalServerError)
			return
		}
		logrus.Info((wd))
	}
	resp := SecondaryResponse{
		Status: "OK",
	}
	writeResponse(w, resp, http.StatusOK)
	return
}
