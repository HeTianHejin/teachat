package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	route "teachat/Route"
	util "teachat/Util"
	"time"

	"github.com/NYTimes/gziphandler" // 压缩
)

func main() {
	// 初始化配置
	if err := util.LoadConfig(); err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
	if err := util.Config.Validate(); err != nil {
		log.Fatalf("配置校验失败: %v", err)
	}
	// 创建路由器
	mux := http.NewServeMux()

	// 静态资源处理
	const staticPrefix = "/v1/static/"
	if _, err := os.Stat(util.Config.Static); os.IsNotExist(err) {
		log.Fatalf("静态资源目录不存在: %s", util.Config.Static)
	}

	// 创建文件处理器（带缓存控制）
	cacheControl := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=31536000") // 一年
			h.ServeHTTP(w, r)
		})
	}

	files := cacheControl(http.FileServer(http.Dir(util.Config.Static)))

	// 添加Gzip压缩
	handler := gziphandler.GzipHandler(files) // 直接使用，无需 nil 检查

	// 注册静态资源处理器
	mux.Handle(staticPrefix, http.StripPrefix(staticPrefix, handler))

	//	mux.HandleFunc("/", route.Index)
	// index
	//测试时使用
	mux.HandleFunc("/v1/", route.ObjectiveSquare)

	// search
	mux.HandleFunc("/v1/search", route.HandleSearch)

	// defined in route_user.go
	mux.HandleFunc("/v1/user/biography", route.Biography)
	mux.HandleFunc("/v1/user/edit", route.EditIntroAndName)
	mux.HandleFunc("/v1/user/forgot", route.Forgot)
	mux.HandleFunc("/v1/user/reset", route.Reset)
	mux.HandleFunc("/v1/user/avatar", route.UserAvatar)
	mux.HandleFunc("/v1/user/invite", route.Invite)

	mux.HandleFunc("/v1/users/connection_friend", route.Friend)
	mux.HandleFunc("/v1/users/connection_follow", route.Follow)
	mux.HandleFunc("/v1/users/connection_fans", route.Fans)

	// defined in route_auth.go
	mux.HandleFunc("/v1/login", route.LoginGet)
	mux.HandleFunc("/v1/logout", route.Logout)
	mux.HandleFunc("/v1/signup", route.SignupGet)
	mux.HandleFunc("/v1/signup_account", route.SignupPost)
	mux.HandleFunc("/v1/authenticate", route.Authenticate)

	// defined in route_team.go
	mux.HandleFunc("/v1/team/new", route.NewTeamGet)
	mux.HandleFunc("/v1/team/create", route.CreateTeamPost)
	mux.HandleFunc("/v1/team/detail", route.TeamDetail)
	mux.HandleFunc("/v1/team/avatar", route.TeamAvatar)
	mux.HandleFunc("/v1/team/invitations", route.InvitationsBrowse)
	mux.HandleFunc("/v1/team/invitation", route.InvitationView)
	mux.HandleFunc("/v1/team/members/fired", route.MemberFired)

	mux.HandleFunc("/v1/team/manage", route.HandleManageTeam)
	mux.HandleFunc("/v1/team/core_manage", route.CoreManage)
	mux.HandleFunc("/v1/team/default", route.SetDefaultTeam)

	mux.HandleFunc("/v1/teams/open", route.OpenTeams)
	mux.HandleFunc("/v1/teams/closed", route.ClosedTeams)
	mux.HandleFunc("/v1/teams/hold", route.HoldTeams)
	mux.HandleFunc("/v1/teams/joined", route.JoinedTeams)
	mux.HandleFunc("/v1/teams/employed", route.EmployedTeams)
	mux.HandleFunc("/v1/teams/application", route.ApplyTeams)

	//mux.HandleFunc("/v1/teams/rejected", route.RejectedTeams)

	//defined in route_team_member.go
	//处理茶团成员个人事务
	mux.HandleFunc("/v1/team_member/application/new", route.HandleNewMemberApplication)
	mux.HandleFunc("/v1/team_member/application/review", route.HandleMemberApplication)
	mux.HandleFunc("/v1/team_member/application/check", route.MemberApplyCheck)
	mux.HandleFunc("/v1/team_member/invite", route.HandleInviteMember)
	mux.HandleFunc("/v1/team_member/invitation", route.HandleMemberInvitation)
	mux.HandleFunc("/v1/team_member/role", route.HandleMemberRole)
	mux.HandleFunc("/v1/team_member/role_changed", route.MemberRoleChanged)
	mux.HandleFunc("/v1/team_member/resign", route.HandleMemberResign)

	//defined in route_family.go
	//处理家庭茶团事务
	mux.HandleFunc("/v1/family/new", route.HandleNewFamily)
	mux.HandleFunc("/v1/family/detail", route.FamilyDetail)
	mux.HandleFunc("/v1/family/default", route.SetDefaultFamily)

	mux.HandleFunc("/v1/families/home", route.HomeFamilies)

	//defined in route_family_member.go
	mux.HandleFunc("/v1/family_member/sign_in_new", route.HandleFamilyMemberSignInNew)
	mux.HandleFunc("/v1/family_member/sign_in", route.HandleFamilyMemberSignIn)
	//mux.HandleFunc("/v1/family_member/sign_out", route.HandleFamilyMemberSignOut)

	//defined in route_objective.go
	mux.HandleFunc("/v1/objective/new", route.HandleNewObjective)
	mux.HandleFunc("/v1/objective/square", route.ObjectiveSquare)
	mux.HandleFunc("/v1/objective/detail", route.ObjectiveDetail)
	//mux.HandleFunc("/v1/objective/edit", editObjective)
	//mux.HandleFunc("/v1/objective/update", updateObjective)

	//defined in route_project.go
	mux.HandleFunc("/v1/project/new", route.HandleNewProject)
	mux.HandleFunc("/v1/project/detail", route.ProjectDetail)
	mux.HandleFunc("/v1/project/approve", route.ProjectApprove)

	// defined in route_thread.go
	mux.HandleFunc("/v1/thread/draft", route.NewDraftThreadHandle)
	mux.HandleFunc("/v1/thread/detail", route.ThreadDetail)
	mux.HandleFunc("/v1/thread/supplement", route.HandleThreadSupplement)
	mux.HandleFunc("/v1/thread/approve", route.ThreadApprove)
	//mux.HandleFunc("/v1/thread/plus", route.PlusThread)

	// defined in route_see-seek.go
	mux.HandleFunc("/v1/see-seek/new", route.HandleNewSeeSeek)
	//mux.HandleFunc("/v1/see-seek/detail", route.SeeSeekDetail)

	//定义在 route_post.go
	mux.HandleFunc("/v1/post/draft", route.NewPostDraft)
	mux.HandleFunc("/v1/post/edit", route.HandleEditPost)
	mux.HandleFunc("/v1/post/detail", route.PostDetail)

	//defined in route_pilot.go
	mux.HandleFunc("/v1/pilot/new", route.NewPilot)
	mux.HandleFunc("/v1/pilot/add", route.AddPilot)
	mux.HandleFunc("/v1/pilot/office", route.OfficePilot)
	//mux.HandleFunc("/v1/pilot/detail", pilotDetail)

	//defined in route_office
	mux.HandleFunc("/v1/office/polite", route.Polite)
	mux.HandleFunc("/v1/office/draftThread", route.ActivateDraftThread)

	//defined in route_message.go
	mux.HandleFunc("/v1/message/letterbox", route.Letterbox)
	mux.HandleFunc("/v1/message/accept", route.AcceptMessages)

	//defined in route_place.go
	mux.HandleFunc("/v1/place/new", route.NewPlace)
	mux.HandleFunc("/v1/place/create", route.CreatePlace)
	mux.HandleFunc("/v1/place/detail", route.PlaceDetail)
	mux.HandleFunc("/v1/place/my", route.MyPlace)
	mux.HandleFunc("/v1/place/collect", route.PlaceCollect)

	// defined in route_goods.go
	mux.HandleFunc("/v1/goods/family_new", route.HandleGoodsFamilyNew)
	//mux.HandleFunc("/v1/goods/family", route.GoodsFamily)
	//mux.HandleFunc("/v1/goods/family_detail", route.GoodsFamilyDetail)
	mux.HandleFunc("/v1/goods/team_new", route.HandleGoodsTeamNew)
	mux.HandleFunc("/v1/goods/team", route.GoodsTeam)
	mux.HandleFunc("/v1/goods/team_detail", route.GoodsTeamDetail)
	mux.HandleFunc("/v1/goods/team_update", route.HandleGoodsTeamUpdate)

	mux.HandleFunc("/v1/goods/collect", route.GoodsCollect)
	mux.HandleFunc("/v1/goods/eye_on", route.GoodsEyeOn)

	//定义在 Route_balance.go
	mux.HandleFunc("/v1/balance/fairnessmug", route.FairnessMug)

	// define in help.go 帮助 文档 信息
	mux.HandleFunc("/v1/help/faq", route.FAQ)
	mux.HandleFunc("/v1/help/doc", route.Doc)
	// about
	mux.HandleFunc("/v1/about", route.About)

	// 创建服务器
	server := &http.Server{
		Addr:           util.Config.Address,
		Handler:        mux,
		ReadTimeout:    time.Duration(util.Config.ReadTimeout) * time.Second, // 修正时间单位
		WriteTimeout:   time.Duration(util.Config.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	// 设置优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("接收到关闭信号，正在停止服务器...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("服务器强制关闭: %v", err)
		} else {
			log.Println("服务器已优雅停止")
		}
	}()
	// 启动服务器
	log.Printf("服务器启动，监听地址: %s", util.Config.Address)
	util.PrintStdout("teachat", util.Version(), "星际茶棚一>开门迎客")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v", err)
	}
	log.Println("服务器已停止")
	log.Println("星际茶棚 --> 打烊休息")
}
