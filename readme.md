# ✅ Golang Chat Application – Features & Development Phases

## 🟢 Phase 1: Basic CLI Chat App
- [x] CLI interface for chat
- [ ] Single server handling multiple clients
- [ ] Broadcast messages to all users
- [ ] Use `net` package and Goroutines for concurrency

---

## 🟡 Phase 2: WebSocket Chat App
- [ ] WebSocket-based real-time messaging
- [ ] Basic web UI (HTML + JavaScript)
- [ ] Display user join/leave notifications

---

## 🟠 Phase 3: Chat Rooms & In-Memory State
- [ ] Support multiple chat rooms
- [ ] Assign usernames or nicknames
- [ ] Maintain in-memory message history per room
- [ ] Use router (`gorilla/mux`, `gin`, or `fiber`)

---

## 🔵 Phase 4: Authentication & Persistence
- [ ] User registration and login system
- [ ] JWT or session-based authentication
- [ ] Store chat messages in a database (PostgreSQL / MongoDB)
- [ ] Optional ORM integration (e.g., GORM)

---

## 🔴 Phase 5: Advanced Features
- [ ] Private messaging (DMs)
- [ ] File sharing support (local or cloud)
- [ ] Typing indicators and read receipts
- [ ] Dockerized deployment
- [ ] Reverse proxy setup (Nginx or Kubernetes)

---

## 🧪 Optional Enhancements
- [ ] Reconnect logic for clients
- [ ] Admin panel for user/room moderation
- [ ] Emoji and markdown support in messages
- [ ] Push notifications

---

# How to Run
There are two commands to run, client and server. 
Server will return any message client send. 
For now, client will send message every second.

```
// Run server
go run ./cmd/server/main.go

// Run client
go run ./cmd/client/main.go
```