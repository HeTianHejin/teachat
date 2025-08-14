package data

import (
	"context"
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

	Outline  string //外形轮廓
	IsDeform bool   //是否变形
	Skin     string //表面皮肤
	IsGraze  bool   //是否破损
	Color    string //颜色
	IsChange bool   //是否变色

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

	Sound      string //声音
	IsAbnormal bool   //是否有异常声音

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

	Odour       string //气味
	IsFoulOdour bool   //是否异味

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

	Temperature string //温度
	IsFever     bool   //是否异常发热
	Stretch     string //弹性
	IsStiff     bool   //是否僵硬
	Shake       string //震动
	IsShake     bool   //是否震动

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

	Ask    string
	Answer string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (ssaa *SeeSeekAskAndAnswer) Create() (err error) {
	statement := `INSERT INTO see_seek_ask_and_answers 
		(uuid, see_seek_id, classify, status, ask, answer, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, uuid`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), ssaa.SeeSeekId, ssaa.Classify, ssaa.Status, ssaa.Ask, ssaa.Answer, time.Now()).Scan(&ssaa.Id, &ssaa.Uuid)
	return err
}
func (ssaa *SeeSeekAskAndAnswer) Get() (err error) {
	statement := `SELECT id, uuid, see_seek_id, classify, status, ask, answer, created_at, updated_at 
		FROM see_seek_ask_and_answers WHERE id=$1`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(ssaa.Id).Scan(&ssaa.Id, &ssaa.Uuid, &ssaa.SeeSeekId, &ssaa.Classify, &ssaa.Status, &ssaa.Ask, &ssaa.Answer, &ssaa.CreatedAt, &ssaa.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (ssaa *SeeSeekAskAndAnswer) GetByUuid() (err error) {
	statement := `SELECT id, uuid, see_seek_id, classify, status, ask, answer, created_at, updated_at 
		FROM see_seek_ask_and_answers WHERE uuid=$1`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(ssaa.Uuid).Scan(&ssaa.Id, &ssaa.Uuid, &ssaa.SeeSeekId, &ssaa.Classify, &ssaa.Status, &ssaa.Ask, &ssaa.Answer, &ssaa.CreatedAt, &ssaa.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (ssaa *SeeSeekAskAndAnswer) Update() (err error) {
	statement := `UPDATE see_seek_ask_and_answers 
		SET see_seek_id=$2, classify=$3, status=$4, ask=$5, answer=$6, updated_at=$7 
		WHERE id=$1`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(ssaa.Id, ssaa.SeeSeekId, ssaa.Classify, ssaa.Status, ssaa.Ask, ssaa.Answer, time.Now())
	if err != nil {
		return
	}
	return
}
