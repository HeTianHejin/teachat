package route

import (
	"net/http"
	dao "teachat/DAO"
)

// 常见问题解答 (Frequently Asked Questions)
func FAQ(w http.ResponseWriter, r *http.Request) {
	report(w, dao.UserUnknown, "报告大王，伶俐虫和精细鬼还在睡懒觉，没有出发去巡山呢。")
}

// 说明文档
func Doc(w http.ResponseWriter, r *http.Request) {
	report(w, dao.UserUnknown, "报告大王，伶俐虫和精细鬼还在睡懒觉，没有出发去抓唐僧呢。")
}
