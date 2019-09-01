package withdrawlnurl

import (
	"math"
	"net/http"
	"regexp"
	"strconv"

	"github.com/sirupsen/logrus"
)

//SecondaryResponse as per LNURL spec for withdrawing
type SecondaryResponse struct {
	Status string
	Reason string
}

//SecondaryHandler as per LNURL specs
func SecondaryHandler(w http.ResponseWriter, r *http.Request) {
	collection, token, err := getCollectionAndToken(r.URL.Path)
	if err != nil {
		logrus.Error(err)
		resp := SecondaryResponse{
			Status: "ERROR",
			Reason: "Bad Request",
		}
		writeResponse(w, resp, http.StatusInternalServerError)
		return
	}
	authorized, euroValue, err := tdb.GetIfTokenAuthorized(token, collection)
	if err != nil {
		logrus.WithField("collection", collection).WithField("Token", token).Error(err)
		resp := SecondaryResponse{
			Status: "ERROR",
			Reason: "Internal Error",
		}
		writeResponse(w, resp, http.StatusInternalServerError)
		return
	}
	if !authorized {
		resp := SecondaryResponse{
			Status: "ERROR",
			Reason: "Token not valid or already claimed",
		}
		writeResponse(w, resp, http.StatusInternalServerError)
		return
	}
	invoices := r.URL.Query()["pr"]
	//LNURL specifies multiple invoices possible
	totalFiatAmt := 0.0
	for _, inv := range invoices {
		value, err := getFiatAmt(inv)
		if err != nil {
			logrus.Error(err)
			resp := SecondaryResponse{
				Status: "ERROR",
				Reason: "Bad Request",
			}
			writeResponse(w, resp, http.StatusInternalServerError)
			return
		}
		totalFiatAmt += value
		//check value with value in db
	}
	if int(math.Round(totalFiatAmt)) != euroValue {
		logrus.WithField("token", token).WithField("invoice amt", math.Round(totalFiatAmt)).WithField("token value", euroValue).Info("Request coming in for wrongly priced invoice")
		resp := SecondaryResponse{
			Status: "ERROR",
			Reason: "Bad Request",
		}
		writeResponse(w, resp, http.StatusInternalServerError)
		return
	}
	for _, inv := range invoices {
		wd, err := on.Withdraw(inv)
		if err != nil {
			resp := SecondaryResponse{
				Status: "ERROR",
				Reason: "Bad Request",
			}
			writeResponse(w, resp, http.StatusInternalServerError)
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

//first calculate the amount of sat, ask on for exchange rate, calculate value in euros
func getFiatAmt(invoice string) (float64, error) {
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
	rate, err := on.GetEuroRate()
	if err != nil {
		return 0, nil
	}
	return rate * satAmt / 1e8, nil
}
