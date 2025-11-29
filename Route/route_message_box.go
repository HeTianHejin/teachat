package route

import (
	"fmt"
	"net/http"
	"strconv"

	data "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/message_box/detail?uuid=xxx 或 ?team_uuid=xxx
// 显示消息盒子详情
func MessageBoxDetail(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		//游客，需登录
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, s_u, "你好，茶博士说久仰大名，请问大名是谁？")
		return
	}

	vals := r.URL.Query()
	uuid := vals.Get("uuid")
	team_uuid := vals.Get("team_uuid")

	var messageBox data.MessageBox
	var team data.Team

	if uuid != "" {
		// 通过消息盒子UUID获取
		err = messageBox.GetMessageBoxByUUID(uuid)
		if err != nil {
			util.Debug(" Cannot get message box by uuid", err)
			report(w, s_u, "你好，茶博士未能找到这个消息盒子，请稍后再试。")
			return
		}

		// 获取关联的团队
		if messageBox.Type == data.MessageBoxTypeTeam {
			team, err = data.GetTeam(messageBox.ObjectId)
			if err != nil {
				util.Debug(" Cannot get team by id", err)
				report(w, s_u, "你好，茶博士未能找到关联的团队，请稍后再试。")
				return
			}
		} else {
			report(w, s_u, "你好，这个消息盒子不是团队消息盒。")
			return
		}
	} else if team_uuid != "" {
		// 通过团队UUID获取
		team, err = data.GetTeamByUUID(team_uuid)
		if err != nil {
			util.Debug(" Cannot get team by uuid", err)
			report(w, s_u, "你好，茶博士未能找到这个团队，请稍后再试。")
			return
		}

		// 获取或创建团队消息盒子（使用安全方法防止并发重复）
		err = messageBox.GetOrCreateMessageBox(data.MessageBoxTypeTeam, team.Id)
		if err != nil {
			util.Debug(" Cannot get or create message box", err)
			report(w, s_u, "你好，茶博士未能获取或创建消息盒子，请稍后再试。")
			return
		}
	} else {
		report(w, s_u, "你好，请提供有效的消息盒子编号或团队编号。")
		return
	}

	// 检查用户是否为团队成员
	isMember := false
	isCoreMember := false
	teamCoreMembers, err := team.CoreMembers()
	if err != nil {
		util.Debug(" Cannot get team core members", err)
		report(w, s_u, "你好，茶博士未能获取团队成员信息，请稍后再试。")
		return
	}

	// 检查核心成员
	for _, member := range teamCoreMembers {
		if member.UserId == s_u.Id {
			isMember = true
			isCoreMember = true
			break
		}
	}
	if !isMember {
		// 检查普通成员
		teamNormalMembers, err := team.NormalMembers()
		if err != nil {
			util.Debug(" Cannot get team normal members", err)
			report(w, s_u, "你好，茶博士未能获取团队成员信息，请稍后再试。")
			return
		}
		if !isMember {
			for _, member := range teamNormalMembers {
				if member.UserId == s_u.Id {
					isMember = true
					break
				}
			}
		}
	}

	if !isMember {
		report(w, s_u, "你好，只有团队成员才能查看团队消息盒。")
		return
	}

	// 获取消息列表（根据用户权限过滤）
	messages, err := messageBox.GetMessagesForUser(s_u.Id)
	if err != nil {
		util.Debug(" Cannot get messages", err)
		report(w, s_u, "你好，茶博士未能获取消息列表，请稍后再试。")
		return
	}

	// 计算用户可见的消息总数
	messageCount := messageBox.AllMessagesCountForUser(s_u.Id)

	// 准备页面数据
	var mBD data.MessageBoxDetail
	mBD.SessUser = s_u
	mBD.MessageBox = messageBox
	mBD.Team = team
	mBD.IsMember = isMember
	mBD.IsCoreMember = isCoreMember
	mBD.Messages = messages
	mBD.MessageCount = messageCount

	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery

	generateHTML(w, &mBD, "layout", "navbar.private", "message_box.detail")
}

// POST /v1/message/delete?id=xxx
// 删除消息
func MessageDelete(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		//游客，需登录
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, s_u, "你好，茶博士说久仰大名，请问大名是谁？")
		return
	}

	vals := r.URL.Query()
	id_str := vals.Get("id")
	if id_str == "" {
		report(w, s_u, "你好，请提供有效的消息编号。")
		return
	}

	id, err := strconv.Atoi(id_str)
	if err != nil {
		report(w, s_u, "你好，消息编号格式不正确。")
		return
	}

	var message data.Message
	err = message.GetMessageById(id)
	if err != nil {
		util.Debug(" Cannot get message by id", err)
		report(w, s_u, "你好，茶博士未能找到这条消息，请稍后再试。")
		return
	}

	// 获取消息盒子并检查团队成员权限
	var messageBox data.MessageBox
	err = messageBox.GetMessageBoxById(message.MessageBoxId)
	if err != nil {
		util.Debug(" Cannot get message box", err)
		report(w, s_u, "你好，茶博士未能找到消息盒子，请稍后再试。")
		return
	}

	var team data.Team
	if messageBox.Type == data.MessageBoxTypeTeam {
		team, err = data.GetTeam(messageBox.ObjectId)
		if err != nil {
			util.Debug(" Cannot get team by id", err)
			report(w, s_u, "你好，茶博士未能找到关联的团队，请稍后再试。")
			return
		}
	} else {
		report(w, s_u, "你好，这个消息盒子不是团队消息盒。")
		return
	}

	// 检查用户是否为团队成员
	isMember := false
	isCoreMember := false
	teamCoreMembers, err := team.CoreMembers()
	if err != nil {
		util.Debug(" Cannot get team core members", err)
		report(w, s_u, "你好，茶博士未能获取团队成员信息，请稍后再试。")
		return
	}

	// 检查核心成员
	for _, member := range teamCoreMembers {
		if member.UserId == s_u.Id {
			isMember = true
			isCoreMember = true
			break
		}
	}
	if !isMember {
		// 检查普通成员
		teamNormalMembers, err := team.NormalMembers()
		if err != nil {
			util.Debug(" Cannot get team normal members", err)
			report(w, s_u, "你好，茶博士未能获取团队成员信息，请稍后再试。")
			return
		}
		for _, member := range teamNormalMembers {
			if member.UserId == s_u.Id {
				isMember = true
				break
			}
		}
	}

	if !isMember {
		report(w, s_u, "你好，只有团队成员才能删除消息。")
		return
	}

	// 检查删除权限：
	// 1. 核心成员可以删除全体可见的消息
	// 2. 消息接收者可以删除自己的消息
	canDelete := false
	if message.ReceiverType == data.MessageReceiverTypeAll && isCoreMember {
		canDelete = true
	} else if message.ReceiverType == data.MessageReceiverTypeMember && message.ReceiverId == s_u.Id {
		canDelete = true
	}

	if !canDelete {
		report(w, s_u, "你好，您没有权限删除这条消息。")
		return
	}

	// 删除消息
	err = message.SoftDelete()
	if err != nil {
		util.Debug(" Cannot delete message", err)
		report(w, s_u, "你好，茶博士未能删除消息，请稍后再试。")
		return
	}

	// 重定向回消息盒子详情页面
	http.Redirect(w, r, fmt.Sprintf("/v1/message_box/detail?uuid=%s", messageBox.Uuid), http.StatusFound)
}

// POST /v1/message/read?id=xxx
// 标记消息为已读
func MessageRead(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		//游客，需登录
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, s_u, "你好，茶博士说久仰大名，请问大名是谁？")
		return
	}

	vals := r.URL.Query()
	id_str := vals.Get("id")
	if id_str == "" {
		report(w, s_u, "你好，请提供有效的消息编号。")
		return
	}

	id, err := strconv.Atoi(id_str)
	if err != nil {
		report(w, s_u, "你好，消息编号格式不正确。")
		return
	}

	var message data.Message
	err = message.GetMessageById(id)
	if err != nil {
		util.Debug(" Cannot get message by id", err)
		report(w, s_u, "你好，茶博士未能找到这条消息，请稍后再试。")
		return
	}

	// 检查用户是否有权限查看和标记这条消息为已读
	// 1. 首先检查消息权限：发送给"全体"的消息所有人可见，发送给"成员"的消息仅指定成员可见
	if message.ReceiverType == data.MessageReceiverTypeMember && message.ReceiverId != s_u.Id {
		report(w, s_u, "你好，这条消息是发送给特定成员的，您无权查看。")
		return
	}

	// 2. 获取消息盒子并检查团队成员权限
	var messageBox data.MessageBox
	err = messageBox.GetMessageBoxById(message.MessageBoxId)
	if err != nil {
		util.Debug(" Cannot get message box", err)
		report(w, s_u, "你好，茶博士未能找到消息盒子，请稍后再试。")
		return
	}

	// 如果是团队消息盒子，检查用户是否为团队成员
	if messageBox.Type == data.MessageBoxTypeTeam {
		team, err := data.GetTeam(messageBox.ObjectId)
		if err != nil {
			util.Debug(" Cannot get team", err)
			report(w, s_u, "你好，茶博士未能找到关联的团队，请稍后再试。")
			return
		}

		isMember := false
		teamCoreMembers, err := team.CoreMembers()
		if err == nil {
			for _, member := range teamCoreMembers {
				if member.UserId == s_u.Id {
					isMember = true
					break
				}
			}
		}

		if !isMember {
			teamNormalMembers, err := team.NormalMembers()
			if err == nil {
				for _, member := range teamNormalMembers {
					if member.UserId == s_u.Id {
						isMember = true
						break
					}
				}
			}
		}

		if !isMember {
			report(w, s_u, "你好，只有团队成员才能标记团队消息为已读。")
			return
		}
	}

	// 标记消息为已读
	err = message.UpdateRead()
	if err != nil {
		util.Debug(" Cannot update message read status", err)
		report(w, s_u, "你好，茶博士未能更新消息状态，请稍后再试。")
		return
	}

	// 重定向回消息盒子详情页面
	if messageBox.Type == data.MessageBoxTypeTeam {
		http.Redirect(w, r, fmt.Sprintf("/v1/message_box/detail?uuid=%s", messageBox.Uuid), http.StatusFound)
	} else {
		http.Redirect(w, r, "/v1/", http.StatusFound)
	}
}

// GET/POST /v1/message/send?team_uuid=xxx&receiver_id=xxx
// 显示发送纸条页面或处理发送请求
func MessageSend(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		messageSendPage(w, r)
	case "POST":
		messageSendPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/message/send?team_uuid=xxx&receiver_id=xxx
// 显示发送纸条页面
func messageSendPage(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		//游客，需登录
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, s_u, "你好，茶博士说久仰大名，请问大名是谁？")
		return
	}

	vals := r.URL.Query()
	team_uuid := vals.Get("team_uuid")
	receiver_id_str := vals.Get("receiver_id")

	if team_uuid == "" || receiver_id_str == "" {
		report(w, s_u, "你好，请提供有效的团队编号和接收者编号。")
		return
	}

	receiver_id, err := strconv.Atoi(receiver_id_str)
	if err != nil {
		report(w, s_u, "你好，接收者编号格式不正确。")
		return
	}

	// 获取团队信息
	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Debug(" Cannot get team by uuid", err)
		report(w, s_u, "你好，茶博士未能找到这个团队，请稍后再试。")
		return
	}

	// 检查用户是否为团队成员
	isMember := false
	teamCoreMembers, err := team.CoreMembers()
	if err != nil {
		util.Debug(" Cannot get team core members", err)
		report(w, s_u, "你好，茶博士未能获取团队成员信息，请稍后再试。")
		return
	}

	teamNormalMembers, err := team.NormalMembers()
	if err != nil {
		util.Debug(" Cannot get team normal members", err)
		report(w, s_u, "你好，茶博士未能获取团队成员信息，请稍后再试。")
		return
	}

	// 检查核心成员
	for _, member := range teamCoreMembers {
		if member.UserId == s_u.Id {
			isMember = true
			break
		}
	}

	// 检查普通成员
	if !isMember {
		for _, member := range teamNormalMembers {
			if member.UserId == s_u.Id {
				isMember = true
				break
			}
		}
	}

	if !isMember {
		report(w, s_u, "你好，只有团队成员才能发送纸条。")
		return
	}

	// 获取接收者用户信息
	receiver, err := data.GetUser(receiver_id)
	if err != nil {
		util.Debug(" Cannot get receiver by id", err)
		report(w, s_u, "你好，茶博士未能找到接收者，请稍后再试。")
		return
	}

	// 检查接收者是否为团队成员
	isReceiverMember := false
	for _, member := range teamCoreMembers {
		if member.UserId == receiver.Id {
			isReceiverMember = true
			break
		}
	}

	if !isReceiverMember {
		for _, member := range teamNormalMembers {
			if member.UserId == receiver.Id {
				isReceiverMember = true
				break
			}
		}
	}

	if !isReceiverMember {
		report(w, s_u, "你好，接收者不是该团队成员。")
		return
	}

	// 获取或创建团队消息盒子
	var messageBox data.MessageBox
	err = messageBox.GetMessageBoxByTypeAndObjectId(data.MessageBoxTypeTeam, team.Id)
	if err != nil {
		// 如果消息盒子不存在，创建一个新的
		messageBox.Uuid = data.Random_UUID()
		messageBox.Type = data.MessageBoxTypeTeam
		messageBox.ObjectId = team.Id
		messageBox.IsEmpty = true
		messageBox.MaxCount = 199
		err = messageBox.Create()
		if err != nil {
			util.Debug(" Cannot create message box", err)
			report(w, s_u, "你好，茶博士未能创建消息盒子，请稍后再试。")
			return
		}
	}

	// 准备页面数据
	var mSPD data.MessageSendPageData
	mSPD.SessUser = s_u
	mSPD.Team = team
	mSPD.Receiver = receiver
	mSPD.MessageBox = messageBox

	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery

	generateHTML(w, &mSPD, "layout", "navbar.private", "message.send")
}

// POST /v1/message/send
// 处理发送纸条请求
func messageSendPost(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		//游客，需登录
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, s_u, "你好，茶博士说久仰大名，请问大名是谁？")
		return
	}

	// 解析表单数据
	team_uuid := r.PostFormValue("team_uuid")
	receiver_id_str := r.PostFormValue("receiver_id")
	content := r.PostFormValue("content")

	if team_uuid == "" || receiver_id_str == "" || content == "" {
		report(w, s_u, "你好，请填写完整的纸条信息。")
		return
	}

	receiver_id, err := strconv.Atoi(receiver_id_str)
	if err != nil {
		report(w, s_u, "你好，接收者编号格式不正确。")
		return
	}

	// 获取团队信息
	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Debug(" Cannot get team by uuid", err)
		report(w, s_u, "你好，茶博士未能找到这个团队，请稍后再试。")
		return
	}

	// 检查用户是否为团队成员
	isMember := false
	teamCoreMembers, err := team.CoreMembers()
	if err == nil {
		for _, member := range teamCoreMembers {
			if member.UserId == s_u.Id {
				isMember = true
				break
			}
		}
	}

	if !isMember {
		teamNormalMembers, err := team.NormalMembers()
		if err == nil {
			for _, member := range teamNormalMembers {
				if member.UserId == s_u.Id {
					isMember = true
					break
				}
			}
		}
	}

	if !isMember {
		report(w, s_u, "你好，只有团队成员才能发送纸条。")
		return
	}

	// 获取或创建团队消息盒子
	var messageBox data.MessageBox
	err = messageBox.GetMessageBoxByTypeAndObjectId(data.MessageBoxTypeTeam, team.Id)
	if err != nil {
		// 如果消息盒子不存在，创建一个新的
		messageBox.Uuid = data.Random_UUID()
		messageBox.Type = data.MessageBoxTypeTeam
		messageBox.ObjectId = team.Id
		messageBox.IsEmpty = true
		messageBox.MaxCount = 199
		err = messageBox.Create()
		if err != nil {
			util.Debug(" Cannot create message box", err)
			report(w, s_u, "你好，茶博士未能创建消息盒子，请稍后再试。")
			return
		}
	}

	// 创建纸条消息
	message := data.Message{
		Uuid:           data.Random_UUID(),
		MessageBoxId:   messageBox.Id,
		SenderType:     data.MessageBoxTypeTeam,
		SenderObjectId: team.Id,
		ReceiverType:   data.MessageReceiverTypeMember,
		ReceiverId:     receiver_id,
		Content:        content,
		IsRead:         false,
	}

	err = message.Create()
	if err != nil {
		util.Debug(" Cannot create message", err)
		report(w, s_u, "你好，茶博士未能发送纸条，请稍后再试。")
		return
	}

	// 更新消息盒子计数

	// 重定向到消息盒子详情页面
	http.Redirect(w, r, fmt.Sprintf("/v1/message_box/detail?uuid=%s", messageBox.Uuid), http.StatusFound)
}

// GET/POST /v1/message/announcement/send?team_uuid=xxx
// 显示发送布告页面或处理发送请求
func MessageAnnouncementSend(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		messageAnnouncementSendPage(w, r)
	case "POST":
		messageAnnouncementSendPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/message/announcement/send?team_uuid=xxx
// 显示发送布告页面
func messageAnnouncementSendPage(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		//游客，需登录
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, s_u, "你好，茶博士说久仰大名，请问大名是谁？")
		return
	}

	vals := r.URL.Query()
	team_uuid := vals.Get("team_uuid")

	if team_uuid == "" {
		report(w, s_u, "你好，请提供有效的团队编号。")
		return
	}

	// 获取团队信息
	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Debug(" Cannot get team by uuid", err)
		report(w, s_u, "你好，茶博士未能找到这个团队，请稍后再试。")
		return
	}

	// 检查团队是否为开放式团队
	if !team.IsOpen() {
		report(w, s_u, "你好，只有开放式团队才能接收布告消息。")
		return
	}

	// 检查用户是否为团队成员（团队成员不需要通过这个页面发送布告，他们有其他方式）
	isMember := false
	teamCoreMembers, err := team.CoreMembers()
	if err == nil {
		for _, member := range teamCoreMembers {
			if member.UserId == s_u.Id {
				isMember = true
				break
			}
		}
	}

	if !isMember {
		teamNormalMembers, err := team.NormalMembers()
		if err == nil {
			for _, member := range teamNormalMembers {
				if member.UserId == s_u.Id {
					isMember = true
					break
				}
			}
		}
	}

	// 如果是团队成员，提示他们使用团队内部的消息功能
	if isMember {
		report(w, s_u, "你好，作为团队成员，请使用团队消息盒子发送消息。")
		return
	}

	// 获取或创建团队消息盒子
	var messageBox data.MessageBox
	err = messageBox.GetOrCreateMessageBox(data.MessageBoxTypeTeam, team.Id)
	if err != nil {
		util.Debug(" Cannot get or create message box", err)
		report(w, s_u, "你好，茶博士未能获取或创建消息盒子，请稍后再试。")
		return
	}

	// 检查消息盒子是否已满
	currentCount := messageBox.AllMessagesCount()
	if currentCount >= messageBox.MaxCount {
		report(w, s_u, "你好，该团队的消息盒子已满，无法发送新的布告。")
		return
	}

	// 准备页面数据
	var mASPD data.MessageAnnouncementSendPageData
	mASPD.SessUser = s_u
	mASPD.Team = team

	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery

	generateHTML(w, &mASPD, "layout", "navbar.private", "message.announcement.send")
}

// POST /v1/message/announcement/send
// 处理发送布告请求
func messageAnnouncementSendPost(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		//游客，需登录
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, s_u, "你好，茶博士说久仰大名，请问大名是谁？")
		return
	}

	// 解析表单数据
	team_uuid := r.PostFormValue("team_uuid")
	content := r.PostFormValue("content")
	sender_type_str := r.PostFormValue("sender_type")
	if team_uuid == "" || content == "" || sender_type_str == "" {
		report(w, s_u, "你好，请填写完整的布告信息。")
		return
	}
	sender_type_int, err := strconv.Atoi(sender_type_str)
	if err != nil {
		report(w, s_u, "你好，请填写完整的布告信息。")
		return
	}
	switch sender_type_int {
	case data.MessageSenderTypeTeam, data.MessageSenderTypeFamily:
		break
	default:
		report(w, s_u, "你好，请填写完整的布告信息。")
		return
	}

	// 获取团队信息
	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Debug(" Cannot get team by uuid", err)
		report(w, s_u, "你好，茶博士未能找到这个团队，请稍后再试。")
		return
	}

	// 检查团队是否为开放式团队
	if !team.IsOpen() {
		report(w, s_u, "你好，只有开放式团队才能接收布告消息。")
		return
	}

	// 检查用户是否为团队成员
	isMember := false
	teamCoreMembers, err := team.CoreMembers()
	if err == nil {
		for _, member := range teamCoreMembers {
			if member.UserId == s_u.Id {
				isMember = true
				break
			}
		}
	}

	if !isMember {
		teamNormalMembers, err := team.NormalMembers()
		if err == nil {
			for _, member := range teamNormalMembers {
				if member.UserId == s_u.Id {
					isMember = true
					break
				}
			}
		}
	}

	// 如果是团队成员，提示他们使用团队内部的消息功能
	if isMember {
		report(w, s_u, "你好，作为团队成员，请使用团队消息盒子发送消息。")
		return
	}

	// 获取或创建团队消息盒子
	var messageBox data.MessageBox
	err = messageBox.GetOrCreateMessageBox(data.MessageBoxTypeTeam, team.Id)
	if err != nil {
		util.Debug(" Cannot get or create message box", err)
		report(w, s_u, "你好，茶博士未能获取或创建消息盒子，请稍后再试。")
		return
	}

	// 检查消息盒子是否已满
	currentCount := messageBox.AllMessagesCount()
	if currentCount >= messageBox.MaxCount {
		report(w, s_u, "你好，该团队的消息盒子已满，无法发送新的布告。")
		return
	}

	// 创建布告消息（发送给全体成员）
	message := data.Message{
		Uuid:           data.Random_UUID(),
		MessageBoxId:   messageBox.Id,
		SenderType:     sender_type_int,
		SenderObjectId: s_u.Id,
		ReceiverType:   data.MessageReceiverTypeAll,
		ReceiverId:     data.UserId_None, // 布告消息发送给全体
		Content:        content,
		IsRead:         false,
	}

	err = message.Create()
	if err != nil {
		util.Debug(" Cannot create announcement message", err)
		report(w, s_u, "你好，茶博士未能发送布告，请稍后再试。")
		return
	}

	// 更新消息盒子状态（如果不为空）
	if messageBox.IsEmpty {
		messageBox.IsEmpty = false
		err = messageBox.Update()
		if err != nil {
			util.Debug(" Cannot update message box", err)
		}
	}

	// 显示成功消息并重定向
	report(w, s_u, fmt.Sprintf("你好，布告已成功发送给 %s 团队全体成员！", team.Name))
}
