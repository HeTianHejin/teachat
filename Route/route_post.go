package route

import (
	"fmt"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// 加水 ，修改回复post的处理器
func HandleEditPost(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case "GET":
		EditPost(w, r)
	case "POST":
		UpdatePost(w, r)
	case "PUT":
		//未开放的窗口
		w.WriteHeader(http.StatusInternalServerError)
		return
	case "DELETE":
		//未开放的窗口
		w.WriteHeader(http.StatusInternalServerError)
		return

	}

}

// get /v1/post/detail？id=
// 品味的详情
func PostDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	vals := r.URL.Query()
	uuid := vals.Get("id")
	var postPD data.PostDetailPageData
	postPD.Post, err = data.GetPostByUuid(uuid)
	if err != nil {
		util.Warning(err, " Cannot get post detail")
		util.Report(w, r, "您好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	// 读取此品味引用的茶议
	postPD.QuoteThread, err = postPD.Post.Thread()
	if err != nil {
		util.Warning(err, " Cannot get thread given post")
		util.Report(w, r, "您好，茶博士失魂鱼，未能读取茶议资料。")
		return
	}
	// 截短此引用的茶议以方便展示
	postPD.QuoteThread.Body = util.Substr(postPD.QuoteThread.Body, 66)
	// 此品味针对的茶议作者资料
	postPD.QuoteThreadAuthor, err = postPD.QuoteThread.User()
	if err != nil {
		util.Warning(err, " Cannot get thread author given post")
		util.Report(w, r, "您好，茶博士失魂鱼，未能读取茶议主人资料。")
		return
	}

	// 读取全部针对此品味的茶议
	postPD.ThreadList, err = postPD.Post.Threads()
	if err != nil {
		util.Warning(err, " Cannot get thread in post list")
		util.Report(w, r, "您好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	// 读取会话
	// 检测pageData.ThreadList数量是否超过一打dozen
	if len(postPD.ThreadList) > 12 {
		postPD.IsOverTwelve = true
	} else {
		postPD.IsOverTwelve = false
	}
	sess, err := util.Session(r)
	if err != nil {
		// 未登录，游客
		postPD.IsAuthor = false
		postPD.IsInput = true
		// 填写页面数据
		postPD.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		util.GenerateHTML(w, &postPD, "layout", "navbar.public", "post.detail")
		return
	}
	// 读取已登陆用户资料
	u, _ := sess.User()
	postPD.SessUser = u
	// 当前会话用户是否此品味作者？
	if u.Id == postPD.Post.UserId {
		postPD.IsAuthor = true
		postPD.IsInput = false
	} else {
		postPD.IsAuthor = false
		postPD.IsInput = true
	}

	util.GenerateHTML(w, &postPD, "layout", "navbar.private", "post.detail")

}

// POST /v1/post/draft
// Create the post 创建品味（跟帖/回复）草稿
func NewPost(w http.ResponseWriter, r *http.Request) {
	sess, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Warning(err, " Cannot parse form")
		util.Report(w, r, "您好，茶博士摸摸头，竟然说今天电脑去热带海岛潜水了。")
		return
	}

	sUser, err := sess.User()
	if err != nil {
		util.Warning(err, " Cannot get user from session")
		util.Report(w, r, "您好，茶博士失魂鱼，未能读取专属资料。")
		return
	}

	//读取用户表态,立场是支持（true）或者反对(false)
	var attitude bool
	a := r.PostFormValue("attitude")
	switch a {
	case "true":
		attitude = true
	case "false":
		attitude = false
	default:
		util.Report(w, r, "您好，茶博士失魂鱼，未能读懂您的表态内容。")
		return
	}

	body := r.PostFormValue("body")
	//检查body的长度，规则是不能少于刘姥姥评价老君眉的品味字数
	if util.CnStrLen(body) <= 17 {
		util.Report(w, r, "您好，戴着厚厚眼镜片的茶博士居然说，请不要用隐形墨水来写品味内容。")
		return
	} else if util.CnStrLen(body) > 456 {
		util.Report(w, r, "您好，彬彬有礼戴着厚厚眼镜片的茶博士居然说，内容太多，茶叶蛋壳都用光了也写不完呀。")
		return
	}
	uuid := r.PostFormValue("uuid")
	//检查uuid是否有效
	thread, err := data.ThreadByUUID(uuid)
	if err != nil {
		util.Report(w, r, "您好，茶博士失魂鱼，未能读取专属茶议。")
		return
	}

	//  检查茶议所在的茶台属性，
	proj, err := thread.Project()
	if err != nil {
		util.Report(w, r, "您好，茶博士失魂鱼，未能读取专属茶台资料。")
		return
	}
	var dPost data.DraftPost
	switch proj.Class {
	case 1:
		// class=1可以品茶，
		if dPost, err = sUser.CreateDraftPost(thread.Id, attitude, body); err != nil {
			util.Report(w, r, "您好，茶博士摸摸头，记录品味失败。")
			return
		}

	case 2:
		// 当前会话用户是否可以入席品茶？需要看台主指定了那些茶团成员可以品茶
		if ok := isUserInvitedByProject(proj, sUser); !ok {
			// Cannot have tea
			util.Report(w, r, "您好，你的大名竟然不在邀请品茶名单上。")
			return
		}

		// Can have tea
		if dPost, err = sUser.CreateDraftPost(thread.Id, attitude, body); err != nil {
			util.Report(w, r, "您好，茶博士摸摸头，竟然说没有墨水，记录品味失败。")
			return
		}

	default:
		// 异常状态的茶台
		util.Report(w, r, "您好，茶博士满头大汗说，陛下你的大名竟然不在邀请品茶名单上。")
		return
	}
	// url := fmt.Sprint("/v1/thread/detail?id=", thread.Uuid)
	// http.Redirect(w, r, url, http.StatusFound)
	// 创建一条友邻盲评,是否接纳 新茶的记录
	aO := data.AcceptObject{
		ObjectId:   dPost.Id,
		ObjectType: 4,
	}
	if err = aO.Create(); err != nil {
		util.Warning(err, "Cannot create accept_object")
		util.Report(w, r, "您好，胭脂洗出秋阶影，冰雪招来露砌魂。")
		return
	}
	// 发送邻座盲评消息
	mess := data.AcceptMessage{
		FromUserId:     1,
		Title:          "新茶语邻座评审邀请",
		Content:        "您好，茶博士隆重宣布：您被茶棚选中为新茶语评审官啦，请及时处理。",
		AcceptObjectId: aO.Id,
	}
	// 发送消息
	if err = AcceptMessageSendExceptUserId(sUser.Id, mess); err != nil {
		util.Report(w, r, "您好，茶博士迷路了，未能发送盲评请求消息。")
		return
	}
	// 提示用户草稿保存成功
	t := fmt.Sprintf("您好，对 %s 发布的品味已准备妥当，稍等有缘茶友评审通过，即可昭告天下。", thread.Title)
	// 提示用户草稿保存成功
	util.Report(w, r, t)
}

// GET /v1/post/accept
// AcceptDraftPost() 友邻盲评审查新品味draftPost是否符合文明发言
func AcceptDraftPost(w http.ResponseWriter, r *http.Request) {
	sess, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Danger(err, " Cannot parse form")
	}
	//从会话中读取用户资料
	sUser, err := sess.User()
	if err != nil {
		util.Danger(err, " Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 读取拟评审的品味草稿
	vals := r.URL.Query()
	id, err := strconv.Atoi(vals.Get("id"))
	if err != nil {
		util.Warning(err, " Cannot convert id to int")
		util.Report(w, r, "茶博士失魂鱼，未能读取新品味草稿资料，请稍后再试。")
		return
	}
	post, err := data.GetDraftPost(id)
	if err != nil {
		util.Warning(err, " Cannot get draft post")
		util.Report(w, r, "茶博士失魂鱼，未能读取新品味草稿资料，请稍后再试。")
		return
	}
	// 检查用户是否受邀请有效状态的审茶官
	if !sUser.CheckHasAcceptMessage(sUser.Id) {
		util.Report(w, r, "您好，友邻盲评邀请已经过期失效啦，感谢你对维护茶棚文明秩序的支持。")
		return
	}
	// 返回品味友邻盲评页面
	util.GenerateHTML(w, &post, "layout", "navbar.private", "post.accept")

}

// POST /post/edit
// update the post
// 更新用户的post，规则是可以补充内容，不能覆载修改之前说的话，不能变更立场（从支持变反对）。
func UpdatePost(w http.ResponseWriter, r *http.Request) {
	//读取请求会话
	sess, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Danger(err, " Cannot parse form")
	}
	//从会话中读取用户资料
	user, err := sess.User()
	if err != nil {
		util.Danger(err, " Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	uuid := r.PostFormValue("uuid")
	post, err := data.GetPostByUuid(uuid)
	if err != nil {
		util.Danger(err, " Cannot read given post")
		util.Report(w, r, "茶博士失魂鱼，未能读取指定表态，请稍后再试。")
		return
	}
	if post.UserId != user.Id {
		util.Danger(err, " Cannot edit other user's post")
		util.Report(w, r, "茶博士提示，目前仅能补充自己的回复")
		return
	} else {
		//可以补充自己的表态内容
		body := r.PostFormValue("body")
		if body != "" {
			//检查补充内容是否有意义，rune 字数>17,总的post字数<456
			if util.CnStrLen(body) > 17 && util.CnStrLen(post.Body)+util.CnStrLen(body) < 456 {
				post.Body += body
			} else {
				//提示用户总字数或者本次提交补充内容超出字数限制
				util.Report(w, r, "您好， 粗鲁的茶博士竟然说字数满了，纸条写不下您的品味。")
				return
			}
			err = post.UpdateBody(body)
			if err != nil {
				util.Danger(err, " Cannot update post")
				util.Report(w, r, "茶博士失魂鱼，未能更新专属资料，请稍后再试。")
				return
			}
			thread, err := data.GetThreadById(post.ThreadId)
			if err != nil {
				util.Danger(err, " Cannot read thread")
				util.Report(w, r, "茶博士失魂鱼，未能读取专属资料，请稍后再试。")
			}
			url := fmt.Sprint("/v1/thread/detail?id=", thread.Uuid)
			http.Redirect(w, r, url, http.StatusFound)
		} else {
			//空白或者一个字被认为是无意义追加内容
			util.Report(w, r, "您好，请勿提供小于17个字的品味补充")
			return
		}

	}

}

// GET /post/edit
// 用户补充自己的表态内容POST的界面
func EditPost(w http.ResponseWriter, r *http.Request) {
	sess, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	} else {

		user, err := sess.User()
		if err != nil {
			util.Danger(err, " Cannot get user from session")
			http.Redirect(w, r, "/v1/login", http.StatusFound)
			return
		}
		vals := r.URL.Query()
		uuid := vals.Get("id")
		post, err := data.GetPostByUuid(uuid)
		if err != nil {
			util.Danger(err, " Cannot read post")
			util.Report(w, r, "您好，茶博士失魂鱼，未能读取专属资料，请稍后再试。")
			return
		}
		if post.UserId == user.Id {
			util.GenerateHTML(w, &post, "layout", "navbar.private", "post.edit")
		} else {
			util.Danger(err, " Cannot edit other user's post")
			util.Report(w, r, "茶博士提示，目前仅能补充自己的回复")
		}

	}

}
