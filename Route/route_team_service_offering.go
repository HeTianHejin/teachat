package route

import (
	"errors"
	"net/http"
	"strconv"
	dao "teachat/DAO"
)

// TeamServiceOfferings displays and processes team service offering management.
func TeamServiceOfferings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		teamServiceOfferingsGet(w, r)
	case http.MethodPost:
		teamServiceOfferingsPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func teamServiceOfferingAccess(r *http.Request, teamUUID string) (*dao.User, *dao.Team, bool, error) {
	sess, err := session(r)
	if err != nil {
		return nil, nil, false, err
	}
	user, err := sess.User()
	if err != nil {
		return nil, nil, false, err
	}
	if teamUUID == "" {
		return &user, nil, false, errors.New("missing team UUID")
	}
	team, err := dao.GetTeamByID(teamUUID)
	if err != nil {
		return &user, nil, false, err
	}
	isManager := user.Id == team.FounderId
	if !isManager {
		ceo, err := team.MemberCEO()
		if err != nil {
			return &user, nil, false, err
		}
		isManager = ceo.UserId == user.Id
	}
	if !isManager && !dao.IsVerifier(user.Id) {
		return &user, nil, false, errors.New("not authorized to access service offerings")
	}
	return &user, &team, isManager, nil
}

func teamServiceOfferingsGet(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.Query().Get("uuid")
	user, team, isManager, err := teamServiceOfferingAccess(r, uuid)
	if err != nil {
		if user == nil {
			http.Redirect(w, r, "/v1/login", http.StatusFound)
			return
		}
		report(w, *user, "你好，您无权管理这个茶团，或茶团资料暂时无法读取。")
		return
	}
	teamBean, err := fetchTeamBean(*team)
	if err != nil {
		report(w, *user, "你好，茶博士未能帮忙查看这个茶团资料，请稍后再试。")
		return
	}
	offerings, err := dao.GetTeamServiceOfferingsByTeamId(team.Id, false, r.Context())
	if err != nil {
		report(w, *user, "你好，茶博士未能帮忙查看服务项目，请稍后再试。")
		return
	}
	isVerifier := dao.IsVerifier(user.Id)
	if !isManager {
		pending := make([]*dao.TeamServiceOffering, 0)
		for _, offering := range offerings {
			if offering.Status == dao.PendingTeamServiceOfferingStatus {
				pending = append(pending, offering)
			}
		}
		offerings = pending
	}

	var editing *dao.TeamServiceOffering
	if offeringUUID := r.URL.Query().Get("offering_uuid"); offeringUUID != "" {
		editing = &dao.TeamServiceOffering{Uuid: offeringUUID}
		if err = editing.GetByIdOrUUID(r.Context()); err != nil || editing.TeamId != team.Id || !editing.IsEditable() {
			report(w, *user, "你好，这个服务项目不存在、所属团队不符或当前状态不可编辑。")
			return
		}
	}
	pageData := struct {
		SessUser   dao.User
		TeamBean   dao.TeamBean
		Offerings  []*dao.TeamServiceOffering
		Editing    *dao.TeamServiceOffering
		IsManager  bool
		IsVerifier bool
	}{SessUser: *user, TeamBean: teamBean, Offerings: offerings, Editing: editing, IsManager: isManager, IsVerifier: isVerifier}
	generateHTML(w, &pageData, "layout", "navbar.private", "team.service_offerings")
}

func teamServiceOfferingsPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	user, team, isManager, err := teamServiceOfferingAccess(r, r.FormValue("uuid"))
	if err != nil {
		if user == nil {
			http.Redirect(w, r, "/v1/login", http.StatusFound)
			return
		}
		report(w, *user, "你好，您无权管理这个茶团，或茶团资料暂时无法读取。")
		return
	}

	action := r.FormValue("action")
	if action == "create" || action == "update" {
		if !isManager {
			report(w, *user, "只有团队创建人或CEO可以创建和修改服务项目。")
			return
		}
		minutes, parseErr := strconv.Atoi(r.FormValue("estimated_minutes"))
		if parseErr != nil || minutes < 0 {
			report(w, *user, "预计耗时必须是大于或等于零的分钟数。")
			return
		}
		price, parseErr := strconv.ParseInt(r.FormValue("price_milligrams"), 10, 64)
		if parseErr != nil || price < 0 {
			report(w, *user, "服务价格必须是大于或等于零的毫克数。")
			return
		}
		offering := &dao.TeamServiceOffering{
			TeamId: team.Id, Name: r.FormValue("name"), Summary: r.FormValue("summary"),
			Description: r.FormValue("description"), TargetProblem: r.FormValue("target_problem"),
			Deliverables: r.FormValue("deliverables"), Requirements: r.FormValue("requirements"),
			EstimatedMinutes: minutes, PriceMilligrams: price, RecorderUserId: user.Id,
		}
		if action == "create" {
			err = offering.Create(r.Context())
		} else {
			offering.Uuid = r.FormValue("offering_uuid")
			err = offering.GetByIdOrUUID(r.Context())
			if err == nil && offering.TeamId != team.Id {
				err = errors.New("offering belongs to another team")
			}
			if err == nil {
				offering.Name = r.FormValue("name")
				offering.Summary = r.FormValue("summary")
				offering.Description = r.FormValue("description")
				offering.TargetProblem = r.FormValue("target_problem")
				offering.Deliverables = r.FormValue("deliverables")
				offering.Requirements = r.FormValue("requirements")
				offering.EstimatedMinutes = minutes
				offering.PriceMilligrams = price
				offering.RecorderUserId = user.Id
				err = offering.Update(r.Context())
			}
		}
		if err != nil {
			report(w, *user, "服务项目保存失败："+err.Error())
			return
		}
		redirectToTeamServiceOfferings(w, r, team.Uuid)
		return
	}

	offering := &dao.TeamServiceOffering{Uuid: r.FormValue("offering_uuid")}
	if err = offering.GetByIdOrUUID(r.Context()); err != nil || offering.TeamId != team.Id {
		report(w, *user, "你好，未能找到属于这个茶团的服务项目。")
		return
	}
	reason := r.FormValue("reason")
	switch action {
	case "submit":
		if !isManager {
			report(w, *user, "只有团队创建人或CEO可以提交服务项目审核。")
			return
		}
		err = offering.Submit(r.Context(), user.Id)
	case "approve":
		if !dao.IsVerifier(user.Id) {
			report(w, *user, "只有见证者团队成员可以审核并上架服务项目。")
			return
		}
		err = offering.Approve(r.Context(), user.Id, user.Id)
	case "reject":
		if !dao.IsVerifier(user.Id) {
			report(w, *user, "只有见证者团队成员可以审核服务项目。")
			return
		}
		err = offering.Reject(r.Context(), user.Id, reason)
	case "pause":
		if !isManager {
			report(w, *user, "只有团队创建人或CEO可以管理服务项目状态。")
			return
		}
		err = offering.Pause(r.Context(), user.Id, reason)
	case "resume":
		if !isManager {
			report(w, *user, "只有团队创建人或CEO可以管理服务项目状态。")
			return
		}
		err = offering.Resume(r.Context(), user.Id)
	case "unpublish":
		if !isManager {
			report(w, *user, "只有团队创建人或CEO可以管理服务项目状态。")
			return
		}
		err = offering.Unpublish(r.Context(), user.Id, reason)
	case "retire":
		if !isManager {
			report(w, *user, "只有团队创建人或CEO可以管理服务项目状态。")
			return
		}
		err = offering.Retire(r.Context(), user.Id, reason)
	default:
		report(w, *user, "不支持的服务项目操作。")
		return
	}
	if err != nil {
		report(w, *user, "服务项目操作失败："+err.Error())
		return
	}
	redirectToTeamServiceOfferings(w, r, team.Uuid)
}

func redirectToTeamServiceOfferings(w http.ResponseWriter, r *http.Request, teamUUID string) {
	http.Redirect(w, r, "/v1/team/service_offerings?uuid="+teamUUID, http.StatusSeeOther)
}

// GET /v1/team/service_offerings/unpublished?uuid=
// 显示团队已下架的服务项目
func TeamUnpublishedServiceOfferings(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，请提交有效的茶团编号，请稍后再试。")
		return
	}
	team, err := dao.GetTeamByID(uuid)
	if err != nil {
		report(w, s_u, "你好，茶博士未能帮忙查看这个茶团资料，请稍后再试。")
		return
	}
	if team.IsPrivate {
		isMember, err := team.IsActiveMember(s_u.Id)
		if err != nil || !isMember {
			report(w, s_u, "你好，这个茶团是私有的，你不能查看。")
			return
		}
	}

	teamBean, err := fetchTeamBean(team)
	if err != nil {
		report(w, s_u, "你好，茶博士未能帮忙查看这个茶团资料，请稍后再试。")
		return
	}
	offerings, err := dao.GetTeamServiceOfferingsByTeamId(team.Id, false, r.Context())
	if err != nil {
		report(w, s_u, "你好，茶博士未能帮忙查看已下架服务项目，请稍后再试。")
		return
	}
	unpublished := make([]*dao.TeamServiceOffering, 0)
	for _, offering := range offerings {
		if offering.Status == dao.UnpublishedTeamServiceOfferingStatus {
			unpublished = append(unpublished, offering)
		}
	}

	pageData := struct {
		SessUser  dao.User
		TeamBean  dao.TeamBean
		Offerings []*dao.TeamServiceOffering
	}{SessUser: s_u, TeamBean: teamBean, Offerings: unpublished}
	generateHTML(w, &pageData, "layout", "navbar.private", "team.service_offerings_unpublished")
}
