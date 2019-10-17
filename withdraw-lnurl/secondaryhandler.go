package withdrawlnurl

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

//SecondaryHandler as per LNURL specs
func SecondaryHandler(w http.ResponseWriter, r *http.Request) {
	collection, token, err := getCollectionAndToken(r.URL.Path)
	if err != nil {
		logrus.Error(err)
		writeErrorResponse(w, "Bad Request", http.StatusBadRequest)
		return
	}
	t, err := tdb.GetIfTokenAuthorized(token, collection)
	if err != nil {
		logrus.WithField("collection", collection).WithField("Token", token).Error(err)
		writeErrorResponse(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	invoices := r.URL.Query()["pr"]
	//LNURL specifies multiple invoices possible
	totalFiatAmt := 0
	for _, inv := range invoices {
		value, err := getFiatAmt(inv, t)
		if err != nil {
			logrus.Error(err)
			writeErrorResponse(w, "Internal Error decoding payreq", http.StatusInternalServerError)
			return
		}
		totalFiatAmt += value
		//check value with value in db
	}
	logrus.Info(totalFiatAmt)
	if totalFiatAmt != t.Value {
		logrus.WithField("token", token).WithField("invoice amt", totalFiatAmt).WithField("token value", t.Value).Info("Request coming in for wrongly priced invoice")
		writeErrorResponse(w, "Bad request", http.StatusBadRequest)
		return
	}
	//TODO add all invoices
	count, err := tdb.SetTokenClaimed(token, invoices[0], fmt.Sprintf("%v", *r), collection)
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
