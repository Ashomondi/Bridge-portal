package notifications

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Identity context key and helpers for notifications
type identityKey struct{}

type Identity struct {
	UserID       string   `json:"user_id"`
	SchoolID     string   `json:"school_id"`
	MembershipID string   `json:"membership_id"`
	Role         string   `json:"role"`
	Permissions  []string `json:"permissions,omitempty"`
}

func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, id)
}

func IdentityFromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey{}).(Identity)
	return id, ok
}

// InAppNotification model
type InAppNotification struct {
	ID              string                 `json:"id"`
	SchoolID        string                 `json:"school_id"`
	RecipientUserID string                 `json:"recipient_user_id"`
	Category        string                 `json:"category"`
	Title           string                 `json:"title"`
	Body            string                 `json:"body"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	IsRead          bool                   `json:"is_read"`
	ReadAt          *time.Time             `json:"read_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// SMSDispatcher interface for transactional SMS (Africa's Talking)
type SMSDispatcher interface {
	SendTransactionalSMS(ctx context.Context, toPhoneNumber, message string) error
}

// AfricasTalkingSMSClient dispatches transactional SMS or logs deterministically in sandbox mode
type AfricasTalkingSMSClient struct {
	Username string
	APIKey   string
	SenderID string
	SentSMS  []struct {
		To      string
		Message string
		SentAt  time.Time
	}
	mu         sync.Mutex
	httpClient *http.Client
}

func NewAfricasTalkingSMSClient(username, apiKey, senderID string) *AfricasTalkingSMSClient {
	if username == "" {
		username = "sandbox"
	}
	return &AfricasTalkingSMSClient{
		Username:   username,
		APIKey:     apiKey,
		SenderID:   senderID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *AfricasTalkingSMSClient) SendTransactionalSMS(ctx context.Context, toPhoneNumber, message string) error {
	if strings.TrimSpace(toPhoneNumber) == "" || strings.TrimSpace(message) == "" {
		return errors.New("SMS recipient and message are required")
	}
	if c.Username != "sandbox" {
		if c.APIKey == "" {
			return errors.New("Africa's Talking API key is required")
		}
		form := url.Values{"username": {c.Username}, "to": {toPhoneNumber}, "message": {message}}
		if c.SenderID != "" {
			form.Set("from", c.SenderID)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.africastalking.com/version1/messaging", strings.NewReader(form.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("apiKey", c.APIKey)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		response, err := c.httpClient.Do(req)
		if err != nil {
			return errors.New("Africa's Talking SMS request failed")
		}
		defer response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return fmt.Errorf("Africa's Talking SMS request returned HTTP %d", response.StatusCode)
		}
		var result struct {
			SMSMessageData struct {
				Recipients []struct {
					Number     string `json:"number"`
					StatusCode int    `json:"statusCode"`
				} `json:"Recipients"`
			} `json:"SMSMessageData"`
		}
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			return errors.New("invalid Africa's Talking SMS response")
		}
		recipients := result.SMSMessageData.Recipients
		if len(recipients) != 1 || recipients[0].Number != toPhoneNumber || recipients[0].StatusCode != 101 {
			return errors.New("Africa's Talking did not accept the SMS recipient")
		}
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SentSMS = append(c.SentSMS, struct {
		To      string
		Message string
		SentAt  time.Time
	}{toPhoneNumber, message, time.Now()})
	log.Print("[Transactional SMS sandbox] Message recorded")
	return nil
}

// Store persistence interface for In-App Notifications
type Store interface {
	CreateInAppNotification(ctx context.Context, notification *InAppNotification) error
	ListInAppNotifications(ctx context.Context, schoolID, recipientUserID string, unreadOnly bool, limit int) ([]InAppNotification, error)
	MarkNotificationRead(ctx context.Context, schoolID, notificationID, recipientUserID string) error
}

// MemoryStore implementation
type MemoryStore struct {
	mu            sync.RWMutex
	Notifications map[string]*InAppNotification
	UserNotifs    map[string][]string // key: schoolID:recipientUserID -> []notificationID
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Notifications: make(map[string]*InAppNotification),
		UserNotifs:    make(map[string][]string),
	}
}

func (m *MemoryStore) CreateInAppNotification(ctx context.Context, n *InAppNotification) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if n.ID == "" {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		n.ID = fmt.Sprintf("notif_%x", b)
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	m.Notifications[n.ID] = n
	key := fmt.Sprintf("%s:%s", n.SchoolID, n.RecipientUserID)
	m.UserNotifs[key] = append(m.UserNotifs[key], n.ID)
	return nil
}

func (m *MemoryStore) ListInAppNotifications(ctx context.Context, schoolID, recipientUserID string, unreadOnly bool, limit int) ([]InAppNotification, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", schoolID, recipientUserID)
	ids := m.UserNotifs[key]
	var res []InAppNotification
	for i := len(ids) - 1; i >= 0; i-- {
		n, ok := m.Notifications[ids[i]]
		if !ok {
			continue
		}
		if unreadOnly && n.IsRead {
			continue
		}
		res = append(res, *n)
		if limit > 0 && len(res) >= limit {
			break
		}
	}
	return res, nil
}

func (m *MemoryStore) MarkNotificationRead(ctx context.Context, schoolID, notificationID, recipientUserID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.Notifications[notificationID]
	if !ok || n.RecipientUserID != recipientUserID {
		return errors.New("notification not found")
	}
	now := time.Now()
	n.IsRead = true
	n.ReadAt = &now
	return nil
}

// Service manages notifications
type Service struct {
	store Store
	sms   SMSDispatcher
}

func NewService(store Store, sms SMSDispatcher) *Service {
	return &Service{
		store: store,
		sms:   sms,
	}
}

// CreateInAppNotification fulfills the billing.InAppNotifier and internal hooks
func (s *Service) CreateInAppNotification(ctx context.Context, schoolID, recipientUserID, category, title, body string, metadata map[string]interface{}) error {
	n := &InAppNotification{
		SchoolID:        schoolID,
		RecipientUserID: recipientUserID,
		Category:        category,
		Title:           title,
		Body:            body,
		Metadata:        metadata,
		IsRead:          false,
		CreatedAt:       time.Now(),
	}
	return s.store.CreateInAppNotification(ctx, n)
}

func (s *Service) GetSMSDispatcher() SMSDispatcher {
	return s.sms
}

func (s *Service) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	wrap := func(h http.HandlerFunc) http.Handler {
		if authMiddleware != nil {
			return authMiddleware(h)
		}
		return h
	}

	mux.Handle("/api/v1/notifications", wrap(s.handleListNotifications))
	mux.HandleFunc("/api/v1/notifications/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/read") && r.Method == http.MethodPatch {
			wrap(s.handleMarkRead).ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
}

func (s *Service) handleListNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, ok := IdentityFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	unreadOnly := r.URL.Query().Get("unread") == "true"
	items, err := s.store.ListInAppNotifications(r.Context(), id.SchoolID, id.UserID, unreadOnly, 50)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"data":   items,
	})
}

func (s *Service) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	id, ok := IdentityFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	clean := strings.TrimPrefix(r.URL.Path, "/api/v1/notifications/")
	notifID := strings.TrimSuffix(clean, "/read")

	if err := s.store.MarkNotificationRead(r.Context(), id.SchoolID, notifID, id.UserID); err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "notification not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
