package data

import "time"

// 手工艺，技能操作，需要集中注意力身体手眼协调平衡配合完成的动作。
// 例如：书法，剪发，和面团，抹墙灰，拧螺丝，雕刻，补牙洞，攀爬作业...
type Handicraft struct {
	Id              int
	Uuid            string
	ProjectId       int // 茶台ID，项目，
	Name            string
	Nickname        string
	ClientTeamId    int    // 需求（甲方）客户,需求团队ID，客户必须是一个组织（team），如果是个体人，则需要注册为单身家庭team（如果这个工程出现单身人身亡事故赔偿，谁是第一受益人？）
	TargetGoodsId   int    // 作业靶子商品Id,例子，1或n首写在白纸上的命题古诗，白纸是作业目标，手工艺内容是在纸上留下美丽的墨迹。如果写了多首，每一份诗（可交易标的物）都可以是一个手艺成品。
	GoodsListId     int    // 消耗品，材料，（货物）商品清单单号
	ToolListId      int    // 装备或工具（短期租赁的商品）清单单号，完成这个部分作业，可能需要多个工具（装备），例如写一首古诗，需要毛笔、墨、纸、砚台和水，书桌等
	Artist          int    // 手艺人ID
	Strength        int    // 体力耗费等级(1-5)
	Intelligence    int    // 智力耗费等级(1-5)Mental effort level required
	DifficultyLevel int    // 掌握工艺的学习课程难度等级(1-5)
	Recorder        int    // 记录人id，ID of the person recording the handicraft details
	Description     string // 手工艺总览，任务综合描述。例如，在绢纸上用毛笔（沾墨）创建一首格律诗词。
	Category        int    // 类型，0:日常普通作业，1:非物质文化遗产？
	Status          int    // 状态
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type StrengthLevel int

const (
	VeryLowStrength  StrengthLevel = 1
	LowStrength      StrengthLevel = 2
	ModerateStrength StrengthLevel = 3
	HighStrength     StrengthLevel = 4
	VeryHighStrength StrengthLevel = 5
)

// handicraft.Create()
func (hc *Handicraft) Create() (err error) {
	statement := "INSERT INTO handicrafts (uuid, project_id, name, nickname, client_team_id, target_goods_id, goods_list_id, tool_list_id, artist, strength, intelligence, difficulty_level, recorder, description, category, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), hc.ProjectId, hc.Name, hc.Nickname, hc.ClientTeamId, hc.TargetGoodsId, hc.GoodsListId, hc.ToolListId, hc.Artist, hc.Strength, hc.Intelligence, hc.DifficultyLevel, hc.Recorder, hc.Description, hc.Category, hc.Status, time.Now(), time.Now()).Scan(&hc.Id)
	return
}

// handicraft.Get()
func (h *Handicraft) Get() (err error) {
	err = Db.QueryRow("SELECT id, uuid, project_id, name, nickname, client_team_id, target_goods_id, goods_list_id, tool_list_id, artist, strength, intelligence, difficulty_level, recorder, description, category, status, created_at, updated_at FROM handicrafts WHERE id = $1", h.Id).
		Scan(&h.Id, &h.Uuid, &h.ProjectId, &h.Name, &h.Nickname, &h.ClientTeamId, &h.TargetGoodsId, &h.GoodsListId, &h.ToolListId, &h.Artist, &h.Strength, &h.Intelligence, &h.DifficultyLevel, &h.Recorder, &h.Description, &h.Category, &h.Status, &h.CreatedAt, &h.UpdatedAt)
	return
}

// handicraft.Update()
func (h *Handicraft) Update() (err error) {
	statement := "UPDATE handicrafts SET project_id = $1, name = $2, nickname = $3, client = $4, target_goods_id = $5, tool_list_id = $6, artist = $7, strength = $8, intelligence = $9, difficulty_level = $10, recorder = $11, description = $12, category = $13, status = $14, updated_at = $15 WHERE id = $16"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(h.ProjectId, h.Name, h.Nickname, h.ClientTeamId, h.TargetGoodsId, h.GoodsListId, h.ToolListId, h.Artist, h.Strength, h.Intelligence, h.DifficultyLevel, h.Recorder, h.Description, h.Category, h.Status, time.Now(), h.Id)
	return
}

// handicraft.Delete()
func (handicraft *Handicraft) Delete() (err error) {
	statement := "DELETE FROM handicrafts WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(handicraft.Id)
	return
}

// handicraft.GetHandicraftByProjectId()
func GetHandicraftsByProjectId(projectId int) (handicrafts []Handicraft, err error) {
	rows, err := Db.Query("SELECT id, uuid, project_id, name, nickname, client_team_id, target_goods_id, goods_list_id, tool_list_id, artist, strength, intelligence, difficulty_level, recorder, description, category, status, created_at, updated_at FROM handicrafts WHERE project_id = $1", projectId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		handicraft := Handicraft{}
		err = rows.Scan(&handicraft.Id, &handicraft.Uuid, &handicraft.ProjectId, &handicraft.Name, &handicraft.Nickname, &handicraft.ClientTeamId, &handicraft.TargetGoodsId, &handicraft.ToolListId, &handicraft.Artist, &handicraft.Strength, &handicraft.Intelligence, &handicraft.DifficultyLevel, &handicraft.Recorder, &handicraft.Description, &handicraft.Category, &handicraft.Status, &handicraft.CreatedAt, &handicraft.UpdatedAt)
		if err != nil {
			return
		}
		handicrafts = append(handicrafts, handicraft)
	}
	return
}

// handicraft.GetHandicraftByUuid()
func GetHandicraftByUuid(uuid string) (handicraft Handicraft, err error) {
	handicraft = Handicraft{}
	err = Db.QueryRow("SELECT id, uuid, project_id, name, nickname, client_team_id, target_goods_id, goods_list_id, tool_list_id, artist, strength, intelligence, difficulty_level, recorder, description, category, status, created_at, updated_at FROM handicrafts WHERE uuid = $1", uuid).
		Scan(&handicraft.Id, &handicraft.Uuid, &handicraft.ProjectId, &handicraft.Name, &handicraft.Nickname, &handicraft.ClientTeamId, &handicraft.TargetGoodsId, &handicraft.ToolListId, &handicraft.Artist, &handicraft.Strength, &handicraft.Intelligence, &handicraft.DifficultyLevel, &handicraft.Recorder, &handicraft.Description, &handicraft.Category, &handicraft.Status, &handicraft.CreatedAt, &handicraft.UpdatedAt)
	return
}

// handicraft.GetHandicraftByTargetGoodsId()
func GetHandicraftByTargetGoodsId(targetGoodsId int) (handicraft Handicraft, err error) {
	handicraft = Handicraft{}
	err = Db.QueryRow("SELECT id, uuid, project_id, name, nickname, client_team_id, target_goods_id, goods_list_id, tool_list_id, artist, strength, intelligence, difficulty_level, recorder, description, category, status, created_at, updated_at FROM handicrafts WHERE target_goods_id = $1", targetGoodsId).
		Scan(&handicraft.Id, &handicraft.Uuid, &handicraft.ProjectId, &handicraft.Name, &handicraft.Nickname, &handicraft.ClientTeamId, &handicraft.TargetGoodsId, &handicraft.ToolListId, &handicraft.Artist, &handicraft.Strength, &handicraft.Intelligence, &handicraft.DifficultyLevel, &handicraft.Recorder, &handicraft.Description, &handicraft.Category, &handicraft.Status, &handicraft.CreatedAt, &handicraft.UpdatedAt)
	return
}

// handicraft.GetHandicraftsByArtist()
func GetHandicraftsByArtist(artist int) (handicrafts []Handicraft, err error) {
	rows, err := Db.Query("SELECT id, uuid, project_id, name, nickname, client_team_id, target_goods_id, goods_list_id, tool_list_id, artist, strength, intelligence, difficulty_level, recorder, description, category, status, created_at, updated_at FROM handicrafts WHERE artist = $1", artist)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		handicraft := Handicraft{}
		err = rows.Scan(&handicraft.Id, &handicraft.Uuid, &handicraft.ProjectId, &handicraft.Name, &handicraft.Nickname, &handicraft.ClientTeamId, &handicraft.TargetGoodsId, &handicraft.ToolListId, &handicraft.Artist, &handicraft.Strength, &handicraft.Intelligence, &handicraft.DifficultyLevel, &handicraft.Recorder, &handicraft.Description, &handicraft.Category, &handicraft.Status, &handicraft.CreatedAt, &handicraft.UpdatedAt)
		if err != nil {
			return
		}
		handicrafts = append(handicrafts, handicraft)
	}
	return
}

// handicraft.GetHandicraftsByRecorder()
func GetHandicraftsByRecorder(recorder int) (handicrafts []Handicraft, err error) {
	rows, err := Db.Query("SELECT id, uuid, project_id, name, nickname, client_team_id, target_goods_id, goods_list_id, tool_list_id, artist, strength, intelligence, difficulty_level, recorder, description, category, status, created_at, updated_at FROM handicrafts WHERE recorder = $1", recorder)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		handicraft := Handicraft{}
		err = rows.Scan(&handicraft.Id, &handicraft.Uuid, &handicraft.ProjectId, &handicraft.Name, &handicraft.Nickname, &handicraft.ClientTeamId, &handicraft.TargetGoodsId, &handicraft.ToolListId, &handicraft.Artist, &handicraft.Strength, &handicraft.Intelligence, &handicraft.DifficultyLevel, &handicraft.Recorder, &handicraft.Description, &handicraft.Category, &handicraft.Status, &handicraft.CreatedAt, &handicraft.UpdatedAt)
		if err != nil {
			return
		}
		handicrafts = append(handicrafts, handicraft)
	}
	return
}

// handicraft.GetHandicraftsByStatus()
func GetHandicraftsByStatus(status int) (handicrafts []Handicraft, err error) {
	rows, err := Db.Query("SELECT id, uuid, project_id, name, nickname, client_team_id, target_goods_id, goods_list_id, tool_list_id, artist, strength, intelligence, difficulty_level, recorder, description, category, status, created_at, updated_at FROM handicrafts WHERE status = $1", status)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		handicraft := Handicraft{}
		err = rows.Scan(&handicraft.Id, &handicraft.Uuid, &handicraft.ProjectId, &handicraft.Name, &handicraft.Nickname, &handicraft.ClientTeamId, &handicraft.TargetGoodsId, &handicraft.ToolListId, &handicraft.Artist, &handicraft.Strength, &handicraft.Intelligence, &handicraft.DifficultyLevel, &handicraft.Recorder, &handicraft.Description, &handicraft.Category, &handicraft.Status, &handicraft.CreatedAt, &handicraft.UpdatedAt)
		if err != nil {
			return
		}
		handicrafts = append(handicrafts, handicraft)
	}
	return
}

// 手工艺开工仪式，到岗准备开工。例如，书法的起手式，准备动手前一刻的快照
type Inauguration struct {
	Id           int
	Uuid         string
	HandicraftId int // 手工艺Id
	Name         string
	Nickname     string
	Artist       int    // 手艺人Id。如果是集团，则填first_team_id作为代表。例如，贾宝玉和他的女仆组成一个作古诗小组，如果一个人自己完成，则为单人成员组。
	Recorder     int    // 记录人id
	Description  string // 作业内容描述
	EvidenceId   int    // 音视频等视觉证据，默认值为 0，表示没有
	Status       int    // 状态
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// inaguration.Create()
func (inauguration *Inauguration) Create() (err error) {
	statement := "INSERT INTO inaugurations (uuid, handicraft_id, name, nickname, artist, recorder, description, category, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), inauguration.HandicraftId, inauguration.Name, inauguration.Nickname, inauguration.Artist, inauguration.Recorder, inauguration.Description, inauguration.EvidenceId, inauguration.Status, time.Now(), time.Now()).Scan(&inauguration.Id)
	return
}

// inauguration.Get()
func (inauguration *Inauguration) Get() (err error) {
	err = Db.QueryRow("SELECT id, uuid, handicraft_id, name, nickname, artist, recorder, description, category, status, created_at, updated_at FROM inaugurations WHERE id = $1", inauguration.Id).
		Scan(&inauguration.Id, &inauguration.Uuid, &inauguration.HandicraftId, &inauguration.Name, &inauguration.Nickname, &inauguration.Artist, &inauguration.Recorder, &inauguration.Description, &inauguration.EvidenceId, &inauguration.Status, &inauguration.CreatedAt, &inauguration.UpdatedAt)
	return
}

// inauguration.Update()
func (inauguration *Inauguration) Update() (err error) {
	statement := "UPDATE inaugurations SET handicraft_id = $1, name = $2, nickname = $3, artist = $4, recorder = $5, description = $6, category = $7, status = $8, updated_at = $9 WHERE id = $10"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(inauguration.HandicraftId, inauguration.Name, inauguration.Nickname, inauguration.Artist, inauguration.Recorder, inauguration.Description, inauguration.EvidenceId, inauguration.Status, time.Now(), inauguration.Id)
	return
}

// inauguration.Delete()
func (inauguration *Inauguration) Delete() (err error) {
	statement := "DELETE FROM inaugurations WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(inauguration.Id)
	return
}

// inauguration.GetInaugurationByHandicraftId()
func GetInaugurationByHandicraftId(handicraftId int) (inauguration Inauguration, err error) {
	inauguration = Inauguration{}
	err = Db.QueryRow("SELECT id, uuid, handicraft_id, name, nickname, artist, recorder, description, category, status, created_at, updated_at FROM inaugurations WHERE handicraft_id = $1", handicraftId).
		Scan(&inauguration.Id, &inauguration.Uuid, &inauguration.HandicraftId, &inauguration.Name, &inauguration.Nickname, &inauguration.Artist, &inauguration.Recorder, &inauguration.Description, &inauguration.EvidenceId, &inauguration.Status, &inauguration.CreatedAt, &inauguration.UpdatedAt)
	return
}

// inauguration.GetInaugurationByUuid()
func GetInaugurationByUuid(uuid string) (inauguration Inauguration, err error) {
	inauguration = Inauguration{}
	err = Db.QueryRow("SELECT id, uuid, handicraft_id, name, nickname, artist, recorder, description, category, status, created_at, updated_at FROM inaugurations WHERE uuid = $1", uuid).
		Scan(&inauguration.Id, &inauguration.Uuid, &inauguration.HandicraftId, &inauguration.Name, &inauguration.Nickname, &inauguration.Artist, &inauguration.Recorder, &inauguration.Description, &inauguration.EvidenceId, &inauguration.Status, &inauguration.CreatedAt, &inauguration.UpdatedAt)
	return
}

// 环节，部分，整个手工艺操作过程的中间一部分，所有环节加起来等于整个手工艺流程。
type Part struct {
	Id              int
	Uuid            string
	HandicraftId    int // 手工艺Id
	Name            string
	Nickname        string
	Artist          int    // 完成这部分作业的手艺人id。每一个人的操作就是一环节，1 part。例如，贾宝玉写下了一首或者几首诗都可以视作一个环节，
	TargetGoodsId   int    // 作业靶子商品Id。例子，假设墨水已经被女仆晴雯磨合准备好，盛在砚台中了，这一部分艺术家贾宝玉的手工艺作业是往毛笔上“沾墨”，那么操作靶子是砚台。
	ToolListId      int    // 完成这个部分作业装备或工具商品Id集合。可能需要多个工具（装备），例如，在写古诗上墨部分艺术家需要毛笔（工具）
	Strength        int    // 体力耗费等级(1-5)
	Intelligence    int    // 智力耗费等级(1-5)
	DifficultyLevel int    // 掌握难度等级(1-5)
	Recorder        int    // 记录人id
	Description     string // 作业内容描述
	EvidenceId      int    // 音视频等视觉证据，默认值为 0，表示没有
	Status          int    // 状态
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
type Tool struct {
	Id           int
	Uuid         string
	HandicraftId int
	PartId       int
	GoodsId      int
	Note         string //备注,特别说明
	Category     int    //类型
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// tool.Create()
func (tool *Tool) Create() (err error) {
	statement := "INSERT INTO tools (uuid, handicraft_id, part_id, goods_id, note, category, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), tool.HandicraftId, tool.PartId, tool.GoodsId, tool.Note, tool.Category, time.Now(), time.Now()).Scan(&tool.Id)
	return
}

// tool.Get()
func (tool *Tool) Get() (err error) {
	err = Db.QueryRow("SELECT id, uuid, handicraft_id, part_id, goods_id, note, category, created_at, updated_at FROM tools WHERE id = $1", tool.Id).
		Scan(&tool.Id, &tool.Uuid, &tool.HandicraftId, &tool.PartId, &tool.GoodsId, &tool.Note, &tool.Category, &tool.CreatedAt, &tool.UpdatedAt)
	return
}

// tool.Update()
func (tool *Tool) Update() (err error) {
	statement := "UPDATE tools SET handicraft_id = $1, part_id = $2, goods_id = $3, note = $4, category = $5, updated_at = $6 WHERE id = $7"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(tool.HandicraftId, tool.PartId, tool.GoodsId, tool.Note, tool.Category, time.Now(), tool.Id)
	return
}

// part.Create()
func (part *Part) Create() (err error) {
	statement := "INSERT INTO parts (uuid, handicraft_id, name, nickname, artist, target_goods_id, goods_list_id, tool_list_id, strength, intelligence, difficulty_level, recorder, description, category, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), part.HandicraftId, part.Name, part.Nickname, part.Artist, part.TargetGoodsId, part.ToolListId, part.Strength, part.Intelligence, part.DifficultyLevel, part.Recorder, part.Description, part.EvidenceId, part.Status, time.Now(), time.Now()).Scan(&part.Id)
	return
}

// part.Get()
func (part *Part) Get() (err error) {
	err = Db.QueryRow("SELECT id, uuid, handicraft_id, name, nickname, artist, target_goods_id, goods_list_id, tool_list_id, strength, intelligence, difficulty_level, recorder, description, category, status, created_at, updated_at FROM parts WHERE id = $1", part.Id).
		Scan(&part.Id, &part.Uuid, &part.HandicraftId, &part.Name, &part.Nickname, &part.Artist, &part.TargetGoodsId, &part.ToolListId, &part.Strength, &part.Intelligence, &part.DifficultyLevel, &part.Recorder, &part.Description, &part.EvidenceId, &part.Status, &part.CreatedAt, &part.UpdatedAt)
	return
}

// part.Update()
func (part *Part) Update() (err error) {
	statement := "UPDATE parts SET handicraft_id = $1, name = $2, nickname = $3, artist = $4, recorder = $10, description = $11, category = $12, status = $13, updated_at = $14 WHERE id = $15"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(part.HandicraftId, part.Name, part.Nickname, part.Artist, part.TargetGoodsId, part.ToolListId, part.Strength, part.Intelligence, part.DifficultyLevel, part.Recorder, part.Description, part.EvidenceId, part.Status, time.Now(), part.Id)
	return
}

// part.Delete()
func (part *Part) Delete() (err error) {
	statement := "DELETE FROM parts WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(part.Id)
	return
}

// part.GetPartByHandicraftId()
func GetPartByHandicraftId(handicraftId int) (part Part, err error) {
	part = Part{}
	err = Db.QueryRow("SELECT id, uuid, handicraft_id, name, nickname, artist, target_goods_id, goods_list_id, tool_list_id, strength, intelligence, difficulty_level, recorder, description, category, status, created_at, updated_at FROM parts WHERE handicraft_id = $1", handicraftId).
		Scan(&part.Id, &part.Uuid, &part.HandicraftId, &part.Name, &part.Nickname, &part.Artist, &part.TargetGoodsId, &part.ToolListId, &part.Strength, &part.Intelligence, &part.DifficultyLevel, &part.Recorder, &part.Description, &part.EvidenceId, &part.Status, &part.CreatedAt, &part.UpdatedAt)
	return
}

// part.GetPartByUuid()
func GetPartByUuid(uuid string) (part Part, err error) {
	part = Part{}
	err = Db.QueryRow("SELECT id, uuid, handicraft_id, name, nickname, artist, target_goods_id, goods_list_id, tool_list_id, strength, intelligence, difficulty_level, recorder, description, category, status, created_at, updated_at FROM parts WHERE uuid = $1", uuid).
		Scan(&part.Id, &part.Uuid, &part.HandicraftId, &part.Name, &part.Nickname, &part.Artist, &part.TargetGoodsId, &part.ToolListId, &part.Strength, &part.Intelligence, &part.DifficultyLevel, &part.Recorder, &part.Description, &part.EvidenceId, &part.Status, &part.CreatedAt, &part.UpdatedAt)
	return
}

// part.GetPartsCountByHandicraftId()
func GetPartsCountByHandicraftId(handicraftId int) (count int, err error) {
	err = Db.QueryRow("SELECT COUNT(*) FROM parts WHERE handicraft_id = $1", handicraftId).Scan(&count)
	return
}

// 收尾，手工艺作业结束仪式，离手（场）快照。
type Ending struct {
	Id           int
	Uuid         string
	HandicraftId int // 手工艺Id
	Name         string
	Nickname     string
	Artist       int    // 完成这部分作业的手艺人id。每一个人的操作就是一环节，1 part。例如，贾宝玉写下了一首或者几首诗都可以视作一个环节，
	Recorder     int    // 记录人id
	Description  string // 作业内容，成就快照描述
	EvidenceId   int    // 默认值为 0，表示没有
	Status       int    // 状态。0:失败作业，1:已完成作业，2:需要延期作业
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ending.Create() 创建一个新的ending
func (ending *Ending) Create() (err error) {
	statement := "INSERT INTO endings (uuid, handicraft_id, name, nickname, artist, recorder, description, category, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), ending.HandicraftId, ending.Name, ending.Nickname, ending.Artist, ending.Recorder, ending.Description, ending.EvidenceId, ending.Status, time.Now(), time.Now()).Scan(&ending.Id)
	return
}

// ending.Get() 通过Id获取一个ending
func (ending *Ending) Get() (err error) {
	err = Db.QueryRow("SELECT id, uuid, handicraft_id, name, nickname, artist, recorder, description, category, status, created_at, updated_at FROM endings WHERE id = $1", ending.Id).
		Scan(&ending.Id, &ending.Uuid, &ending.HandicraftId, &ending.Name, &ending.Nickname, &ending.Artist, &ending.Recorder, &ending.Description, &ending.EvidenceId, &ending.Status, &ending.CreatedAt, &ending.UpdatedAt)
	return
}

// ending.Update() 更新一个ending
func (ending *Ending) Update() (err error) {
	statement := "UPDATE endings SET handicraft_id = $1, name = $2, nickname = $3, artist = $4, recorder = $5, description = $6, category=$7 , status = $8, updated_at = $9 WHERE id = $10"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(ending.HandicraftId, ending.Name, ending.Nickname, ending.Artist, ending.Recorder, ending.Description, ending.EvidenceId, ending.Status, time.Now(), ending.Id)
	return
}

// ending.Delete() 删除一个ending
func (ending *Ending) Delete() (err error) {
	statement := "DELETE FROM endings WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(ending.Id)
	return
}

// ending.GetEndingByHandicraftId() 通过手工艺Id获取1 ending
func GetEndingsByHandicraftId(handicraftId int) (ending Ending, err error) {
	ending = Ending{}
	err = Db.QueryRow("SELECT id, uuid, handicraft_id, name, nickname, artist, recorder, description, category, status, created_at, updated_at FROM endings WHERE handicraft_id = $1", handicraftId).
		Scan(&ending.Id, &ending.Uuid, &ending.HandicraftId, &ending.Name, &ending.Nickname, &ending.Artist, &ending.Recorder, &ending.Description, &ending.EvidenceId, &ending.Status, &ending.CreatedAt, &ending.UpdatedAt)
	return
}

// 证据，依据，指音视频等视觉证据，证明手工艺作业符合描述的资料,
// 最好能反映成就。或者人力消耗、工具的折旧情况。
type Evidence struct {
	Id           int
	Uuid         string
	HandicraftId int // 标记属于那一个手工艺，
	Recorder     int
	Description  string
	Images       string // 图片(可选)
	Video        string // 视频(可选)
	Audio        string // 音频(可选)
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// evidence.Create() 创建一个新的evidence
func (evidence *Evidence) Create() (err error) {
	statement := "INSERT INTO evidences (uuid, handicraft_id, recorder, description, images, video, audio, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), evidence.HandicraftId, evidence.Recorder, evidence.Description, evidence.Images, evidence.Video, evidence.Audio, time.Now(), time.Now()).Scan(&evidence.Id)
	return
}

// evidence.Get() 通过Id获取一个evidence
func (evidence *Evidence) Get() (err error) {
	err = Db.QueryRow("SELECT id, uuid, handicraft_id, recorder, description, images, video, audio, created_at, updated_at FROM evidences WHERE id = $1", evidence.Id).
		Scan(&evidence.Id, &evidence.Uuid, &evidence.HandicraftId, &evidence.Recorder, &evidence.Description, &evidence.Images, &evidence.Video, &evidence.Audio, &evidence.CreatedAt, &evidence.UpdatedAt)
	return
}

// evidence.Update() 更新一个evidence
func (evidence *Evidence) Update() (err error) {
	statement := "UPDATE evidences SET handicraft_id = $1, recorder = $2, description = $3, images = $4, video = $5, audio = $6, updated_at = $7 WHERE id = $8"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(evidence.HandicraftId, evidence.Recorder, evidence.Description, evidence.Images, evidence.Video, evidence.Audio, time.Now(), evidence.Id)
	return
}

// evidence.Delete() 删除一个evidence
func (evidence *Evidence) Delete() (err error) {
	statement := "DELETE FROM evidences WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(evidence.Id)
	return
}

// evidence.GetEvidencesByHandicraftId() 通过手工艺Id获取全部 evidences
func GetEvidencesByHandicraftId(handicraftId int) (evidences []Evidence, err error) {
	rows, err := Db.Query("SELECT id, uuid, handicraft_id, recorder, description, images, video, audio, created_at, updated_at FROM evidences WHERE handicraft_id = $1", handicraftId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		evidence := Evidence{}
		err = rows.Scan(&evidence.Id, &evidence.Uuid, &evidence.HandicraftId, &evidence.Recorder, &evidence.Description, &evidence.Images, &evidence.Video, &evidence.Audio, &evidence.CreatedAt, &evidence.UpdatedAt)
		if err != nil {
			return
		}
		evidences = append(evidences, evidence)
	}
	err = rows.Err()
	return
}
