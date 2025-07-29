package route

import "net/http"

func HandleNewHandicraft(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		NewHandicraftGet(w, r)
	case http.MethodPost:
		NewHandicraftPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/handicraft/new
func NewHandicraftGet(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// POST /v1/handicraft/new
func NewHandicraftPost(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// GET /v1/handicraft/detail?uuid=xXx
func HandicraftDetail(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}
