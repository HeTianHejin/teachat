package route

import "net/http"

func HandleNewSuggestion(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		NewSuggestionGet(w, r)
	case http.MethodPost:
		NewSuggestionPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/suggestion/new
func NewSuggestionGet(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// POST /v1/suggestion/new
func NewSuggestionPost(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// GET /v1/suggestion/detail?uuid=xXx
func SuggestionDetail(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}
