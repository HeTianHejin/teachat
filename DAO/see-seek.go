package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// 看看，睇睇，
// 一个为解决某个具体问题而举行的茶会“探察环节“记录
type SeeSeek struct {
	Id          int
	Uuid        string
	Name        string //例如：洗手盆故障检查 或者 踏勘厨房蟑螂出没情况？
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

	Category int           //分类：0、公开，1、保密，仅当事家庭/团队可见内容
	Status   SeeSeekStatus //状态：0、未开始，1、进行中，2、暂停，3、已终止，4、已结束
	Step     int           //当前步骤：1、环境条件，2、场所隐患，3、风险评估，4、感官观察，5、检测报告

	CreatedAt time.Time
	UpdatedAt *time.Time
}

const (
	SeeSeekCategoryPublic = iota // 公开
	SeeSeekCategorySecret        // 保密
)

type SeeSeekStatus int

const (
	SeeSeekStatusNotStarted = iota // 未开始
	SeeSeekStatusInProgress        // 进行中
	SeeSeekStatusPaused            // 已暂停
	SeeSeekStatusAborted           // 已半途终止（异常）
	SeeSeekStatusCompleted         // 已完成（顺利结束）
)

const (
	SeeSeekStepEnvironment = iota + 1 // 1、环境条件
	SeeSeekStepHazard                 // 2、场所隐患
	SeeSeekStepRisk                   // 3、风险评估
	SeeSeekStepObservation            // 4、感官观察
	SeeSeekStepReport                 // 5、检测报告
)

func GetSeeSeekStepTitle(step int) string {
	switch step {
	case SeeSeekStepEnvironment:
		return "环境条件"
	case SeeSeekStepHazard:
		return "场所隐患"
	case SeeSeekStepRisk:
		return "风险评估"
	case SeeSeekStepObservation:
		return "感官观察"
	case SeeSeekStepReport:
		return "检测报告"
	default:
		return "未知步骤"
	}
}

// SeeSeek.Create() // 创建一个SeeSeek
// 编写postgreSQL语句，插入新纪录，return （err error）
func (s *SeeSeek) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO see_seeks 
		(uuid, name, nickname, description, place_id, project_id, start_time, end_time, 
		 payer_user_id, payer_team_id, payer_family_id, payee_user_id, payee_team_id, payee_family_id, 
		 verifier_user_id, verifier_team_id, verifier_family_id, category, status, step) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20) 
		RETURNING id, uuid`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, Random_UUID(), s.Name, s.Nickname, s.Description,
		s.PlaceId, s.ProjectId, s.StartTime, s.EndTime, s.PayerUserId, s.PayerTeamId,
		s.PayerFamilyId, s.PayeeUserId, s.PayeeTeamId, s.PayeeFamilyId, s.VerifierUserId,
		s.VerifierTeamId, s.VerifierFamilyId, s.Category, s.Status, s.Step).Scan(&s.Id, &s.Uuid)
	return err
}

// SeeSeek.GetByIdOrUUID()
func (s *SeeSeek) GetByIdOrUUID(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if s.Id < 0 || s.Uuid == "" {
		return errors.New("invalid SeeSeek ID or UUID")
	}
	statement := `SELECT id, uuid, name, nickname, description, place_id, project_id,
		payer_user_id, payer_team_id, payer_family_id, payee_user_id, payee_team_id, payee_family_id,
		verifier_user_id, verifier_team_id, verifier_family_id, category, status, step,
		start_time, end_time, created_at, updated_at
		FROM see_seeks WHERE id=$1 OR uuid=$2`
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, s.Id, s.Uuid).Scan(&s.Id, &s.Uuid, &s.Name, &s.Nickname, &s.Description,
		&s.PlaceId, &s.ProjectId, &s.PayerUserId, &s.PayerTeamId, &s.PayerFamilyId,
		&s.PayeeUserId, &s.PayeeTeamId, &s.PayeeFamilyId, &s.VerifierUserId,
		&s.VerifierTeamId, &s.VerifierFamilyId, &s.Category, &s.Status, &s.Step,
		&s.StartTime, &s.EndTime, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
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

// “看看”作业场所自然环境条件
type SeeSeekEnvironment struct {
	Id            int
	Uuid          string
	SeeSeekId     int
	EnvironmentId int
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

// SeeSeekEnvironment.Create()
func (sse *SeeSeekEnvironment) Create() (err error) {
	statement := `INSERT INTO see_seek_environments 
		(uuid, see_seek_id, environment_id, created_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, uuid`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), sse.SeeSeekId, sse.EnvironmentId, time.Now()).Scan(&sse.Id, &sse.Uuid)
	return err
}

// SeeSeekEnvironment.Get()
func (sse *SeeSeekEnvironment) Get() (err error) {
	statement := `SELECT id, uuid, see_seek_id, environment_id, created_at, updated_at
		FROM see_seek_environments WHERE id=$1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(sse.Id).Scan(&sse.Id, &sse.Uuid, &sse.SeeSeekId, &sse.EnvironmentId, &sse.CreatedAt, &sse.UpdatedAt)
	if err != nil {
		return
	}
	return
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

// SeeSeekHazard.Create()
func (ssh *SeeSeekHazard) Create() (err error) {
	statement := `INSERT INTO see_seek_hazards
		(uuid, see_seek_id, hazard_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, uuid`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), ssh.SeeSeekId, ssh.HazardId, time.Now()).Scan(&ssh.Id, &ssh.Uuid)
	return err
}

// SeeSeekHazard.Get()
func (ssh *SeeSeekHazard) Get() (err error) {
	statement := `SELECT id, uuid, see_seek_id, hazard_id, created_at, updated_at
		FROM see_seek_hazards WHERE id=$1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(ssh.Id).Scan(&ssh.Id, &ssh.Uuid, &ssh.SeeSeekId, &ssh.HazardId, &ssh.CreatedAt, &ssh.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// 验证隐患ID是否存在
func IsHazardIdExists(hazardId int) (bool, error) {
	var exists bool
	statement := `SELECT EXISTS(SELECT 1 FROM hazards WHERE id=$1)`
	err := db.QueryRow(statement, hazardId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// SeeSeekHazard.DeleteBySeeSeekId()
func DeleteSeeSeekHazardsBySeeSeekId(seeSeekId int) error {
	statement := `DELETE FROM see_seek_hazards WHERE see_seek_id=$1`
	_, err := db.Exec(statement, seeSeekId)
	return err
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

// SeeSeekRisk.Create()
func (ssr *SeeSeekRisk) Create() (err error) {
	statement := `INSERT INTO see_seek_risks
		(uuid, see_seek_id, risk_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, uuid`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), ssr.SeeSeekId, ssr.RiskId, time.Now()).Scan(&ssr.Id, &ssr.Uuid)
	return err
}

// SeeSeekRisk.Get()
func (ssr *SeeSeekRisk) Get() (err error) {
	statement := `SELECT id, uuid, see_seek_id, risk_id, created_at, updated_at
		FROM see_seek_risks WHERE id=$1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(ssr.Id).Scan(&ssr.Id, &ssr.Uuid, &ssr.SeeSeekId, &ssr.RiskId, &ssr.CreatedAt, &ssr.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// SeeSeekRisk.DeleteBySeeSeekId()
func DeleteSeeSeekRisksBySeeSeekId(seeSeekId int) error {
	statement := `DELETE FROM see_seek_risks WHERE see_seek_id=$1`
	_, err := db.Exec(statement, seeSeekId)
	return err
}

// 望，观察
type SeeSeekLook struct {
	Id        int
	Uuid      string
	SeeSeekId int
	Classify  int
	Status    int //0、未开始，1、已完成

	Outline  string //外形轮廓
	IsDeform bool   //是否变形
	Skin     string //表面皮肤
	IsGraze  bool   //是否破损
	Color    string //颜色
	IsChange bool   //是否变色

	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (ssl *SeeSeekLook) Create() (err error) {
	statement := `INSERT INTO see_seek_looks 
		(uuid, see_seek_id, classify, status, outline, is_deform, skin, is_graze, color, is_change, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		RETURNING id, uuid`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), ssl.SeeSeekId, ssl.Classify, ssl.Status, ssl.Outline, ssl.IsDeform, ssl.Skin, ssl.IsGraze, ssl.Color, ssl.IsChange, time.Now()).Scan(&ssl.Id, &ssl.Uuid)
	return err
}

// SeeSeekLook.DeleteBySeeSeekId()
func DeleteSeeSeekLooksBySeeSeekId(seeSeekId int) error {
	statement := `DELETE FROM see_seek_looks WHERE see_seek_id=$1`
	_, err := db.Exec(statement, seeSeekId)
	return err
}

// 听，声音
type SeeSeekListen struct {
	Id        int
	Uuid      string
	SeeSeekId int
	Classify  int
	Status    int

	Sound      string //声音
	IsAbnormal bool   //是否有异常声音

	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (ssl *SeeSeekListen) Create() (err error) {
	statement := `INSERT INTO see_seek_listens 
		(uuid, see_seek_id, classify, status, sound, is_abnormal, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, uuid`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), ssl.SeeSeekId, ssl.Classify, ssl.Status, ssl.Sound, ssl.IsAbnormal, time.Now()).Scan(&ssl.Id, &ssl.Uuid)
	return err
}

// SeeSeekListen.DeleteBySeeSeekId()
func DeleteSeeSeekListensBySeeSeekId(seeSeekId int) error {
	statement := `DELETE FROM see_seek_listens WHERE see_seek_id=$1`
	_, err := db.Exec(statement, seeSeekId)
	return err
}

// 嗅，气味
type SeeSeekSmell struct {
	Id        int
	Uuid      string
	SeeSeekId int
	Classify  int
	Status    int

	Odour       string //气味
	IsFoulOdour bool   //是否异味

	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (sss *SeeSeekSmell) Create() (err error) {
	statement := `INSERT INTO see_seek_smells 
		(uuid, see_seek_id, classify, status, odour, is_foul_odour, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, uuid`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), sss.SeeSeekId, sss.Classify, sss.Status, sss.Odour, sss.IsFoulOdour, time.Now()).Scan(&sss.Id, &sss.Uuid)
	return err
}

// SeeSeekSmell.DeleteBySeeSeekId()
func DeleteSeeSeekSmellsBySeeSeekId(seeSeekId int) error {
	statement := `DELETE FROM see_seek_smells WHERE see_seek_id=$1`
	_, err := db.Exec(statement, seeSeekId)
	return err
}

// 触，触摸，按压，感受
type SeeSeekTouch struct {
	Id        int
	Uuid      string
	SeeSeekId int
	Classify  int

	Status int //0

	Temperature string //温度
	IsFever     bool   //是否异常发热
	Stretch     string //弹性
	IsStiff     bool   //是否僵硬
	Shake       string //震动
	IsShake     bool   //是否震动过大

	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (sst *SeeSeekTouch) Create() (err error) {
	statement := `INSERT INTO see_seek_touches 
		(uuid, see_seek_id, classify, status, temperature, is_fever, stretch, is_stiff, shake, is_shake, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		RETURNING id, uuid`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), sst.SeeSeekId, sst.Classify, sst.Status, sst.Temperature, sst.IsFever, sst.Stretch, sst.IsStiff, sst.Shake, sst.IsShake, time.Now()).Scan(&sst.Id, &sst.Uuid)
	return err
}

// SeeSeekTouch.DeleteBySeeSeekId()
func DeleteSeeSeekTouchesBySeeSeekId(seeSeekId int) error {
	statement := `DELETE FROM see_seek_touches WHERE see_seek_id=$1`
	_, err := db.Exec(statement, seeSeekId)
	return err
}

// 根据project_id查找SeeSeek记录
func GetSeeSeekByProjectId(projectId int, ctx context.Context) (SeeSeek, error) {
	//cancel after 5 seconds
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var s SeeSeek
	statement := `SELECT id, uuid, name, nickname, description, place_id, project_id, 
		payer_user_id, payer_team_id, payer_family_id, payee_user_id, payee_team_id, payee_family_id,
		verifier_user_id, verifier_team_id, verifier_family_id, category, status, step, created_at, updated_at
		FROM see_seeks WHERE project_id = $1 ORDER BY created_at DESC LIMIT 1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return s, err
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, projectId).Scan(&s.Id, &s.Uuid, &s.Name, &s.Nickname, &s.Description,
		&s.PlaceId, &s.ProjectId, &s.PayerUserId, &s.PayerTeamId, &s.PayerFamilyId,
		&s.PayeeUserId, &s.PayeeTeamId, &s.PayeeFamilyId, &s.VerifierUserId,
		&s.VerifierTeamId, &s.VerifierFamilyId, &s.Category, &s.Status, &s.Step, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		// 没有找到记录，返回明确空记录错误信息
		s.Id = 0
		return s, err
	} else if err != nil {
		return s, err // 发生其他错误
	}

	return s, err
}

// SeeSeek.Update() 更新SeeSeek记录
func (s *SeeSeek) Update() error {
	statement := `UPDATE see_seeks SET name = $2, nickname = $3, description = $4, 
		status = $5, step = $6, updated_at = $7 WHERE id = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(s.Id, s.Name, s.Nickname, s.Description, s.Status, s.Step, time.Now())
	return err
}

// 获取SeeSeek的环境记录
func (s *SeeSeek) GetEnvironments() ([]SeeSeekEnvironment, error) {
	rows, err := db.Query("SELECT id, uuid, see_seek_id, environment_id, created_at, updated_at FROM see_seek_environments WHERE see_seek_id = $1", s.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envs []SeeSeekEnvironment
	for rows.Next() {
		var env SeeSeekEnvironment
		err := rows.Scan(&env.Id, &env.Uuid, &env.SeeSeekId, &env.EnvironmentId, &env.CreatedAt, &env.UpdatedAt)
		if err != nil {
			return nil, err
		}
		envs = append(envs, env)
	}
	return envs, nil
}

// 获取SeeSeek的隐患记录
func (s *SeeSeek) GetHazards() ([]SeeSeekHazard, error) {
	rows, err := db.Query("SELECT id, uuid, see_seek_id, hazard_id, created_at, updated_at FROM see_seek_hazards WHERE see_seek_id = $1", s.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hazards []SeeSeekHazard
	for rows.Next() {
		var hazard SeeSeekHazard
		err := rows.Scan(&hazard.Id, &hazard.Uuid, &hazard.SeeSeekId, &hazard.HazardId, &hazard.CreatedAt, &hazard.UpdatedAt)
		if err != nil {
			return nil, err
		}
		hazards = append(hazards, hazard)
	}
	return hazards, nil
}

// 获取SeeSeek的风险记录
func (s *SeeSeek) GetRisks() ([]SeeSeekRisk, error) {
	rows, err := db.Query("SELECT id, uuid, see_seek_id, risk_id, created_at, updated_at FROM see_seek_risks WHERE see_seek_id = $1", s.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var risks []SeeSeekRisk
	for rows.Next() {
		var risk SeeSeekRisk
		err := rows.Scan(&risk.Id, &risk.Uuid, &risk.SeeSeekId, &risk.RiskId, &risk.CreatedAt, &risk.UpdatedAt)
		if err != nil {
			return nil, err
		}
		risks = append(risks, risk)
	}
	return risks, nil
}

// 获取SeeSeek的视觉观察记录
func (s *SeeSeek) GetLooks() ([]SeeSeekLook, error) {
	rows, err := db.Query("SELECT id, uuid, see_seek_id, classify, status, outline, is_deform, skin, is_graze, color, is_change, created_at, updated_at FROM see_seek_looks WHERE see_seek_id = $1", s.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var looks []SeeSeekLook
	for rows.Next() {
		var look SeeSeekLook
		err := rows.Scan(&look.Id, &look.Uuid, &look.SeeSeekId, &look.Classify, &look.Status, &look.Outline, &look.IsDeform, &look.Skin, &look.IsGraze, &look.Color, &look.IsChange, &look.CreatedAt, &look.UpdatedAt)
		if err != nil {
			return nil, err
		}
		looks = append(looks, look)
	}
	return looks, nil
}

// 获取SeeSeek的听觉观察记录
func (s *SeeSeek) GetListens() ([]SeeSeekListen, error) {
	rows, err := db.Query("SELECT id, uuid, see_seek_id, classify, status, sound, is_abnormal, created_at, updated_at FROM see_seek_listens WHERE see_seek_id = $1", s.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listens []SeeSeekListen
	for rows.Next() {
		var listen SeeSeekListen
		err := rows.Scan(&listen.Id, &listen.Uuid, &listen.SeeSeekId, &listen.Classify, &listen.Status, &listen.Sound, &listen.IsAbnormal, &listen.CreatedAt, &listen.UpdatedAt)
		if err != nil {
			return nil, err
		}
		listens = append(listens, listen)
	}
	return listens, nil
}

// 获取SeeSeek的嗅觉观察记录
func (s *SeeSeek) GetSmells() ([]SeeSeekSmell, error) {
	rows, err := db.Query("SELECT id, uuid, see_seek_id, classify, status, odour, is_foul_odour, created_at, updated_at FROM see_seek_smells WHERE see_seek_id = $1", s.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var smells []SeeSeekSmell
	for rows.Next() {
		var smell SeeSeekSmell
		err := rows.Scan(&smell.Id, &smell.Uuid, &smell.SeeSeekId, &smell.Classify, &smell.Status, &smell.Odour, &smell.IsFoulOdour, &smell.CreatedAt, &smell.UpdatedAt)
		if err != nil {
			return nil, err
		}
		smells = append(smells, smell)
	}
	return smells, nil
}

// 获取SeeSeek的触觉观察记录
func (s *SeeSeek) GetTouches() ([]SeeSeekTouch, error) {
	rows, err := db.Query("SELECT id, uuid, see_seek_id, classify, status, temperature, is_fever, stretch, is_stiff, shake, is_shake, created_at, updated_at FROM see_seek_touches WHERE see_seek_id = $1", s.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var touches []SeeSeekTouch
	for rows.Next() {
		var touch SeeSeekTouch
		err := rows.Scan(&touch.Id, &touch.Uuid, &touch.SeeSeekId, &touch.Classify, &touch.Status, &touch.Temperature, &touch.IsFever, &touch.Stretch, &touch.IsStiff, &touch.Shake, &touch.IsShake, &touch.CreatedAt, &touch.UpdatedAt)
		if err != nil {
			return nil, err
		}
		touches = append(touches, touch)
	}
	return touches, nil
}

// 获取SeeSeek的检测报告记录
func (s *SeeSeek) GetExaminationReports() ([]SeeSeekExaminationReport, error) {
	rows, err := db.Query("SELECT id, uuid, see_seek_id, classify, status, name, nickname, description, sample_type, sample_order, instrument_goods_id, report_title, report_content, master_user_id, reviewer_user_id, report_date, attachment, tags, created_at, updated_at FROM see_seek_examination_reports WHERE see_seek_id = $1", s.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []SeeSeekExaminationReport
	for rows.Next() {
		var report SeeSeekExaminationReport
		err := rows.Scan(&report.ID, &report.Uuid, &report.SeeSeekID, &report.Classify, &report.Status, &report.Name, &report.Nickname, &report.Description, &report.SampleType, &report.SampleOrder, &report.InstrumentGoodsID, &report.ReportTitle, &report.ReportContent, &report.MasterUserId, &report.ReviewerUserId, &report.ReportDate, &report.Attachment, &report.Tags, &report.CreatedAt, &report.UpdatedAt)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	return reports, nil
}
