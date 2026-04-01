package httpserver

import (
	"net/http"

	"github.com/fishkaoff/ts-backend/internal/transport"
	"github.com/go-chi/chi/v5"
)

func (s *HTTPServer) AdminRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/upload-price", s.HandleUploadPrice)

	return r
}

func (s *HTTPServer) HandleUploadPrice(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(30 << 20)
	if err != nil {
		http.Error(w, "cannot parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("price")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = s.productsService.UpdatePrice(r.Context(), file)
	if err != nil {
		writeError(w, *transport.InternalError.AddDetails(err.Error()))
		return
	}

	writeJSON(w, 200, map[string]string{
		"status": "processing",
	})
}
