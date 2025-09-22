package data

// 查询功能，根据关键词查找数据库记录，所得到的数据集合，页面数据
type SearchPageData struct {
	SessUser User

	IsEmpty bool //查询结果为空?
	Count   int  //查询结果个数

	UserBeanSlice      []UserBean //茶友（用户）资料夹队列
	TeamBeanSlice      []TeamBean //茶团资料夹队列
	ThreadBeanSlice    []ThreadBean
	ProjectBeanSlice   []ProjectBean
	ObjectiveBeanSlice []ObjectiveBean
	PlaceSlice         []Place       //品茶地点集合
	EnvironmentSlice   []Environment //环境条件集合
	HazardSlice        []Hazard      //隐患集合
	RiskSlice          []Risk        //风险集合
	GoodsSlice         []Goods       // 物资集合
}

// 我的地盘我做主
type PlaceSlice struct {
	SessUser   User
	PlaceSlice []Place
}
type PlaceDetail struct {
	SessUser User
	Place    Place
	IsAuthor bool
}

// 好东西，物资清单
type GoodsTeamSlice struct {
	SessUser User
	IsAdmin  bool

	Team         Team
	Availability []GoodsAvailability

	GoodsSlice []Goods
}

// 茶团物资详情
type GoodsTeamDetail struct {
	SessUser  User
	IsAdmin   bool
	GoodsTeam GoodsTeam

	Team Team

	Goods Goods
}

type GoodsFamilySlice struct {
	SessUser User
	IsAdmin  bool

	Family       Family
	Availability []GoodsAvailability

	GoodsSlice []Goods
}
type GoodsFamilyDetail struct {
	SessUser    User
	IsAdmin     bool
	GoodsFamily GoodsFamily

	Family Family

	Goods Goods
}
type GoodsUserSlice struct {
	SessUser User

	GoodsSlice []Goods
}
type GoodsDetail struct {
	SessUser User
	Goods    Goods
}

type GoodsProjectSlice struct {
	SessUser User
	IsAdmin  bool

	Project Project

	GoodsSlicePayee []Goods //收茶方可用物资
	GoodsSlicePayer []Goods //出茶方可用物资
}

type GoodsProjectList struct {
	SessUser   User
	IsAdmin    bool
	IsMaster   bool
	IsVerifier bool
	IsInvited  bool

	ProjectBean        ProjectBean
	QuoteObjectiveBean ObjectiveBean

	GoodsProjectList      []GoodsProject
	GoodsProjectReadiness GoodsProjectReadiness
	HasReadinessRecord    bool

	GoodsList []Goods
}
