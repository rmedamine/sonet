package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"social/api"
	"social/config"
	"social/utils"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow frontend domain (Adjust as needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Change "*" if needed

		// Allow credentials (cookies, Authorization header)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Allow all headers
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		// Allow common HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// ✅ Handle OPTIONS preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r) // Continue to next handler
	})
}

func main() {
	if err := utils.InitServices(); err != nil {
		log.Fatal(err)
	}
	defer config.DB.Close()
	go api.NotificationWorker()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/users", api.UserSearch)
	mux.HandleFunc("/api/home", api.IndexHandler)
	mux.HandleFunc("/api/register", api.RegisterApi)
	mux.HandleFunc("/api/profile/{userId}/upload/avatar", api.Upload)
	mux.HandleFunc("/api/login", api.LoginApi)
	mux.HandleFunc("/api/logout", api.LogoutApi)
	mux.HandleFunc("/api/auth/me", api.GetUser)
	mux.HandleFunc("POST /api/post", api.PostApi)
	mux.HandleFunc("/api/posts", api.GetPostsApi)
	mux.HandleFunc("/api/post/{id}", api.GetPostApi)
	mux.HandleFunc("/api/react/post/{postId}", api.ReactToPost)
	mux.HandleFunc("/api/react/comment/{commentId}", api.ReactToComment)
	mux.HandleFunc("/api/add/comment", api.AddComment)
	mux.HandleFunc("/api/chat/", api.MessageApi)
	mux.HandleFunc("/api/chats/", api.PrivateChat)

	mux.HandleFunc("/api/profile/{userId}", api.Profile)
	mux.HandleFunc("/api/profile/update", api.Update)

	//--------------- follow/unfollow ----------------
	mux.HandleFunc("POST /api/follow/{userId}", api.FollowUser)
	mux.HandleFunc("/api/follow/requests/{userId}/accept", api.AcceptFollowRequest)
	mux.HandleFunc("/api/follow/requests/{userId}/reject", api.RejectRequest)
	mux.HandleFunc("/api/unfollow/{userId}", api.Unfollow)
	mux.HandleFunc("/api/remove/follower/{userId}", api.RemoveFollower)
	mux.HandleFunc("/api/followers/{userId}", api.GetFollowers)
	mux.HandleFunc("/api/followings/{userId}", api.GetFollowing)
	mux.HandleFunc("GET /api/follow/requests", api.GetFollowRequests)
	//  ------------- GROUPS ----------------
	mux.HandleFunc("/api/group/create", api.CreateGroup)
	mux.HandleFunc("/api/group/{groupId}", api.GetGroup)
	mux.HandleFunc("/api/groups", api.GetGroups)
	mux.HandleFunc("/api/group/{groupId}/request", api.RequestJoinGroup)
	mux.HandleFunc("/api/group/{groupId}/accept/{userId}", api.AcceptJoinRequest)
	mux.HandleFunc("/api/group/{groupId}/reject/{userId}", api.RejectJoinRequest)
	mux.HandleFunc("/api/group/{groupId}/leave", api.LeaveGroup)
	mux.HandleFunc("/api/group/{groupId}/invite/{userId}", api.InviteToGroup)
	mux.HandleFunc("/api/group/{groupId}/invitations/{userId}/accept", api.AcceptInvite)
	mux.HandleFunc("/api/group/{groupId}/invitations/{userId}/reject", api.RejectInvite)
	mux.HandleFunc("/api/users/{userId}/group-invites", api.GetUserGroupInvites)
	mux.HandleFunc("/api/group/{groupId}/members/{userId}/kick", api.KickMember)
	mux.HandleFunc("/api/group/{groupId}/delete", api.DeleteGroup)
	mux.HandleFunc("/api/group/{groupId}/events", api.GetGroupEvents)
	mux.HandleFunc("/api/group/{groupId}/event", api.CreateGroupEvent)
	// mux.HandleFunc("/api/group/{groupId}/event/{eventId}", api.GetGroupEvent)
	mux.HandleFunc("/api/group/{groupId}/event/respond", api.RespondToGroupEvent)
	mux.HandleFunc("/api/notification", api.GetNotifications)
	mux.HandleFunc("/api/notification/{notificationId}/read", api.ReadNotification)
	mux.HandleFunc("/api/notification/{notificationId}/clear", api.ClearNotification)
	mux.HandleFunc("/api/notification/clear_all", api.ClearAllNotifications)
	mux.HandleFunc("/api/notification/read_all", api.MarkAllAsRead)

	//  ------------- GROUPS POSTS ----------------

	// Group post endpoints
	mux.HandleFunc("/api/group-posts/create/{groupId}", api.HandleGroupPost) // Create a new post in a group
	mux.HandleFunc("/api/group-posts/list/{groupId}", api.GetGroupPostsApi)  // Get all posts for a group with pagination
	mux.HandleFunc("/api/group-posts/{id}", api.GetGroupPostApi)
	mux.HandleFunc("/api/group-posts/post/{id}/react", api.ReactToGroupPost)
	// Group post comment endpoints
	mux.HandleFunc("/api/group-posts/add/comment", api.AddGroupComment)
	// mux.HandleFunc("/api/group-posts/{postId}/{commentId}", api.ReactToComment)

	// Group post reaction endpoints
	mux.HandleFunc("/api/group-posts/comment/{id}/react", api.ReactToGroupComment)
	mux.HandleFunc("/api/users/search", api.UsersSearch)

	mux.HandleFunc("/ws", api.WsEndpoint)

	// serve uploads folder
	mux.HandleFunc("/uploads/", func(w http.ResponseWriter, r *http.Request) {
		filePath := "." + r.URL.Path
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			utils.WriteJSON(w, http.StatusNotFound, "File Not found", nil)
			return
		} else if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Internal Server Error", nil)
			return
		}
		http.ServeFile(w, r, filePath)
	})

	fmt.Printf("Server running on http://localhost%v\n", config.ADDRS)
	err := http.ListenAndServe(config.ADDRS, corsMiddleware(mux))
	if err != nil {
		log.Fatal(err)
	}
}
