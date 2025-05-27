package pachca

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"time"
)

// Messages
// Объект для работы с сообщениями
type Messages struct {
	client *resty.Client
}

// Message
// Объект для сообщения pachca
type Message struct {
	EntityType         string            `json:"entity_type,omitempty"`
	EntityID           int               `json:"entity_id,omitempty"`
	Content            string            `json:"content,omitempty"`
	Files              []MessageFile     `json:"files,omitempty"`
	Buttons            [][]MessageButton `json:"buttons,omitempty"`
	ParentMessageID    *int              `json:"parent_message_id,omitempty"`
	DisplayAvatarURL   string            `json:"display_avatar_url"`
	DisplayName        string            `json:"display_name,omitempty"`
	SkipInviteMentions bool              `json:"skip_invite_mentions,omitempty"`
	LinkPreview        bool              `json:"link_preview,omitempty"`
}

// OutgoingMessage
// Объект для исходящего сообщения в pachca
type OutgoingMessage struct {
	Message *Message `json:"message"`
}

// MessageFile
// Вложенный объект для исходящего сообщения описывающий файл
type MessageFile struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	FileType string `json:"file_type"`
	Size     int    `json:"size"`
}

// MessageButton
// Вложенный объект для исходящего сообщения описывающий кнопку
type MessageButton struct {
	Text string `json:"text"`
	Data string `json:"data,omitempty"`
	URL  string `json:"url,omitempty"`
}

// MessageResponseRaw
// Объект для хранения ответа от API возвращающих одно сообщение
type MessageResponseRaw struct {
	Data MessageResponse `json:"data"`
}

// MessagesResponseRaw
// Объект для хранения ответа от API возвращающий список сообщений
type MessagesResponseRaw struct {
	Data []MessageResponse `json:"data"`
}

type MessageResponse struct {
	ID         int       `json:"id"`
	EntityType string    `json:"entity_type"`
	EntityID   int       `json:"entity_id"`
	ChatID     int       `json:"chat_id"`
	Content    string    `json:"content"`
	UserID     int       `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
	URL        string    `json:"url"`
	Files      []struct {
		ID       int    `json:"id"`
		Key      string `json:"key"`
		Name     string `json:"name"`
		FileType string `json:"file_type"`
		URL      string `json:"url"`
	} `json:"files"`
	Buttons [][]MessageButton
	Thread  *struct {
		ID     int `json:"id"`
		ChatID int `json:"chat_id"`
	} `json:"thread"`
	Forwarding *struct {
		OriginalMessageID          int  `json:"original_message_id"`
		OriginalChatID             int  `json:"original_chat_id"`
		AuthorID                   int  `json:"author_id"`
		OriginalCreatedAt          int  `json:"original_created_at"`
		OriginalThreadID           *int `json:"original_thread_id"`
		OriginalThreadMessageID    *int `json:"original_thread_message_id"`
		OriginalThreadParentChatID *int `json:"original_thread_parent_chat_id"`
	} `json:"forwarding"`
	ParentMessageID *int `json:"parent_message_id"`
}

func (m *Messages) Get(ctx context.Context, messageID int) (*MessageResponse, *resty.Response, error) {
	if messageID <= 0 {
		return nil, nil, fmt.Errorf("%w: incorrect message ID %d", ErrInvalidInput, messageID)
	}

	url := fmt.Sprintf("%v/%v", messagesURL, messageID)
	resp, err := m.client.R().
		SetContext(ctx).
		Get(url)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%v: %v", ErrResponseCode, resp.StatusCode())
	}

	var r *MessageResponseRaw
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, resp, fmt.Errorf("%v: %v", ErrResponseDecode, err)
	}

	return &r.Data, resp, nil
}

func (m *Messages) GetAll(ctx context.Context, chatID int) ([]MessageResponse, *resty.Response, error) {
	if chatID <= 0 {
		return nil, nil, fmt.Errorf("%v: incorrect chat ID %d", ErrInvalidInput, chatID)
	}

	var messages *MessagesResponseRaw

	resp, err := m.client.R().
		SetContext(ctx).
		SetQueryString(fmt.Sprintf("chat_id=%v", chatID)).
		SetHeader("Content-Type", "application/json").
		Get(messagesURL)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%v: %v", ErrResponseCode, resp.StatusCode())
	}

	err = json.Unmarshal(resp.Body(), &messages)
	if err != nil {
		return nil, resp, fmt.Errorf("%v: %v", ErrResponseDecode, err)
	}

	return messages.Data, resp, nil
}

func (m *Messages) Edit(ctx context.Context, messageID int, newMessage *Message) (*MessageResponse, *resty.Response, error) {
	if messageID <= 0 {
		return nil, nil, fmt.Errorf("%v: incorrect message ID %v", ErrInvalidInput, messageID)
	}
	if newMessage == nil {
		return nil, nil, fmt.Errorf("%v: new message is nil", ErrInvalidInput)
	}
	if newMessage.EntityID <= 0 {
		return nil, nil, fmt.Errorf("%v: incorrect entity ID %d", ErrInvalidInput, newMessage.EntityID)
	}
	if newMessage.Content == "" {
		return nil, nil, fmt.Errorf("%v: message content is empty", ErrInvalidInput)
	}

	body := OutgoingMessage{
		Message: newMessage,
	}

	url := fmt.Sprintf("%v/%v", messagesURL, messageID)
	resp, err := m.client.R().
		SetContext(ctx).
		SetBody(body).
		Put(url)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 200 {
		return nil, resp, fmt.Errorf("%v: %v", ErrResponseCode, resp.StatusCode())
	}

	var r *MessageResponseRaw
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, resp, fmt.Errorf("%v: %v", ErrResponseDecode, err)
	}

	return &r.Data, resp, nil
}

func (m *Messages) Send(ctx context.Context, msg *Message) (*MessageResponse, *resty.Response, error) {
	if msg == nil {
		return nil, nil, fmt.Errorf("%v: message is nil", ErrInvalidInput)
	}
	if msg.EntityID <= 0 {
		return nil, nil, fmt.Errorf("%v: incorrect entity ID %d", ErrInvalidInput, msg.EntityID)
	}
	if msg.Content == "" {
		return nil, nil, fmt.Errorf("%v: message content is empty", ErrInvalidInput)
	}

	body := OutgoingMessage{
		Message: msg,
	}

	resp, err := m.client.R().
		SetBody(body).
		SetContext(ctx).
		Post(messagesURL)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != 201 {
		return nil, resp, fmt.Errorf("%w: %d", ErrResponseCode, resp.StatusCode())
	}

	var r MessageResponseRaw
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, resp, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return &r.Data, resp, nil
}

func (m *Messages) Delete(ctx context.Context, messageID int) (*resty.Response, error) {
	if messageID <= 0 {
		return nil, fmt.Errorf("%w: incorrect message ID %d", ErrInvalidInput, messageID)
	}

	url := fmt.Sprintf("%v/%v", messagesURL, messageID)
	resp, err := m.client.R().
		SetContext(ctx).
		Delete(url)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 204 {
		return resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	return resp, nil
}

func (m *Messages) Pin(ctx context.Context, messageID int) (*resty.Response, error) {
	if messageID <= 0 {
		return nil, fmt.Errorf("%w: incorrect message ID %d", ErrInvalidInput, messageID)
	}

	url := fmt.Sprintf("%v/%v/pin", messagesURL, messageID)
	resp, err := m.client.R().
		SetContext(ctx).
		Post(url)
	if err != nil {
		return resp, err
	}

	// 201 при успешном пине сообщения
	// 409 если сообщение уже было запинено
	if resp.StatusCode() != 201 && resp.StatusCode() != 409 {
		return resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	return resp, nil
}

func (m *Messages) Unpin(ctx context.Context, messageID int) (*resty.Response, error) {
	if messageID <= 0 {
		return nil, fmt.Errorf("%w: incorrect message ID %d", ErrInvalidInput, messageID)
	}

	url := fmt.Sprintf("%v/%v/pin", messagesURL, messageID)
	resp, err := m.client.R().
		SetContext(ctx).
		Delete(url)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 204 {
		return resp, fmt.Errorf("%w: %v", ErrResponseCode, resp.StatusCode())
	}

	return resp, nil
}
