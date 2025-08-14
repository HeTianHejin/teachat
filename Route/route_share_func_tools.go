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
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
	"text/template"
	"unicode/utf8"
)

// 处理器把页面模版和需求数据揉合后，由这个方法，将填写好的页面“制作“成HTML格式，调用http响应方法，发送给浏览器端客户
func renderHTML(w http.ResponseWriter, page_data any, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("templates/%s.go.html", file))
	}

	// 创建模板并添加自定义函数
	tmpl := template.New("layout").Funcs(template.FuncMap{
		"GetEnvironmentLevelDescription": data.GetEnvironmentLevelDescription,
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
	if err = templates.ExecuteTemplate(w, "layout", page_data); err != nil {
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

// // 验证提交的string是否 1 正整数？
// func verifyPositiveIntegerFormat(str string) bool {
// 	if str == "" {
// 		return false
// 	}
// 	pattern := `^[1-9]\d*$`
// 	reg := regexp.MustCompile(pattern)
// 	return reg.MatchString(str)
// }

// 验证team_id_slice，必需是正整数的逗号分隔的"2,19,87..."字符串格式是否正确，正确返回true，错误返回false。
func verifyIdSliceFormat(team_id_slice string) bool {
	if team_id_slice == "" {
		return false
	}
	// 使用双引号显式声明正则表达式，避免隐藏字符
	pattern := "^[0-9]+(,[0-9]+)*$"
	reg, err := regexp.Compile(pattern)
	if err != nil {
		// 实际生产环境应记录该错误
		return false
	}
	return reg.MatchString(team_id_slice)
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

	// if numerator > denominator {
	// 	// 分子大于分母时，表示统计数据输入错误，返回一个中间值
	// 	return 50
	// } else if ratio < 0 {
	// 	// 分子小于分母且比例为负数，表示统计数据输入错误，返回一个中间值
	// 	return 50
	// } else if ratio < 1 {
	// 	// 比例小于1时，返回最低限度值1
	// 	return 1
	// }

	// 其他情况，使用math.Floor确保向下取整，然后四舍五入
	//return int(math.Floor(ratio + 0.5))
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

// randomInt() 生成count个随机且不重复的整数，范围在[start, end)之间，按升序排列
// func randomInt(start, end, count int) []int {
// 	// 检查参数有效性
// 	if count <= 0 || start >= end {
// 		return nil
// 	}

// 	// 初始化包含所有可能随机数的切片
// 	nums := make([]int, end-start)
// 	for i := range nums {
// 		nums[i] = start + i
// 	}

// 	// 使用Fisher-Yates洗牌算法打乱切片顺序
// 	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
// 	for i := len(nums) - 1; i > 0; i-- {
// 		j := r.Intn(i + 1)
// 		nums[i], nums[j] = nums[j], nums[i]
// 	}

// 	// 切片只需要前count个元素
// 	nums = nums[:count]

// 	// 对切片进行排序
// 	sort.Ints(nums)

// 	return nums
// }

// // 生成“火星文”替换下标队列
// func staRepIntSlice(str_len, ratio int) (numSlice []int, err error) {

// 	half := str_len / 2
// 	substandard := str_len * ratio / 100
// 	// 存放结果的slice
// 	numSlice = make([]int, str_len)

// 	// 随机生成替换下标
// 	switch {
// 	case ratio < 50:
// 		numSlice = []int{}
// 		return numSlice, errors.New("ratio must be not less than 50")
// 	case ratio == 50:
// 		numSlice = randomInt(0, str_len, half)
// 	case ratio > 50:
// 		numSlice = randomInt(0, str_len, substandard)
// 	}

// 	return
// }

// 1. 校验茶议已有内容是否不超限,false == 超限
func submitAdditionalContent(w http.ResponseWriter, r *http.Request, body, additional string) bool {
	if cnStrLen(body) >= int(util.Config.ThreadMaxWord) {
		report(w, r, "已有内容已超过最大字数限制，无法补充。")
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
		report(w, r, errMsg)
		return false
	}
	// 3. 校验补充内容是否包含敏感词

	return true
}

// 计算中文字符串长度
func cnStrLen(str string) int {
	return utf8.RuneCountInString(str)
}

// // 对未经蒙评的草稿进行“火星文”遮盖隐秘处理，即用星号替换50%或者指定更高比例文字
// func marsString(str string, ratio int) string {
// 	len := cnStrLen(str)
// 	// 获取替换字符的下标队列
// 	nslice, err := staRepIntSlice(len, ratio)
// 	if err != nil {
// 		return str
// 	}
// 	// 把字符串转换为[]rune
// 	rstr := []rune(str)
// 	// 遍历替换字符的下标队列

// 	for _, n := range nslice {
// 		// 替换下标指定的字符为星号
// 		rstr[n] = '*'
// 	}

// 	// 将[]rune转换为字符串

// 	return string(rstr)
// }

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

// 截取一段指定开始和结束位置的文字，用range迭代方法。入参string，返回string“...”
// 注意，输入负数=最大值
// func subStr2(str string, start, end int) string {

// 	//str += "." //这是根据range的特性加的，如果不加，截取不到最后一个字（end+1=意外，因为1中文=3字节！）

// 	var cnt, s, e int
// 	for s = range str {
// 		if cnt == start {
// 			break
// 		}
// 		cnt++
// 	}
// 	cnt = 0
// 	for e = range str {
// 		if cnt == end {
// 			break
// 		}
// 		cnt++
// 	}
// 	return str[s:e]
// }

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
func validateCnStrLen(value string, min int, max int, fieldName string, w http.ResponseWriter, r *http.Request) bool {
	if cnStrLen(value) < min {
		report(w, r, fmt.Sprintf("你好，茶博士竟然说该茶议%s为空或太短，请确认后再试一次。", fieldName))
		return false
	}
	if cnStrLen(value) > max {
		report(w, r, fmt.Sprintf("你好，茶博士竟然说该茶议%s过长，请确认后再试一次。", fieldName))
		return false
	}
	return true
}

// 处理头像图片上传方法，图片要求为jpeg格式，size<30kb,宽高尺寸是64，32像素之间
func processUploadAvatar(w http.ResponseWriter, r *http.Request, uuid string) error {
	// 从请求中解包出单个上传文件
	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		report(w, r, "获取头像文件失败，请稍后再试。")
		return err
	}
	// 确保文件在函数执行完毕后关闭
	defer file.Close()

	// 获取文件大小，注意：客户端提供的文件大小可能不准确
	size := fileHeader.Size
	if size > 30*1024 {
		report(w, r, "文件大小超过30kb,茶博士接不住。")
		return errors.New("the file size over 30kb")
	}
	// 实际读取文件大小进行校验，以防止客户端伪造
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		report(w, r, "读取头像文件失败，请稍后再试。")
		return err
	}
	if len(fileBytes) > 30*1024 {
		report(w, r, "文件大小超过30kb,茶博士接不住。")
		return errors.New("the file size over 30kb")
	}

	// 获取文件名和检查文件后缀
	filename := fileHeader.Filename
	ext := strings.ToLower(path.Ext(filename))
	if ext != ".jpeg" && ext != ".jpg" {
		report(w, r, "注意头像图片文件类型, 目前仅限jpeg格式图片上传。")
		return errors.New("the file type is not jpeg")
	}

	// 获取文件类型，注意：客户端提供的文件类型可能不准确
	fileType := http.DetectContentType(fileBytes)
	if fileType != "image/jpeg" {
		report(w, r, "注意图片文件类型,目前仅限jpeg格式。")
		return errors.New("the file type is not jpeg")
	}

	// 检测图片尺寸宽高和图像格式,判断是否合适
	width, height, err := getWidthHeightForJpeg(fileBytes)
	if err != nil {
		report(w, r, "注意图片文件格式, 目前仅限jpeg格式。")
		return err
	}
	if width < 32 || width > 64 || height < 32 || height > 64 {
		report(w, r, "注意图片尺寸, 宽高需要在32-64像素之间。")
		return errors.New("the image size is not between 32 and 64")
	}

	// 创建新文件，无需切换目录，直接使用完整路径，减少安全风险
	newFilePath := util.Config.ImageDir + uuid + util.Config.ImageExt
	newFile, err := os.Create(newFilePath)
	if err != nil {
		util.Debug("创建头像文件名失败", err)
		report(w, r, "创建头像文件失败，请稍后再试。")
		return err
	}
	// 确保文件在函数执行完毕后关闭
	defer newFile.Close()

	// 通过缓存方法写入硬盘
	buff := bufio.NewWriter(newFile)
	if _, err = buff.Write(fileBytes); err != nil {
		util.Debug("fail to write avatar image", err)
		report(w, r, "你好，茶博士居然说没有墨水了， 未能写完头像文件，请稍后再试。")
		return err
	}
	if err = buff.Flush(); err != nil {
		util.Debug("fail to write avatar image", err)
		report(w, r, "你好，茶博士居然说没有墨水了，写入头像文件不成功，请稍后再试。")
		return err
	}

	return nil
}

// 茶博士——古时专指陆羽。陆羽著《茶经》，唐德宗李适曾当面称陆羽为“茶博士”。
// 茶博士-teaOffice，是古代中华传统文化对茶馆工作人员的昵称，如：富家宴会，犹有专供茶事之人，谓之茶博士。——唐代《西湖志馀》
// 现在多指精通茶艺的师傅，尤其是四川的长嘴壶茶艺，茶博士个个都是身怀绝技的“高手”。
// 茶博士向茶客报告信息的方法，包括但不限于意外事件和通知、感谢等等提示。
func report(w http.ResponseWriter, r *http.Request, msg ...any) {
	var userBPD data.UserBean
	var b strings.Builder
	for i, arg := range msg {
		if i > 0 {
			b.WriteByte(' ') // 参数间添加空格
		}
		fmt.Fprint(&b, arg)
	}
	userBPD.Message = b.String()

	s, err := session(r)
	if err != nil {
		userBPD.SessUser = data.User{
			Id:   data.UserId_None,
			Name: "游客",
		}
		renderHTML(w, &userBPD, "layout", "navbar.public", "feedback")
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", s.Email, err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	userBPD.SessUser = s_u

	renderHTML(w, &userBPD, "layout", "navbar.private", "feedback")
}

// Checks if the user is logged in and has a session, if not err is not nil
func session(r *http.Request) (data.Session, error) {
	cookie, err := r.Cookie("_cookie")
	if err != nil {
		return data.Session{}, fmt.Errorf("cookie not found: %w", err)
	}

	sess := data.Session{Uuid: cookie.Value}
	ok, checkErr := sess.Check()
	if checkErr != nil {
		return data.Session{}, fmt.Errorf("session check failed: %w", checkErr)
	}
	if !ok {
		return data.Session{}, errors.New("invalid or expired session")
	}

	return sess, nil
}

// parse HTML templates
// pass in a slice of file names, and get a template
// func parseTemplateFiles(filenames ...string) *template.Template {
// 	var files []string
// 	t := template.New("layout")
// 	for _, file := range filenames {
// 		// 使用 filepath.Join 安全拼接路径,unix+windows
// 		filePath := filepath.Join(util.Config.TemplateExt, file+util.Config.TemplateExt)
// 		files = append(files, filePath)
// 	}
// 	t = template.Must(t.ParseFiles(files...))
// 	return t
// }

// 记录用户最后的查询路径和参数
// func recordLastQueryPath(sess_user_id int, path, raw_query string) (err error) {
// 	lq := data.LastQuery{
// 		UserId: sess_user_id,
// 		Path:   path,
// 		Query:  raw_query,
// 	}
// 	if err = lq.Create(); err != nil {
// 		return err
// 	}
// 	return
// }

func moveDefaultTeamToFront(teamSlice []data.TeamBean, defaultTeamID int) ([]data.TeamBean, error) {
	newSlice := make([]data.TeamBean, 0, len(teamSlice))
	var defaultTeam *data.TeamBean

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
	return append([]data.TeamBean{*defaultTeam}, newSlice...), nil
}

// validateTeamAndFamilyParams 验证团队和家庭ID参数的合法性
// 返回: (是否有效, 错误) ---deepseek协助优化
func validateTeamAndFamilyParams(is_private bool, team_id int, family_id int, currentUserID int, w http.ResponseWriter, r *http.Request) (bool, error) {

	// 基本参数检查（这些检查不涉及数据库操作）
	//非法id组合
	if family_id == data.FamilyIdUnknown && team_id == data.TeamIdNone {
		report(w, r, "你好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
		return false, nil
	}
	if team_id == data.TeamIdNone || team_id == data.TeamIdSpaceshipCrew {
		report(w, r, "指定的团队编号是保留编号，不能使用。")
		return false, nil
	}

	if team_id < 0 || family_id < 0 {
		report(w, r, "团队ID不合法。")
		return false, nil
	}

	// 茶语管理权限归属是.IsPrivate 属性声明的，
	//所以可以同时指定两者,符合任何人必然有某个家庭，但不一定有事业团队背景的实际情况
	if is_private {
		// 管理权属于家庭
		if family_id == data.FamilyIdUnknown {
			report(w, r, "你好，四海为家者今天不能发布新茶语，请明天再试。")
			return false, fmt.Errorf("unknown family #%d cannot do this", family_id)
		}
		family := data.Family{Id: family_id}
		// if err := family.Get(); err != nil {
		// 	return false, err // 数据库错误，返回error
		// }
		isOnlyOne, err := family.IsOnlyOneMember()
		if err != nil {
			util.Debug("Cannot count family member given id", family.Id, err)
			report(w, r, "你好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
			return false, err
		}
		if isOnlyOne {
			report(w, r, "根据“慎独”约定，单独成员家庭目前暂时不能品茶噢，请向船长抗议。")
			return false, fmt.Errorf("onlyone member family #%d cannot do this", family_id)
		}

		is_member, err := family.IsMember(currentUserID)
		if err != nil {
			return false, err // 数据库错误，返回error
		}
		if !is_member {
			report(w, r, "你好，家庭成员资格检查失败，请确认后再试。")
			return false, fmt.Errorf(" team %d id_member check failed", team_id)
		}
	} else {
		// 管理权属于团队
		if team_id == data.TeamIdNone || team_id == data.TeamIdSpaceshipCrew {
			report(w, r, "你好，特殊团队今天还不能创建茶话会，请稍后再试。")
			return false, fmt.Errorf("special team #%d cannot do this", team_id)
		}
		//声明是四海为家【与家庭背景（责任）无关】
		if team_id == data.TeamIdFreelancer {
			//既隐藏家庭背景，也不声明团队的“独狼”
			// 违背了“慎独”原则
			report(w, r, "你好，茶博士查阅了天书黄页，四海为家的自由人，今天不适宜发表茶话。")
			return false, nil
		}

		team := data.Team{Id: team_id}
		// if err := team.Get(); err != nil {
		// 	return false, err // 数据库错误，返回error
		// }
		is_member, err := team.IsMember(currentUserID)
		if err != nil {
			return false, err // 数据库错误，返回error
		}
		if !is_member {
			report(w, r, "你好，眼前无路想回头，您是什么团成员？什么茶话会？请稍后再试。")
			return false, nil
		}

	}

	return true, nil
}
