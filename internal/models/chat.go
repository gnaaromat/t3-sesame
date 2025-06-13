package models

import (
	"database/sql"
	"time"
)

type MessageTree struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	AIID      *int      `json:"ai_id" db:"ai_id"`
	Title     string    `json:"title" db:"title"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Message struct {
	ID            int       `json:"id" db:"id"`
	MessageTreeID int       `json:"message_tree_id" db:"message_tree_id"`
	Content       string    `json:"content" db:"content"`
	IsIncoming    bool      `json:"is_incoming" db:"is_incoming"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type ChatService struct {
	db *sql.DB
}

func NewChatService(db *sql.DB) *ChatService {
	return &ChatService{db: db}
}

func (s *ChatService) GetUserMessageTrees(userID int) ([]MessageTree, error) {
	query := `
        SELECT id, user_id, ai_id, title, created_at, updated_at
        FROM message_trees 
        WHERE user_id = $1 
        ORDER BY updated_at DESC
    `

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trees []MessageTree
	for rows.Next() {
		var tree MessageTree
		err := rows.Scan(&tree.ID, &tree.UserID, &tree.AIID, &tree.Title,
			&tree.CreatedAt, &tree.UpdatedAt)
		if err != nil {
			return nil, err
		}
		trees = append(trees, tree)
	}

	return trees, nil
}

func (s *ChatService) CreateMessageTree(userID int) (*MessageTree, error) {
	tree := &MessageTree{
		UserID: userID,
		Title:  "New Chat",
	}

	query := `
        INSERT INTO message_trees (user_id, title)
        VALUES ($1, $2)
        RETURNING id, created_at, updated_at
    `

	err := s.db.QueryRow(query, tree.UserID, tree.Title).
		Scan(&tree.ID, &tree.CreatedAt, &tree.UpdatedAt)

	return tree, err
}

func (s *ChatService) GetMessagesByTreeID(treeID int) ([]Message, error) {
	query := `
        SELECT id, message_tree_id, content, is_incoming, created_at
        FROM messages 
        WHERE message_tree_id = $1 
        ORDER BY created_at ASC
    `

	rows, err := s.db.Query(query, treeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.MessageTreeID, &msg.Content,
			&msg.IsIncoming, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (s *ChatService) GetMessageTree(treeID, userID int) (*MessageTree, error) {
	tree := &MessageTree{}
	query := `
        SELECT id, user_id, ai_id, title, created_at, updated_at
        FROM message_trees 
        WHERE id = $1 AND user_id = $2
    `

	err := s.db.QueryRow(query, treeID, userID).Scan(
		&tree.ID, &tree.UserID, &tree.AIID, &tree.Title,
		&tree.CreatedAt, &tree.UpdatedAt)

	return tree, err
}

func (s *ChatService) SaveMessage(treeID int, content string, isIncoming bool) (*Message, error) {
	msg := &Message{
		MessageTreeID: treeID,
		Content:       content,
		IsIncoming:    isIncoming,
	}

	query := `
        INSERT INTO messages (message_tree_id, content, is_incoming)
        VALUES ($1, $2, $3)
        RETURNING id, created_at
    `

	err := s.db.QueryRow(query, msg.MessageTreeID, msg.Content, msg.IsIncoming).
		Scan(&msg.ID, &msg.CreatedAt)

	// Update the message tree's updated_at timestamp
	if err == nil {
		s.db.Exec("UPDATE message_trees SET updated_at = NOW() WHERE id = $1", treeID)
	}

	return msg, err
}
