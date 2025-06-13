package handlers

import (
	"net/http"
	"strconv"
	"t3sesame/internal/models"
	"t3sesame/internal/templates"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type ChatHandler struct {
	chatService *models.ChatService
}

func NewChatHandler(chatService *models.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) ShowMainInterface(c echo.Context) error {
	sess, _ := session.Get("session", c)
	userID := sess.Values["user_id"].(int)
	username := sess.Values["username"].(string)

	trees, err := h.chatService.GetUserMessageTrees(userID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load conversations")
	}

	return templates.MainLayout(username, trees).Render(c.Request().Context(), c.Response().Writer)
}

func (h *ChatHandler) GetChatMessages(c echo.Context) error {
	sess, _ := session.Get("session", c)
	userID := sess.Values["user_id"].(int)

	treeIDStr := c.Param("tree_id")
	treeID, err := strconv.Atoi(treeIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid tree ID")
	}

	// Verify ownership
	tree, err := h.chatService.GetMessageTree(treeID, userID)
	if err != nil {
		return c.String(http.StatusNotFound, "Conversation not found")
	}

	messages, err := h.chatService.GetMessagesByTreeID(treeID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load messages")
	}

	return templates.MessageDisplay(*tree, messages).Render(c.Request().Context(), c.Response().Writer)
}

func (h *ChatHandler) CreateNewChat(c echo.Context) error {
	sess, _ := session.Get("session", c)
	userID := sess.Values["user_id"].(int)

	tree, err := h.chatService.CreateMessageTree(userID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create new chat")
	}

	return templates.NewChatCreated(*tree).Render(c.Request().Context(), c.Response().Writer)
}

func (h *ChatHandler) SendMessage(c echo.Context) error {
	sess, _ := session.Get("session", c)
	userID := sess.Values["user_id"].(int)

	treeIDStr := c.Param("tree_id")
	treeID, err := strconv.Atoi(treeIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid tree ID")
	}

	// Verify ownership
	_, err = h.chatService.GetMessageTree(treeID, userID)
	if err != nil {
		return c.String(http.StatusNotFound, "Conversation not found")
	}

	content := c.FormValue("content")
	if content == "" {
		return c.String(http.StatusBadRequest, "Message content is required")
	}

	// Save user message
	userMsg, err := h.chatService.SaveMessage(treeID, content, false)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save message")
	}

	// For now, just echo back a simple response
	// In the next user story, we'll integrate with AI
	aiResponse := "I received your message: \"" + content + "\". AI integration coming soon!"
	aiMsg, err := h.chatService.SaveMessage(treeID, aiResponse, true)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save AI response")
	}

	// Return both messages
	c.Response().Writer.Write([]byte(`<div>`))
	templates.MessageBubble(*userMsg).Render(c.Request().Context(), c.Response().Writer)
	templates.MessageBubble(*aiMsg).Render(c.Request().Context(), c.Response().Writer)
	c.Response().Writer.Write([]byte(`</div>`))

	return nil
}
