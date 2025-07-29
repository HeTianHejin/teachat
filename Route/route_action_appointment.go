package route

import "net/http"

func HandleNewAppointment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		NewAppointmentGet(w, r)
	case http.MethodPost:
		NewAppointmentPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/appointment/new
func NewAppointmentGet(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// POST /v1/appointment/new
func NewAppointmentPost(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// Get /v1/appointment/detail?uuid=xXx
func AppointmentDetail(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}
