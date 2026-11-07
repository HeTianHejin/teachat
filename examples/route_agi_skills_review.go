package main

// 独立客户端脚本：以 sql/seed_data.sql 中预设的 AGI 见证者（users.id=3）身份
// 登录 TeaChat，读取「新登记技能」待审核列表并打印数量与技能名称。
//
// 在仓库根目录运行：
//
//	go run examples/route_agi_skills_review.go
//	go run examples/route_agi_skills_review.go -password '<其他密码>' -base http://192.168.200.19:8000/v1/
//	go run examples/route_agi_skills_review.go -probe-only   # 只做网络/会话/JSON 预检
//
// 默认使用本地 seed 测试账号的明文口令（太空船长 / 见证者1号 / 见证者2号共用），
// 可用 -password 或环境变量 TEACHAT_AGI_PASSWORD 覆盖；该口令仅限测试环境。
// 脚本先用 bcrypt 离线核对口令与预设哈希一致，再向服务端提交登录表单。
// seed 数据只保存 bcrypt 哈希；不依赖 DAO/Route 包，也不连数据库。
// 唯一外部依赖是 go.mod 已有的 x/crypto/bcrypt。

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	defaultBaseURL    = "http://192.168.200.19:8000/v1/"
	watchword         = "闻香识茶"
	sessionCookieName = "_cookie"
	passwordEnvVar    = "TEACHAT_AGI_PASSWORD" // 运行时修改为预设用户填充的明文密码
)

// 预设身份：sql/seed_data.sql users 表 id=3 的记录，以及 agent_profiles 的申报范围。
type presetAGI struct {
	ID           int
	UUID         string
	Name         string
	FamilyName   string
	GivenName    string
	Alias        string
	Email        string
	PasswordHash string
	Biography    string
	Role         string
	Gender       string // seed 中为 NULL
	Avatar       string
	AgencyScope  string
}

var agi = presetAGI{
	ID:           3,
	UUID:         "38be3046-b192-44b6-7afb-070a7e98",
	Name:         "见证者2号",
	FamilyName:   "",
	GivenName:    "",
	Alias:        "AGI见证者",
	Email:        "teachat-witness@storetravels.com",
	PasswordHash: "$2a$10$JWf84LSbH.88qAELA/ofM.OaybTDqJaG7bzovD0o8/OT7GvJaGBQ6",
	Biography:    "我是AGI见证者，欢迎来到星际茶会！",
	Role:         "AGI",
	Gender:       "NULL",
	Avatar:       "teaSet",
	AgencyScope:  "read,post,review",
}

// GET /v1/api/verifier/skills 的响应外层结构（Route.ApiResponse）。
type apiEnvelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    *skillsData `json:"data"`
}

type skillsData struct {
	Skills []skillItem `json:"skills"`
	Total  int         `json:"total"`
}

// Route.verifierSkillItem 中与本脚本相关的字段。
type skillItem struct {
	ID                int    `json:"id"`
	UUID              string `json:"uuid"`
	UserID            int    `json:"user_id"`
	Name              string `json:"name"`
	Nickname          string `json:"nickname"`
	CategoryLabel     string `json:"category_label"`
	ReviewStatusLabel string `json:"review_status_label"`
	CreatedAt         string `json:"created_at"`
}

func main() {
	baseFlag := flag.String("base", defaultBaseURL, "站点根地址，需以 /v1/ 结尾")
	passwordFlag := flag.String("password", "", "预设 AGI 账号的明文密码；默认用 seed 测试口令，也可用环境变量 "+passwordEnvVar)
	probeOnly := flag.Bool("probe-only", false, "只做网络、会话与 JSON 响应预检，不登录")
	flag.Parse()

	password := *passwordFlag
	if password == "" {
		password = os.Getenv(passwordEnvVar)
	}
	if password == "" {
		fatal("no password")
	}
	base, err := parseBase(*baseFlag)
	if err != nil {
		fatal("%v", err)
	}

	step := 0
	begin := func(title string) {
		step++
		fmt.Printf("[%d] %s\n", step, title)
	}
	ok := func(format string, args ...any) {
		fmt.Printf("    OK: %s\n", fmt.Sprintf(format, args...))
	}

	skillsURL := resolve(base, "api/verifier/skills")
	fmt.Printf("身份: id=%d uuid=%s 昵称=%s 花名=%s 邮箱=%s role=%s gender=%s avatar=%s\n",
		agi.ID, agi.UUID, agi.Name, agi.Alias, agi.Email, agi.Role, agi.Gender, agi.Avatar)
	fmt.Printf("申报代理范围: %s（含 review）；目标: %s\n\n", agi.AgencyScope, base.String())

	// 预检阶段使用不带 CookieJar 的客户端，避免污染后续会话。
	probe := newClient(nil)

	begin("验证网络与登录入口")
	resp, body, err := fetch(probe, http.MethodGet, resolve(base, "login"), nil, "")
	if err != nil {
		fatal("无法访问 %s：%v", base.String(), err)
	}
	if resp.StatusCode != http.StatusOK {
		fatal("GET %s 返回 %d（期望 200），请检查服务地址与网络", base.String(), resp.StatusCode)
	}
	page := string(body)
	if !strings.Contains(page, `action="/v1/authenticate"`) ||
		!strings.Contains(page, `name="email"`) ||
		!strings.Contains(page, `name="password"`) ||
		!strings.Contains(page, `name="watchword"`) {
		fatal("登录页缺少预期表单字段，确认 -base 是否指向 /v1/：%s", snippet(body))
	}
	ok("GET %s -> 200，登录表单 POST /v1/authenticate 字段齐全", base.String())

	begin("验证 cookie 会话校验与 JSON 错误响应解析")
	for _, probeCase := range []struct {
		name    string
		cookies []*http.Cookie
	}{
		{"无 cookie", nil},
		{"伪造 cookie", []*http.Cookie{{Name: sessionCookieName, Value: "invalid-session-probe"}}},
	} {
		resp, body, err = fetch(probe, http.MethodGet, skillsURL, nil, "", probeCase.cookies...)
		if err != nil {
			fatal("%s 请求失败：%v", probeCase.name, err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			fatal("%s 应返回 401，实际 %d：%s", probeCase.name, resp.StatusCode, snippet(body))
		}
		env, err := decodeEnvelope(resp.Header.Get("Content-Type"), body)
		if err != nil {
			fatal("%s 的 401 响应解析失败：%v", probeCase.name, err)
		}
		if env.Success || env.Message == "" {
			fatal("%s 的 401 JSON 结构异常：success=%t message=%q", probeCase.name, env.Success, env.Message)
		}
		ok("%s -> 401 %q，JSON 外层结构可解析", probeCase.name, env.Message)
	}

	begin("验证 API 方法校验（另一种错误响应）")
	resp, body, err = fetch(probe, http.MethodPost, skillsURL, nil, "")
	if err != nil {
		fatal("方法校验请求失败：%v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		fatal("POST 列表接口应返回 405，实际 %d：%s", resp.StatusCode, snippet(body))
	}
	env, err := decodeEnvelope(resp.Header.Get("Content-Type"), body)
	if err != nil {
		fatal("405 响应解析失败：%v", err)
	}
	if env.Success || env.Message == "" {
		fatal("405 JSON 结构异常：success=%t message=%q", env.Success, env.Message)
	}
	ok("POST -> 405 %q，错误响应解析一致", env.Message)

	if *probeOnly {
		fmt.Println("\n仅预检模式：网络、会话校验与 API 响应解析均已通过，未尝试登录。")
		return
	}

	begin("离线核对账号凭据与预设身份")
	if err := bcrypt.CompareHashAndPassword([]byte(agi.PasswordHash), []byte(password)); err != nil {
		fatal("提供的密码与预设 id=%d 的 bcrypt 哈希不匹配（%v）；请确认账号与密码", agi.ID, err)
	}
	ok("密码与预设哈希匹配，账号为 %s（role=%s）", agi.Email, agi.Role)

	begin("登录并捕获 _cookie 会话")
	jar, err := cookiejar.New(nil)
	if err != nil {
		fatal("创建 cookiejar 失败：%v", err)
	}
	client := newClient(jar)
	authURL := resolve(base, "authenticate")
	form := url.Values{
		"email":     {agi.Email},
		"password":  {password},
		"watchword": {watchword},
		"footprint": {"/v1/"}, // 登录成功后站内跳转目标
		"query":     {""},
	}
	resp, body, err = fetch(client, http.MethodPost, authURL, strings.NewReader(form.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		fatal("登录请求失败：%v", err)
	}
	if resp.StatusCode != http.StatusFound {
		// 登录失败返回的是 200 HTML 提示页，不是 JSON，也不是 302。
		fatal("登录未成功：HTTP %d（成功应为 302），页面片段：%s", resp.StatusCode, snippet(body))
	}
	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == sessionCookieName && c.Value != "" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		fatal("302 响应中没有有效的 Set-Cookie: %s", sessionCookieName)
	}
	ok("POST %s -> 302 %s；Set-Cookie %s=<已隐藏> HttpOnly=%t MaxAge=%d",
		authURL, resp.Header.Get("Location"), sessionCookieName, sessionCookie.HttpOnly, sessionCookie.MaxAge)

	// 服务端设置 cookie 时未显式给 Path，浏览器/cookiejar 会按 /v1/authenticate 推导出 /v1。
	// 显式写入一次，确保后续 /v1/api/... 请求一定带上会话。
	authParsed, err := url.Parse(authURL)
	if err != nil {
		fatal("解析登录 URL 失败：%v", err)
	}
	jar.SetCookies(authParsed, []*http.Cookie{sessionCookie})
	skillsParsed, err := url.Parse(skillsURL)
	if err != nil {
		fatal("解析列表 URL 失败：%v", err)
	}
	if !hasCookie(jar.Cookies(skillsParsed), sessionCookieName) {
		fatal("会话 cookie 未能绑定到 %s，请检查 -base 路径", skillsURL)
	}
	ok("cookiejar 已把 %s 附加到 %s", sessionCookieName, skillsURL)

	begin("验证账号权限并读取待审核技能列表")
	resp, body, err = fetch(client, http.MethodGet, skillsURL, nil, "")
	if err != nil {
		fatal("读取待审核列表失败：%v", err)
	}
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		fatal("会话已失效（401）：%s", snippet(body))
	case http.StatusForbidden:
		fatal("账号权限不足（403，不是见证者团队成员）：%s", snippet(body))
	case http.StatusInternalServerError:
		fatal("服务端读取待审核列表失败（500）：%s", snippet(body))
	case http.StatusOK:
	default:
		fatal("读取列表返回意外状态 %d：%s", resp.StatusCode, snippet(body))
	}
	env, err = decodeEnvelope(resp.Header.Get("Content-Type"), body)
	if err != nil {
		fatal("列表响应解析失败：%v", err)
	}
	if !env.Success || env.Data == nil {
		fatal("列表响应不是成功结构：success=%t message=%q data=%v", env.Success, env.Message, env.Data != nil)
	}
	if env.Data.Total != len(env.Data.Skills) {
		fatal("响应自相矛盾：total=%d，但 skills 实际 %d 条", env.Data.Total, len(env.Data.Skills))
	}
	ok("200 success=true，见证者成员资格有效；total=%d，skills=%d 条，message=%q",
		env.Data.Total, len(env.Data.Skills), env.Message)

	fmt.Printf("\n待审核技能数量: %d\n", env.Data.Total)
	for i, s := range env.Data.Skills {
		fmt.Printf("  %d. %s (id=%d, uuid=%s, 分类=%s, 状态=%s)\n",
			i+1, s.Name, s.ID, s.UUID, s.CategoryLabel, s.ReviewStatusLabel)
	}
	if env.Data.Total == 0 {
		fmt.Println("  （暂无待审核技能）")
	}
}

func parseBase(raw string) (*url.URL, error) {
	if !strings.HasSuffix(raw, "/") {
		raw += "/"
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("非法的 -base 地址 %q", raw)
	}
	if !strings.HasSuffix(u.Path, "/v1/") {
		return nil, fmt.Errorf("-base 需以 /v1/ 结尾，当前路径为 %q", u.Path)
	}
	return u, nil
}

func resolve(base *url.URL, rel string) string {
	return base.ResolveReference(&url.URL{Path: rel}).String()
}

func newClient(jar http.CookieJar) *http.Client {
	return &http.Client{
		Jar:     jar,
		Timeout: 15 * time.Second,
		// 不自动跟随 302：登录成功与否要看原始响应里的 Set-Cookie。
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func fetch(client *http.Client, method, rawURL string, body io.Reader, contentType string, extraCookies ...*http.Cookie) (*http.Response, []byte, error) {
	req, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "teachat-agi-skills-review/1.0")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for _, c := range extraCookies {
		req.AddCookie(c)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return resp, nil, err
	}
	return resp, data, nil
}

func decodeEnvelope(contentType string, body []byte) (apiEnvelope, error) {
	var env apiEnvelope
	if !strings.Contains(strings.ToLower(contentType), "application/json") {
		return env, fmt.Errorf("响应不是 JSON（Content-Type=%q）：%s", contentType, snippet(body))
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return env, fmt.Errorf("JSON 解码失败：%v；响应片段：%s", err, snippet(body))
	}
	return env, nil
}

func hasCookie(cookies []*http.Cookie, name string) bool {
	for _, c := range cookies {
		if c.Name == name && c.Value != "" {
			return true
		}
	}
	return false
}

func snippet(body []byte) string {
	s := strings.Join(strings.Fields(string(body)), " ")
	runes := []rune(s)
	if len(runes) > 200 {
		s = string(runes[:200]) + "..."
	}
	if s == "" {
		return "<空响应>"
	}
	return s
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "失败: "+format+"\n", args...)
	os.Exit(1)
}
