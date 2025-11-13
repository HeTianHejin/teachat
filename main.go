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
	handler := gziphandler.GzipHandler(files)

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

	mux.HandleFunc("/v1/users/connection_follow", route.Follow)

	// defined in route_auth.go
	mux.HandleFunc("/v1/login", route.LoginGet)
	mux.HandleFunc("/v1/logout", route.Logout)
	mux.HandleFunc("/v1/signup", route.SignupGet)
	mux.HandleFunc("/v1/signup_account", route.SignupPost)
	mux.HandleFunc("/v1/authenticate", route.Authenticate)

	// defined in route_team.go
	// 团队杂项
	mux.HandleFunc("/v1/team/new", route.NewTeamGet)
	mux.HandleFunc("/v1/team/create", route.CreateTeamPost)
	mux.HandleFunc("/v1/team/detail", route.TeamDetail)
	mux.HandleFunc("/v1/team/avatar", route.TeamAvatar)
	mux.HandleFunc("/v1/team/invitations", route.TeamInvitations)
	mux.HandleFunc("/v1/team/applications", route.TeamApplications)

	mux.HandleFunc("/v1/team/manage", route.HandleManageTeam)
	mux.HandleFunc("/v1/team/edit", route.HandleEditTeam)
	mux.HandleFunc("/v1/team/member_add", route.TeamMemberAddGet)
	mux.HandleFunc("/v1/team/search_user", route.HandleTeamSearchUser)
	mux.HandleFunc("/v1/team/core_manage", route.CoreManage)
	mux.HandleFunc("/v1/team/new_applications/check", route.TeamNewApplicationsCheck)

	mux.HandleFunc("/v1/team_member/role", route.HandleMemberRole)
	mux.HandleFunc("/v1/team_member/role_changed", route.MemberRoleChanged)
	mux.HandleFunc("/v1/team_members/fired", route.MemberFired)

	// 茶友个人与团队关系
	mux.HandleFunc("/v1/team/default", route.SetDefaultTeam)
	mux.HandleFunc("/v1/teams/hold", route.HoldTeams)
	mux.HandleFunc("/v1/teams/joined", route.JoinedTeams)
	mux.HandleFunc("/v1/teams/employed", route.EmployedTeams)
	//mux.HandleFunc("/v1/teams/rejected", route.RejectedTeams)

	//defined in route_team_member.go
	//处理茶团成员事务
	// 申请加盟书-1团队管理员
	mux.HandleFunc("/v1/team_member/application/new", route.HandleNewMemberApplication)
	mux.HandleFunc("/v1/team_member/application/review", route.HandleMemberApplicationReview)
	mux.HandleFunc("/v1/team_member/application/detail", route.MemberApplicationDetail)
	// 申请加盟书-2成员个人
	mux.HandleFunc("/v1/applications/member", route.ApplyTeams)
	// 邀请函相关
	mux.HandleFunc("/v1/team_member/invite", route.HandleInviteMember)
	mux.HandleFunc("/v1/team_member/invitation/read", route.HandleMemberInvitationRead)
	mux.HandleFunc("/v1/team_member/invitation/detail", route.MemberInvitationDetail)

	mux.HandleFunc("/v1/team_member/resign", route.HandleMemberResign)

	// 集团管理路由
	mux.HandleFunc("/v1/group/new", route.NewGroupGet)
	mux.HandleFunc("/v1/group/create", route.CreateGroupPost)
	mux.HandleFunc("/v1/groups", route.GroupsGet)
	mux.HandleFunc("/v1/group/read", route.GroupReadGet)
	mux.HandleFunc("/v1/group/detail", route.GroupDetailGet)
	mux.HandleFunc("/v1/group/manage", route.GroupManageGet)
	mux.HandleFunc("/v1/group/invitations", route.GroupInvitationsGet)
	mux.HandleFunc("/v1/group/add_team", route.AddTeamToGroupPost)
	mux.HandleFunc("/v1/group/edit", route.HandleEditGroup)
	mux.HandleFunc("/v1/group/delete", route.DeleteGroupPost)
	// 集团成员管理路由
	mux.HandleFunc("/v1/group/member_add", route.GroupMemberAddGet)
	mux.HandleFunc("/v1/group/search_team", route.HandleGroupSearchTeam)
	mux.HandleFunc("/v1/group/member_remove", route.HandleGroupMemberRemove)
	mux.HandleFunc("/v1/group/member_invite", route.HandleGroupMemberInvite)
	mux.HandleFunc("/v1/group/member_invitation", route.HandleGroupMemberInvitation)

	//defined in route_family.go
	//家庭茶团杂项
	mux.HandleFunc("/v1/family/new", route.HandleNewFamily)
	mux.HandleFunc("/v1/family/detail", route.FamilyDetail)
	mux.HandleFunc("/v1/family/default", route.SetDefaultFamily)

	mux.HandleFunc("/v1/families/home", route.HomeFamilies)

	//defined in route_family_member.go
	mux.HandleFunc("/v1/family_member/sign_in_new", route.HandleFamilyMemberSignInNew)
	mux.HandleFunc("/v1/family_member/sign_in", route.HandleFamilyMemberSignIn)
	//mux.HandleFunc("/v1/family_member/sign_out", route.HandleFamilyMemberSignOut)

	//defined in route_talk_objective.go
	mux.HandleFunc("/v1/objective/new", route.HandleNewObjective)
	mux.HandleFunc("/v1/objective/square", route.ObjectiveSquare)
	mux.HandleFunc("/v1/objective/detail", route.ObjectiveDetail)
	mux.HandleFunc("/v1/objective/supplement", route.HandleObjectiveSupplement)

	//defined in route_talk_project.go
	mux.HandleFunc("/v1/project/new", route.HandleNewProject)
	mux.HandleFunc("/v1/project/detail", route.ProjectDetail)
	mux.HandleFunc("/v1/project/approve", route.ProjectApprove)
	mux.HandleFunc("/v1/project/place_update", route.HandleProjectPlace)

	// defined in route_talk_thread.go
	mux.HandleFunc("/v1/thread/draft", route.NewDraftThreadHandle)
	mux.HandleFunc("/v1/thread/detail", route.ThreadDetail)
	mux.HandleFunc("/v1/thread/approve", route.ThreadApprove)
	mux.HandleFunc("/v1/thread/supplement", route.HandleThreadSupplement)

	//定义在 route_talk_post.go
	mux.HandleFunc("/v1/post/draft", route.NewPostDraft)
	mux.HandleFunc("/v1/post/supplement", route.HandleSupplementPost)
	mux.HandleFunc("/v1/post/detail", route.PostDetail)
	//mux.HandleFunc("/v1/post/depth", route.PostDepth)

	//defined in route_action_appointment.go
	mux.HandleFunc("/v1/appointment/new", route.HandleNewAppointment)
	mux.HandleFunc("/v1/appointment/accept", route.AppointmentAccept)
	mux.HandleFunc("/v1/appointment/reject", route.AppointmentReject)
	mux.HandleFunc("/v1/appointment/detail", route.AppointmentDetail)

	// defined in route_action_see-seek.go
	mux.HandleFunc("/v1/see-seek/new", route.HandleNewSeeSeek)
	mux.HandleFunc("/v1/see-seek/detail", route.HandleSeeSeekDetail)
	// defined in route_see-seek_step.go
	mux.HandleFunc("/v1/see-seek/step2", route.HandleSeeSeekStep2)
	mux.HandleFunc("/v1/see-seek/step3", route.HandleSeeSeekStep3)
	mux.HandleFunc("/v1/see-seek/step4", route.HandleSeeSeekStep4)
	mux.HandleFunc("/v1/see-seek/step5", route.HandleSeeSeekStep5)

	// defined in route_action_brain-fire.go
	mux.HandleFunc("/v1/brain-fire/new", route.HandleNewBrainFire)
	mux.HandleFunc("/v1/brain-fire/detail", route.HandleBrainFireDetail)

	//defined in route_action_suggestion.go
	mux.HandleFunc("/v1/suggestion/new", route.HandleNewSuggestion)
	mux.HandleFunc("/v1/suggestion/detail", route.SuggestionDetail)

	// defined in route_goods.go
	mux.HandleFunc("/v1/goods/family_new", route.HandleGoodsFamilyNew)
	mux.HandleFunc("/v1/goods/family", route.GoodsFamily)
	mux.HandleFunc("/v1/goods/family_detail", route.GoodsFamilyDetail)
	mux.HandleFunc("/v1/goods/team_new", route.HandleGoodsTeamNew)
	mux.HandleFunc("/v1/goods/team", route.GoodsTeam)
	mux.HandleFunc("/v1/goods/team_detail", route.GoodsTeamDetail)
	mux.HandleFunc("/v1/goods/team_update", route.HandleGoodsTeamUpdate)
	mux.HandleFunc("/v1/goods/collect", route.GoodsCollect)
	mux.HandleFunc("/v1/goods/uncollect", route.GoodsUncollect)
	mux.HandleFunc("/v1/goods/eye_on", route.GoodsEyeOn)
	mux.HandleFunc("/v1/goods/detail", route.GoodsDetail)
	// defined in route_action_goods.go
	mux.HandleFunc("/v1/goods/project_new", route.HandleGoodsProjectNew)
	mux.HandleFunc("/v1/goods/project_detail", route.HandleGoodsProjectDetail)
	mux.HandleFunc("/v1/goods/project_readiness", route.HandleGoodsProjectReadiness)

	//defined in route_pilot.go
	mux.HandleFunc("/v1/pilot/new", route.NewPilot)
	mux.HandleFunc("/v1/pilot/add", route.AddPilot)
	mux.HandleFunc("/v1/pilot/office", route.OfficePilot)
	//mux.HandleFunc("/v1/pilot/detail", pilotDetail)

	//defined in route_office
	mux.HandleFunc("/v1/office/polite", route.Polite)
	mux.HandleFunc("/v1/office/draftThread", route.ActivateDraftThread)

	//defined in route_message.go
	mux.HandleFunc("/v1/message/accept", route.AcceptMessages)
	mux.HandleFunc("/v1/message/invitation_team", route.InvitationsTeam)
	mux.HandleFunc("/v1/message/invitation_group", route.InvitationGroup)

	//defined in route_place.go
	mux.HandleFunc("/v1/place/new", route.NewPlace)
	mux.HandleFunc("/v1/place/create", route.CreatePlace)
	mux.HandleFunc("/v1/place/detail", route.PlaceDetail)
	mux.HandleFunc("/v1/place/my", route.MyPlace)
	mux.HandleFunc("/v1/place/collect", route.PlaceCollect)

	//defined in route_action_environment.go
	mux.HandleFunc("/v1/environment/new", route.HandleNewEnvironment)
	mux.HandleFunc("/v1/environment/detail", route.HandleEnvironmentDetail)

	//defined in route_action_hazard.go
	mux.HandleFunc("/v1/hazard/new", route.HandleNewHazard)
	mux.HandleFunc("/v1/hazard/detail", route.HandleHazardDetail)

	//defined in route_action_skill.go
	mux.HandleFunc("/v1/skill/new", route.HandleNewSkill)
	mux.HandleFunc("/v1/skill/detail", route.HandleSkillDetail)
	mux.HandleFunc("/v1/skill_user/edit", route.HandleSkillUserEdit)
	mux.HandleFunc("/v1/skill_team/edit", route.HandleSkillTeamEdit)

	mux.HandleFunc("/v1/skills/user_list", route.HandleSkillsUserList)
	mux.HandleFunc("/v1/skills/team_list", route.HandleSkillsTeamList)

	//defined in route_action_magic.go
	mux.HandleFunc("/v1/magic/new", route.HandleNewMagic)
	mux.HandleFunc("/v1/magic/detail", route.HandleMagicDetail)
	mux.HandleFunc("/v1/magic/list", route.HandleMagicList)
	mux.HandleFunc("/v1/magic_user/edit", route.HandleMagicUserEdit)
	mux.HandleFunc("/v1/magic_team/edit", route.HandleMagicTeamEdit)

	mux.HandleFunc("/v1/magics/user_list", route.HandleMagicsUserList)
	mux.HandleFunc("/v1/magics/team_list", route.HandleMagicsTeamList)

	//defined in route_action_handicraft.go
	mux.HandleFunc("/v1/handicraft/new", route.HandleNewHandicraft)
	mux.HandleFunc("/v1/handicraft/detail", route.HandleHandicraftDetail)
	mux.HandleFunc("/v1/handicraft/list", route.HandleHandicraftList)
	// defined in route_handicraft_step.go
	mux.HandleFunc("/v1/handicraft/step2", route.HandleHandicraftStep2)
	mux.HandleFunc("/v1/handicraft/step3", route.HandleHandicraftStep3)
	mux.HandleFunc("/v1/handicraft/step4", route.HandleHandicraftStep4)
	mux.HandleFunc("/v1/handicraft/step5", route.HandleHandicraftStep5)

	//defined in route_evidence.go
	mux.HandleFunc("/v1/evidence/new", route.HandleNewEvidence)
	mux.HandleFunc("/v1/evidence/detail", route.HandleEvidenceDetail)

	//defined in route_action_risk.go
	mux.HandleFunc("/v1/risk/new", route.HandleNewRisk)
	mux.HandleFunc("/v1/risk/detail", route.HandleRiskDetail)

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
