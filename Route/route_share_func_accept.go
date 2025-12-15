package route

import (
	"context"
	"errors"
	"fmt"
	dao "teachat/DAO"
	util "teachat/Util"
)

// 创建AcceptObject并发送邻座蒙评通知
func createAndSendAcceptNotification(objectId int, objectType int, excludeUserId int, ctx context.Context) error {
	// 创建AcceptObject
	aO := dao.AcceptObject{
		ObjectId:   objectId,
		ObjectType: objectType,
	}
	if err := aO.Create(); err != nil {
		util.Debug("Cannot create accept_object given objectId", objectId)
		return fmt.Errorf("创建AcceptObject失败: %w", err)
	}

	// 创建通知
	mess := dao.AcceptNotification{
		FromUserId:     dao.UserId_Captain_Spaceship,
		Title:          "新茶语邻座评审邀请",
		Content:        "您被茶棚选中为新茶语评审官啦，请及时去审理。",
		AcceptObjectId: aO.Id,
	}

	// 发送通知
	if err := TwoAcceptNotificationsSendExceptUserId(excludeUserId, mess, ctx); err != nil {
		return fmt.Errorf("发送通知失败: %w", err)
	}

	// 返回提示信息
	return nil
}

// 接纳文明新茶围
func acceptNewObjective(objectId int) (*dao.Objective, error) {
	ob := dao.Objective{
		Id: objectId,
	}
	if err := ob.Get(); err != nil {
		util.Debug("Cannot get objective", objectId, err)
		return nil, errors.New("你好，茶博士失魂鱼，竟然说没有找到新茶茶叶的资料未必是怪事。")
	}
	// 检查当前茶围的状态
	switch ob.Class {
	case dao.ObClassOpenDraft:
		ob.Class = dao.ObClassOpen
	case dao.ObClassCloseDraft:
		ob.Class = dao.ObClassClose
	default:
		return nil, errors.New("你好，茶博士失魂鱼，竟然说有时找茶叶也是一种修养的过程。")
	}

	if err := ob.UpdateClass(); err != nil {
		util.Debug("Cannot update ob class", objectId, err)
		return nil, errors.New("你好，一畦春韭绿，十里稻花香。")
	}
	return &ob, nil
}

// 接纳文明新茶台
func acceptNewProject(objectId int) error {
	pr := dao.Project{
		Id: objectId,
	}
	if err := pr.Get(); err != nil {
		util.Debug("Cannot get project", objectId, err)
		return errors.New("你好，茶博士失魂鱼，竟然说有时找茶叶也是一种修养的过程。")
	}

	switch pr.Class {
	case dao.PrClassOpenDraft:
		pr.Class = dao.PrClassOpen
	case dao.PrClassCloseDraft:
		pr.Class = dao.PrClassClose
	default:
		return errors.New("你好，茶博士失魂鱼，竟然说有时找茶叶也是一种修养的过程。")
	}

	if err := pr.UpdateClass(); err != nil {
		util.Debug("Cannot update pr class", objectId, err)
		return errors.New("你好，一畦春韭绿，十里稻花香。")
	}
	return nil
}

// 接纳文明新茶议功能函数
func acceptNewDraftThread(objectId int) (*dao.Thread, error) {
	dThread := dao.DraftThread{Id: objectId}
	if err := dThread.Get(); err != nil {
		return nil, fmt.Errorf("获取茶议草稿失败: %w", err)
	}

	if err := dThread.UpdateStatus(dao.DraftThreadStatusAccepted); err != nil {
		return nil, fmt.Errorf("更新茶议草稿状态失败: %w", err)
	}

	thread := dao.Thread{
		Body:      dThread.Body,
		UserId:    dThread.UserId,
		Class:     dThread.Class,
		Title:     dThread.Title,
		ProjectId: dThread.ProjectId,
		FamilyId:  dThread.FamilyId,
		Type:      dThread.Type,
		PostId:    dThread.PostId,
		TeamId:    dThread.TeamId,
		IsPrivate: dThread.IsPrivate,
		Category:  dThread.Category,
	}

	if err := thread.Create(); err != nil {
		return nil, fmt.Errorf("创建新茶议失败: %w", err)
	}

	return &thread, nil
}

// 接纳文明新茶语之品味
func acceptNewDraftPost(objectId int) (*dao.Post, error) {
	dPost := dao.DraftPost{Id: objectId}
	if err := dPost.Get(); err != nil {
		return nil, fmt.Errorf("获取品味草稿失败: %w", err)
	}
	new_post := dao.Post{
		Body:      dPost.Body,
		UserId:    dPost.UserId,
		FamilyId:  dPost.FamilyId,
		TeamId:    dPost.TeamId,
		ThreadId:  dPost.ThreadId,
		IsPrivate: dPost.IsPrivate,
		Attitude:  dPost.Attitude,
		Class:     dPost.Class,
	}
	if err := new_post.Create(); err != nil {
		return nil, fmt.Errorf("创建新品味失败: %w", err)
	}
	return &new_post, nil
}

// 接纳文明新茶团功能函数
func acceptNewTeam(objectId int) (*dao.Team, error) {
	t := dao.Team{Id: objectId}
	if err := t.Get(); err != nil {
		util.Debug("Cannot get team", objectId, err)
		return nil, errors.New("你好，茶博士失魂鱼，竟然说没有找到新茶茶叶的资料未必是怪事。")
	}
	switch t.Class {
	case dao.TeamClassOpenDraft:
		t.Class = dao.TeamClassOpen
	case dao.TeamClassCloseDraft:
		t.Class = dao.TeamClassClose
	default:
		return nil, errors.New("你好，茶博士失魂鱼，竟然说有时找茶叶也是一种修养的过程。")
	}
	if err := t.UpdateClass(); err != nil {
		util.Debug("Cannot update t class", objectId, err)
		return nil, errors.New("你好，一畦春韭绿，十里稻花香。")
	}
	
	// 确保团队茶叶账户存在
	if err := dao.EnsureTeamTeaAccountExists(t.Id); err != nil {
		util.Debug("Cannot ensure team tea account exists", objectId, err)
		return nil, errors.New("你好，茶博士失魂鱼，竟然说团队茶叶账户准备失败。")
	}
	
	return &t, nil
}

// 接纳文明新集团的功能函数
func acceptNewGroup(objectId int) (*dao.Group, error) {
	g := dao.Group{Id: objectId}
	if err := g.Get(); err != nil {
		util.Debug("Cannot get group", objectId, err)
		return nil, errors.New("你好，茶博士失魂鱼，竟然说没有找到新茶茶叶的资料未必是怪事。")
	}
	switch g.Class {
	case dao.GroupClassOpenDraft:
		g.Class = dao.GroupClassOpen
	case dao.GroupClassCloseDraft:
		g.Class = dao.GroupClassClose
	default:
		return nil, errors.New("你好，茶博士失魂鱼，竟然说有时找茶叶也是一种修养的过程。")
	}
	if err := g.Update(); err != nil {
		util.Debug("Cannot update g class", objectId,

			err)
		return nil, errors.New("你好，一畦春韭绿，十里稻花香。")
	}
	return &g, nil
}
