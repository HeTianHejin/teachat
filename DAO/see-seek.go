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

//SeeSeek.Delete() // 删除一个SeeSeek

// 与特定人举行一个结构化交流
type SeeSeekMaster struct {
	Id        int
	Uuid      string
	SeeSeekId int
	Classify  int

	RecorderUserId int
	UserId         int
	Status         int

	RequestTitle   string
	RequestContent string
	RequestHistory string
	RequestRemark  string //特殊情况表述

	MasterTitle   string
	MasterContent string
	MasterHistory string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// SeeSeekMaster.Create() // 创建一个SeeSeekMaster
func (see_seek_master *SeeSeekMaster) Create() (err error) {
	statement := "INSERT INTO see_seek_masters (uuid, see_seek_id, classify, recorder_user_id, user_id, status, request_title, request_content, request_history, requst_remark, master_title, master_content, master_history, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), see_seek_master.SeeSeekId, see_seek_master.Classify, see_seek_master.RecorderUserId, see_seek_master.UserId, see_seek_master.Status, see_seek_master.RequestTitle, see_seek_master.RequestContent, see_seek_master.RequestHistory, see_seek_master.RequestRemark, see_seek_master.MasterTitle, see_seek_master.MasterContent, see_seek_master.MasterHistory, time.Now()).Scan(&see_seek_master.Id, &see_seek_master.Uuid)
	if err != nil {
		return
	}
	return
}

// SeeSeekMaster.Get() // 读取一个SeeSeekMaster
func (see_seek_master *SeeSeekMaster) Get() (err error) {
	statement := "SELECT id, uuid, see_seek_id, classify, recorder_user_id, user_id, status, request_title, request_content, request_history, requst_remark, master_title, master_content, master_history, created_at, updated_at FROM see_seek_masters WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_master.Id).Scan(&see_seek_master.Id, &see_seek_master.Uuid, &see_seek_master.SeeSeekId, &see_seek_master.Classify, &see_seek_master.RecorderUserId, &see_seek_master.UserId, &see_seek_master.Status, &see_seek_master.RequestTitle, &see_seek_master.RequestContent, &see_seek_master.RequestHistory, &see_seek_master.RequestRemark, &see_seek_master.MasterTitle, &see_seek_master.MasterContent, &see_seek_master.MasterHistory, &see_seek_master.CreatedAt, &see_seek_master.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// SeeSeekMaster.GetByUuid() // 读取一个SeeSeekMaster
func (see_seek_master *SeeSeekMaster) GetByUuid() (err error) {
	statement := "SELECT id, uuid, see_seek_id, classify, recorder_user_id, user_id, status, request_title, request_content, request_history, requst_remark, master_title, master_content, master_history, created_at, updated_at FROM see_seek_masters WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_master.Uuid).Scan(&see_seek_master.Id, &see_seek_master.Uuid, &see_seek_master.SeeSeekId, &see_seek_master.Classify, &see_seek_master.RecorderUserId, &see_seek_master.UserId, &see_seek_master.Status, &see_seek_master.RequestTitle, &see_seek_master.RequestContent, &see_seek_master.RequestHistory, &see_seek_master.RequestRemark, &see_seek_master.MasterTitle, &see_seek_master.MasterContent, &see_seek_master.MasterHistory, &see_seek_master.CreatedAt, &see_seek_master.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// SeeSeekMaster.Update() // 更新一个SeeSeekMaster
func (see_seek_master *SeeSeekMaster) Update() (err error) {
	statement := "UPDATE see_seek_masters SET see_seek_id=$2, classify=$3, recorder_user_id=$4, user_id=$5, status=$6, request_title=$7, request_content=$8, request_history=$9, requst_remark=$10, master_title=$11, master_content=$12, master_history=$13, updated_at=$14 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(see_seek_master.Id, see_seek_master.SeeSeekId, see_seek_master.Classify, see_seek_master.RecorderUserId, see_seek_master.UserId, see_seek_master.Status, see_seek_master.RequestTitle, see_seek_master.RequestContent, see_seek_master.RequestHistory, see_seek_master.RequestRemark, see_seek_master.MasterTitle, see_seek_master.MasterContent, see_seek_master.MasterHistory, time.Now())
	if err != nil {
		return
	}
	return
}

// 望，观察
type SeeSeekMasterLook struct {
	Id              int
	Uuid            string
	SeeSeekMasterId int
	Classify        int
	Status          int

	RequestOutline     string //外形轮廓
	IsDeform           bool
	RequestSkin        string //表面皮肤
	IsGraze            bool
	RequestColor       string //颜色
	IsChange           bool
	RequestLookHistory string

	MasterOutline     string
	MasterIsDeform    bool
	MasterSkin        string
	MasterIsGraze     bool
	MasterColor       string
	MasterIsChange    bool
	MasterLookHistory string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// SeeSeekMasterLook.Create() // 创建一个SeeSeekMasterLook
func (see_seek_master_look *SeeSeekMasterLook) Create() (err error) {
	statement := "INSERT INTO see_seek_master_looks (uuid, see_seek_master_id, classify, status, request_outline, is_deform, request_skin, is_graze, request_color, is_change, request_look_history, master_outline, master_is_deform, master_skin, master_is_graze, master_color, master_is_change, master_look_history, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), see_seek_master_look.SeeSeekMasterId, see_seek_master_look.Classify, see_seek_master_look.Status, see_seek_master_look.RequestOutline, see_seek_master_look.IsDeform, see_seek_master_look.RequestSkin, see_seek_master_look.IsGraze, see_seek_master_look.RequestColor, see_seek_master_look.IsChange, see_seek_master_look.RequestLookHistory, see_seek_master_look.MasterOutline, see_seek_master_look.MasterIsDeform, see_seek_master_look.MasterSkin, see_seek_master_look.MasterIsGraze, see_seek_master_look.MasterColor, see_seek_master_look.MasterIsChange, see_seek_master_look.MasterLookHistory, time.Now()).Scan(&see_seek_master_look.Id, &see_seek_master_look.Uuid)
	if err != nil {
		return
	}
	return
}

// SeeSeekMasterLook.Get() // 读取一个SeeSeekMasterLook
func (see_seek_master_look *SeeSeekMasterLook) Get() (err error) {
	statement := "SELECT id, uuid, see_seek_master_id, classify, status, request_outline, is_deform, request_skin, is_graze, request_color, is_change, request_look_history, master_outline, master_is_deform, master_skin, master_is_graze, master_color, master_is_change, master_look_history, created_at, updated_at FROM see_seek_master_looks WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_master_look.Id).Scan(&see_seek_master_look.Id, &see_seek_master_look.Uuid, &see_seek_master_look.SeeSeekMasterId, &see_seek_master_look.Classify, &see_seek_master_look.Status, &see_seek_master_look.RequestOutline, &see_seek_master_look.IsDeform, &see_seek_master_look.RequestSkin, &see_seek_master_look.IsGraze, &see_seek_master_look.RequestColor, &see_seek_master_look.IsChange, &see_seek_master_look.RequestLookHistory, &see_seek_master_look.MasterOutline, &see_seek_master_look.MasterIsDeform, &see_seek_master_look.MasterSkin, &see_seek_master_look.MasterIsGraze, &see_seek_master_look.MasterColor, &see_seek_master_look.MasterIsChange, &see_seek_master_look.MasterLookHistory, &see_seek_master_look.CreatedAt, &see_seek_master_look.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_look *SeeSeekMasterLook) GetByUuid() (err error) {
	statement := "SELECT id, uuid, see_seek_master_id, classify, status, request_outline, is_deform, request_skin, is_graze, request_color, is_change, request_look_history, master_outline, master_is_deform, master_skin, master_is_graze, master_color, master_is_change, master_look_history, created_at, updated_at FROM see_seek_master_looks WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_master_look.Uuid).Scan(&see_seek_master_look.Id, &see_seek_master_look.Uuid, &see_seek_master_look.SeeSeekMasterId, &see_seek_master_look.Classify, &see_seek_master_look.Status, &see_seek_master_look.RequestOutline, &see_seek_master_look.IsDeform, &see_seek_master_look.RequestSkin, &see_seek_master_look.IsGraze, &see_seek_master_look.RequestColor, &see_seek_master_look.IsChange, &see_seek_master_look.RequestLookHistory, &see_seek_master_look.MasterOutline, &see_seek_master_look.MasterIsDeform, &see_seek_master_look.MasterSkin, &see_seek_master_look.MasterIsGraze, &see_seek_master_look.MasterColor, &see_seek_master_look.MasterIsChange, &see_seek_master_look.MasterLookHistory, &see_seek_master_look.CreatedAt, &see_seek_master_look.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// SeeSeekMasterLook.Update() // 更新一个SeeSeekMasterLook
func (see_seek_master_look *SeeSeekMasterLook) Update() (err error) {
	statement := "UPDATE see_seek_master_looks SET see_seek_master_id=$2, classify=$3, status=$4, request_outline=$5, is_deform=$6, request_skin=$7, is_graze=$8, request_color=$9, is_change=$10, request_look_history=$11, master_outline=$12, master_is_deform=$13, master_skin=$14, master_is_graze=$15, master_color=$16, master_is_change=$17, master_look_history=$18, updated_at=$19 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(see_seek_master_look.Id, see_seek_master_look.SeeSeekMasterId, see_seek_master_look.Classify, see_seek_master_look.Status, see_seek_master_look.RequestOutline, see_seek_master_look.IsDeform, see_seek_master_look.RequestSkin, see_seek_master_look.IsGraze, see_seek_master_look.RequestColor, see_seek_master_look.IsChange, see_seek_master_look.RequestLookHistory, see_seek_master_look.MasterOutline, see_seek_master_look.MasterIsDeform, see_seek_master_look.MasterSkin, see_seek_master_look.MasterIsGraze, see_seek_master_look.MasterColor, see_seek_master_look.MasterIsChange, see_seek_master_look.MasterLookHistory, time.Now())
	if err != nil {
		return
	}
	return
}

// 听，声音
type SeeSeekMasterListen struct {
	Id              int
	Uuid            string
	SeeSeekMasterId int
	Classify        int
	Status          int

	RequestSound        string
	IsAbnormal          bool
	RequestSoundHistory string

	MasterSound        string
	MasterIsAbnormal   bool
	MasterSoundHistory string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (see_seek_master_listen *SeeSeekMasterListen) Create() (err error) {
	statement := "INSERT INTO see_seek_master_listens (uuid, see_seek_master_id, classify, status, request_sound, is_abnormal, request_sound_history, master_sound, master_is_abnormal, master_sound_history, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), see_seek_master_listen.SeeSeekMasterId, see_seek_master_listen.Classify, see_seek_master_listen.Status, see_seek_master_listen.RequestSound, see_seek_master_listen.IsAbnormal, see_seek_master_listen.RequestSoundHistory, see_seek_master_listen.MasterSound, see_seek_master_listen.MasterIsAbnormal, see_seek_master_listen.MasterSoundHistory, time.Now()).Scan(&see_seek_master_listen.Id, &see_seek_master_listen.Uuid)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_listen *SeeSeekMasterListen) Get() (err error) {
	statement := "SELECT id, uuid, see_seek_master_id, classify, status, request_sound, is_abnormal, request_sound_history, master_sound, master_is_abnormal, master_sound_history, created_at, updated_at FROM see_seek_master_listens WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_master_listen.Id).Scan(&see_seek_master_listen.Id, &see_seek_master_listen.Uuid, &see_seek_master_listen.SeeSeekMasterId, &see_seek_master_listen.Classify, &see_seek_master_listen.Status, &see_seek_master_listen.RequestSound, &see_seek_master_listen.IsAbnormal, &see_seek_master_listen.RequestSoundHistory, &see_seek_master_listen.MasterSound, &see_seek_master_listen.MasterIsAbnormal, &see_seek_master_listen.MasterSoundHistory, &see_seek_master_listen.CreatedAt, &see_seek_master_listen.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_listen *SeeSeekMasterListen) GetByUuid() (err error) {
	statement := "SELECT id, uuid, see_seek_master_id, classify, status, request_sound, is_abnormal, request_sound_history, master_sound, master_is_abnormal, master_sound_history, created_at, updated_at FROM see_seek_master_listens WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_master_listen.Uuid).Scan(&see_seek_master_listen.Id, &see_seek_master_listen.Uuid, &see_seek_master_listen.SeeSeekMasterId, &see_seek_master_listen.Classify, &see_seek_master_listen.Status, &see_seek_master_listen.RequestSound, &see_seek_master_listen.IsAbnormal, &see_seek_master_listen.RequestSoundHistory, &see_seek_master_listen.MasterSound, &see_seek_master_listen.MasterIsAbnormal, &see_seek_master_listen.MasterSoundHistory, &see_seek_master_listen.CreatedAt, &see_seek_master_listen.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_listen *SeeSeekMasterListen) Update() (err error) {
	statement := "UPDATE see_seek_master_listens SET see_seek_master_id=$2, classify=$3, status=$4, request_sound=$5, is_abnormal=$6, request_sound_history=$7, master_sound=$8, master_is_abnormal=$9, master_sound_history=$10, updated_at=$11 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(see_seek_master_listen.Id, see_seek_master_listen.SeeSeekMasterId, see_seek_master_listen.Classify, see_seek_master_listen.Status, see_seek_master_listen.RequestSound, see_seek_master_listen.IsAbnormal, see_seek_master_listen.RequestSoundHistory, see_seek_master_listen.MasterSound, see_seek_master_listen.MasterIsAbnormal, see_seek_master_listen.MasterSoundHistory, time.Now())
	if err != nil {
		return
	}
	return
}

// 嗅，气味
type SeeSeekMasterSmell struct {
	Id              int
	Uuid            string
	SeeSeekMasterId int
	Classify        int
	Status          int

	RequestOdour        string
	IsFoulOdour         bool
	RequestOdourHistory string

	MasterOdour        string
	MasterIsFoulOdour  bool
	MasterOdourHistory string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (see_seek_master_smell *SeeSeekMasterSmell) Create() (err error) {
	statement := "INSERT INTO see_seek_master_smells (uuid, see_seek_master_id, classify, status, request_odour, is_foul_odour, request_odour_history, master_odour, master_is_foul_odour, master_odour_history, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), see_seek_master_smell.SeeSeekMasterId, see_seek_master_smell.Classify, see_seek_master_smell.Status, see_seek_master_smell.RequestOdour, see_seek_master_smell.IsFoulOdour, see_seek_master_smell.RequestOdourHistory, see_seek_master_smell.MasterOdour, see_seek_master_smell.MasterIsFoulOdour, see_seek_master_smell.MasterOdourHistory, time.Now()).Scan(&see_seek_master_smell.Id, &see_seek_master_smell.Uuid)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_smell *SeeSeekMasterSmell) Get() (err error) {
	statement := "SELECT id, uuid, see_seek_master_id, classify, status, request_odour, is_foul_odour, request_odour_history, master_odour, master_is_foul_odour, master_odour_history, created_at, updated_at FROM see_seek_master_smells WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_master_smell.Id).Scan(&see_seek_master_smell.Id, &see_seek_master_smell.Uuid, &see_seek_master_smell.SeeSeekMasterId, &see_seek_master_smell.Classify, &see_seek_master_smell.Status, &see_seek_master_smell.RequestOdour, &see_seek_master_smell.IsFoulOdour, &see_seek_master_smell.RequestOdourHistory, &see_seek_master_smell.MasterOdour, &see_seek_master_smell.MasterIsFoulOdour, &see_seek_master_smell.MasterOdourHistory, &see_seek_master_smell.CreatedAt, &see_seek_master_smell.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_smell *SeeSeekMasterSmell) GetByUuid() (err error) {
	statement := "SELECT id, uuid, see_seek_master_id, classify, status, request_odour, is_foul_odour, request_odour_history, master_odour, master_is_foul_odour, master_odour_history, created_at, updated_at FROM see_seek_master_smells WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_master_smell.Uuid).Scan(&see_seek_master_smell.Id, &see_seek_master_smell.Uuid, &see_seek_master_smell.SeeSeekMasterId, &see_seek_master_smell.Classify, &see_seek_master_smell.Status, &see_seek_master_smell.RequestOdour, &see_seek_master_smell.IsFoulOdour, &see_seek_master_smell.RequestOdourHistory, &see_seek_master_smell.MasterOdour, &see_seek_master_smell.MasterIsFoulOdour, &see_seek_master_smell.MasterOdourHistory, &see_seek_master_smell.CreatedAt, &see_seek_master_smell.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_smell *SeeSeekMasterSmell) Update() (err error) {
	statement := "UPDATE see_seek_master_smells SET see_seek_master_id=$2, classify=$3, status=$4, request_odour=$5, is_foul_odour=$6, request_odour_history=$7, master_odour=$8, master_is_foul_odour=$9, master_odour_history=$10, updated_at=$11 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(see_seek_master_smell.Id, see_seek_master_smell.SeeSeekMasterId, see_seek_master_smell.Classify, see_seek_master_smell.Status, see_seek_master_smell.RequestOdour, see_seek_master_smell.IsFoulOdour, see_seek_master_smell.RequestOdourHistory, see_seek_master_smell.MasterOdour, see_seek_master_smell.MasterIsFoulOdour, see_seek_master_smell.MasterOdourHistory, time.Now())
	if err != nil {
		return
	}
	return
}

// 触摸，
type SeeSeekMasterTouch struct {
	Id              int
	Uuid            string
	SeeSeekMasterId int
	Classify        int

	Status int

	RequestTemperature  string
	IsFever             bool
	RequestStretch      string
	IsStiff             bool
	RequestShake        string
	IsShake             bool
	RequestTouchHistory string

	MasterTemperature  string
	MasterIsFever      bool
	MasterStretch      string
	MasterIsStiff      bool
	MasterShake        string
	MasterIsShake      bool
	MasterTouchHistory string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (see_seek_master_touch *SeeSeekMasterTouch) Create() (err error) {
	statement := "INSERT INTO see_seek_master_touches (uuid, see_seek_master_id, classify, status, request_temperature, is_fever, request_stretch, is_stiff, request_shake, is_shake, request_touch_history, master_temperature, master_is_fever, master_stretch, master_is_stiff, master_shake, master_is_shake, master_touch_history, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), see_seek_master_touch.SeeSeekMasterId, see_seek_master_touch.Classify, see_seek_master_touch.Status, see_seek_master_touch.RequestTemperature, see_seek_master_touch.IsFever, see_seek_master_touch.RequestStretch, see_seek_master_touch.IsStiff, see_seek_master_touch.RequestShake, see_seek_master_touch.IsShake, see_seek_master_touch.RequestTouchHistory, see_seek_master_touch.MasterTemperature, see_seek_master_touch.MasterIsFever, see_seek_master_touch.MasterStretch, see_seek_master_touch.MasterIsStiff, see_seek_master_touch.MasterShake, see_seek_master_touch.MasterIsShake, see_seek_master_touch.MasterTouchHistory, time.Now()).Scan(&see_seek_master_touch.Id, &see_seek_master_touch.Uuid)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_touch *SeeSeekMasterTouch) Get() (err error) {
	statement := "SELECT id, uuid, see_seek_master_id, classify, status, request_temperature, is_fever, request_stretch, is_stiff, request_shake, is_shake, request_touch_history, master_temperature, master_is_fever, master_stretch, master_is_stiff, master_shake, master_is_shake, master_touch_history, created_at, updated_at FROM see_seek_master_touches WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_master_touch.Id).Scan(&see_seek_master_touch.Id, &see_seek_master_touch.Uuid, &see_seek_master_touch.SeeSeekMasterId, &see_seek_master_touch.Classify, &see_seek_master_touch.Status, &see_seek_master_touch.RequestTemperature, &see_seek_master_touch.IsFever, &see_seek_master_touch.RequestStretch, &see_seek_master_touch.IsStiff, &see_seek_master_touch.RequestShake, &see_seek_master_touch.IsShake, &see_seek_master_touch.RequestTouchHistory, &see_seek_master_touch.MasterTemperature, &see_seek_master_touch.MasterIsFever, &see_seek_master_touch.MasterStretch, &see_seek_master_touch.MasterIsStiff, &see_seek_master_touch.MasterShake, &see_seek_master_touch.MasterIsShake, &see_seek_master_touch.MasterTouchHistory, &see_seek_master_touch.CreatedAt, &see_seek_master_touch.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_touch *SeeSeekMasterTouch) GetByUuid() (err error) {
	statement := "SELECT id, uuid, see_seek_master_id, classify, status, request_temperature, is_fever, request_stretch, is_stiff, request_shake, is_shake, request_touch_history, master_temperature, master_is_fever, master_stretch, master_is_stiff, master_shake, master_is_shake, master_touch_history, created_at, updated_at FROM see_seek_master_touches WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(see_seek_master_touch.Uuid).Scan(&see_seek_master_touch.Id, &see_seek_master_touch.Uuid, &see_seek_master_touch.SeeSeekMasterId, &see_seek_master_touch.Classify, &see_seek_master_touch.Status, &see_seek_master_touch.RequestTemperature, &see_seek_master_touch.IsFever, &see_seek_master_touch.RequestStretch, &see_seek_master_touch.IsStiff, &see_seek_master_touch.RequestShake, &see_seek_master_touch.IsShake, &see_seek_master_touch.RequestTouchHistory, &see_seek_master_touch.MasterTemperature, &see_seek_master_touch.MasterIsFever, &see_seek_master_touch.MasterStretch, &see_seek_master_touch.MasterIsStiff, &see_seek_master_touch.MasterShake, &see_seek_master_touch.MasterIsShake, &see_seek_master_touch.MasterTouchHistory, &see_seek_master_touch.CreatedAt, &see_seek_master_touch.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (see_seek_master_touch *SeeSeekMasterTouch) Update() (err error) {
	statement := "UPDATE see_seek_master_touches SET see_seek_master_id=$2, classify=$3, status=$4, request_temperature=$5, is_fever=$6, request_stretch=$7, is_stiff=$8, request_shake=$9, is_shake=$10, request_touch_history=$11, master_temperature=$12, master_is_fever=$13, master_stretch=$14, master_is_stiff=$15, master_shake=$16, master_is_shake=$17, master_touch_history=$18, updated_at=$19 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(see_seek_master_touch.Id, see_seek_master_touch.SeeSeekMasterId, see_seek_master_touch.Classify, see_seek_master_touch.Status, see_seek_master_touch.RequestTemperature, see_seek_master_touch.IsFever, see_seek_master_touch.RequestStretch, see_seek_master_touch.IsStiff, see_seek_master_touch.RequestShake, see_seek_master_touch.IsShake, see_seek_master_touch.RequestTouchHistory, see_seek_master_touch.MasterTemperature, see_seek_master_touch.MasterIsFever, see_seek_master_touch.MasterStretch, see_seek_master_touch.MasterIsStiff, see_seek_master_touch.MasterShake, see_seek_master_touch.MasterIsShake, see_seek_master_touch.MasterTouchHistory, time.Now())
	if err != nil {
		return
	}
	return
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
	UpdatedAt      time.Time
}
