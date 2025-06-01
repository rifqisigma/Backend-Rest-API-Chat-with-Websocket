package dto

import "time"

//auth
type RegisterReq struct {
	Name     string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

//chat
type IncomingMessage struct {
	Action  string `json:"action"`
	Content string `json:"content"`
	ID      uint   `json:"id"`
}

type CreateChatReq struct {
	GroupId  uint   `json:"group_id"`
	MemberId uint   `json:"member_id"`
	Message  string `json:"message"`
}

type UpdateChatReq struct {
	ChatId   uint   `json:"-"`
	MemberId uint   `json:"-"`
	Message  string `json:"-"`
}

type ResponseChat struct {
	ID        uint             `json:"chat_id"`
	MemberId  uint             `json:"member_id"`
	Message   string           `json:"message"`
	CreatedAt time.Time        `json:"created_at"`
	Status    []StatusChatRead `json:"status"`
}

type StatusChatRead struct {
	MemberId uint `json:"member_id"`
	IsRead   bool `json:"is_read"`
}

//group
type CreateGroupReq struct {
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	UserId uint   `json:"-"`
}

type UpdateGroupReq struct {
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	MemberId uint   `json:"-"`
	GroupId  uint   `json:"-"`
}

//group members
type AddMemberReq struct {
	AdminId uint   `json:"-"`
	GroupId uint   `json:"-"`
	UserIds []uint `json:"user_id"`
}

type RemoveMemberReq struct {
	AdminId uint   `json:"-"`
	UserIds []uint `json:"user_id"`
}

type UpdateRoleMember struct {
	AdminId  uint   `json:"-"`
	GroupId  uint   `json:"-"`
	MemberId uint   `json:"member_id"`
	Role     string `json:"role"`
}

//chat read
type MemberStatus struct {
	MemberId uint `json:"member_id"`
	Status   bool `json:"is_read"`
}
