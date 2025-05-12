package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"social/models"

	"social/config"
	"social/utils"

	"github.com/gorilla/websocket"
)

var NC *ConnectionMap

type Response struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

var NotificationQueue = make(chan *models.Notification, 100)

func NotificationWorker() {
	for notification := range NotificationQueue {
		notificationRepo := models.NewNotificationRepository()
		err := notificationRepo.CreateNotification(notification)
		if err != nil {
			fmt.Println("Error creating notification:", err)
			continue
		}
		SendNotification(notification)
	}
}

var (
	groupRepo  = models.NewGroupRepository()
	chatRepo   = models.NewChatRepository()
	userRepo   = models.NewUserRepository()
	followRepo = models.NewFollowRepository()
)

func init() {
	NC = NewConnectionMap()
}

type ConnectionMap struct {
	connections map[int][]*websocket.Conn
	mu          sync.Mutex
}

func NewConnectionMap() *ConnectionMap {
	return &ConnectionMap{
		connections: make(map[int][]*websocket.Conn),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewResponse(event string, data map[string]any) []byte {
	response := Response{Type: event, Data: data}
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Println("Error marshalling response:", err)
		return []byte{}
	}
	return jsonData
}

func (NC *ConnectionMap) AddConnection(userID int, conn *websocket.Conn) {
	NC.mu.Lock()
	defer NC.mu.Unlock()

	_, exists := NC.connections[userID]
	NC.connections[userID] = append(NC.connections[userID], conn)
	if !exists {
		// Notify other users that this user is online
		onlineMessage := NewResponse("online", map[string]any{"id": []int{userID}})
		for key := range NC.connections {
			NC.BroadcastMessage(key, onlineMessage, websocket.TextMessage)
		}
	}
	// Send the list of connected users to the new user
	connectedUsers := []int{}
	for user := range NC.connections {
		if user != userID {
			connectedUsers = append(connectedUsers, user)
		}
	}
	NC.SendToConn(conn, NewResponse("online", map[string]any{"id": connectedUsers}), websocket.TextMessage)
}

func (NC *ConnectionMap) RemoveConnection(userID int, conn *websocket.Conn) {
	NC.mu.Lock()
	defer NC.mu.Unlock()

	if conns, exists := NC.connections[userID]; exists {
		for i, c := range conns {
			if c == conn {
				NC.connections[userID] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
		if len(NC.connections[userID]) == 0 {
			delete(NC.connections, userID)
			offlineMessage := NewResponse("offline", map[string]any{"id": userID})
			for key := range NC.connections {
				NC.BroadcastMessage(key, offlineMessage, websocket.TextMessage)
			}
		}
	}
}

func (NC *ConnectionMap) SendToConn(conn *websocket.Conn, msg []byte, messageType int) error {
	return conn.WriteMessage(messageType, msg)
}

func (NC *ConnectionMap) BroadcastMessage(userID int, msg []byte, messageType int) error {
	conns, exists := NC.connections[userID]
	if !exists {
		return nil
	}
	for _, conn := range conns {
		if err := conn.WriteMessage(messageType, msg); err != nil {
			log.Println("Error broadcasting message:", err)
		}
	}
	return nil
}

func (NC *ConnectionMap) BroadcastGroupMessage(groupID int, msg []byte, messageType int) error {
	NC.mu.Lock()
	defer NC.mu.Unlock()

	// Get all members of the group
	groupRepo := models.NewGroupRepository()
	members, err := groupRepo.GetGroupMembers(int64(groupID))
	if err != nil {
		log.Println("Error getting group members:", err)
		return err
	}

	// Send message to all online group members
	for _, member := range members {
		if conns, exists := NC.connections[int(member.ID)]; exists {
			for _, conn := range conns {
				if err := conn.WriteMessage(messageType, msg); err != nil {
					log.Println("Error broadcasting group message:", err)
				}
			}
		}
	}
	return nil
}

func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		token = utils.GetSessionCookie(r)
	}
	session, err := config.SESSION.GetSession(token)
	if err != nil || session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupRepo = models.NewGroupRepository()
	chatRepo = models.NewChatRepository()
	userRepo = models.NewUserRepository()
	followRepo = models.NewFollowRepository()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	userID := int(session.UserId)
	NC.AddConnection(userID, conn)
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket error:", err)
			break
		}
		var readData Response
		if err := json.Unmarshal(msg, &readData); err != nil {
			log.Println("Error parsing message:", err)
			continue
		}
		switch readData.Type {
		case "privatetyping", "groupTyping":
			readData.Data["typer"] = session.Email
			readData.Data["typerId"] = userID
			typingMessage, _ := json.Marshal(readData)
			if receiverID, ok := readData.Data["receiver"].(float64); ok {
				NC.BroadcastMessage(int(receiverID), typingMessage, messageType)
			} else if groupID, ok := readData.Data["group"].(float64); ok {
				NC.BroadcastMessage(int(groupID), typingMessage, messageType)
			}
		case "privateMessage":
			receiverID, ok := readData.Data["receiver"].(float64)
			if !ok {
				log.Println("Invalid receiver ID")
				continue
			}
			fmt.Println("receiverid", receiverID)
			if !isValidChat(userID, int(receiverID)) {
				continue
			}
			readData.Data["sender"] = userID
			readData.Data["timestamp"] = time.Now()

			messageContent := readData.Data["message"].(string)
			if len(messageContent) == 0 || strings.TrimSpace(messageContent) == "" {
				NC.SendToConn(conn, []byte("Errro"), websocket.TextMessage)
				return
			}

			msgStruct := savePrivateMessage(userID, int(receiverID), readData.Data["message"].(string))

			bytes, _ := json.Marshal(msgStruct)
			json.Unmarshal(bytes, &readData.Data)

			msg, _ = json.Marshal(readData)
			NC.BroadcastMessage(int(receiverID), msg, messageType)
			NC.BroadcastMessage(userID, msg, messageType)
			MessageNotify(int64(receiverID), int64(userID))
		case "groupMessage":
			groupID, ok := readData.Data["group"].(float64)
			// print the data received
			fmt.Println("groupID", readData.Data)
			if !ok {
				log.Println("Invalid group ID")
				continue
			}
			if !isGroupMember(userID, int(groupID)) {
				continue
			}

			messageContent := readData.Data["message"].(string)
			if len(messageContent) == 0 || strings.TrimSpace(messageContent) == "" {
				NC.SendToConn(conn, []byte("Errro"), websocket.TextMessage)
				return
			}

			msg := saveGroupMessage(userID, int(groupID), readData.Data["message"].(string))

			// Create a new response with the complete message data
			response := Response{
				Type: "groupMessage",
				Data: map[string]interface{}{
					"id":        msg.Id,
					"senderId":  userID,
					"group":     groupID,
					"message":   msg.Message,
					"timestamp": msg.SentAt,
				},
			}

			msgBytes, _ := json.Marshal(response)
			NC.BroadcastGroupMessage(int(groupID), msgBytes, messageType)
		case "groupJoin":
			groupID, ok := readData.Data["group"].(float64)
			if !ok {
				log.Println("Invalid group ID")
				continue
			}
			NC.AddConnection(int(groupID), conn)
		case "groupLeave":
			groupID, ok := readData.Data["group"].(float64)
			if !ok {
				log.Println("Invalid group ID")
				continue
			}
			NC.RemoveConnection(int(groupID), conn)
		}
	}
	NC.RemoveConnection(userID, conn)
}

func isValidChat(senderID, receiverID int) bool {
	if _, err := userRepo.UserExistsById(int64(receiverID)); err != nil {
		log.Println("Receiver does not exist")
		return false
	}
	isFollower, _ := followRepo.IsFollowing(int64(senderID), int64(receiverID))
	if !isFollower {
		isFollower, _ = followRepo.IsFollowing(int64(receiverID), int64(senderID))
	}
	return isFollower
}

func isGroupMember(userID, groupID int) bool {
	memberRepo := models.NewGroupRepository()
	memberExists, err := memberRepo.IsUserMember(int64(userID), int64(groupID))
	if err != nil {
		log.Println("Error checking group membership:", err)
	}
	return memberExists
}

func savePrivateMessage(senderID, receiverID int, message string) models.PrivateMessage {
	msg := models.PrivateMessage{
		SenderId:   int64(senderID),
		ReceiverId: int64(receiverID),
		Message:    message,
		CreatedAt:  time.Now().Format("2006-01-02 15:04:05"),
	}
	chatRepo.CreatePrivateMessage(&msg)
	return msg
}

func saveGroupMessage(senderID, groupID int, message string) models.GroupChat {
	msg := models.GroupChat{
		SenderId: int64(senderID),
		GroupId:  int64(groupID),
		Message:  message,
		SentAt:   time.Now().Format("2006-01-02 15:04:05"),
	}
	chatRepo.CreateGroupChat(&msg)
	return msg
}

func SendNotification(n *models.Notification) {
	notificationData := NewResponse("notification", map[string]any{"notification": n})
	NC.mu.Lock()
	defer NC.mu.Unlock()
	for _, conn := range NC.connections[int(n.UserID)] {
		conn.WriteMessage(websocket.TextMessage, notificationData)
	}
}
