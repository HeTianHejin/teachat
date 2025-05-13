package data

import "time"

// 看看，睇睇，
type SeeSeek struct {
	Id          int
	Uuid        string
	Name        string
	Nickname    string
	Description string
	Category    int
	Status      int
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

// SeeSeek.Create() // 创建一个SeeSeek
// 编写postgreSQL语句，插入新纪录，return （err error）
func (see_seek *SeeSeek) Create() (err error) {
	statement := "INSERT INTO see_seeks (uuid, name, nickname, description, category, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), see_seek.Name, see_seek.Nickname, see_seek.Description, see_seek.Category, see_seek.Status, time.Now()).Scan(&see_seek.Id, &see_seek.Uuid)

	return
}

// SeeSeek.Get() // 读取一个SeeSeek
func (see_seek *SeeSeek) Get() (err error) {
	statement := "SELECT id, uuid, name, nickname, description, category, status, created_at, updated_at FROM see_seeks WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek.Id).Scan(&see_seek.Id, &see_seek.Uuid, &see_seek.Name, &see_seek.Nickname, &see_seek.Description, &see_seek.Category, &see_seek.Status, &see_seek.CreatedAt, &see_seek.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// SeeSeek.GetByUuid() // 读取一个SeeSeek
func (see_seek *SeeSeek) GetByUuid() (err error) {
	statement := "SELECT id, uuid, name, nickname, description, category, status, created_at, updated_at FROM see_seeks WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek.Uuid).Scan(&see_seek.Id, &see_seek.Uuid, &see_seek.Name, &see_seek.Nickname, &see_seek.Description, &see_seek.Category, &see_seek.Status, &see_seek.CreatedAt, &see_seek.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// SeeSeek.Update() // 更新一个SeeSeek
func (see_seek *SeeSeek) Update() (err error) {
	statement := "UPDATE see_seeks SET name=$2, nickname=$3, description=$4, category=$5, status=$6, updated_at=$7 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(see_seek.Id, see_seek.Name, see_seek.Nickname, see_seek.Description, see_seek.Category, see_seek.Status, time.Now())
	if err != nil {
		return
	}
	return
}

// 与特定团队举行一个结构化交流
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
