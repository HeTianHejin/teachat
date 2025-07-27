package data

import "time"

// 对作业环境模糊（口头）记录，用于茶话会交流
// 作业环境属性
type Environment struct {
	Id      int
	Uuid    string
	Summary string //概述

	// 1: "极热（Scorching）",    >40℃
	// 2: "炎热（Hot）",          30-40℃
	// 3: "舒适（Comfortable）",  18-30℃
	// 4: "微凉（Cool）",         5-18℃
	// 5: "寒冷（Freezing）",     <5℃
	Temperature int //温度

	// 1: "极湿（Suffocating）", // >90%
	// 2: "潮湿（Humid）",       // 70-90%
	// 3: "适宜（Balanced）",    // 40-70%
	// 4: "干燥（Dry）",         // 20-40%
	// 5: "极干（Arid）",        // <20%
	Humidity int //湿度

	// 1: "毒雾（Hazardous）",       // >250
	// 2: "重污染（Very Unhealthy）", // 150-250
	// 3: "中度污染（Unhealthy）",     // 55-150
	// 4: "轻度污染（Moderate）",      // 35-55
	// 5: "优良（Good）",            // <35
	PM25 int //粉尘

	// 1: "震耳欲聋（Deafening）", / >90dB
	// 2: "嘈杂（Noisy）",       / 70-90dB
	// 3: "一般（Moderate）",    / 50-70dB
	// 4: "安静（Quiet）",       / 30-50dB
	// 5: "极静（Silent）",      / <30dB
	Noise int //噪声

	//  1: "刺眼（Blinding）", // >10,000
	// 	2: "明亮（Bright）",   // 1,000-10,000
	// 	3: "适中（Normal）",   // 300-1,000
	// 	4: "昏暗（Dim）",      // 50-300
	// 	5: "黑暗（Dark）",     // <50
	Light int //光照

	// 1: "飓风（Hurricane）", // >32.7
	// 2: "强风（Gale）",      // 10.8-32.7
	// 3: "微风（Breeze）",    // 3.3-10.8
	// 4: "轻风（Light）",     // 1.5-3.3
	// 5: "无风（Calm）",      // <1.5
	Wind int //风力

	// 1: "湍急（Rapid）",    // >3
	// 2: "较快（Swift）",    // 1-3
	// 3: "平稳（Steady）",   // 0.3-1
	// 4: "缓慢（Slow）",     // 0.1-0.3
	// 5: "静止（Stagnant）", // <0.1
	Flow int //流速

	// 1: "暴雨（Torrential）", // >50
	// 2: "大雨（Heavy）",      // 15-50
	// 3: "中雨（Moderate）",   // 5-15
	// 4: "小雨（Light）",      // 1-5
	// 5: "无雨（None）",       // <1
	Rain int //雨量

	// 1: "极高（Very High）", // >1030
	// 2: "偏高（High）",      // 1015-1030
	// 3: "正常（Normal）",    // 990-1015
	// 4: "偏低（Low）",       // 970-990
	// 5: "极低（Very Low）",  // <970
	Pressure int //气压

	// 1: "严重烟雾（Dense）", // 高浓度
	// 2: "明显烟雾（Thick）", // 中高浓度
	// 3: "轻度烟雾（Hazy）",  // 可察觉
	// 4: "微量烟雾（Trace）", // 轻微
	// 5: "无烟雾（Clear）",  // 无
	Smoke int //烟雾

	// 1: "沙尘暴（Dust Storm）", // >500
	// 2: "重度扬尘（Heavy）",     // 200-500
	// 3: "中度扬尘（Moderate）",  // 100-200
	// 4: "轻度扬尘（Light）",     // 50-100
	// 5: "无尘（Clean）",       // <50
	Dust int //扬尘

	//  1: "极臭（Extreme Stench）",
	// 	2: "浓烈臭味（Strong Odor）",
	// 	3: "明显异味（Noticeable Smell）",
	// 	4: "轻微气味（Faint Odor）",
	// 	5: "无异味（Odorless）",
	Odor int //异味:

	// 1: "极差（Zero）",      // <0.1
	// 2: "很差（Poor）",      // 0.1-1
	// 3: "一般（Fair）",      // 1-5
	// 4: "良好（Good）",      // 5-10
	// 5: "极佳（Excellent）", // >10
	Visibility int //能见度

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 作业环境属性分级映射
var LevelMaps = map[string]map[int]string{
	// 异味
	"Odor": {
		1: "极臭（Extreme Stench）",
		2: "浓烈臭味（Strong Odor）",
		3: "明显异味（Noticeable Smell）",
		4: "轻微气味（Faint Odor）",
		5: "无异味（Odorless）",
	},
	// 噪声（分贝逻辑：数字越小越安静）
	"Noise": {
		1: "震耳欲聋（Deafening）", // >90dB
		2: "嘈杂（Noisy）",       // 70-90dB
		3: "一般（Moderate）",    // 50-70dB
		4: "安静（Quiet）",       // 30-50dB
		5: "极静（Silent）",      // <30dB
	},
	// 温度（℃）
	"Temperature": {
		1: "极热（Scorching）",   // >40℃
		2: "炎热（Hot）",         // 30-40℃
		3: "舒适（Comfortable）", // 18-30℃
		4: "微凉（Cool）",        // 5-18℃
		5: "寒冷（Freezing）",    // <5℃
	},
	// 湿度（%RH）
	"Humidity": {
		1: "极湿（Suffocating）", // >90%
		2: "潮湿（Humid）",       // 70-90%
		3: "适宜（Balanced）",    // 40-70%
		4: "干燥（Dry）",         // 20-40%
		5: "极干（Arid）",        // <20%
	},
	// PM2.5（μg/m³）
	"PM25": {
		1: "毒雾（Hazardous）",       // >250
		2: "重污染（Very Unhealthy）", // 150-250
		3: "中度污染（Unhealthy）",     // 55-150
		4: "轻度污染（Moderate）",      // 35-55
		5: "优良（Good）",            // <35
	},
	// 光照（Lux）
	"Light": {
		1: "刺眼（Blinding）", // >10,000
		2: "明亮（Bright）",   // 1,000-10,000
		3: "适中（Normal）",   // 300-1,000
		4: "昏暗（Dim）",      // 50-300
		5: "黑暗（Dark）",     // <50
	},
	// 风力（m/s）
	"Wind": {
		1: "飓风（Hurricane）", // >32.7
		2: "强风（Gale）",      // 10.8-32.7
		3: "微风（Breeze）",    // 3.3-10.8
		4: "轻风（Light）",     // 1.5-3.3
		5: "无风（Calm）",      // <1.5
	},
	// 流速（m/s，通用流体）
	"Flow": {
		1: "湍急（Rapid）",    // >3
		2: "较快（Swift）",    // 1-3
		3: "平稳（Steady）",   // 0.3-1
		4: "缓慢（Slow）",     // 0.1-0.3
		5: "静止（Stagnant）", // <0.1
	},
	// 雨量（mm/h）
	"Rain": {
		1: "暴雨（Torrential）", // >50
		2: "大雨（Heavy）",      // 15-50
		3: "中雨（Moderate）",   // 5-15
		4: "小雨（Light）",      // 1-5
		5: "无雨（None）",       // <1
	},
	// 气压（hPa）
	"Pressure": {
		1: "极高（Very High）", // >1030
		2: "偏高（High）",      // 1015-1030
		3: "正常（Normal）",    // 990-1015
		4: "偏低（Low）",       // 970-990
		5: "极低（Very Low）",  // <970
	},
	// 烟雾（浓度指数）
	"Smoke": {
		1: "严重烟雾（Dense）", // 高浓度
		2: "明显烟雾（Thick）", // 中高浓度
		3: "轻度烟雾（Hazy）",  // 可察觉
		4: "微量烟雾（Trace）", // 轻微
		5: "无烟雾（Clear）",  // 无
	},
	// 扬尘（μg/m³）
	"Dust": {
		1: "沙尘暴（Dust Storm）", // >500
		2: "重度扬尘（Heavy）",     // 200-500
		3: "中度扬尘（Moderate）",  // 100-200
		4: "轻度扬尘（Light）",     // 50-100
		5: "无尘（Clean）",       // <50
	},
	// 能见度（km）
	"Visibility": {
		1: "极差（Zero）",      // <0.1
		2: "很差（Poor）",      // 0.1-1
		3: "一般（Fair）",      // 1-5
		4: "良好（Good）",      // 5-10
		5: "极佳（Excellent）", // >10
	},
}

// 根据字段名和等级返回作业环境描述
// 安全获取分级描述（处理无效字段或等级）
func GetLevelDescription(field string, level int) string {
	if level < 1 || level > 5 {
		return "无效等级"
	}
	if m, ok := LevelMaps[field]; ok {
		return m[level]
	}
	return "未知字段：" + field
}

// 常见作业环境
// 家庭--普通家庭
var EnvironmentHome = Environment{
	Id:          1,
	Uuid:        "62c06442-2b60-418a-6493-a91bd03ae4k6",
	Summary:     "普通家庭",
	Temperature: 3,
	Humidity:    3,
	PM25:        5,
	Noise:       4,
	Light:       2,
	Wind:        5,
	Flow:        5,
	Rain:        5,
	Pressure:    3,
	Smoke:       5,
	Dust:        5,
	Odor:        5,
	Visibility:  4,
}

// Environment.Create() 创建作业环境
func (e *Environment) Create() (err error) {
	statement := "INSERT INTO environments (uuid, summary, temperature, humidity, pm25, noise, light, wind, flow, rain, pressure, smoke, dust, odor, visibility, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), e.Summary, e.Temperature, e.Humidity, e.PM25, e.Noise, e.Light, e.Wind, e.Flow, e.Rain, e.Pressure, e.Smoke, e.Dust, e.Odor, e.Visibility, time.Now()).Scan(&e.Id, &e.Uuid)
	if err != nil {
		return
	}
	return
}

// Environment.Get() 获取作业环境
func (e *Environment) Get() (err error) {
	statement := "SELECT id, uuid, summary, temperature, humidity, pm25, noise, light, wind, flow, rain, pressure, smoke, dust, odor, visibility, created_at, updated_at FROM environments WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(e.Id).Scan(&e.Id, &e.Uuid, &e.Summary, &e.Temperature, &e.Humidity, &e.PM25, &e.Noise, &e.Light, &e.Wind, &e.Flow, &e.Rain, &e.Pressure, &e.Smoke, &e.Dust, &e.Odor, &e.Visibility, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func GetEnvironment(id int) (e Environment, err error) {
	statement := "SELECT id, uuid, summary, temperature, humidity, pm25, noise, light, wind, flow, rain, pressure, smoke, dust, odor, visibility, created_at, updated_at FROM environments WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(id).Scan(&e.Id, &e.Uuid, &e.Summary, &e.Temperature, &e.Humidity, &e.PM25, &e.Noise, &e.Light, &e.Wind, &e.Flow, &e.Rain, &e.Pressure, &e.Smoke, &e.Dust, &e.Odor, &e.Visibility, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return
	}
	return
}

func (e *Environment) GetByUuid() (err error) {
	statement := "SELECT id, uuid, summary, temperature, humidity, pm25, noise, light, wind, flow, rain, pressure, smoke, dust, odor, visibility, created_at, updated_at FROM environments WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(e.Uuid).Scan(&e.Id, &e.Uuid, &e.Summary, &e.Temperature, &e.Humidity, &e.PM25, &e.Noise, &e.Light, &e.Wind, &e.Flow, &e.Rain, &e.Pressure, &e.Smoke, &e.Dust, &e.Odor, &e.Visibility, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return
	}
	return
}
