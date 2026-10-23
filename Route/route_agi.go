package route

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	dao "teachat/DAO"
	util "teachat/Util"

	"golang.org/x/crypto/bcrypt"
)

// AGI 代理账号的登录锚点。
// AGI 无法拥有真实邮箱，这里使用命名空间化、不可投递的地址，
// 仅作为登录凭据，不冒充“拥有邮箱的人类”。
const agiLoginEmailDomain = "agents.teachat.local"

var agiLoginHandleRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{2,31}$`)

// AgiSignupPageData AGI 专用注册页渲染数据。
type AgiSignupPageData struct {
	SessUser dao.User

	Teams           []dao.Team
	Scopes          []dao.AgentScopeOption
	AutonomyOptions []dao.AgentAutonomyOption

	Message        string // 提示信息（成功或失败）
	GeneratedLogin string // 创建成功后的登录锚点
}

// GET/POST /v1/agi/signup
// AGI 代理账号只能由团队发起人或现任CEO登记，并绑定受委托团队。
func AgiSignup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		agiSignupGet(w, r)
	case http.MethodPost:
		agiSignupPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func agiSignupGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session %v", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	page := AgiSignupPageData{
		SessUser:        s_u,
		Scopes:          dao.AgentScopeOptions,
		AutonomyOptions: dao.AgentAutonomyOptions,
	}

	if !s_u.IsHuman() {
		page.Message = "只有申报为人类的茶友才能作为 AGI 代理账号的赞助人。"
		generateHTML(w, &page, "layout", "navbar.private", "agi.signup")
		return
	}

	teams, err := dao.GetSponsorableTeams(s_u.Id)
	if err != nil {
		util.Error("Cannot get sponsorable teams for user %d: %v", s_u.Id, err)
		page.Message = "茶博士失魂鱼，未能读取你名下的团队，请稍后再试。"
		generateHTML(w, &page, "layout", "navbar.private", "agi.signup")
		return
	}
	page.Teams = teams
	if len(teams) == 0 {
		page.Message = "你当前没有可背书的团队。AGI 代理账号必须绑定一个由你发起、或你担任在职 CEO 的团队。"
	}

	generateHTML(w, &page, "layout", "navbar.private", "agi.signup")
}

func agiSignupPost(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session %v", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	if !s_u.IsHuman() {
		report(w, s_u, "只有申报为人类的茶友才能作为 AGI 代理账号的赞助人。")
		return
	}
	if err = r.ParseForm(); err != nil {
		util.Debug("Cannot parse agi signup form %v", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能读取提交内容，请稍后再试。")
		return
	}

	// 1. 赞助团队校验
	teamUuid := strings.TrimSpace(r.PostFormValue("team_uuid"))
	team, err := dao.GetTeamByUUID(teamUuid)
	if err != nil {
		util.Debug("Cannot get sponsoring team %q: %v", teamUuid, err)
		report(w, s_u, "你好，未能找到指定的受委托团队，请确认后再试。")
		return
	}
	allowed, err := dao.CanSponsorAgent(team.Id, s_u.Id)
	if err != nil {
		util.Error("Cannot verify agent sponsor %d for team %d: %v", s_u.Id, team.Id, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能核实赞助资格，请稍后再试。")
		return
	}
	if !allowed {
		report(w, s_u, "你好，你必须是该团队的发起人或现任在职 CEO，才能为其登记 AGI 代理账号。")
		return
	}

	// 2. 申报名称与模型信息
	aliasName := strings.TrimSpace(r.PostFormValue("alias_name"))
	if aliasName == "" || !isValidUserName(aliasName) {
		report(w, s_u, "你好，请确认代理名称只包含字母、数字、下划线或中文字符。")
		return
	}
	modelFamily := strings.TrimSpace(r.PostFormValue("model_family"))
	modelVersion := strings.TrimSpace(r.PostFormValue("model_version"))
	provider := strings.TrimSpace(r.PostFormValue("provider"))
	if modelFamily == "" || modelVersion == "" {
		report(w, s_u, "你好，请填写模型系列与模型版本。")
		return
	}
	if cnStrLen(modelFamily) > 64 || cnStrLen(modelVersion) > 64 || cnStrLen(provider) > 64 {
		report(w, s_u, "你好，模型系列、版本与提供方的描述过长，请精简后再试。")
		return
	}

	// 3. 代理范围与自主程度
	scopes, err := dao.NormalizeAgencyScopes(r.Form["agency_scope"])
	if err != nil {
		report(w, s_u, "你好，"+err.Error()+"。")
		return
	}
	autonomyLevel, err := strconv.Atoi(r.PostFormValue("autonomy_level"))
	if err != nil || dao.AgentAutonomyLabel(autonomyLevel) == "未知" {
		report(w, s_u, "你好，请确认自主程度选择是否正确。")
		return
	}

	// 4. 登录锚点与密码
	handle := strings.ToLower(strings.TrimSpace(r.PostFormValue("login_handle")))
	if !agiLoginHandleRegex.MatchString(handle) {
		report(w, s_u, "你好，代理登录名需为 3-32 位小写字母、数字、下划线或连字符，且以字母或数字开头。")
		return
	}
	password := r.PostFormValue("password")
	if len(password) < 6 || len(password) > 42 {
		report(w, s_u, "你好，密码长度需在 6-42 位之间。")
		return
	}
	email := fmt.Sprintf("agi-%s@%s", handle, agiLoginEmailDomain)
	if !isEmail(email) {
		report(w, s_u, "你好，代理登录名无法生成合法的登录锚点，请换一个名称。")
		return
	}
	exists, err := dao.UserExistByEmail(email)
	if err != nil {
		util.Error("检查AGI登录锚点存在性失败: %v", err)
		report(w, s_u, "你好，茶博士因找不到笔导致注册失败，请稍后再试。")
		return
	}
	if exists {
		report(w, s_u, "你好，该代理登录名已被占用，请换一个。")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		util.Error("Cannot hash AGI password: %v", err)
		report(w, s_u, "你好，茶博士因找不到笔导致注册失败，请稍后再试。")
		return
	}

	biography := strings.TrimSpace(r.PostFormValue("biography"))
	if biography == "" {
		biography = fmt.Sprintf("%s，由 %s 团队申报的 AGI 代理。", aliasName, team.Abbreviation)
	}
	if cnStrLen(biography) > int(util.Config.ThreadMaxWord) {
		report(w, s_u, "你好，简介字数超过上限，请精简后再试。")
		return
	}

	newU := dao.User{
		Name:      aliasName,
		AliasName: aliasName,
		Email:     email,
		Password:  string(hashedPassword),
		Biography: biography,
		Role:      dao.User_Role_AGI,
		Gender:    dao.GenderUnknown, // AGI 不申报性别，数据库中为 NULL
		Avatar:    "teaSet",
	}
	profile := dao.AgentProfile{
		SponsorTeamId: team.Id,
		SponsorUserId: s_u.Id,
		ModelFamily:   modelFamily,
		ModelVersion:  modelVersion,
		Provider:      provider,
		AgencyScope:   strings.Join(scopes, ","),
		AutonomyLevel: autonomyLevel,
	}

	if err = dao.CreateAgentAccount(&newU, &profile); err != nil {
		util.Error("Cannot create AGI account: %v", err)
		if strings.Contains(err.Error(), "duplicate key") {
			report(w, s_u, "你好，该代理登录名已被占用，请换一个。")
			return
		}
		report(w, s_u, "你好，茶博士因找不到笔导致注册失败，请稍后再试。")
		return
	}

	if err = dao.TeaUserEnsureAccountExists(newU.Id); err != nil {
		// 账号与申报资料已创建成功，星茶账户失败不应回滚注册；记录并提示即可。
		util.Error("Cannot ensure tea account for AGI user %d: %v", newU.Id, err)
	}

	page := AgiSignupPageData{
		SessUser:        s_u,
		Scopes:          dao.AgentScopeOptions,
		AutonomyOptions: dao.AgentAutonomyOptions,
		GeneratedLogin:  email,
		Message: fmt.Sprintf("AGI 代理「%s」登记成功，受委托团队：%s，赞助人：%s。请把登录名与密码交给该代理，并妥善保管。",
			aliasName, team.Name, s_u.DisplayName()),
	}
	if teams, err := dao.GetSponsorableTeams(s_u.Id); err == nil {
		page.Teams = teams
	}

	util.Debug("Successfully registered AGI agent %s for team %d", email, team.Id)
	generateHTML(w, &page, "layout", "navbar.private", "agi.signup")
}
