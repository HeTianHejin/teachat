package data

import "time"

// 社区？ 寓意村落/镇/家族/大企业生活区...等等的人们生活聚积区域
// 谁来负责编辑其内容？实际控制团队的ceo核心成员，
// 问题：社区历史是由谁来写的？
type Community struct {
	Id              int
	Uuid            string
	Name            string
	Introduction    string
	FamilyIdSet     []int // 构成社区生活的家庭Id集合
	InfluenceTeamId int   // 主要的影响（驱动）社区生活形式思路变化的团体id，实际控制长老会或者董事会
	EditedUserIdSet []int // 编辑过的用户id集合
	StateIndex      int   // 状态指数
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Flag            string // 旗帜图片
}

// Save() 保存一个社区/部落/大家族...
func (c *Community) Save() error {
	statement := "INSERT INTO communities (uuid, name, introduction, family_id_set, influence_team_id, edited_user_id_set, state_index, created_at, updated_at, flag) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(CreateUuid(), c.Name, c.Introduction, c.FamilyIdSet, c.InfluenceTeamId, c.EditedUserIdSet, c.StateIndex, time.Now(), time.Now(), c.Flag)
	if err != nil {
		return err
	}
	return nil
}

// Update() 更新一个社区/部落/大家族...资料，根据其ID
func (c *Community) Update() error {
	statement := "UPDATE communities SET name=$1, introduction=$2, family_id_set=$3, influence_team_id=$4, edited_user_id_set=$5, state_index=$6, updated_at=$7 WHERE id=$8"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(c.Name, c.Introduction, c.FamilyIdSet, c.InfluenceTeamId, c.EditedUserIdSet, c.StateIndex, time.Now(), c.Id)
	if err != nil {
		return err
	}
	return nil
}

// UpdateFlag() 更新一个社区/部落/大家族...的标志旗帜图片名，根据其ID
func (c *Community) UpdateFlag() error {
	statement := "UPDATE communities SET flag=$1 WHERE id=$2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(c.Flag, c.Id)
	if err != nil {
		return err
	}
	return nil
}

// Get() 获取一个社区/部落/大家族...资料根据其ID
func (c *Community) Get() error {
	statement := "SELECT id, uuid, name, introduction, family_id_set, influence_team_id, edited_user_id_set, state_index, created_at, updated_at, flag FROM communities WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(c.Id).Scan(&c.Id, &c.Uuid, &c.Name, &c.Introduction, &c.FamilyIdSet, &c.InfluenceTeamId, &c.EditedUserIdSet, &c.StateIndex, &c.CreatedAt, &c.UpdatedAt, &c.Flag)
	if err != nil {
		return err
	}
	return nil
}
