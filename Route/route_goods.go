package route

import (
	"net/http"
)

// 处理新物资的办理窗口
func HandleNewGoods(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		GoodsGet(w, r)
	case "POST":
		GoodsPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func GoodsPost(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func GoodsGet(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}
