package dao

import "time"

// 专项检查报告
type SeeSeekExaminationReport struct {
	ID        int    `json:"id"`
	Uuid      string `json:"uuid"`
	SeeSeekID int    `json:"see_seek_id"`
	Classify  int    `json:"classify"` // 1: 设备, 2: 宠物, 3: 动植物，4: 矿物
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

	MasterUserId   int `json:"master_user_id"` // 操作人员
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
	stmt, err := DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
	Classify int    `json:"classify"` // 1: 设备, 2: 宠物, 3: 动植物，4: 矿物

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
	stmt, err := DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
