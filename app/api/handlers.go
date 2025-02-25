package api

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/telegram"
	"amArbaoui/yaggptbot/app/user"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	uservice    *user.UserServiceImpl
	chatService telegram.ChatService
}

func (usr *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var newUserRequest UserRequest
	silent := r.URL.Query().Get("silent")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&newUserRequest)
	if err != nil {
		ErrorResponse(err, "unprocessable request", http.StatusUnprocessableEntity, w, r)
		return
	}
	err = usr.uservice.SaveUser(NewUserFromAddUserRequest(newUserRequest))
	if err != nil {
		errResp := fmt.Sprintf("failed to create user: %s, error: %s", newUserRequest.TgUsername, err)
		ErrorResponse(err, errResp, http.StatusInternalServerError, w, r)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "user %s created\n", newUserRequest.TgUsername)
	if silent != "true" {
		usr.greetUser(newUserRequest)
	}

}

func (usr *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	userDetails, err := usr.uservice.GetUsersDetails()
	errResp := "failed to get users"
	if err != nil {
		ErrorResponse(err, errResp, http.StatusInternalServerError, w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserDetailsListFrom(userDetails))

}

func (usr *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var newUserRequest UserRequest
	errResp := fmt.Sprintf("failed to update user: %s", newUserRequest.TgUsername)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&newUserRequest)
	if err != nil {
		ErrorResponse(err, "unprocessable request", http.StatusUnprocessableEntity, w, r)
		return
	}
	err = usr.uservice.UpdateUser(NewUserFromUpdateUserRequest(newUserRequest))
	if err != nil {
		ErrorResponse(err, errResp, http.StatusUnprocessableEntity, w, r)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "user with id %f updated\n", newUserRequest.TgId)
}

func (usr *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	tgIdText := chi.URLParam(r, "TgId")
	errResp := fmt.Sprintf("failed to delete user %s", tgIdText)
	tgId, err := strconv.ParseInt(tgIdText, 10, 64)
	if err != nil {
		ErrorResponse(err, "invalid user ID", http.StatusInternalServerError, w, r)
		return
	}
	_, err = usr.uservice.GetUserByTgId(tgId)
	if err != nil {
		ErrorResponse(err, "user not found", http.StatusInternalServerError, w, r)
		return
	}

	err = usr.uservice.DeleteUser(tgId)

	if err != nil {
		ErrorResponse(err, errResp, http.StatusInternalServerError, w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "user with tgId %s deleted\n", tgIdText)
}

func (usr *UserHandler) greetUser(newUserRequest UserRequest) {
	_, err := usr.chatService.SendMessage(
		telegram.MessageOut{
			Text:     tgbotapi.EscapeText("Markdown", config.GreetUserMessage),
			ChatId:   int64(newUserRequest.ChatId),
			RepyToId: 0,
		},
	)
	if err != nil {
		log.Printf("[WARN] failed to send greeting message to %f due to %s", newUserRequest.TgId, err)
	}
	_, err = usr.chatService.SendMessage(
		telegram.MessageOut{
			Text:     tgbotapi.EscapeText("Markdown", config.HowToUseItMessage),
			ChatId:   int64(newUserRequest.ChatId),
			RepyToId: 0,
		},
	)
	if err != nil {
		log.Printf("[WARN] failed to send usage message to %f due to %s", newUserRequest.TgId, err)
	}

}

type LlmHandler struct {
	llmService *llm.LlmService
}

func (l *LlmHandler) GetCompletion(w http.ResponseWriter, r *http.Request) {
	var completionRequestHolder CompletionRequestListData
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&completionRequestHolder)
	if err != nil {
		ErrorResponse(err, "unprocessable request", http.StatusUnprocessableEntity, w, r)
		return
	}
	llmReq := NewCompletionRequestMessageSlice(&completionRequestHolder)
	resp, err := l.llmService.GetCompletionMessage(llmReq, "", config.DefaultModel) // TODO: select model
	if err != nil {
		ErrorResponse(err, "failed to get completion", http.StatusUnprocessableEntity, w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, resp)

}

func ErrorResponse(err error, errMsg string, status int, w http.ResponseWriter, r *http.Request) {
	log.Printf("%s, %s\n", errMsg, err)
	http.Error(w, errMsg, http.StatusInternalServerError)
}
