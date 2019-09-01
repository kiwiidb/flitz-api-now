package withdrawlnurl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func getCollectionAndToken(path string) (collection string, token string, err error) {
	//path is /{lnurl-primary,lnurl-secondary}/collection/token, so 2 and 3
	splittedRoute := strings.Split(path, "/")
	if len(splittedRoute) < 3 {
		return "", "", fmt.Errorf("Wrong number of route parameters in url %s", path)
	}
	return strings.Split(path, "/")[2], strings.Split((path), "/")[3], nil
}
func writeErrorResponse(w http.ResponseWriter, message string, status int) {
	resp := SecondaryResponse{
		Status: "ERROR",
		Reason: message,
	}
	writeResponse(w, resp, status)
}
func writeResponse(w http.ResponseWriter, resp interface{}, status int) {
	respBytes, err := json.Marshal(resp)
	if err != nil {
		logrus.Error(err.Error())
		http.Error(w, "something wrong", http.StatusInternalServerError)
		return
	}
	w.Write(respBytes)
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
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
