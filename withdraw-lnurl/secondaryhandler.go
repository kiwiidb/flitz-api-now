package withdrawlnurl

import (
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

	//TODO
	//Extract token, collection from request route parameters
	//Extract invoice from request query parameters
	//Check if invoice(s) have the correct amt
	//initiate opennode withdraw
	//(optionally) check for opennode status of withdraw
	//construct response and reply

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
