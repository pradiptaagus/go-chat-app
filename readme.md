# ✅ Golang Chat Application

# How to Run
There are two commands to run, client and server. 
Server will return any message client send. 
For now, client will send message every second.

```
// Run server
go run ./cmd/server/main.go

// Run client (optional)
// Client will send a message every tick and then a receive message
go run ./cmd/client/main.go
```
After running the server, you can visit chat client interface at [http://localhost:8080](http://localhost:8080).
try to open the connection first, then start to send a message.
You can open browser console to see more information about the log.

# Development Roadmap

## 🟢 Phase 1: Basic CLI Chat App
- [x] CLI interface for chat
- [x] Single server handling multiple clients
- [x] Broadcast messages to all users
- [x] Use `net` package and Goroutines for concurrency

---

## 🟡 Phase 2: WebSocket Chat App
- [x] WebSocket-based real-time messaging
- [x] Basic web UI (HTML + JavaScript)
- [x] Display user join/leave notifications

---

## 🟠 Phase 3: Chat Rooms & In-Memory State
- [x] Support multiple chat rooms
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