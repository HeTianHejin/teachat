package main

import (
	"net/http"
	route "teachat/Route"
	util "teachat/Util"
	"time"
)

func main() {

	// 在控制台输出一些运行地址及应用版本信息
	util.PrintStdout("teachat", util.Version(), "server at", util.Config.Address)

	// handle static assets
	mux := http.NewServeMux()
	files := http.FileServer(http.Dir(util.Config.Static))
	mux.Handle("/v1/static/", http.StripPrefix("/v1/static/", files))

	// index
	mux.HandleFunc("/", route.Index)
	mux.HandleFunc("/v1/", route.Index)

	// search
	mux.HandleFunc("/v1/search", route.Search)

	// defined in route_user.go
	mux.HandleFunc("/v1/user/biography", route.Biography)
	mux.HandleFunc("/v1/user/edit", route.EditIntroAndName)
	mux.HandleFunc("/v1/user/forgot", route.Forgot)
	mux.HandleFunc("/v1/user/reset", route.Reset)
	mux.HandleFunc("/v1/user/avatar", route.UserAvatar)
	mux.HandleFunc("/v1/users/connection_friend", route.Friend)
	mux.HandleFunc("/v1/users/connection_follow", route.Follow)
	mux.HandleFunc("/v1/users/connection_fans", route.Fans)

	// defined in route_auth.go
	mux.HandleFunc("/v1/login", route.Login)
	mux.HandleFunc("/v1/logout", route.Logout)
	mux.HandleFunc("/v1/signup", route.SignupForm)
	mux.HandleFunc("/v1/signup_account", route.SignupAccount)
	mux.HandleFunc("/v1/authenticate", route.Authenticate)

	// defined in route_team.go
	mux.HandleFunc("/v1/team/new", route.NewTeam)
	mux.HandleFunc("/v1/team/create", route.CreateTeam)
	mux.HandleFunc("/v1/team/detail", route.TeamDetail)
	mux.HandleFunc("/v1/team/open", route.OpenTeams)
	mux.HandleFunc("/v1/team/closed", route.ClosedTeams)
	mux.HandleFunc("/v1/team/hold", route.HoldTeams)
	mux.HandleFunc("/v1/team/joined", route.JoinedTeam)
	mux.HandleFunc("/v1/team/employed", route.EmployedTeam)
	mux.HandleFunc("/v1/team/manage", route.HandleManageTeam)
	mux.HandleFunc("/v1/team/core_manage", route.CoreManage)
	mux.HandleFunc("/v1/team/avatar", route.TeamAvatar)
	mux.HandleFunc("/v1/team/invitations", route.ReviewInvitations)

	//defined in route_team_member.go
	mux.HandleFunc("/v1/team/team_member/join", route.HandleJoinTeam)
	mux.HandleFunc("/v1/team/team_member/invite_page", route.HandleInviteMember)
	mux.HandleFunc("/v1/team/team_member/invitation", route.HandleInvitation)
	mux.HandleFunc("/v1/team/team_member/quit", route.HandleMemberQuit)

	//defined in route_objective.go
	mux.HandleFunc("/v1/objective/new", route.HandleNewObjective)
	mux.HandleFunc("/v1/objective/square", route.ObjectiveSquare)
	mux.HandleFunc("/v1/objective/detail", route.ObjectiveDetail)
	//mux.HandleFunc("/v1/objective/edit", editObjective)
	//mux.HandleFunc("/v1/objective/update", updateObjective)

	//defined in route_project.go
	mux.HandleFunc("/v1/project/new", route.HandleNewProject)
	mux.HandleFunc("/v1/project/detail", route.ProjectDetail)

	// defined in route_thread.go
	mux.HandleFunc("/v1/thread/new", route.NewThread)
	mux.HandleFunc("/v1/thread/draft", route.DraftThread)
	mux.HandleFunc("/v1/thread/accept", route.AcceptDraftThread)
	mux.HandleFunc("/v1/thread/detail", route.ThreadDetail)
	mux.HandleFunc("/v1/thread/edit", route.EditThread)
	mux.HandleFunc("/v1/thread/update", route.UpdateThread)
	mux.HandleFunc("/v1/thread/plus", route.PlusThread)

	//定义在 route_post.go
	mux.HandleFunc("/v1/post/draft", route.NewPost)
	mux.HandleFunc("/v1/post/accept", route.AcceptDraftPost)
	mux.HandleFunc("/v1/post/edit", route.HandleEditPost)
	mux.HandleFunc("/v1/post/detail", route.PostDetail)

	//defined in route_pilot.go
	mux.HandleFunc("/v1/pilot/new", route.NewPilot)
	mux.HandleFunc("/v1/pilot/add", route.AddPilot)
	mux.HandleFunc("/v1/pilot/office", route.OfficePilot)
	//mux.HandleFunc("/v1/pilot/detail", pilotDetail)
	mux.HandleFunc("/v1/pilot/invitepage", route.InvitePage)

	//defined in route_office
	mux.HandleFunc("/v1/office/polite", route.Polite)
	mux.HandleFunc("/v1/office/draftThread", route.ActivateDraftThread)

	//defined in route_message.go
	mux.HandleFunc("/v1/message/letterbox", route.Letterbox)
	mux.HandleFunc("/v1/message/accept", route.AcceptMessages)

	//定义在 Route_balance.go
	mux.HandleFunc("/v1/balance/fairnessmug", route.FairnessMug)

	// define in help.go 帮助 文档 信息
	mux.HandleFunc("/v1/help/faq", route.FAQ)
	mux.HandleFunc("/v1/help/doc", route.Doc)
	// about
	mux.HandleFunc("/v1/about", route.About)

	// starting up the server
	server := &http.Server{
		Addr:           util.Config.Address,
		Handler:        mux,
		ReadTimeout:    time.Duration(util.Config.ReadTimeout * int64(time.Second)),
		WriteTimeout:   time.Duration(util.Config.WriteTimeout * int64(time.Second)),
		MaxHeaderBytes: 1 << 20,
	}
	server.ListenAndServe()
}
