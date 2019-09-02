package withdrawlnurl

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

//PrimaryHandler main handler for this lambda
//redeem a token and withdraw your sats!
func PrimaryHandler(w http.ResponseWriter, r *http.Request) {
	collection, token, err := GetCollectionAndToken(r.URL.Path)
	if err != nil {
		logrus.Error(err)
		writeErrorResponse(w, "Bad Request", http.StatusInternalServerError)
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
	//TODO add this whole thing to on library
	btcPrice, err := on.GetEuroRate()
	if err != nil {
		logrus.Error(err.Error())
		writeErrorResponse(w, "Error getting fiat rates", http.StatusInternalServerError)
		return
	}
	milliSatoshiValue := int(float64(euroValue)/btcPrice*1e8) * 1e3
	secondaryRoute := fmt.Sprintf("%s/%s/%s", "/lnurl-secondary", collection, token)
	resp := PrimaryResponse{
		Callback:        fmt.Sprintf("https://%s%s", r.Host, secondaryRoute),
		K1:              "", //not needed
		MinWithdrawable: milliSatoshiValue,
		MaxWithdrawable: milliSatoshiValue,
		Tag:             "withdrawRequest",
	}
	writeResponse(w, resp, http.StatusOK)
	return
}
