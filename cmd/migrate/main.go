package main

import (
	"api_chat_ws/cmd/database"
	"api_chat_ws/model"
	"log"
)

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.ChatGroup{}, &model.GroupMember{}, &model.Chat{}); err != nil {
		log.Fatalf("error migrasi : %v", err)
	}

	log.Println("Migrasi berhasil")
}
