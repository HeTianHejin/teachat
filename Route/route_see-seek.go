package route

import "net/http"

// Handler() /v1/see-seek/new
func SeeSeekNew(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		SeeSeekNewGet(w, r)
	case http.MethodPost:
		SeeSeekNewPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func SeeSeekNewPost(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func SeeSeekNewGet(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}
