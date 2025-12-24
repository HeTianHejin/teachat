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
	mux.HandleFunc("/v1/team/members/left", route.TeamMembersLeft)

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
	// 邀请函-成员个人
	mux.HandleFunc("/v1/invitations/member", route.InvitationsReceived)
	// 退出声明-成员个人
	mux.HandleFunc("/v1/resignations/member", route.ResignationsReceived)
	// 邀请函相关
	mux.HandleFunc("/v1/team_member/invite", route.HandleInviteMember)
	mux.HandleFunc("/v1/team_member/invitation/read", route.HandleMemberInvitationRead)
	mux.HandleFunc("/v1/team_member/invitation/detail", route.MemberInvitationDetail)

	mux.HandleFunc("/v1/team_member/resign", route.HandleMemberResign)
	mux.HandleFunc("/v1/team_member/resigned", route.TeamMemberResigned)
	mux.HandleFunc("/v1/team_member/resignation/detail", route.TeamMemberResignationDetail)

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
	mux.HandleFunc("/v1/family/edit", route.HandleEditFamily)
	mux.HandleFunc("/v1/family/default", route.SetDefaultFamily)

	mux.HandleFunc("/v1/family/home", route.HomeFamilies)
	mux.HandleFunc("/v1/family/home/private", route.HomePrivateFamilies)
	mux.HandleFunc("/v1/family/tree", route.FamilyTree)
	mux.HandleFunc("/v1/families/parent", route.ParentFamilies)
	mux.HandleFunc("/v1/families/parent/private", route.ParentPrivateFamilies)
	mux.HandleFunc("/v1/families/child", route.ChildFamilies)
	mux.HandleFunc("/v1/families/in-laws", route.InLawsFamilies)
	mux.HandleFunc("/v1/families/in-laws/private", route.InLawsPrivateFamilies)
	mux.HandleFunc("/v1/families/gone", route.GoneFamilies)
	mux.HandleFunc("/v1/families/gone/private", route.GonePrivateFamilies)

	//defined in route_family_member.go
	mux.HandleFunc("/v1/family_member/sign_in_new", route.HandleFamilyMemberSignInNew)
	mux.HandleFunc("/v1/family_member/sign_in", route.HandleFamilyMemberSignIn)
	mux.HandleFunc("/v1/family_member/detail", route.FamilyMemberDetail)
	mux.HandleFunc("/v1/family_member/edit", route.HandleFamilyMemberEdit)
	//mux.HandleFunc("/v1/family_member/sign_out", route.HandleFamilyMemberSignOut)

	//defined in route_family.go - family member add
	mux.HandleFunc("/v1/family/member_add", route.FamilyMemberAddGet)
	mux.HandleFunc("/v1/family/search_user", route.HandleFamilySearchUser)

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

	//defined in route_notification.go
	mux.HandleFunc("/v1/notification/accept", route.AcceptNotifications)
	mux.HandleFunc("/v1/notification/invitation_team", route.InvitationsTeam)
	mux.HandleFunc("/v1/notification/invitation_group", route.InvitationGroup)

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

	//defined in route_message_box.go
	mux.HandleFunc("/v1/message_box/detail", route.MessageBoxDetail)
	mux.HandleFunc("/v1/message/read", route.MessageRead)
	mux.HandleFunc("/v1/message/delete", route.MessageDelete)
	mux.HandleFunc("/v1/message/send", route.HandleMessageTeamSend)
	mux.HandleFunc("/v1/message/announcement/send", route.MessageAnnouncementSend)

	//定义在 Route_balance.go
	mux.HandleFunc("/v1/balance/fairnessmug", route.FairnessMug)

	// 用户茶叶账户系统路由
	mux.HandleFunc("/v1/desk", route.HandleDesk)                                                                      //用户茶叶账户入口页面
	mux.HandleFunc("/v1/tea/user/account", route.GetTeaUserAccount)                                                   // 用户茶叶账户信息API
	mux.HandleFunc("/v1/tea/user/account/freeze", route.FreezeTeaUserAccount)                                         // 用户茶叶账户冻结API
	mux.HandleFunc("/v1/tea/user/account/unfreeze", route.UnfreezeTeaUserAccount)                                     // 用户茶叶账户解冻API
	mux.HandleFunc("/v1/tea/user/transfer/user_to_user", route.CreateTeaUserToUserTransferAPI)                        // 用户对用户创建转账API
	mux.HandleFunc("/v1/tea/user/transfer/user_to_team", route.CreateTeaUserToTeamTransferAPI)                        // 用户对团队创建转账API
	mux.HandleFunc("/v1/tea/user/transfers/outs/user_to_user", route.GetTeaUserToUserTransferOutsAPI)                 // 用户对用户转出记录API
	mux.HandleFunc("/v1/tea/user/transfers/outs/user_to_user/page", route.GetTeaUserToTeamTransferOutsAPI)            // 用户对团队转出记录API
	mux.HandleFunc("/v1/tea/user/transfers/pending/user_to_user", route.GetTeaUserPendingUserToUserTransfersAPI)      // 用户待确认用户对用户转账API
	mux.HandleFunc("/v1/tea/user/transfers/pending/user_to_team", route.GetTeaUserPendingUserToTeamTransfersAPI)      // 用户待确认用户对团队转账API
	mux.HandleFunc("/v1/tea/user/transfers/pending/user_to_user/page", route.HandleTeaUserPendingUserToUserTransfers) // 用户待确认用户对用户转账页面路由
	mux.HandleFunc("/v1/tea/user/transfers/pending/user_to_team/page", route.HandleTeaUserPendingUserToTeamTransfers) // 用户待确认用户对团队转账页面路由
	mux.HandleFunc("/v1/tea/user/transfers/history/user_to_user", route.GetTeaUserToUserTransferHistoryAPI)           // 用户对用户转账历史API
	mux.HandleFunc("/v1/tea/user/transfers/history/user_to_team", route.GetTeaUserToTeamTransferHistoryAPI)           // 用户对团队转账历史API
	mux.HandleFunc("/v1/tea/user/transfers/history/user_to_user/page", route.HandleTeaUserTransferHistory)            // 用户对用户转账历史页面路由
	mux.HandleFunc("/v1/tea/user/transfers/history/user_to_team/page", route.HandleTeaTeamTransferHistory)            // 用户对团队转账历史页面路由
	mux.HandleFunc("/v1/tea/user/transfers/ins/user_from_user", route.GetTeaUserFromUserTransferInsAPI)               // 用户接收用户转入记录API - 接收历史（所有状态）
	mux.HandleFunc("/v1/tea/user/transfers/ins/user_from_user/completed", route.GetTeaUserCompletedTransferInsAPI)     // 用户接收用户转入记录API - 收入记录（仅已完成）
	mux.HandleFunc("/v1/tea/user/transfers/ins/user_from_team", route.GetTeaUserFromTeamTransferInsAPI)               // 用户接收团队转入记录API
	mux.HandleFunc("/v1/tea/user/transfers/ins/user_from_user/page", route.HandleTeaUserFromUserTransferIns)          // 用户接收用户转入记录页面路由 - 接收历史（所有状态）
	mux.HandleFunc("/v1/tea/user/transfers/ins/user_from_user/completed/page", route.HandleTeaUserCompletedTransferIns) // 用户接收用户转入记录页面路由 - 收入记录（仅已完成）
	mux.HandleFunc("/v1/tea/user/transfers/ins/user_from_team/page", route.HandleTeaUserFromTeamTransferIns)          // 用户接收团队转入记录页面路由
	// 新增：区分用户对用户和用户对团队转账的确认/拒绝路由
	mux.HandleFunc("/v1/tea/user/transfer/confirm/user_to_user", route.ConfirmTeaUserToUserTransferAPI) // 确认接收用户对用户转账API
	mux.HandleFunc("/v1/tea/user/transfer/reject/user_to_user", route.RejectTeaUserToUserTransferAPI)   // 拒绝接收用户对用户转账API
	mux.HandleFunc("/v1/tea/user/transfer/confirm/user_to_team", route.ConfirmTeaUserToTeamTransferAPI) // 确认接收用户对团队转账API
	mux.HandleFunc("/v1/tea/user/transfer/reject/user_to_team", route.RejectTeaUserToTeamTransferAPI)   // 拒绝接收用户对团队转账API

	// 团队茶叶账户系统路由（对应用户版功能）
	mux.HandleFunc("/v1/tea/team/account/api", route.GetTeaTeamAccount)                                                   // 团队茶叶账户API
	mux.HandleFunc("/v1/tea/team/transactions/api", route.GetTeaTeamTransactionHistory)                                 // 团队交易历史API
	mux.HandleFunc("/v1/tea/team/account/freeze", route.FreezeTeaTeamAccount)                                           // 团队茶叶账户冻结API
	mux.HandleFunc("/v1/tea/team/account/unfreeze", route.UnfreezeTeaTeamAccount)                                       // 团队茶叶账户解冻API
	mux.HandleFunc("/v1/tea/team/transfer/team_to_user", route.CreateTeaTeamToUserTransferAPI)                          // 团队对用户创建转账API
	mux.HandleFunc("/v1/tea/team/transfer/team_to_team", route.CreateTeaTeamToTeamTransferAPI)                          // 团队对团队创建转账API
	mux.HandleFunc("/v1/tea/team/transfers/outs/team_to_user", route.GetTeaTeamToUserTransferOutsAPI)                   // 团队对用户转出记录API
	mux.HandleFunc("/v1/tea/team/transfers/outs/team_to_team", route.GetTeaTeamToTeamTransferOutsAPI)                   // 团队对团队转出记录API
	mux.HandleFunc("/v1/tea/team/transfers/pending/team_to_user", route.GetTeaTeamPendingTeamToUserTransfersAPI)        // 团队待确认团队对用户转账API
	mux.HandleFunc("/v1/tea/team/transfers/pending/team_to_team", route.GetTeaTeamPendingTeamToTeamTransfersAPI)        // 团队待确认团队对团队转账API
	mux.HandleFunc("/v1/tea/team/transfers/pending/team_to_user/page", route.HandleTeaTeamPendingTeamToUserTransfers)   // 团队待确认团队对用户转账页面路由
	mux.HandleFunc("/v1/tea/team/transfers/pending/team_to_team/page", route.HandleTeaTeamPendingTeamToTeamTransfers)   // 团队待确认团队对团队转账页面路由
	mux.HandleFunc("/v1/tea/team/transfers/history/team_to_user", route.GetTeaTeamToUserTransferHistoryAPI)             // 团队对用户转账历史API
	mux.HandleFunc("/v1/tea/team/transfers/history/team_to_team", route.GetTeaTeamToTeamTransferHistoryAPI)             // 团队对团队转账历史API
	mux.HandleFunc("/v1/tea/team/transfers/history/team_to_user/page", route.HandleTeaTeamToUserTransferHistory)        // 团队对用户转账历史页面路由
	mux.HandleFunc("/v1/tea/team/transfers/history/team_to_team/page", route.HandleTeaTeamToTeamTransferHistory)        // 团队对团队转账历史页面路由
	mux.HandleFunc("/v1/tea/team/transfers/ins/team_from_user", route.GetTeaTeamFromUserTransferInsAPI)                 // 团队接收用户转入记录API
	mux.HandleFunc("/v1/tea/team/transfers/ins/team_from_team", route.GetTeaTeamFromTeamTransferInsAPI)                 // 团队接收团队转入记录API
	mux.HandleFunc("/v1/tea/team/transfers/ins/team_from_user/page", route.HandleTeaTeamFromUserTransferIns)            // 团队接收用户转入记录页面路由
	mux.HandleFunc("/v1/tea/team/transfers/ins/team_from_team/page", route.HandleTeaTeamFromTeamTransferIns)            // 团队接收团队转入记录页面路由
	// 新增：区分团队对用户和团队对团队转账的确认/拒绝路由
	mux.HandleFunc("/v1/tea/team/transfer/confirm/team_to_user", route.ConfirmTeaTeamToUserTransferAPI)                 // 确认接收团队对用户转账API
	mux.HandleFunc("/v1/tea/team/transfer/reject/team_to_user", route.RejectTeaTeamToUserTransferAPI)                   // 拒绝接收团队对用户转账API
	mux.HandleFunc("/v1/tea/team/transfer/confirm/team_to_team", route.ConfirmTeaTeamToTeamTransferAPI)                 // 确认接收团队对团队转账API
	mux.HandleFunc("/v1/tea/team/transfer/reject/team_to_team", route.RejectTeaTeamToTeamTransferAPI)                   // 拒绝接收团队对团队转账API
	// 团队转账审批路由
	mux.HandleFunc("/v1/tea/team/transfer/approve/team_to_user", route.ApproveTeaTeamToUserTransferAPI)                 // 审批团队对用户转账API
	mux.HandleFunc("/v1/tea/team/transfer/approve/team_to_team", route.ApproveTeaTeamToTeamTransferAPI)                 // 审批团队对团队转账API
	mux.HandleFunc("/v1/tea/team/transfer/reject_approval/team_to_user", route.RejectTeaTeamToUserTransferApprovalAPI) // 拒绝审批团队对用户转账API
	mux.HandleFunc("/v1/tea/team/transfer/reject_approval/team_to_team", route.RejectTeaTeamToTeamTransferApprovalAPI) // 拒绝审批团队对团队转账API
	mux.HandleFunc("/v1/tea/team/account", route.HandleTeaTeamTeaAccount)                              // 团队茶叶账户页面路由
	mux.HandleFunc("/v1/tea/team/transfers/pending/page", route.HandleTeaTeamPendingIncomingTransfers) // 团队待确认转入转账页面路由
	mux.HandleFunc("/v1/tea/team/transfers/pending/api", route.GetTeaTeamPendingIncomingTransfers)     // 团队待确认转入转账API
	mux.HandleFunc("/v1/tea/team/transactions/page", route.HandleTeaTeamTransactionHistory)            // 团队交易流水页面路由
	mux.HandleFunc("/v1/tea/team/operations/history/page", route.HandleTeaTeamOperationsHistory)       // 团队操作历史页面路由

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
	// 启动定时处理过期转账的任务
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // 每5分钟检查一次
		defer ticker.Stop()

		for range ticker.C {
			log.Println("开始处理过期转账...")
			if err := route.ProcessExpiredTransfersJob(); err != nil {
				log.Printf("处理过期转账失败: %v", err)
			} else {
				log.Println("过期转账处理完成")
			}
		}
	}()

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
