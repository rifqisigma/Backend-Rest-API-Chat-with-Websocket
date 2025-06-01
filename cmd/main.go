package main

import (
	"api_chat_ws/cmd/database"
	"api_chat_ws/cmd/route"
	"api_chat_ws/internal/handler"
	"api_chat_ws/internal/repository"
	"api_chat_ws/internal/usecase"
	"api_chat_ws/ws"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("env : %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("db : %v", err)
	}

	userRepo := repository.NewAuthRepo(db)
	userUsecase := usecase.NewAuthUsecase(userRepo)
	userHandler := handler.NewAuthHandler(userUsecase)
	chatRepo := repository.NewChatRepository(db)
	chatUsecase := usecase.NewChatUsecase(chatRepo)

	hub := ws.NewHub()
	go hub.Run()

	ChatHandler := handler.NewChatHandler(
		hub,
		chatUsecase,
	)

	r := route.SetupRoute(userHandler, ChatHandler)

	port := os.Getenv("PORT")
	fmt.Println("server berjalan pada port:" + port)
	http.ListenAndServe(":"+port, r)

}
