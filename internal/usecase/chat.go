package usecase

import (
	"api_chat_ws/dto"
	"api_chat_ws/internal/repository"

	"encoding/json"
	"fmt"
)

type ChatUsecase interface {
	CreateGroup(req *dto.CreateGroupReq) error
	UpdateGroup(req *dto.UpdateGroupReq) error
	DeleteGroup(adminId, groupId uint) error
	AddMember(req *dto.AddMemberReq) error
	RemoveMember(req []uint, adminId uint) error
	ExitGroup(memberId uint) error
	UpdateRoleUser(req *dto.UpdateRoleMember) error

	GetMemberId(id, groupId uint) (uint, error)
	LoadGroupChat(groupId uint) ([]byte, error)
	GetMembers(groupId uint) ([]uint, error)
	CreateChat(userID, groupID uint, message string, status []dto.MemberStatus) ([]byte, error)
	UpdateChat(chatId, memberId uint, message string) ([]byte, error)
	DeleteChat(memberId, id uint) ([]byte, error)
	UpdateStatusChat(memberId uint) error
}

type chatUsecase struct {
	repo repository.ChatRepo
}

func NewChatUsecase(r repository.ChatRepo) ChatUsecase {
	return &chatUsecase{r}
}

func (u *chatUsecase) CreateGroup(req *dto.CreateGroupReq) error {
	return u.repo.CreateGroup(req)
}

func (u *chatUsecase) UpdateGroup(req *dto.UpdateGroupReq) error {
	valid, err := u.repo.IsMemberAdmin(req.MemberId)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("you arent admin")
	}

	return u.repo.UpdateGroup(req)
}

func (u *chatUsecase) DeleteGroup(adminId, groupId uint) error {
	valid, err := u.repo.IsMemberAdmin(adminId)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("you arent admin")
	}

	return u.repo.DeleteGroup(groupId)

}

func (u *chatUsecase) AddMember(req *dto.AddMemberReq) error {
	valid, err := u.repo.IsMemberAdmin(req.AdminId)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("you arent admin")
	}

	return u.repo.AddMember(req)
}

func (u *chatUsecase) RemoveMember(req []uint, adminId uint) error {
	valid, err := u.repo.IsMemberAdmin(adminId)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("you arent admin")
	}

	return u.repo.RemoveMember(req)
}

func (u *chatUsecase) ExitGroup(memberId uint) error {
	return u.repo.ExitGroup(memberId)
}

func (u *chatUsecase) UpdateRoleUser(req *dto.UpdateRoleMember) error {
	valid, err := u.repo.IsMemberAdmin(req.AdminId)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("you arent admin")
	}

	return u.repo.UpdateRoleUser(req.MemberId, req.Role)
}

func (u *chatUsecase) GetMembers(groupId uint) ([]uint, error) {
	return u.repo.GetGroupMembers(groupId)
}

func (u *chatUsecase) CreateChat(memberId, groupID uint, message string, status []dto.MemberStatus) ([]byte, error) {

	chat, err := u.repo.CreateChat(memberId, groupID, message, status)
	if err != nil {
		return nil, err
	}

	response, _ := json.Marshal(&chat)
	return response, nil
}

func (u *chatUsecase) UpdateChat(chatId, memberId uint, message string) ([]byte, error) {
	chat, err := u.repo.UpdateChat(chatId, memberId, message)
	if err != nil {
		return nil, err
	}

	response, _ := json.Marshal(&chat)
	return response, nil
}
func (u *chatUsecase) DeleteChat(memberId, id uint) ([]byte, error) {
	chat, err := u.repo.DeleteChat(memberId, id)
	if err != nil {
		return nil, err
	}
	response, _ := json.Marshal(&chat)
	return response, nil
}

func (u *chatUsecase) LoadGroupChat(groupId uint) ([]byte, error) {
	chats, err := u.repo.LoadGroupChat(groupId)
	if err != nil {
		return nil, err
	}

	response, _ := json.Marshal(&chats)
	return response, nil
}

func (u *chatUsecase) GetMemberId(id, groupId uint) (uint, error) {
	return u.repo.GetMemberId(id, groupId)
}

func (u *chatUsecase) UpdateStatusChat(memberId uint) error {
	return u.repo.UpdateStatusChat(memberId)
}
