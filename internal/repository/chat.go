package repository

import (
	"api_chat_ws/dto"
	"api_chat_ws/helper/utils"
	"api_chat_ws/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ChatRepo interface {
	CreateGroup(req *dto.CreateGroupReq) error
	UpdateGroup(req *dto.UpdateGroupReq) error
	DeleteGroup(groupId uint) error
	AddMember(req *dto.AddMemberReq) error
	RemoveMember(req []uint) error
	ExitGroup(memberId uint) error
	UpdateRoleUser(memberId uint, role string) error

	IsMemberAdmin(memberId uint) (bool, error)
	GetMemberId(id, groupId uint) (uint, error)
	LoadGroupChat(groupId uint) ([]dto.ResponseChat, error)
	GetGroupMembers(groupID uint) ([]uint, error)

	CreateChat(memberId, groupId uint, message string, status []dto.MemberStatus) (*dto.ResponseChat, error)
	UpdateChat(chatId, memberId uint, message string) (*dto.ResponseChat, error)
	DeleteChat(memberId, id uint) (*dto.ResponseChat, error)

	UpdateStatusChat(memberId uint) error
}

type chatRepo struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepo {
	return &chatRepo{db}
}

// write group & member
func (r *chatRepo) CreateGroup(req *dto.CreateGroupReq) error {
	tx := r.db.Begin()

	newGroup := model.ChatGroup{
		Name:        req.Name,
		Description: req.Name,
	}
	if err := tx.Create(&newGroup).Error; err != nil {
		tx.Rollback()
		return err
	}

	admin := model.GroupMember{
		GroupID: newGroup.ID,
		UserID:  req.UserId,
		Role:    "admin",
	}
	if err := tx.Create(&admin).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (r *chatRepo) UpdateGroup(req *dto.UpdateGroupReq) error {
	updated := make(map[string]interface{})

	if req.Desc != "" {
		updated["desc"] = req.Desc
	}
	if req.Name != "" {
		updated["name"] = req.Name
	}
	return r.db.Model(&model.ChatGroup{}).Where("id = ?", req.GroupId).Updates(updated).Error
}

func (r *chatRepo) DeleteGroup(groupId uint) error {
	return r.db.Model(&model.ChatGroup{}).Where("id = ?", groupId).Delete(&model.ChatGroup{}).Error
}

func (r *chatRepo) AddMember(req *dto.AddMemberReq) error {

	newMembers := make([]model.GroupMember, 0, len(req.UserIds))
	for _, nm := range req.UserIds {
		newMembers = append(newMembers, model.GroupMember{
			GroupID: req.GroupId,
			UserID:  nm,
			Role:    "member",
		})
	}

	return r.db.Model(&model.GroupMember{}).CreateInBatches(&newMembers, 2).Error
}

func (r *chatRepo) RemoveMember(req []uint) error {
	return r.db.Model(&model.GroupMember{}).Where("id IN ?", req).Delete(&model.GroupMember{}).Error
}

func (r *chatRepo) ExitGroup(memberId uint) error {
	return r.db.Model(&model.GroupMember{}).Where("id = ?", memberId).Delete(&model.GroupMember{}).Error
}

func (r *chatRepo) UpdateRoleUser(memberId uint, role string) error {
	return r.db.Model(&model.GroupMember{}).Where("id = ?", memberId).Update("role", role).Error
}

func (r *chatRepo) IsMemberAdmin(memberId uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.GroupMember{}).Where("id = ? AND role = ?", memberId, "admin").Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *chatRepo) LoadGroupChat(groupId uint) ([]dto.ResponseChat, error) {
	var chats []dto.ResponseChat
	if err := r.db.Model(&model.Chat{}).Select("id AS chat_id,group_member_id  AS member_id, message, created_at").Where("group_id = ?", groupId).Find(&chats).Order("-created_at").Error; err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *chatRepo) CreateChat(memberId, groupId uint, message string, status []dto.MemberStatus) (*dto.ResponseChat, error) {
	tx := r.db.Begin()
	newChat := model.Chat{
		GroupMemberID: &memberId,
		GroupID:       groupId,
		Message:       message,
	}
	if err := r.db.Create(&newChat).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	membersStatus := make([]model.ChatRead, 0, len(status))
	membersStatusResponse := make([]dto.StatusChatRead, 0, len(status))
	for _, m := range status {
		if m.MemberId == memberId {
			continue
		}
		membersStatus = append(membersStatus, model.ChatRead{
			ChatId:   newChat.ID,
			MemberId: m.MemberId,
			IsRead:   m.Status,
		})
		membersStatusResponse = append(membersStatusResponse, dto.StatusChatRead{
			MemberId: m.MemberId,
			IsRead:   m.Status,
		})
	}

	if len(membersStatus) > 0 {
		if err := tx.Model(&model.ChatRead{}).Create(&membersStatus).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()
	response := dto.ResponseChat{
		ID:        newChat.ID,
		MemberId:  *newChat.GroupMemberID,
		Message:   newChat.Message,
		CreatedAt: newChat.CreatedAt,
		Status:    membersStatusResponse,
	}
	return &response, nil
}

func (r *chatRepo) UpdateChat(chatId, memberId uint, message string) (*dto.ResponseChat, error) {
	err := r.db.Model(&model.Chat{}).Where("id = ? AND member_id = ?", chatId, memberId).Updates(map[string]interface{}{
		"message":    message,
		"created_at": time.Now(),
	}).Error
	if err != nil {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrrnotChat
	}

	var status []dto.StatusChatRead
	if err := r.db.Model(&model.ChatRead{}).Select("member_id, is_read").Where("chat_id = ?", chatId).Scan(&status).Error; err != nil {
		return nil, err
	}
	response := dto.ResponseChat{
		ID:        chatId,
		MemberId:  memberId,
		Message:   message,
		CreatedAt: time.Now(),
		Status:    status,
	}
	return &response, nil
}

func (r *chatRepo) DeleteChat(memberId, id uint) (*dto.ResponseChat, error) {
	err := r.db.Model(&model.Chat{}).Where("id = ? AND member_id = ?", id, memberId).Delete(&model.Chat{}).Error
	if err != nil {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrrnotChat
	}

	response := dto.ResponseChat{
		ID:        id,
		MemberId:  memberId,
		Message:   "has been deleted",
		CreatedAt: time.Now(),
	}
	return &response, nil
}

func (r *chatRepo) GetMemberId(id, groupId uint) (uint, error) {
	var member model.GroupMember
	err := r.db.Model(&model.GroupMember{}).Select("id").Where("user_id = ? AND group_id = ?", id, groupId).First(&member).Error
	if err != nil {
		return 0, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	return member.ID, nil
}

func (r *chatRepo) GetGroupMembers(groupID uint) ([]uint, error) {
	var members []uint
	err := r.db.Model(&model.GroupMember{}).Where("group_id = ?", groupID).Pluck("id", &members).Error
	return members, err
}

func (r *chatRepo) UpdateStatusChat(memberId uint) error {
	if err := r.db.Model(&model.ChatRead{}).Where("member_id = ?  AND is_read = ?", memberId, false).Update("is_read", true).Error; err != nil {
		return err
	}

	return nil
}
