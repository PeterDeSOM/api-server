package controllers

import (
	"net/http"

	s "github.com/crypblorm/bitcoin/api-server/services"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// GetAddressInfo Returns information about the given bitcoin address.
// Some information requires the address to be in the wallet.
// func GetAddressInfo(w http.ResponseWriter, r *http.Request) {
// }

// GetBalance Returns the total available balance.
// The available balance is what the wallet considers currently spendable, and is
// thus affected by options which limit spendability such as -spendzeroconfchange.
func GetBalance(w http.ResponseWriter, r *http.Request) {

	accountname := chi.URLParam(r, "accountname")
	if len(accountname) == 0 {
		accountname = "*"
	}
	acctBalance := s.GetBalance(accountname)
	// A chi router helper for serializing and returning json
	render.JSON(w, r, acctBalance)
}
