package data

import (
	"context"
	"time"
)

// 看看，睇睇，
// 一个为解决某个具体问题而举行的茶会记录
type SeeSeek struct {
	Id          int
	Uuid        string
	Name        string //例如：洗手盆故障检查 或者 厨房蟑螂出没？
	Nickname    string
	Description string

	ProjectId int
	// 日期时间
	StartTime time.Time // 开始时间，默认为当前时间
	EndTime   time.Time // 结束时间，默认为开始时间+1小时
	PlaceId   int       // 约茶地方ID
	// 相关团队和家庭信息
	PayerUserId   int //出茶叶代表人Id
	PayerTeamId   int //出茶叶团队Id
	PayerFamilyId int //出茶叶家庭Id

	PayeeUserId   int //收茶叶代表人Id
	PayeeTeamId   int //收茶叶团队Id
	PayeeFamilyId int //收茶叶家庭Id

	VerifierUserId   int
	VerifierFamilyId int
	VerifierTeamId   int

	Category int //分类：0、公开，1、保密，仅当事家庭/团队可见内容
	Status   int //状态：0、未开始，1、进行中，2、暂停，3、已终止，4、已结束

	CreatedAt time.Time
	UpdatedAt *time.Time
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

// SeeSeek.Create() // 创建一个SeeSeek
// 编写postgreSQL语句，插入新纪录，return （err error）
func (s *SeeSeek) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO see_seeks 
		(uuid, name, nickname, description, place_id, project_id, start_time, end_time, 
		 payer_user_id, payer_team_id, payer_family_id, payee_user_id, payee_team_id, payee_family_id, 
		 verifier_user_id, verifier_team_id, verifier_family_id, category, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19) 
		RETURNING id, uuid`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, Random_UUID(), s.Name, s.Nickname, s.Description,
		s.PlaceId, s.ProjectId, s.StartTime, s.EndTime, s.PayerUserId, s.PayerTeamId,
		s.PayerFamilyId, s.PayeeUserId, s.PayeeTeamId, s.PayeeFamilyId, s.VerifierUserId,
		s.VerifierTeamId, s.VerifierFamilyId, s.Category, s.Status).Scan(&s.Id, &s.Uuid)
	return err
}

// SeeSeek.Get()
func (s *SeeSeek) Get(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, name, nickname, description, user_id, project_id, category, status, created_at, updated_at 
		FROM see_seeks WHERE id=$1`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, s.Id).Scan(&s.Id, &s.Uuid, &s.Name, &s.Nickname, &s.Description, &s.PlaceId, &s.ProjectId, &s.Category, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// SeeSeek.CreateAtDate() string
func (s *SeeSeek) CreatedDateTime() string {
	return s.CreatedAt.Format(FMT_DATE_TIME_CN)
}

func (s *SeeSeek) StatusString() string {
	switch s.Status {
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

// “看看”场所安全隐患，风险的源
type SeeSeekHazard struct {
	Id        int
	Uuid      string
	SeeSeekId int
	HazardId  int
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 执行“看看”作业安全风险，风险考验因素
type SeeSeekRisk struct {
	Id        int
	Uuid      string
	SeeSeekId int
	RiskId    int
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 望，观察
type SeeSeekLook struct {
	Id        int
	Uuid      string
	SeeSeekId int
	Classify  int
	Status    int

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
type SeeSeekListen struct {
	Id        int
	Uuid      string
	SeeSeekId int
	Classify  int
	Status    int

	RequestSound        string //声音
	IsAbnormal          bool   //是否有异常声音
	RequestSoundHistory string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 嗅，气味
type SeeSeekSmell struct {
	Id        int
	Uuid      string
	SeeSeekId int
	Classify  int
	Status    int

	RequestOdour        string //气味
	IsFoulOdour         bool   //是否异味
	RequestOdourHistory string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 触摸，
type SeeSeekTouch struct {
	Id        int
	Uuid      string
	SeeSeekId int
	Classify  int

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
type SeeSeekAskAndAnswer struct {
	Id        int
	Uuid      string
	SeeSeekId int
	Classify  int

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

func (ssaa *SeeSeekAskAndAnswer) Create() (err error) {
	statement := `INSERT INTO see_seek_ask_and_answers 
		(uuid, see_seek_id, classify, status, request_title, request_content, request_history, 
		 master_title, master_content, master_history, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		RETURNING id, uuid`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), ssaa.SeeSeekId, ssaa.Classify, ssaa.Status, ssaa.RequestTitle, ssaa.RequestContent, ssaa.RequestHistory, ssaa.MasterTitle, ssaa.MasterContent, ssaa.MasterHistory, time.Now()).Scan(&ssaa.Id, &ssaa.Uuid)
	if err != nil {
		return
	}
	return
}
func (ssaa *SeeSeekAskAndAnswer) Get() (err error) {
	statement := `SELECT id, uuid, see_seek_id, classify, status, request_title, request_content, request_history, 
		 master_title, master_content, master_history, created_at, updated_at 
		FROM see_seek_ask_and_answers WHERE id=$1`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(ssaa.Id).Scan(&ssaa.Id, &ssaa.Uuid, &ssaa.SeeSeekId, &ssaa.Classify, &ssaa.Status, &ssaa.RequestTitle, &ssaa.RequestContent, &ssaa.RequestHistory, &ssaa.MasterTitle, &ssaa.MasterContent, &ssaa.MasterHistory, &ssaa.CreatedAt, &ssaa.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (ssaa *SeeSeekAskAndAnswer) GetByUuid() (err error) {
	statement := `SELECT id, uuid, see_seek_id, classify, status, request_title, request_content, request_history, 
		 master_title, master_content, master_history, created_at, updated_at 
		FROM see_seek_ask_and_answers WHERE uuid=$1`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(ssaa.Uuid).Scan(&ssaa.Id, &ssaa.Uuid, &ssaa.SeeSeekId, &ssaa.Classify, &ssaa.Status, &ssaa.RequestTitle, &ssaa.RequestContent, &ssaa.RequestHistory, &ssaa.MasterTitle, &ssaa.MasterContent, &ssaa.MasterHistory, &ssaa.CreatedAt, &ssaa.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (ssaa *SeeSeekAskAndAnswer) Update() (err error) {
	statement := `UPDATE see_seek_ask_and_answers 
		SET see_seek_id=$2, classify=$3, status=$4, request_title=$5, request_content=$6, request_history=$7, 
		    master_title=$8, master_content=$9, master_history=$10, updated_at=$11 
		WHERE id=$1`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(ssaa.Id, ssaa.SeeSeekId, ssaa.Classify, ssaa.Status, ssaa.RequestTitle, ssaa.RequestContent, ssaa.RequestHistory, ssaa.MasterTitle, ssaa.MasterContent, ssaa.MasterHistory, time.Now())
	if err != nil {
		return
	}
	return
}

// 报告
type SeeSeekExaminationReport struct {
	ID        int    `json:"id"`
	Uuid      string `json:"uuid"`
	SeeSeekID int    `json:"see_seek_id"`
	Classify  int    `json:"classify"` // 1: Device, 2: Pet
	Status    int    `json:"status"`   // 0: Draft, 1: Completed, 2: Reviewed

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

// SeeSeekExaminationReport.Create() // 创建一个 SeeSeekExaminationReport
func (s *SeeSeekExaminationReport) Create() (err error) {
	statement := `INSERT INTO see_seek_examination_reports 
		(uuid, see_seek_id, classify, status, name, nickname, description, sample_type, sample_order, 
		 instrument_goods_id, report_title, report_content, created_at, master_user_id, reviewer_user_id, 
		 report_date, attachment, tags) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) 
		RETURNING id, uuid`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), s.SeeSeekID, s.Classify, s.Status, s.Name, s.Nickname,
		s.Description, s.SampleType, s.SampleOrder, s.InstrumentGoodsID, s.ReportTitle,
		s.ReportContent, time.Now(), s.MasterUserId, s.ReviewerUserId, s.ReportDate,
		s.Attachment, s.Tags).Scan(&s.ID, &s.Uuid)
	return err
}
func (s *SeeSeekExaminationReport) Update() (err error) {
	statement := "UPDATE see_seek_examination_reports SET see_seek_id=$2, classify=$3, status=$4, name=$5, nickname=$6, description=$7, sample_type=$8, sample_order=$9, instrument_goods_id=$10, report_title=$11, report_content=$12, updated_at=$13, master_user_id=$14, reviewer_user_id=$15, report_date=$16, attachment=$17, tags=$18 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(s.ID, s.SeeSeekID, s.Classify, s.Status, s.Name, s.Nickname, s.Description, s.SampleType, s.SampleOrder, s.InstrumentGoodsID, s.ReportTitle, s.ReportContent, time.Now(), s.MasterUserId, s.ReviewerUserId, s.ReportDate, s.Attachment, s.Tags)
	if err != nil {
		return
	}
	return
}

// SeeSeekExaminationReport.Get() // 读取一个 SeeSeekExaminationReport
func (s *SeeSeekExaminationReport) Get() (err error) {
	statement := "SELECT id, uuid, see_seek_id, classify, status, name, nickname, description, sample_type, sample_order, instrument_goods_id, report_title, report_content, created_at, updated_at, master_user_id, reviewer_user_id, report_date, attachment, tags FROM see_seek_examination_reports WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(s.ID).Scan(&s.ID, &s.Uuid, &s.SeeSeekID, &s.Classify, &s.Status, &s.Name, &s.Nickname, &s.Description, &s.SampleType, &s.SampleOrder, &s.InstrumentGoodsID, &s.ReportTitle, &s.ReportContent, &s.CreatedAt, &s.UpdatedAt, &s.MasterUserId, &s.ReviewerUserId, &s.ReportDate, &s.Attachment, &s.Tags)
	if err != nil {
		return
	}
	return
}
func (s *SeeSeekExaminationReport) GetByUuid() (err error) {
	statement := "SELECT id, uuid, see_seek_id, classify, status, name, nickname, description, sample_type, sample_order, instrument_goods_id, report_title, report_content, created_at, updated_at, master_user_id, reviewer_user_id, report_date, attachment, tags FROM see_seek_examination_reports WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(s.Uuid).Scan(&s.ID, &s.Uuid, &s.SeeSeekID, &s.Classify, &s.Status, &s.Name, &s.Nickname, &s.Description, &s.SampleType, &s.SampleOrder, &s.InstrumentGoodsID, &s.ReportTitle, &s.ReportContent, &s.CreatedAt, &s.UpdatedAt, &s.MasterUserId, &s.ReviewerUserId, &s.ReportDate, &s.Attachment, &s.Tags)
	if err != nil {
		return
	}
	return
}

// 报告项目
type SeeSeekExaminationItem struct {
	ID       int    `json:"id"`
	Uuid     string `json:"uuid"`
	Classify int    `json:"classify"` // 1: Device, 2: Pet

	SeeSeekExaminationReportID int `json:"see_seek_examination_report_id"`

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

// SeeSeekExaminationItem.Create() // 创建一个 SeeSeekExaminationItem
func (s *SeeSeekExaminationItem) Create() (err error) {
	statement := "INSERT INTO see_seek_examination_items (uuid, classify, see_seek_examination_report_id, item_code, item_name, result, result_unit, reference_min, reference_max, remark, abnormal_flag, method, operator, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), s.Classify, s.SeeSeekExaminationReportID, s.ItemCode, s.ItemName, s.Result, s.ResultUnit, s.ReferenceMin, s.ReferenceMax, s.Remark, s.AbnormalFlag, s.Method, s.Operator, s.Status, time.Now()).Scan(&s.ID, &s.Uuid)
	if err != nil {
		return
	}
	return
}
func (s *SeeSeekExaminationItem) Update() (err error) {
	statement := "UPDATE see_seek_examination_items SET classify=$2, see_seek_examination_report_id=$3, item_code=$4, item_name=$5, result=$6, result_unit=$7, reference_min=$8, reference_max=$9, remark=$10, abnormal_flag=$11, method=$12, operator=$13, status=$14, updated_at=$15 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(s.ID, s.Classify, s.SeeSeekExaminationReportID, s.ItemCode, s.ItemName, s.Result, s.ResultUnit, s.ReferenceMin, s.ReferenceMax, s.Remark, s.AbnormalFlag, s.Method, s.Operator, s.Status, time.Now())
	if err != nil {
		return
	}
	return
}
func (s *SeeSeekExaminationItem) Get() (err error) {
	statement := "SELECT id, uuid, classify, see_seek_examination_report_id, item_code, item_name, result, result_unit, reference_min, reference_max, remark, abnormal_flag, method, operator, status, created_at, updated_at FROM see_seek_examination_items WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(s.ID).Scan(&s.ID, &s.Uuid, &s.Classify, &s.SeeSeekExaminationReportID, &s.ItemCode, &s.ItemName, &s.Result, &s.ResultUnit, &s.ReferenceMin, &s.ReferenceMax, &s.Remark, &s.AbnormalFlag, &s.Method, &s.Operator, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (s *SeeSeekExaminationItem) GetByUuid() (err error) {
	statement := "SELECT id, uuid, classify, see_seek_examination_report_id, item_code, item_name, result, result_unit, reference_min, reference_max, remark, abnormal_flag, method, operator, status, created_at, updated_at FROM see_seek_examination_items WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(s.Uuid).Scan(&s.ID, &s.Uuid, &s.Classify, &s.SeeSeekExaminationReportID, &s.ItemCode, &s.ItemName, &s.Result, &s.ResultUnit, &s.ReferenceMin, &s.ReferenceMax, &s.Remark, &s.AbnormalFlag, &s.Method, &s.Operator, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return
	}
	return
}
