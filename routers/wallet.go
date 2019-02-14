package routers

import (
	c "github.com/crypblorm/bitcoin/api-server/controllers"
	"github.com/go-chi/chi"
)

// rWallet return router
func rWallet() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/addressinfo/{address}", c.GetBalance)
	router.Get("/addressbalance", c.GetBalance)
	router.Get("/addressbalance/{accountname}", c.GetBalance)
	return router
}
