package api

import (
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/models"
)

type UserRequest struct {
	TgId       float64 `json:"tg_id"`
	ChatId     float64 `json:"chat_id"`
	TgUsername string  `json:"tg_username"`
}

type UserDetails struct {
	TgId       int64  `json:"tg_id"`
	ChatId     int64  `json:"chat_id"`
	TgUsername string `json:"tg_username"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  *int64 `json:"updated_at"`
}

type CompletionRequestListData struct {
	Completions []CompletionRequest `json:"completions"`
}
type CompletionRequest struct {
	Text string `json:"text"`
	Role string `json:"role"`
}

func NewUserFromAddUserRequest(newUserRequest UserRequest) *models.User {
	return &models.User{
		Id:     int64(newUserRequest.TgId),
		ChatId: int64(newUserRequest.ChatId),
		TgName: newUserRequest.TgUsername,
	}
}

func NewUserFromUpdateUserRequest(newUserRequest UserRequest) *models.User {
	return NewUserFromAddUserRequest(newUserRequest)
}

type UserDetailsList struct {
	Users []UserDetails `json:"users"`
}

func UserDetailsListFrom(ud []models.UserDetails) UserDetailsList {
	userDetails := []UserDetails{}
	for _, userDetail := range ud {
		userDetails = append(userDetails, *NewUserDetails(&userDetail))
	}
	return UserDetailsList{Users: userDetails}

}

func NewUserDetails(ud *models.UserDetails) *UserDetails {
	return &UserDetails{TgId: ud.ChatId,
		ChatId:     ud.ChatId,
		TgUsername: ud.TgName,
		CreatedAt:  ud.CreatedAt,
		UpdatedAt:  ud.UpdatedAt,
	}
}

func NewCompletionRequestMessageSlice(req *CompletionRequestListData) []llm.CompletionRequestMessage {
	completions := []llm.CompletionRequestMessage{}
	for _, r := range req.Completions {
		completions = append(completions, *NewCompletionRequestMessage(&r))
	}
	return completions
}

func NewCompletionRequestMessage(req *CompletionRequest) *llm.CompletionRequestMessage {
	return &llm.CompletionRequestMessage{Text: req.Text, Role: req.Role}
}
