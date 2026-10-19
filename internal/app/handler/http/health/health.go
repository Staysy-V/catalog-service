package rhealth

import (
	rhandler "github.com/Staysy-V/catalog-service/internal/app/handler/http"
	"net/http"
)

type handler struct{}

func NewHandler() rhandler.Health {
	return &handler{}
}

func (h *handler) LastCheck(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}
