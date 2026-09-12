package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/util"
)

// Message 实时消息结构。
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Hub 社区实时消息广播中心（gorilla/websocket）。
type Hub struct {
	mu        sync.Mutex
	clients   map[*websocket.Conn]bool
	logger    *slog.Logger
	jwtSecret string
}

// NewHub 构造消息中心。
func NewHub(logger *slog.Logger, jwtSecret string) *Hub {
	return &Hub{clients: make(map[*websocket.Conn]bool), logger: logger, jwtSecret: jwtSecret}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Broadcast 向所有客户端广播消息。
func (h *Hub) Broadcast(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error(constants.LogInternalError, "err", err)
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			h.logger.Warn(constants.LogInternalError, "err", err)
			conn.Close()
			delete(h.clients, conn)
		}
	}
}

// Handle 处理 WebSocket 连接（token 通过 ?token= 查询参数校验，浏览器 WebSocket 无法携带请求头）。
func (h *Hub) Handle(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": constants.CodeUnauthorized, "message": constants.MsgLoginRequired})
		return
	}
	claims, err := util.ParseToken(h.jwtSecret, token)
	if err != nil {
		h.logger.Warn(constants.LogAuthTokenInvalid, "request_id", util.GetRequestID(c), "err", err)
		c.JSON(http.StatusUnauthorized, gin.H{"code": constants.CodeUnauthorized, "message": constants.ErrorText[constants.CodeUnauthorized]})
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Warn(constants.LogInternalError, "err", err)
		return
	}
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()
	h.logger.Info(constants.LogPostCreated, "user_id", claims.UserID, "username", claims.Username, "type", "ws_connected")
	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()
	}()
	// 保持连接并读取客户端消息（客户端发送 ping 维持心跳）
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
