package data

import (
	"context"
	"errors"
	util "teachat/Util"
	"time"
)

// 看看，睇睇，
// 例如一个“问诊设备故障”的交流会谈记录
type SeeSeek struct {
	Id          int
	Uuid        string
	Name        string
	Nickname    string
	Description string

	RequesterId       int // 需求方茶团代表人ID（直接负责人），注意需求方是家庭&团队组合
	RequesterFamilyId int // 需求方家庭id【如果需要声明与家庭无关，选ID=UnknownFamilyID（0）】
	RequesterTeamId   int // 需求方团队Id【如果需要声明与团队无关，选id=TeamIdFreelancer（2）】

	ProviderId       int // 服务方茶团代表人ID（直接负责人），注意服务方是家庭&团队组合
	ProviderFamilyId int // 服务方家庭id【如果需要声明与家庭无关，选ID=UnknownFamilyID（0）】
	ProviderTeamId   int //服务方团队Id【如果需要声明与团队无关，选id=TeamIdFreelancer（2）】

	VerifierId       int // 监护、见证，审核人,监护方代表id（直接负责人）
	VerifierFamilyId int // 监护方家庭id【如果需要声明与家庭无关，选ID=UnknownFamilyID（0）】
	VerifierTeamId   int // 监护方团队Id【如果需要声明与团队无关，选id=TeamIdFreelancer（2）】

	PlaceId           int // 事发地点ID
	EnvironmentId     int //看看环境条件Id
	RiskSeverityLevel int //看看风险等级

	Category  int //分类：0、公开，1、保密，仅当事家庭/团队可见内容
	Status    int //状态：0、未开始，1、进行中，2、暂停，3、已终止，4、已结束
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// SeeSeek.Create() // 创建一个SeeSeek
// 编写postgreSQL语句，插入新纪录，return （err error）
func (ss *SeeSeek) Create(ctx context.Context) (err error) {
	// 设置一个 5 秒的超时
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // 确保在函数退出时取消上下文

	statement := `INSERT INTO see_seek (uuid, name, nickname, description, requester_id, requester_family_id, requester_team_id, provider_id, provider_family_id, provider_team_id, verifier_id, verifier_family_id, verifier_team_id, place_id, environment_id, risk_severity_level, category, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19) RETURNING id, uuid`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, Random_UUID(), ss.Name, ss.Nickname, ss.Description, ss.RequesterId, ss.RequesterFamilyId, ss.RequesterTeamId, ss.ProviderId, ss.ProviderFamilyId, ss.ProviderTeamId, ss.VerifierId, ss.VerifierFamilyId, ss.VerifierTeamId, ss.PlaceId, ss.EnvironmentId, ss.RiskSeverityLevel, ss.Category, ss.Status, time.Now()).Scan(&ss.Id, &ss.Uuid)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			util.Debug("Query timed out")
		}
		return
	}
	return
}
func (ss *SeeSeek) Get(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // 确保在函数退出时取消上下文

	statement := `SELECT id, uuid, name, nickname, description, requester_id, requester_family_id, requester_team_id, provider_id, provider_family_id, provider_team_id, verifier_id, verifier_family_id, verifier_team_id, place_id, environment_id, risk_severity_level, category, status, created_at, updated_at FROM see_seek WHERE id = $1`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, ss.Id).Scan(&ss.Id, &ss.Uuid, &ss.Name, &ss.Nickname, &ss.Description, &ss.RequesterId, &ss.RequesterFamilyId, &ss.RequesterTeamId, &ss.ProviderId, &ss.ProviderFamilyId, &ss.ProviderTeamId, &ss.VerifierId, &ss.VerifierFamilyId, &ss.VerifierTeamId, &ss.PlaceId, &ss.EnvironmentId, &ss.RiskSeverityLevel, &ss.Category, &ss.Status, &ss.CreatedAt, &ss.UpdatedAt)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			util.Debug("Query timed out")
		}
		return
	}
	return
}
func (ss *SeeSeek) Update(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // 确保在函数退出时取消上下文

	statement := `UPDATE see_seek SET name = $1, nickname = $2, description = $3, requester_id = $4, requester_family_id = $5, requester_team_id = $6, provider_id = $7, provider_family_id = $8, provider_team_id = $9, verifier_id = $10, verifier_family_id = $11, verifier_team_id = $12, place_id = $13, environment_id = $14, risk_severity_level = $15, category = $16, status = $17, updated_at = $18 WHERE id = $19`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, ss.Name, ss.Nickname, ss.Description, ss.RequesterId, ss.RequesterFamilyId, ss.RequesterTeamId, ss.ProviderId, ss.ProviderFamilyId, ss.ProviderTeamId, ss.VerifierId, ss.VerifierFamilyId, ss.VerifierTeamId, ss.PlaceId, ss.EnvironmentId, ss.RiskSeverityLevel, ss.Category, ss.Status, time.Now(), ss.Id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			util.Debug("Query timed out")
		}
		return
	}
	return
}

// SeeSeek.CreateAtDate() string
func (ss *SeeSeek) CreatedDateTime() string {
	return ss.CreatedAt.Format(FMT_DATE_TIME_CN)
}

const (
	SeeSeekCategoryPublic = iota // 公开
	SeeSeekCategorySecret        // 保密
)

const (
	SeeSeekStatusNotStarted = iota // 未开始
	SeeSeekStatusInProgress        // 进行中
	SeeSeekStatusPaused            // 已暂停
	SeeSeekStatusAborted           // 已半途终止（异常）
	SeeSeekStatusCompleted         // 已完成（顺利结束）
)

func (see_seek *SeeSeek) StatusString() string {
	switch see_seek.Status {
	case SeeSeekStatusNotStarted:
		return "未开始"
	case SeeSeekStatusInProgress:
		return "进行中"
	case SeeSeekStatusPaused:
		return "已暂停"
	case SeeSeekStatusAborted:
		return "已半途终止"
	case SeeSeekStatusCompleted:
		return "已完成"
	default:
		return "未知状态"
	}
}

// “看看”作业自然（物理）环境条件
type SeeSeekEnvironment struct {
	Id            int
	Uuid          string
	SeeSeekId     int
	EnvironmentId int
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

// “看看”作业场所安全风险，风险因素
type SeeSeekRisk struct {
	Id        int
	Uuid      string
	SeeSeekId int
	RiskId    int
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 团队or家庭与特定团队举行一个结构化交流
type SeeSeekMaster struct {
	Id        int
	Uuid      string
	SeeSeekId int
	Classify  int

	RecorderUserId int
	UserId         int
	Status         int

	RequesterTitle   string
	RequesterContent string
	RequesterHistory string
	RequesterRemark  string //特殊情况表述

	MasterTitle   string
	MasterContent string
	MasterHistory string
	MasterRemark  string //特殊情况表述

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 望，观察
type SeeSeekMasterLook struct {
	Id              int
	Uuid            string
	SeeSeekMasterId int
	Classify        int
	Status          int

	RequestOutline     string //外形轮廓
	IsDeform           bool   //是否变形
	RequestSkin        string //表面皮肤
	IsGraze            bool   //是否破损
	RequestColor       string //颜色
	IsChange           bool   //是否变色
	RequestLookHistory string //过往历史

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 听，声音
type SeeSeekMasterListen struct {
	Id              int
	Uuid            string
	SeeSeekMasterId int
	Classify        int
	Status          int

	RequestSound        string //声音
	IsAbnormal          bool   //是否有异常声音
	RequestSoundHistory string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 嗅，气味
type SeeSeekMasterSmell struct {
	Id              int
	Uuid            string
	SeeSeekMasterId int
	Classify        int
	Status          int

	RequestOdour        string //气味
	IsFoulOdour         bool   //是否异味
	RequestOdourHistory string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 触摸，
type SeeSeekMasterTouch struct {
	Id              int
	Uuid            string
	SeeSeekMasterId int
	Classify        int

	Status int

	RequestTemperature  string //温度
	IsFever             bool   //是否异常发热
	RequestStretch      string //弹性
	IsStiff             bool   //是否僵硬
	RequestShake        string //震动
	IsShake             bool   //是否震动
	RequestTouchHistory string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 问答
type SeeSeekMasterAskAndAnswer struct {
	Id              int
	Uuid            string
	SeeSeekMasterId int
	Classify        int

	Status int

	RequestTitle   string
	RequestContent string
	RequestHistory string

	MasterTitle   string
	MasterContent string
	MasterHistory string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (see_seek_master_ask_and_answer *SeeSeekMasterAskAndAnswer) Create() (err error) {
	statement := "INSERT INTO see_seek_master_ask_and_answers (uuid, see_seek_master_id, classify, status, request_title, request_content, request_history, master_title, master_content, master_history, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), see_seek_master_ask_and_answer.SeeSeekMasterId, see_seek_master_ask_and_answer.Classify, see_seek_master_ask_and_answer.Status, see_seek_master_ask_and_answer.RequestTitle, see_seek_master_ask_and_answer.RequestContent, see_seek_master_ask_and_answer.RequestHistory, see_seek_master_ask_and_answer.MasterTitle, see_seek_master_ask_and_answer.MasterContent, see_seek_master_ask_and_answer.MasterHistory, time.Now()).Scan(&see_seek_master_ask_and_answer.Id, &see_seek_master_ask_and_answer.Uuid)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_ask_and_answer *SeeSeekMasterAskAndAnswer) Get() (err error) {
	statement := "SELECT id, uuid, see_seek_master_id, classify, status, request_title, request_content, request_history, master_title, master_content, master_history, created_at, updated_at FROM see_seek_master_ask_and_answers WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_master_ask_and_answer.Id).Scan(&see_seek_master_ask_and_answer.Id, &see_seek_master_ask_and_answer.Uuid, &see_seek_master_ask_and_answer.SeeSeekMasterId, &see_seek_master_ask_and_answer.Classify, &see_seek_master_ask_and_answer.Status, &see_seek_master_ask_and_answer.RequestTitle, &see_seek_master_ask_and_answer.RequestContent, &see_seek_master_ask_and_answer.RequestHistory, &see_seek_master_ask_and_answer.MasterTitle, &see_seek_master_ask_and_answer.MasterContent, &see_seek_master_ask_and_answer.MasterHistory, &see_seek_master_ask_and_answer.CreatedAt, &see_seek_master_ask_and_answer.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_ask_and_answer *SeeSeekMasterAskAndAnswer) GetByUuid() (err error) {
	statement := "SELECT id, uuid, see_seek_master_id, classify, status, request_title, request_content, request_history, master_title, master_content, master_history, created_at, updated_at FROM see_seek_master_ask_and_answers WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_master_ask_and_answer.Uuid).Scan(&see_seek_master_ask_and_answer.Id, &see_seek_master_ask_and_answer.Uuid, &see_seek_master_ask_and_answer.SeeSeekMasterId, &see_seek_master_ask_and_answer.Classify, &see_seek_master_ask_and_answer.Status, &see_seek_master_ask_and_answer.RequestTitle, &see_seek_master_ask_and_answer.RequestContent, &see_seek_master_ask_and_answer.RequestHistory, &see_seek_master_ask_and_answer.MasterTitle, &see_seek_master_ask_and_answer.MasterContent, &see_seek_master_ask_and_answer.MasterHistory, &see_seek_master_ask_and_answer.CreatedAt, &see_seek_master_ask_and_answer.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_ask_and_answer *SeeSeekMasterAskAndAnswer) Update() (err error) {
	statement := "UPDATE see_seek_master_ask_and_answers SET see_seek_master_id=$2, classify=$3, status=$4, request_title=$5, request_content=$6, request_history=$7, master_title=$8, master_content=$9, master_history=$10, updated_at=$11 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(see_seek_master_ask_and_answer.Id, see_seek_master_ask_and_answer.SeeSeekMasterId, see_seek_master_ask_and_answer.Classify, see_seek_master_ask_and_answer.Status, see_seek_master_ask_and_answer.RequestTitle, see_seek_master_ask_and_answer.RequestContent, see_seek_master_ask_and_answer.RequestHistory, see_seek_master_ask_and_answer.MasterTitle, see_seek_master_ask_and_answer.MasterContent, see_seek_master_ask_and_answer.MasterHistory, time.Now())
	if err != nil {
		return
	}
	return
}

// 报告
type SeeSeekMasterExaminationReport struct {
	ID              int    `json:"id"`
	Uuid            string `json:"uuid"`
	SeeSeekMasterID int    `json:"see_seek_master_id"`
	Classify        int    `json:"classify"` // 1: Device, 2: Pet
	Status          int    `json:"status"`   // 0: Draft, 1: Completed, 2: Reviewed

	Name        string `json:"name"`
	Nickname    string `json:"nickname"`
	Description string `json:"description"`

	SampleType        string `json:"sample_type"`
	SampleOrder       string `json:"sample_order"`
	InstrumentGoodsID int    `json:"instrument_goods_id"`

	ReportTitle   string `json:"report_title"`
	ReportContent string `json:"report_content"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`

	MasterUserId   int `json:"master_user_id"`
	ReviewerUserId int `json:"reviewer_user_id"`

	ReportDate time.Time `json:"report_date"`
	Attachment string    `json:"attachment"` //用于存储报告的附件（如图片、PDF 等）
	Tags       string    `json:"tags"`       //用于标记报告的分类或关键词
}

// SeeSeekMasterExaminationReport.Create() // 创建一个SeeSeekMasterExaminationReport
func (s *SeeSeekMasterExaminationReport) Create() (err error) {
	statement := "INSERT INTO see_seek_master_examination_reports (uuid, see_seek_master_id, classify, status, name, nickname, description, sample_type, sample_order, instrument_goods_id, report_title, report_content, created_at, master_user_id, reviewer_user_id, report_date, attachment, tags) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), s.SeeSeekMasterID, s.Classify, s.Status, s.Name, s.Nickname, s.Description, s.SampleType, s.SampleOrder, s.InstrumentGoodsID, s.ReportTitle, s.ReportContent, time.Now(), s.MasterUserId, s.ReviewerUserId, s.ReportDate, s.Attachment, s.Tags).Scan(&s.ID, &s.Uuid)
	if err != nil {
		return
	}
	return
}
func (s *SeeSeekMasterExaminationReport) Update() (err error) {
	statement := "UPDATE see_seek_master_examination_reports SET see_seek_master_id=$2, classify=$3, status=$4, name=$5, nickname=$6, description=$7, sample_type=$8, sample_order=$9, instrument_goods_id=$10, report_title=$11, report_content=$12, updated_at=$13, master_user_id=$14, reviewer_user_id=$15, report_date=$16, attachment=$17, tags=$18 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(s.ID, s.SeeSeekMasterID, s.Classify, s.Status, s.Name, s.Nickname, s.Description, s.SampleType, s.SampleOrder, s.InstrumentGoodsID, s.ReportTitle, s.ReportContent, time.Now(), s.MasterUserId, s.ReviewerUserId, s.ReportDate, s.Attachment, s.Tags)
	if err != nil {
		return
	}
	return
}

// SeeSeekMasterExaminationReport.Get() // 读取一个SeeSeekMasterExaminationReport
func (s *SeeSeekMasterExaminationReport) Get() (err error) {
	statement := "SELECT id, uuid, see_seek_master_id, classify, status, name, nickname, description, sample_type, sample_order, instrument_goods_id, report_title, report_content, created_at, updated_at, master_user_id, reviewer_user_id, report_date, attachment, tags FROM see_seek_master_examination_reports WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(s.ID).Scan(&s.ID, &s.Uuid, &s.SeeSeekMasterID, &s.Classify, &s.Status, &s.Name, &s.Nickname, &s.Description, &s.SampleType, &s.SampleOrder, &s.InstrumentGoodsID, &s.ReportTitle, &s.ReportContent, &s.CreatedAt, &s.UpdatedAt, &s.MasterUserId, &s.ReviewerUserId, &s.ReportDate, &s.Attachment, &s.Tags)
	if err != nil {
		return
	}
	return
}
func (s *SeeSeekMasterExaminationReport) GetByUuid() (err error) {
	statement := "SELECT id, uuid, see_seek_master_id, classify, status, name, nickname, description, sample_type, sample_order, instrument_goods_id, report_title, report_content, created_at, updated_at, master_user_id, reviewer_user_id, report_date, attachment, tags FROM see_seek_master_examination_reports WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(s.Uuid).Scan(&s.ID, &s.Uuid, &s.SeeSeekMasterID, &s.Classify, &s.Status, &s.Name, &s.Nickname, &s.Description, &s.SampleType, &s.SampleOrder, &s.InstrumentGoodsID, &s.ReportTitle, &s.ReportContent, &s.CreatedAt, &s.UpdatedAt, &s.MasterUserId, &s.ReviewerUserId, &s.ReportDate, &s.Attachment, &s.Tags)
	if err != nil {
		return
	}
	return
}

// 报告项目
type SeeSeekMasterExaminationItem struct {
	ID       int    `json:"id"`
	Uuid     string `json:"uuid"`
	Classify int    `json:"classify"` // 1: Device, 2: Pet

	SeeSeekMasterExaminationReportID int `json:"see_seek_master_examination_report_id"`

	ItemCode     string   `json:"item_code"`
	ItemName     string   `json:"item_name"`
	Result       string   `json:"result"` // Can be string, number, or boolean
	ResultUnit   string   `json:"result_unit"`
	ReferenceMin *float64 `json:"reference_min"`
	ReferenceMax *float64 `json:"reference_max"`
	Remark       string   `json:"remark"`
	AbnormalFlag bool     `json:"abnormal_flag"` //（如布尔值或枚举），用于标记检查结果是否异常

	Method   string `json:"method"`   //方法
	Operator string `json:"operator"` //操作员

	Status    int        `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// SeeSeekMasterExaminationItem.Create() // 创建一个SeeSeekMasterExaminationItem
func (s *SeeSeekMasterExaminationItem) Create() (err error) {
	statement := "INSERT INTO see_seek_master_examination_items (uuid, classify, see_seek_master_examination_report_id, item_code, item_name, result, result_unit, reference_min, reference_max, remark, abnormal_flag, method, operator, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), s.Classify, s.SeeSeekMasterExaminationReportID, s.ItemCode, s.ItemName, s.Result, s.ResultUnit, s.ReferenceMin, s.ReferenceMax, s.Remark, s.AbnormalFlag, s.Method, s.Operator, s.Status, time.Now()).Scan(&s.ID, &s.Uuid)
	if err != nil {
		return
	}
	return
}
func (s *SeeSeekMasterExaminationItem) Update() (err error) {
	statement := "UPDATE see_seek_master_examination_items SET classify=$2, see_seek_master_examination_report_id=$3, item_code=$4, item_name=$5, result=$6, result_unit=$7, reference_min=$8, reference_max=$9, remark=$10, abnormal_flag=$11, method=$12, operator=$13, status=$14, updated_at=$15 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(s.ID, s.Classify, s.SeeSeekMasterExaminationReportID, s.ItemCode, s.ItemName, s.Result, s.ResultUnit, s.ReferenceMin, s.ReferenceMax, s.Remark, s.AbnormalFlag, s.Method, s.Operator, s.Status, time.Now())
	if err != nil {
		return
	}
	return
}
func (s *SeeSeekMasterExaminationItem) Get() (err error) {
	statement := "SELECT id, uuid, classify, see_seek_master_examination_report_id, item_code, item_name, result, result_unit, reference_min, reference_max, remark, abnormal_flag, method, operator, status, created_at, updated_at FROM see_seek_master_examination_items WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(s.ID).Scan(&s.ID, &s.Uuid, &s.Classify, &s.SeeSeekMasterExaminationReportID, &s.ItemCode, &s.ItemName, &s.Result, &s.ResultUnit, &s.ReferenceMin, &s.ReferenceMax, &s.Remark, &s.AbnormalFlag, &s.Method, &s.Operator, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (s *SeeSeekMasterExaminationItem) GetByUuid() (err error) {
	statement := "SELECT id, uuid, classify, see_seek_master_examination_report_id, item_code, item_name, result, result_unit, reference_min, reference_max, remark, abnormal_flag, method, operator, status, created_at, updated_at FROM see_seek_master_examination_items WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(s.Uuid).Scan(&s.ID, &s.Uuid, &s.Classify, &s.SeeSeekMasterExaminationReportID, &s.ItemCode, &s.ItemName, &s.Result, &s.ResultUnit, &s.ReferenceMin, &s.ReferenceMax, &s.Remark, &s.AbnormalFlag, &s.Method, &s.Operator, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// 凭据，依据，指音视频等视觉证据，
type SeeSeekEvidence struct {
	Id             int
	Uuid           string
	SeeSeekId      int    // 标记属于那一个“看看”，
	Description    string // 描述记录
	RecorderUserId int    // 记录人id
	Note           string //备注,特别说明
	Category       int    //分类：1、图片，2、视频，3、音频，4、其他
	Link           string // 储存链接（地址）
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

func (see_seek_evidence *SeeSeekEvidence) Create() (err error) {
	statement := "INSERT INTO see_seek_evidences (uuid, see_seek_id, description, recorder_user_id, note, category, link, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), see_seek_evidence.SeeSeekId, see_seek_evidence.Description, see_seek_evidence.RecorderUserId, see_seek_evidence.Note, see_seek_evidence.Category, see_seek_evidence.Link, time.Now()).Scan(&see_seek_evidence.Id, &see_seek_evidence.Uuid)
	if err != nil {
		return
	}
	return
}
func (see_seek_evidence *SeeSeekEvidence) Update() (err error) {
	statement := "UPDATE see_seek_evidences SET see_seek_id=$2, description=$3, recorder_user_id=$4, note=$5, category=$6, link=$7, updated_at=$8 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(see_seek_evidence.Id, see_seek_evidence.SeeSeekId, see_seek_evidence.Description, see_seek_evidence.RecorderUserId, see_seek_evidence.Note, see_seek_evidence.Category, see_seek_evidence.Link, time.Now())
	if err != nil {
		return
	}
	return
}
func (see_seek_evidence *SeeSeekEvidence) Get() (err error) {
	statement := "SELECT id, uuid, see_seek_id, description, recorder_user_id, note, category, link, created_at, updated_at FROM see_seek_evidences WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_evidence.Id).Scan(&see_seek_evidence.Id, &see_seek_evidence.Uuid, &see_seek_evidence.SeeSeekId, &see_seek_evidence.Description, &see_seek_evidence.RecorderUserId, &see_seek_evidence.Note, &see_seek_evidence.Category, &see_seek_evidence.Link, &see_seek_evidence.CreatedAt, &see_seek_evidence.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (see_seek_evidence *SeeSeekEvidence) GetByUuid() (err error) {
	statement := "SELECT id, uuid, see_seek_id, description, recorder_user_id, note, category, link, created_at, updated_at FROM see_seek_evidences WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_evidence.Uuid).Scan(&see_seek_evidence.Id, &see_seek_evidence.Uuid, &see_seek_evidence.SeeSeekId, &see_seek_evidence.Description, &see_seek_evidence.RecorderUserId, &see_seek_evidence.Note, &see_seek_evidence.Category, &see_seek_evidence.Link, &see_seek_evidence.CreatedAt, &see_seek_evidence.UpdatedAt)
	if err != nil {
		return
	}
	return
}
