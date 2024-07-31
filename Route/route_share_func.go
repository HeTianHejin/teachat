package route

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	mrand "math/rand"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
	"text/template"
	"time"
	"unicode/utf8"
)

/*
   存放各个路由文件共享的一些方法
*/

// 处理头像图片上传方法，图片要求为jpeg格式，size<30kb,宽高尺寸是64，32像素之间
func ProcessUploadAvatar(w http.ResponseWriter, r *http.Request, uuid string) error {
	// 从请求中解包出单个上传文件
	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		Report(w, r, "获取头像文件失败，请稍后再试。")
		return err
	}
	// 确保文件在函数执行完毕后关闭
	defer file.Close()

	// 获取文件大小，注意：客户端提供的文件大小可能不准确
	size := fileHeader.Size
	if size > 30*1024 {
		Report(w, r, "文件大小超过30kb,茶博士接不住。")
		return errors.New("the file size over 30kb")
	}
	// 实际读取文件大小进行校验，以防止客户端伪造
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		Report(w, r, "读取头像文件失败，请稍后再试。")
		return err
	}
	if len(fileBytes) > 30*1024 {
		Report(w, r, "文件大小超过30kb,茶博士接不住。")
		return errors.New("the file size over 30kb")
	}

	// 获取文件名和检查文件后缀
	filename := fileHeader.Filename
	ext := strings.ToLower(path.Ext(filename))
	if ext != ".jpeg" {
		Report(w, r, "注意头像图片文件类型, 目前仅限jpeg格式图片上传。")
		return errors.New("the file type is not jpeg")
	}

	// 获取文件类型，注意：客户端提供的文件类型可能不准确
	fileType := http.DetectContentType(fileBytes)
	if fileType != "image/jpeg" {
		Report(w, r, "注意图片文件类型,目前仅限jpeg格式。")
		return errors.New("the file type is not jpeg")
	}

	// 检测图片尺寸宽高和图像格式,判断是否合适
	width, height, err := GetWidthHeightForJpeg(fileBytes)
	if err != nil {
		Report(w, r, "注意图片文件格式, 目前仅限jpeg格式。")
		return err
	}
	if width < 32 || width > 64 || height < 32 || height > 64 {
		Report(w, r, "注意图片尺寸, 宽高需要在32-64像素之间。")
		return errors.New("the image size is not between 32 and 64")
	}

	// 创建新文件，无需切换目录，直接使用完整路径，减少安全风险
	newFilePath := data.ImageDir + uuid + data.ImageExt
	newFile, err := os.Create(newFilePath)
	if err != nil {
		util.Danger(err, "创建头像文件名失败")
		Report(w, r, "创建头像文件失败，请稍后再试。")
		return err
	}
	// 确保文件在函数执行完毕后关闭
	defer newFile.Close()

	// 通过缓存方法写入硬盘
	buff := bufio.NewWriter(newFile)
	buff.Write(fileBytes)
	err = buff.Flush()
	if err != nil {
		util.Warning(err, "fail to write avatar image")
		Report(w, r, "您好，茶博士居然说没有墨水了，写入头像文件不成功，请稍后再试。")
		return err
	}

	// _, err = newFile.Write(fileBytes)
	return nil
}

// 茶博士向茶客的报告信息，通常是一些错误提示
func Report(w http.ResponseWriter, r *http.Request, msg string) {
	var userBPD data.UserBiographyPageData
	userBPD.Message = msg
	s, err := Session(r)
	if err != nil {
		userBPD.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		GenerateHTML(w, &userBPD, "layout", "navbar.public", "feedback")
		return
	}
	userBPD.SessUser, _ = s.User()
	GenerateHTML(w, &userBPD, "layout", "navbar.private", "feedback")
}

// Checks if the user is logged in and has a Session, if not err is not nil
func Session(r *http.Request) (sess data.Session, err error) {
	cookie, err := r.Cookie("_cookie")
	if err == nil {
		sess = data.Session{Uuid: cookie.Value}
		if ok, _ := sess.Check(); !ok {
			err = errors.New("invalid session")
		}
	}
	return
}

// parse HTML templates
// pass in a list of file names, and get a template
func ParseTemplateFiles(filenames ...string) (t *template.Template) {
	var files []string
	t = template.New("layout")
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("templates/%s.go.html", file))
	}
	t = template.Must(t.ParseFiles(files...))
	return
}

// 处理器把页面模版和需求数据揉合后，由这个方法，将填写好的页面“制作“成HTML格式，调用http响应方法，发送给浏览器端客户
func GenerateHTML(w http.ResponseWriter, data interface{}, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("templates/%s.go.html", file))
	}

	templates := template.Must(template.ParseFiles(files...))
	templates.ExecuteTemplate(w, "layout", data)
}

// 验证邮箱格式是否正确，正确返回true，错误返回false。
func VerifyEmailFormat(email string) bool {
	pattern := `^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// 验证team_id_list:"2,19,87..."字符串格式是否正确，正确返回true，错误返回false。
func VerifyTeamIdListFormat(teamIdList string) bool {
	if teamIdList == "" {
		return false
	}
	pattern := `^[0-9]+(,[0-9]+)*$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(teamIdList)
}

// 输入两个统计数（辩论的正方累积得分数，辩论总得分数）（整数），计算前者与后者比值，结果浮点数向上四舍五入取整,
// 返回百分数的分子整数
func ProgressRound(numerator, denominator int) int {
	if denominator == 0 {
		// 分母为0时，视作未有记录，即未进行表决状态，返回100
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
	return int(math.Floor(ratio + 0.5))
}

/*
* 入参： JPG 图片文件的二进制数据
* 出参：JPG 图片的宽和高
* Author Mr.YF https://www.cnblogs.com/voipman
 */
func GetWidthHeightForJpeg(imgBytes []byte) (int, int, error) {
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

// RandomInt() 生成count个随机且不重复的整数，范围在[start, end)之间，按升序排列
func RandomInt(start, end, count int) []int {
	// 检查参数有效性
	if count <= 0 || start >= end {
		return nil
	}

	// 初始化包含所有可能随机数的切片
	nums := make([]int, end-start)
	for i := range nums {
		nums[i] = start + i
	}

	// 使用Fisher-Yates洗牌算法打乱切片顺序
	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	for i := len(nums) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		nums[i], nums[j] = nums[j], nums[i]
	}

	// 切片只需要前count个元素
	nums = nums[:count]

	// 对切片进行排序
	sort.Ints(nums)

	return nums
}

// 生成“火星文”替换下标队列
func StaRepIntList(str_len, ratio int) (numList []int, err error) {

	half := str_len / 2
	substandard := str_len * ratio / 100
	// 存放结果的slice
	numList = make([]int, str_len)

	// 随机生成替换下标
	switch {
	case ratio < 50:
		numList = []int{}
		return numList, errors.New("ratio must be not less than 50")
	case ratio == 50:
		numList = RandomInt(0, str_len, half)
	case ratio > 50:
		numList = RandomInt(0, str_len, substandard)
	}

	return
}

// 计算中文字符串长度
func CnStrLen(str string) int {
	return utf8.RuneCountInString(str)
}

// 对未经盲评的草稿进行“火星文”遮盖隐秘处理，即用星号替换50%或者指定更高比例文字
func MarsString(str string, ratio int) string {
	len := CnStrLen(str)
	// 获取替换字符的下标队列
	nlist, err := StaRepIntList(len, ratio)
	if err != nil {
		return str
	}
	// 把字符串转换为[]rune
	rstr := []rune(str)
	// 遍历替换字符的下标队列

	for _, n := range nlist {
		// 替换下标指定的字符为星号
		rstr[n] = '*'
	}

	// 将[]rune转换为字符串

	return string(rstr)
}

// 入参string，截取前面一段指定长度文字，返回string
// 注意，输入负数=最大值
// 参考https://blog.thinkeridea.com/201910/go/efficient_string_truncation.html
func Substr(s string, length int) string {
	//这是根据range的特性加的，如果不加，截取不到最后一个字（end+1=意外，因为1中文=3字节！）
	//str += "."
	var n, i int
	for i = range s {
		if n == length {
			break
		}
		n++
	}

	return s[:i]
}

// 截取一段指定开始和结束位置的文字，用range迭代方法。入参string，返回string“...”
// 注意，输入负数=最大值
func Substr2(str string, start, end int) string {

	//str += "." //这是根据range的特性加的，如果不加，截取不到最后一个字（end+1=意外，因为1中文=3字节！）

	var cnt, s, e int
	for s = range str {
		if cnt == start {
			break
		}
		cnt++
	}
	cnt = 0
	for e = range str {
		if cnt == end {
			break
		}
		cnt++
	}
	return str[s:e]
}
