package routers

import (
	c "github.com/crypblorm/bitcoin/api-server/controllers"
	"github.com/go-chi/chi"
)

// rWallet return router
func rRawtx() *chi.Mux {
	router := chi.NewRouter()
	router.Post("/decode", c.DecodePsbt)
	return router
}
