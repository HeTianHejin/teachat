package route

import (
	"net/http"
)

// 常见问题解答 (Frequently Asked Questions)
func FAQ(w http.ResponseWriter, r *http.Request) {
	Report(w, r, "报告大王，伶俐虫和精细鬼还在睡懒觉，没有出发去巡山呢。")
}

// 说明文档
func Doc(w http.ResponseWriter, r *http.Request) {
	Report(w, r, "报告大王，伶俐虫和精细鬼还在睡懒觉，没有出发去抓唐僧呢。")
}
