package route

import (
	"api_chat_ws/helper/middleware"
	"api_chat_ws/internal/handler"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRoute(userHandler *handler.AuthHandler, ChatHandler *handler.WebSocketHandler) *mux.Router {
	r := mux.NewRouter()
	user := r.PathPrefix("/user").Subrouter()
	user.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost)
	user.HandleFunc("/register", userHandler.Register).Methods(http.MethodPost)

	chatws := r.PathPrefix("/x").Subrouter()

	chatws.HandleFunc("/ws/{group_id}/{user_id}", ChatHandler.ServeWS)

	chatM := r.PathPrefix("/chat").Subrouter()
	chatM.Use(middleware.AuthMiddleware)
	chatM.HandleFunc("/stream/{group_id}", ChatHandler.ServeWS)

	chatG := chatM.PathPrefix("/group").Subrouter()
	chatG.HandleFunc("/create", ChatHandler.CreateGroup).Methods(http.MethodPost)
	chatG.HandleFunc("/update/{groupId}", ChatHandler.UpdateGroup).Methods(http.MethodPut)
	chatG.HandleFunc("/delete/{groupId}", ChatHandler.DeleteGroup).Methods(http.MethodDelete)
	chatG.HandleFunc("/add-members/{groupId}", ChatHandler.AddMembers).Methods(http.MethodPost)
	chatG.HandleFunc("/remove-members/{groupId}", ChatHandler.RemoveMembers).Methods(http.MethodPost)
	chatG.HandleFunc("/exit-group/{groupId}", ChatHandler.ExitGroup).Methods(http.MethodDelete)
	chatG.HandleFunc("/update-role-members/{groupId}", ChatHandler.UpdateRoleMembers).Methods(http.MethodPut)

	return r
}
