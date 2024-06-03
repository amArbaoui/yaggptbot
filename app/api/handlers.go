package api

import (
	"amArbaoui/yaggptbot/app/user"
	"encoding/json"
	"fmt"
	"net/http"
)

type UserHandler struct {
	uservice *user.UserServiceImpl
}

func (usr *UserHandler) NewUser(w http.ResponseWriter, r *http.Request) {
	var newUserRequest NewUserRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&newUserRequest)
	if err != nil {
		http.Error(w, "Unprocessable request", http.StatusUnprocessableEntity)
		return
	}
	err = usr.uservice.SaveUser(NewUserFromAddUserRequest(newUserRequest))
	if err != nil {
		errResp := fmt.Sprintf("failed to create user: %s, error: %s", newUserRequest.TgUsername, err.Error())
		http.Error(w, errResp, http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "user %s created\n", newUserRequest.TgUsername)

}
