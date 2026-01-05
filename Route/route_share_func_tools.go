package route

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	dao "teachat/DAO"
	util "teachat/Util"
	"text/template"
	"unicode/utf8"
)

// 处理器把页面模版和需求数据揉合后，由这个方法，将填写好的页面“制作“成HTML格式，调用http响应方法，发送给浏览器端客户
func generateHTML(w http.ResponseWriter, template_data any, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("templates/%s.go.html", file))
	}

	// 创建模板并添加自定义函数
	tmpl := template.New("layout").Funcs(template.FuncMap{
		"GetEnvironmentLevelDescription": dao.GetEnvironmentLevelDescription,
		"GetStarIcons": func(level int) string {
			if level < 1 || level > 5 {
				return ""
			}
			stars := ""
			for i := 0; i < level; i++ {
				stars += `<span class="glyphicon glyphicon-star" style="color: #f39c12;"></span>`
			}
			return stars
		},
		"RiskSeverityLevelString": dao.RiskSeverityLevelString,
		"AvailabilityString":      dao.GoodsAvailabilityString,
		"mul": func(a, b int) int {
			return a * b
		},
		"sub": func(a, b float64) float64 {
			return a - b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"max": func(a, b int) int {
			if a > b {
				return a
			}
			return b
		},
		"min": func(a, b int) int {
			if a < b {
				return a
			}
			return b
		},
		"seq": func(start, end int) []int {
			var items []int
			for i := start; i <= end; i++ {
				items = append(items, i)
			}
			return items
		},
		"iterate": func(count int) []int {
			var items []int
			for i := 0; i < count; i++ {
				items = append(items, i)
			}
			return items
		},
		"split": func(s, sep string) []string {
			return strings.Split(s, sep)
		},
		"trim": func(s string) string {
			return strings.TrimSpace(s)
		},
		"FormatFloat": util.FormatFloat,
	})

	// 手动解析模板并处理错误
	templates, err := tmpl.ParseFiles(files...)
	if err != nil {
		// 添加详细的错误日志和HTTP错误响应
		util.PrintStdout("模板解析错误: ", err)
		http.Error(w, "*** 茶博士: 茶壶不见了，无法烧水冲茶，陛下稍安勿躁 ***", http.StatusInternalServerError)
		return
	}

	// 安全增强：设置内容类型为HTML并添加XSS防护头
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// 执行模板渲染
	if err = templates.ExecuteTemplate(w, "layout", template_data); err != nil {
		// 添加详细的错误日志
		util.PrintStdout("模板渲染错误: ", err)
		// 避免在错误响应中泄露敏感信息
		http.Error(w, "*** 茶博士: 茶壶不见了，无法烧水冲茶，陛下稍安勿躁 ***", http.StatusInternalServerError)
	}
}

// 验证邮箱地址，格式是否正确，正确返回true，错误返回false。
func isEmail(email string) bool {
	pattern := `^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// 验证用户名，只允许字母、数字、下划线或中文字符，正确返回true，错误返回false。
func isValidUserName(name string) bool {
	pattern := `^[a-zA-Z0-9_\p{Han}]+$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(name)
}

// 验证id_slice，必需是非零正整数而且不重复的逗号分隔的"2,19,87..."字符串格式，是否正确，正确返回true，错误返回false。
// 预编译正则表达式提高性能（deepSeek.com）
var idSliceRegex = regexp.MustCompile(`^[1-9][0-9]*(,[1-9][0-9]*)*$`)

// verifyIdSliceFormat 验证ID切片格式，必须是正整数（不允许0）且不重复的逗号分隔字符串
// 格式示例: "2,19,87", "1"
// 正确返回true，错误返回false
func verifyIdSliceFormat(idSlice string) bool {
	idSlice = strings.TrimSpace(idSlice)
	if idSlice == "" {
		return false
	}

	if !idSliceRegex.MatchString(idSlice) {
		return false
	}

	ids := strings.Split(idSlice, ",")
	seen := make(map[string]bool)

	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr) // 确保去除可能的前后空格
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			return false
		}

		if seen[idStr] {
			return false
		}
		seen[idStr] = true
	}

	return true
}

// 修改verifyIdSliceFormat函数，格式正确返回[]int，错误返回error
func parseIdSlice(idSlice string) ([]int, error) {
	idSlice = strings.TrimSpace(idSlice)
	if idSlice == "" {
		return []int{}, nil
	}

	if !verifyIdSliceFormat(idSlice) {
		return nil, fmt.Errorf("invalid hazard ID format")
	}

	var ids []int
	for _, idStr := range strings.Split(idSlice, ",") {
		idStr = strings.TrimSpace(idStr)
		id, _ := strconv.Atoi(idStr)
		ids = append(ids, id)
	}

	return ids, nil
}

// 比较两个ID切片内容是否一样
func compareIdsSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	countMap := make(map[int]int)
	for _, id := range a {
		countMap[id]++
	}
	for _, id := range b {
		if countMap[id] == 0 {
			return false
		}
		countMap[id]--
	}
	return true
}

// 输入两个统计数（辩论的正方累积得分数，辩论总得分数）（整数），计算前者与后者比值，结果浮点数向上四舍五入取整,
// 返回百分数的分子整数
func progressRound(numerator, denominator int) int {
	if denominator == 0 {
		// 分母为0时，视作未有记录，即未进行表决状态，返回默认值100
		return 100
	}
	if numerator == denominator {
		// 分子等于分母时，表示100%正方
		return 100
	}
	ratio := float64(numerator) / float64(denominator) * 100
	return int(math.Round(ratio))
}

/*
* 入参： JPG 图片文件的二进制数据
* 出参：JPG 图片的宽和高
* Author Mr.YF https://www.cnblogs.com/voipman
 */
func getWidthHeightForJpeg(imgBytes []byte) (int, int, error) {
	var offset int
	imgByteLen := len(imgBytes)
	for i := 0; i < imgByteLen-1; i++ {
		if imgBytes[i] != 0xff {
			continue
		}
		if imgBytes[i+1] == 0xC0 || imgBytes[i+1] == 0xC1 || imgBytes[i+1] == 0xC2 {
			offset = i
			break
		}
	}
	offset += 5
	if offset >= imgByteLen {
		return 0, 0, errors.New("unknown format")
	}
	height := int(imgBytes[offset])<<8 + int(imgBytes[offset+1])
	width := int(imgBytes[offset+2])<<8 + int(imgBytes[offset+3])
	return width, height, nil
}

// 1. 校验茶议已有内容是否不超限,false == 超限
func submitAdditionalContent(w http.ResponseWriter, s_u dao.User, body, additional string) bool {
	if cnStrLen(body) >= int(util.Config.ThreadMaxWord) {
		report(w, s_u, "已有内容已超过最大字数限制，无法补充。")
		return false
	}

	// 2. 校验补充内容字数
	min := int(util.Config.ThreadMinWord)
	max := int(util.Config.ThreadMaxWord) - cnStrLen(body)
	current := cnStrLen(additional)

	if current < min || current > max {
		errMsg := fmt.Sprintf(
			"茶博士提示：补充内容需满足：%d ≤ 字数 ≤ %d（当前：%d）。",
			min, max, current,
		)
		report(w, s_u, errMsg)
		return false
	}
	// 3. 校验补充内容是否包含敏感词

	return true
}

// 计算中文字符串长度
func cnStrLen(str string) int {
	return utf8.RuneCountInString(str)
}

// 入参string，截取前面一段指定长度文字，返回string，作为预览文字
// CodeBuddy修改
func subStr(s string, length int) string {
	if length <= 0 {
		return ""
	}
	var count int //统计字符数（而非字节数）
	end := 0      //记录最后一个字符的起始字节位置
	for i := range s {
		if count == length {
			break
		}
		count++
		end = i
	}
	if count < length {
		return s
	}
	_, size := utf8.DecodeRuneInString(s[end:])
	return s[:end+size]
}

// sanitizeRedirectPath 只允许站内路径（如 /v1/home），禁止外部域名
// --- DeeSeek
func sanitizeRedirectPath(inputPath string) string {
	if inputPath == "" {
		return "/v1/" // 默认路径
	}

	// 检查是否以 "/" 开头（相对路径）
	if len(inputPath) > 0 && inputPath[0] == '/' {
		// 可选：进一步校验路径格式（避免路径遍历攻击，如 /../）
		cleanedPath := path.Clean(inputPath)
		if !strings.HasPrefix(cleanedPath, "/v1/") {
			return "/v1/" // 强制限制到特定前缀
		}
		return cleanedPath
	}

	// 非相对路径（如http://）则返回默认路径
	return "/v1/"
}

// Helper function for validating string length
func validateCnStrLen(value string, min int, max int, fieldName string, w http.ResponseWriter, s_u dao.User) bool {
	if cnStrLen(value) < min {
		report(w, s_u, fmt.Sprintf("你好，茶博士竟然说该茶议%s为空或太短，请确认后再试一次。", fieldName))
		return false
	}
	if cnStrLen(value) > max {
		report(w, s_u, fmt.Sprintf("你好，茶博士竟然说该茶议%s过长，请确认后再试一次。", fieldName))
		return false
	}
	return true
}

// 处理头像图片上传方法，图片要求为jpeg格式，size<30kb,宽高尺寸是64，32像素之间
func processUploadAvatar(w http.ResponseWriter, r *http.Request, s_u dao.User, uuid string) error {
	// 从请求中解包出单个上传文件
	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		report(w, s_u, "获取头像文件失败，请稍后再试。")
		return err
	}
	// 确保文件在函数执行完毕后关闭
	defer file.Close()

	// 获取文件大小，注意：客户端提供的文件大小可能不准确
	size := fileHeader.Size
	if size > 30*1024 {
		report(w, s_u, "文件大小超过30kb,茶博士接不住。")
		return errors.New("the file size over 30kb")
	}
	// 实际读取文件大小进行校验，以防止客户端伪造
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		report(w, s_u, "读取头像文件失败，请稍后再试。")
		return err
	}
	if len(fileBytes) > 30*1024 {
		report(w, s_u, "文件大小超过30kb,茶博士接不住。")
		return errors.New("the file size over 30kb")
	}

	// 获取文件名和检查文件后缀
	filename := fileHeader.Filename
	ext := strings.ToLower(path.Ext(filename))
	if ext != ".jpeg" && ext != ".jpg" {
		report(w, s_u, "注意头像图片文件类型, 目前仅限jpeg格式图片上传。")
		return errors.New("the file type is not jpeg")
	}

	// 获取文件类型，注意：客户端提供的文件类型可能不准确
	fileType := http.DetectContentType(fileBytes)
	if fileType != "image/jpeg" {
		report(w, s_u, "注意图片文件类型,目前仅限jpeg格式。")
		return errors.New("the file type is not jpeg")
	}

	// 检测图片尺寸宽高和图像格式,判断是否合适
	width, height, err := getWidthHeightForJpeg(fileBytes)
	if err != nil {
		report(w, s_u, "注意图片文件格式, 目前仅限jpeg格式。")
		return err
	}
	if width < 32 || width > 64 || height < 32 || height > 64 {
		report(w, s_u, "注意图片尺寸, 宽高需要在32-64像素之间。")
		return errors.New("the image size is not between 32 and 64")
	}

	// 创建新文件，无需切换目录，直接使用完整路径，减少安全风险
	newFilePath := util.Config.ImageDir + uuid + util.Config.ImageExt
	newFile, err := os.Create(newFilePath)
	if err != nil {
		util.Debug("创建头像文件名失败", err)
		report(w, s_u, "创建头像文件失败，请稍后再试。")
		return err
	}
	// 确保文件在函数执行完毕后关闭
	defer newFile.Close()

	// 通过缓存方法写入硬盘
	buff := bufio.NewWriter(newFile)
	if _, err = buff.Write(fileBytes); err != nil {
		util.Debug("fail to write avatar image", err)
		report(w, s_u, "你好，茶博士居然说没有墨水了， 未能写完头像文件，请稍后再试。")
		return err
	}
	if err = buff.Flush(); err != nil {
		util.Debug("fail to write avatar image", err)
		report(w, s_u, "你好，茶博士居然说没有墨水了，写入头像文件不成功，请稍后再试。")
		return err
	}

	return nil
}

// 茶博士向茶客回话的方法，包括但不限于意外事件和通知、感谢等等提示。
func report(w http.ResponseWriter, s_u dao.User, msg ...any) {
	type uM struct {
		SessUser dao.User
		Message  string
	}
	m := uM{}
	var b strings.Builder
	for i, arg := range msg {
		if i > 0 {
			b.WriteByte(' ') // 参数间添加空格
		}
		fmt.Fprint(&b, arg)
	}
	m.Message = b.String()

	if s_u.Id > 0 {
		m.SessUser = s_u
		generateHTML(w, &m, "layout", "navbar.private", "feedback")
	} else {
		m.SessUser = dao.UserUnknown
		generateHTML(w, &m, "layout", "navbar.public", "feedback")
	}

}

// Checks if the user is logged in and has a session, if not err is not nil
func session(r *http.Request) (dao.Session, error) {
	cookie, err := r.Cookie("_cookie")
	if err != nil {
		return dao.Session{}, fmt.Errorf("cookie not found: %w", err)
	}

	sess := dao.Session{Uuid: cookie.Value}
	ok, checkErr := sess.Check()
	if checkErr != nil {
		return dao.Session{}, fmt.Errorf("session check failed: %w", checkErr)
	}
	if !ok {
		return dao.Session{}, errors.New("invalid or expired session")
	}

	return sess, nil
}

func moveDefaultTeamToFront(teamSlice []dao.TeamBean, defaultTeamID int) ([]dao.TeamBean, error) {
	newSlice := make([]dao.TeamBean, 0, len(teamSlice))
	var defaultTeam *dao.TeamBean

	// 分离默认团队和其他团队
	for _, tb := range teamSlice {
		if tb.Team.Id == defaultTeamID {
			defaultTeam = &tb
			continue
		}
		newSlice = append(newSlice, tb)
	}

	if defaultTeam == nil {
		return nil, fmt.Errorf("默认团队 %d 未找到", defaultTeamID)
	}

	// 合并结果（默认团队在前）
	return append([]dao.TeamBean{*defaultTeam}, newSlice...), nil
}

// validateTeamAndFamilyParams 验证团队和家庭ID参数的合法性
// 返回: (是否有效, 错误) ---deepseek协助优化
func validateTeamAndFamilyParams(is_private bool, team_id int, family_id int, s_u dao.User, w http.ResponseWriter) (bool, error) {

	// 基本参数检查（这些检查不涉及数据库操作）
	//非法id组合
	if family_id == dao.FamilyIdUnknown && team_id == dao.TeamIdNone {
		report(w, s_u, "你好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
		return false, nil
	}
	if team_id == dao.TeamIdNone || team_id == dao.TeamIdSpaceshipCrew {
		report(w, s_u, "指定的团队编号是保留编号，不能使用。")
		return false, nil
	}

	if team_id < 0 || family_id < 0 {
		report(w, s_u, "团队ID不合法。")
		return false, nil
	}

	// 茶语管理权限归属是.IsPrivate 属性声明的，
	//所以可以同时指定两者,符合任何人必然有某个家庭，但不一定有事业团队背景的实际情况
	if is_private {
		// 管理权属于家庭
		if family_id == dao.FamilyIdUnknown {
			report(w, s_u, "你好，四海为家者今天不能发布新茶语，请明天再试。")
			return false, fmt.Errorf("unknown family #%d cannot do this", family_id)
		}
		family := dao.Family{Id: family_id}
		// if err := family.Get(); err != nil {
		// 	return false, err // 数据库错误，返回error
		// }
		isOnlyOne, err := family.IsOnlyOneMember()
		if err != nil {
			util.Debug("Cannot count family member given id", family.Id, err)
			report(w, s_u, "你好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
			return false, err
		}
		if isOnlyOne {
			report(w, s_u, "根据“慎独”约定，单独成员家庭目前暂时不能品茶噢，请向船长抗议。")
			return false, fmt.Errorf("onlyone member family #%d cannot do this", family_id)
		}

		is_member, err := family.IsMember(s_u.Id)
		if err != nil {
			return false, err // 数据库错误，返回error
		}
		if !is_member {
			report(w, s_u, "你好，家庭成员资格检查失败，请确认后再试。")
			return false, fmt.Errorf(" team %d id_member check failed", team_id)
		}
	} else {
		// 管理权属于团队
		if team_id == dao.TeamIdNone || team_id == dao.TeamIdSpaceshipCrew {
			report(w, s_u, "你好，特殊团队今天还不能创建茶话会，请稍后再试。")
			return false, fmt.Errorf("special team #%d cannot do this", team_id)
		}
		//声明是四海为家【与家庭背景（责任）无关】
		if team_id == dao.TeamIdFreelancer {
			//既隐藏家庭背景，也不声明团队的“独狼”
			// 违背了“慎独”原则
			report(w, s_u, "你好，茶博士查阅了天书黄页，四海为家的自由人，今天不适宜发表茶话。")
			return false, nil
		}

		team := dao.Team{Id: team_id}
		// if err := team.Get(); err != nil {
		// 	return false, err // 数据库错误，返回error
		// }
		is_member, err := team.IsMember(s_u.Id)
		if err != nil {
			return false, err // 数据库错误，返回error
		}
		if !is_member {
			report(w, s_u, "你好，眼前无路想回头，您是什么团成员？什么茶话会？请稍后再试。")
			return false, nil
		}

	}

	return true, nil
}
