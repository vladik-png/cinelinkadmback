package services

import (
	"admin-aws/internal/database"
	"admin-aws/internal/models"
	"errors"
	"time"
)

type ChatDTO struct {
	ID              uint        `json:"id"`
	Name            string      `json:"name"`
	Avatar          string      `json:"avatar"`
	ChatType        string      `json:"chat_type"`
	CreatorID       uint        `json:"creator_id"`
	LastMessage     interface{} `json:"last_message"`
	ParticipantsIds []uint      `json:"participants_ids"`
}

type ChatMemberDTO struct {
	EmployeeID uint   `json:"employee_id"`
	FullName   string `json:"full_name"`
	AvatarUrl  string `json:"avatar_url"`
	Role       string `json:"role"`
}

type ChatMessageDTO struct {
	MessageID   uint   `json:"message_id"`
	ChatID      uint   `json:"chat_id"`
	EmployeeID  uint   `json:"user_id"`
	MessageType string `json:"message_type"`
	Message     string `json:"message"`
	Timestamp   string `json:"timestamp"`
	Status      string `json:"status"`
}

func GetUserChats(employeeID uint) ([]ChatDTO, error) {
	var members []models.CorporateChatMember
	if err := database.ChatDB.Where("employee_id = ?", employeeID).Find(&members).Error; err != nil {
		return nil, err
	}

	var chatIDs []uint
	for _, m := range members {
		chatIDs = append(chatIDs, m.ChatID)
	}

	var chats []models.CorporateChat
	if len(chatIDs) > 0 {
		database.ChatDB.Where("id IN ?", chatIDs).Find(&chats)
	}

	var result []ChatDTO
	for _, c := range chats {
		var lastMsg models.CorporateChatMessage
		database.ChatDB.Where("chat_id = ?", c.ID).Order("created_at desc").First(&lastMsg)

		var participants []models.CorporateChatMember
		database.ChatDB.Where("chat_id = ?", c.ID).Find(&participants)
		var pIds []uint
		for _, p := range participants {
			pIds = append(pIds, p.EmployeeID)
		}

		var lastMsgDTO interface{}
		if lastMsg.ID != 0 {
			lastMsgDTO = ChatMessageDTO{
				MessageID:   lastMsg.ID,
				ChatID:      lastMsg.ChatID,
				EmployeeID:  lastMsg.EmployeeID,
				MessageType: string(lastMsg.MessageType),
				Message:     lastMsg.MessageContent,
				Timestamp:   lastMsg.AddedAt.Format("2006-01-02T15:04:05Z07:00"),
				Status:      "sent",
			}
		}

		result = append(result, ChatDTO{
			ID:              c.ID,
			Name:            c.Name,
			ChatType:        string(c.ChatType),
			CreatorID:       c.CreatorID,
			ParticipantsIds: pIds,
			LastMessage:     lastMsgDTO,
		})
	}

	return result, nil
}

func GetChatDetails(chatID uint) (*ChatDTO, error) {
	var c models.CorporateChat
	if err := database.ChatDB.First(&c, chatID).Error; err != nil {
		return nil, err
	}
	var participants []models.CorporateChatMember
	database.ChatDB.Where("chat_id = ?", c.ID).Find(&participants)
	var pIds []uint
	for _, p := range participants {
		pIds = append(pIds, p.EmployeeID)
	}

	return &ChatDTO{
		ID:              c.ID,
		Name:            c.Name,
		ChatType:        string(c.ChatType),
		CreatorID:       c.CreatorID,
		ParticipantsIds: pIds,
	}, nil
}

func GetOrCreateChat(employeeID, friendID uint) (uint, error) {
	var existingChatID uint
	query := `
		SELECT c.id
		FROM corporate_chats c
		JOIN corporate_chat_members cm1 ON c.id = cm1.chat_id
		JOIN corporate_chat_members cm2 ON c.id = cm2.chat_id
		WHERE c.chat_type = 'direct' AND cm1.employee_id = ? AND cm2.employee_id = ?
		LIMIT 1
	`
	err := database.ChatDB.Raw(query, employeeID, friendID).Scan(&existingChatID).Error
	if err == nil && existingChatID != 0 {
		return existingChatID, nil
	}

	newChat := models.CorporateChat{
		ChatType:  models.ChatTypeDirect,
		CreatorID: employeeID,
	}
	if err := database.ChatDB.Create(&newChat).Error; err != nil {
		return 0, err
	}

	database.ChatDB.Create(&models.CorporateChatMember{ChatID: newChat.ID, EmployeeID: employeeID, Role: "admin"})
	if employeeID != friendID {
		database.ChatDB.Create(&models.CorporateChatMember{ChatID: newChat.ID, EmployeeID: friendID, Role: "user"})
	}

	return newChat.ID, nil
}

func GetChatMembers(chatID uint) ([]ChatMemberDTO, error) {
	var members []models.CorporateChatMember
	if err := database.ChatDB.Where("chat_id = ?", chatID).Find(&members).Error; err != nil {
		return nil, err
	}

	var result []ChatMemberDTO
	for _, m := range members {
		result = append(result, ChatMemberDTO{
			EmployeeID: m.EmployeeID,
			Role:       m.Role,
			FullName:   "Employee " + string(rune(m.EmployeeID)),
		})
	}
	return result, nil
}

func AddChatMember(chatID, employeeID uint) error {
	var count int64
	database.ChatDB.Model(&models.CorporateChatMember{}).Where("chat_id = ? AND employee_id = ?", chatID, employeeID).Count(&count)
	if count > 0 {
		return errors.New("already a member")
	}
	member := models.CorporateChatMember{
		ChatID:     chatID,
		EmployeeID: employeeID,
		Role:       "user",
	}
	return database.ChatDB.Create(&member).Error
}

func RemoveChatMember(chatID, employeeID uint) error {
	return database.ChatDB.Where("chat_id = ? AND employee_id = ?", chatID, employeeID).Delete(&models.CorporateChatMember{}).Error
}

func GetChatMessages(chatID uint) ([]ChatMessageDTO, error) {
	var messages []models.CorporateChatMessage
	if err := database.ChatDB.Where("chat_id = ?", chatID).Order("created_at asc").Find(&messages).Error; err != nil {
		return nil, err
	}

	var members []models.CorporateChatMember
	database.ChatDB.Where("chat_id = ?", chatID).Find(&members)

	memberOnline := make(map[uint]bool)
	memberLastSeenMsg := make(map[uint]uint)

	for _, m := range members {
		memberLastSeenMsg[m.EmployeeID] = m.LastSeenMessageID
		var empStatus models.EmployeeStatus
		database.ChatDB.First(&empStatus, m.EmployeeID)
		memberOnline[m.EmployeeID] = empStatus.IsOnline
	}

	var result []ChatMessageDTO
	for _, msg := range messages {
		status := "sent"
		maxLastSeen := uint(0)
		anyOnline := false

		for _, m := range members {
			if m.EmployeeID != msg.EmployeeID {
				if memberLastSeenMsg[m.EmployeeID] > maxLastSeen {
					maxLastSeen = memberLastSeenMsg[m.EmployeeID]
				}
				if memberOnline[m.EmployeeID] {
					anyOnline = true
				}
			}
		}

		if msg.ID <= maxLastSeen {
			status = "seen"
		} else if anyOnline {
			status = "delivered"
		}

		result = append(result, ChatMessageDTO{
			MessageID:   msg.ID,
			ChatID:      msg.ChatID,
			EmployeeID:  msg.EmployeeID,
			MessageType: string(msg.MessageType),
			Message:     msg.MessageContent,
			Timestamp:   msg.AddedAt.Format("2006-01-02T15:04:05Z07:00"),
			Status:      status,
		})
	}
	return result, nil
}

func SendMessage(chatID, employeeID uint, content, msgType string) (ChatMessageDTO, error) {
	msg := models.CorporateChatMessage{
		ChatID:         chatID,
		EmployeeID:     employeeID,
		MessageType:    models.MessageType(msgType),
		MessageContent: content,
		AddedAt:        time.Now().UTC(),
	}
	if err := database.ChatDB.Create(&msg).Error; err != nil {
		return ChatMessageDTO{}, err
	}

	return ChatMessageDTO{
		MessageID:   msg.ID,
		ChatID:      msg.ChatID,
		EmployeeID:  msg.EmployeeID,
		MessageType: string(msg.MessageType),
		Message:     msg.MessageContent,
		Timestamp:   msg.AddedAt.Format("2006-01-02T15:04:05Z07:00"),
		Status:      "sent",
	}, nil
}

func DeleteChat(chatID uint) error {
	if err := database.ChatDB.Where("chat_id = ?", chatID).Delete(&models.CorporateChatMember{}).Error; err != nil {
		return err
	}
	if err := database.ChatDB.Where("chat_id = ?", chatID).Delete(&models.CorporateChatMessage{}).Error; err != nil {
		return err
	}
	if err := database.ChatDB.Where("id = ?", chatID).Delete(&models.CorporateChat{}).Error; err != nil {
		return err
	}
	return nil
}
