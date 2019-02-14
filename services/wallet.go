package services

import (
	"log"

	"github.com/btcsuite/btcutil"
)

// GetAddressInfo Returns information about the given bitcoin address.
// Some information requires the address to be in the wallet.
// func GetAddressInfo(address string)  {
// }

// GetBalance Returns the total available balance.
// The available balance is what the wallet considers currently spendable, and is
// thus affected by options which limit spendability such as -spendzeroconfchange.
func GetBalance(accountname string) btcutil.Amount {

	client := HTTPClient()
	acctBalance, err := client.GetBalance(accountname)
	if err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}
	client.Shutdown()
	return acctBalance
}
