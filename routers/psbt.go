package routers

import (
	c "github.com/crypblorm/bitcoin/api-server/controllers"
	"github.com/go-chi/chi"
)

// rWallet return router
func rPsbt() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/list/{address}/{target}/{by}", c.PsbtList)
	router.Put("/create", c.CreatePsbt)
	router.Put("/sign", c.SignPsbt)
	router.Post("/push", c.PushPsbt)
	router.Post("/cancel", c.CancelPsbt)
	router.Post("/getcomplete", c.GetPsbtComplete)
	return router
}
