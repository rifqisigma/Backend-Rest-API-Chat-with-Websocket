package handler

import (
	"api_chat_ws/dto"
	"api_chat_ws/helper/middleware"
	"api_chat_ws/helper/utils"
	"api_chat_ws/internal/usecase"
	"api_chat_ws/ws"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	hub     *ws.Hub
	usecase usecase.ChatUsecase
}

func NewChatHandler(hub *ws.Hub, usecase usecase.ChatUsecase) *WebSocketHandler {
	return &WebSocketHandler{hub, usecase}
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("err %v", err)
		return
	}

	params := mux.Vars(r)
	groupID, err := strconv.Atoi(params["group_id"])
	if err != nil {
		log.Println("Gagal parse group_id:", err)
		return
	}

	userId, err := strconv.Atoi(params["user_id"])
	if err != nil {
		log.Println("Gagal parse group_id:", err)

		return
	}

	memberId, err := h.usecase.GetMemberId(uint(userId), uint(groupID))
	if err != nil {
		switch err {
		case utils.ErrNotMember:
			fmt.Printf("err %v", err)
			return

		default:
			fmt.Printf("err %v", err)

			return
		}
	}

	client := &ws.Client{
		MemberId: memberId,
		GroupID:  uint(groupID),
		Conn:     conn,
		Send:     make(chan []byte, 256), // buffer biar nggak nge-block
	}

	fmt.Print(client.MemberId)

	h.hub.Register <- client

	// Jalankan pump tulis dan baca
	go client.WritePump()
	go client.ReadPump(h.hub, h.usecase)
}

func (h *WebSocketHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, valid := claimsRaw.(*utils.JWTCLAIMS)
	if !valid {
		utils.WriteError(w, http.StatusUnauthorized, "invalid jwt")
		return
	}
	var req dto.CreateGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	req.UserId = claims.UserID
	if err := h.usecase.CreateGroup(&req); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())

		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *WebSocketHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, valid := claimsRaw.(*utils.JWTCLAIMS)
	if !valid {
		utils.WriteError(w, http.StatusUnauthorized, "invalid jwt")
		return
	}

	params := mux.Vars(r)
	paramsGroupid, _ := strconv.Atoi(params["groupId"])

	var req dto.UpdateGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	memberId, err := h.usecase.GetMemberId(claims.UserID, uint(paramsGroupid))
	if err != nil {
		switch err {
		case utils.ErrNotMember:
			fmt.Printf("err %v", err)
			return

		default:
			fmt.Printf("err %v", err)

			return
		}
	}

	req.GroupId = uint(paramsGroupid)
	req.MemberId = memberId
	if err := h.usecase.UpdateGroup(&req); err != nil {
		switch err {
		case utils.ErrNotAdmin:
			utils.WriteError(w, http.StatusForbidden, err.Error())
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *WebSocketHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, valid := claimsRaw.(*utils.JWTCLAIMS)
	if !valid {
		utils.WriteError(w, http.StatusUnauthorized, "invalid jwt")
		return
	}

	params := mux.Vars(r)
	paramsGroupid, _ := strconv.Atoi(params["groupId"])

	memberId, err := h.usecase.GetMemberId(claims.UserID, uint(paramsGroupid))
	if err != nil {
		switch err {
		case utils.ErrNotMember:
			fmt.Printf("err %v", err)
			return

		default:
			fmt.Printf("err %v", err)

			return
		}
	}

	if err := h.usecase.DeleteGroup(memberId, uint(paramsGroupid)); err != nil {
		switch err {
		case utils.ErrNotAdmin:
			utils.WriteError(w, http.StatusForbidden, err.Error())
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *WebSocketHandler) AddMembers(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, valid := claimsRaw.(*utils.JWTCLAIMS)
	if !valid {
		utils.WriteError(w, http.StatusUnauthorized, "invalid jwt")
		return
	}

	params := mux.Vars(r)
	paramsGroupid, _ := strconv.Atoi(params["groupId"])

	var req dto.AddMemberReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	memberId, err := h.usecase.GetMemberId(claims.UserID, uint(paramsGroupid))
	if err != nil {
		switch err {
		case utils.ErrNotMember:
			fmt.Printf("err %v", err)
			return

		default:
			fmt.Printf("err %v", err)

			return
		}
	}

	req.AdminId = memberId
	req.GroupId = uint(paramsGroupid)
	if err := h.usecase.AddMember(&req); err != nil {
		switch err {
		case utils.ErrNotAdmin:
			utils.WriteError(w, http.StatusForbidden, err.Error())
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *WebSocketHandler) RemoveMembers(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, valid := claimsRaw.(*utils.JWTCLAIMS)
	if !valid {
		utils.WriteError(w, http.StatusUnauthorized, "invalid jwt")
		return
	}

	params := mux.Vars(r)
	paramsGroupid, _ := strconv.Atoi(params["groupId"])

	var req dto.RemoveMemberReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	memberId, err := h.usecase.GetMemberId(claims.UserID, uint(paramsGroupid))
	if err != nil {
		switch err {
		case utils.ErrNotMember:
			fmt.Printf("err %v", err)
			return

		default:
			fmt.Printf("err %v", err)

			return
		}
	}

	req.AdminId = memberId
	if err := h.usecase.RemoveMember(req.UserIds, req.AdminId); err != nil {
		switch err {
		case utils.ErrNotAdmin:
			utils.WriteError(w, http.StatusForbidden, err.Error())
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *WebSocketHandler) ExitGroup(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, valid := claimsRaw.(*utils.JWTCLAIMS)
	if !valid {
		utils.WriteError(w, http.StatusUnauthorized, "invalid jwt")
		return
	}

	params := mux.Vars(r)
	paramsGroupid, _ := strconv.Atoi(params["groupId"])

	memberId, err := h.usecase.GetMemberId(claims.UserID, uint(paramsGroupid))
	if err != nil {
		switch err {
		case utils.ErrNotMember:
			fmt.Printf("err %v", err)
			return

		default:
			fmt.Printf("err %v", err)

			return
		}
	}
	if err := h.usecase.ExitGroup(memberId); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *WebSocketHandler) UpdateRoleMembers(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.UserContextKey)
	claims, valid := claimsRaw.(*utils.JWTCLAIMS)
	if !valid {
		utils.WriteError(w, http.StatusUnauthorized, "invalid jwt")
		return
	}

	params := mux.Vars(r)
	paramsGroupid, _ := strconv.Atoi(params["groupId"])

	var req dto.UpdateRoleMember
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	memberId, err := h.usecase.GetMemberId(claims.UserID, uint(paramsGroupid))
	if err != nil {
		switch err {
		case utils.ErrNotMember:
			fmt.Printf("err %v", err)
			return

		default:
			fmt.Printf("err %v", err)

			return
		}
	}

	req.AdminId = memberId
	req.GroupId = uint(paramsGroupid)
	if err := h.usecase.UpdateRoleUser(&req); err != nil {
		switch err {
		case utils.ErrNotAdmin:
			utils.WriteError(w, http.StatusForbidden, err.Error())
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}
