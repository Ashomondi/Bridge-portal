package notifications

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"unicode/utf8"
)

type SMSPurpose string

const (
	AppLoginOTP        SMSPurpose = "app_login"
	WhatsAppSessionOTP SMSPurpose = "whatsapp_session"
	PhoneChangeOTP     SMSPurpose = "phone_change"
	PhoneChangeNotice  SMSPurpose = "phone_change_notice"
)

var (
	ErrInvalidPhone = errors.New("invalid recipient phone number")
	ErrInvalidSMS   = errors.New("invalid transactional SMS")
)

// SMSMessage contains sensitive data; dispatchers must not log Body or To.
type SMSMessage struct {
	To      string     `json:"to"`
	Body    string     `json:"body"`
	Purpose SMSPurpose `json:"purpose"`
}

// SMSDispatcher accepts only transactional messages. A nil error means the
// dispatcher accepted the message, not that the handset received it.
type SMSDispatcher interface {
	DispatchSMS(context.Context, SMSMessage) error
}

// NormalizePhone accepts Kenyan local numbers and international E.164 numbers.
// Validation errors intentionally exclude the input to avoid exposing numbers.
func NormalizePhone(phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	var normalized strings.Builder
	for _, c := range phone {
		switch {
		case c >= '0' && c <= '9':
			normalized.WriteRune(c)
		case c == '+' && normalized.Len() == 0:
			normalized.WriteRune(c)
		case c == ' ' || c == '-' || c == '(' || c == ')':
		default:
			return "", ErrInvalidPhone
		}
	}
	phone = normalized.String()
	if len(phone) == 10 && (strings.HasPrefix(phone, "07") || strings.HasPrefix(phone, "01")) {
		phone = "+254" + phone[1:]
	} else if len(phone) == 12 && strings.HasPrefix(phone, "254") {
		phone = "+" + phone
	}
	if len(phone) < 9 || len(phone) > 16 || phone[0] != '+' || phone[1] == '0' {
		return "", ErrInvalidPhone
	}
	for _, c := range phone[1:] {
		if c < '0' || c > '9' {
			return "", ErrInvalidPhone
		}
	}
	if strings.HasPrefix(phone, "+254") && (len(phone) != 13 || (phone[4] != '7' && phone[4] != '1')) {
		return "", ErrInvalidPhone
	}
	return phone, nil
}

func (m SMSMessage) normalized() (SMSMessage, error) {
	switch m.Purpose {
	case AppLoginOTP, WhatsAppSessionOTP, PhoneChangeOTP, PhoneChangeNotice:
	default:
		return SMSMessage{}, ErrInvalidSMS
	}
	if strings.TrimSpace(m.Body) == "" || !utf8.ValidString(m.Body) {
		return SMSMessage{}, ErrInvalidSMS
	}
	phone, err := NormalizePhone(m.To)
	if err != nil {
		return SMSMessage{}, err
	}
	m.To = phone
	return m, nil
}

type sandboxSMSDispatcher struct {
	logger *slog.Logger
}

// NewSandboxSMSDispatcher records only redacted delivery metadata, never sends
// a network request, and never logs the message body or the full recipient.
func NewSandboxSMSDispatcher(logger *slog.Logger) SMSDispatcher {
	if logger == nil {
		logger = slog.Default()
	}
	return &sandboxSMSDispatcher{logger: logger}
}

func (d *sandboxSMSDispatcher) DispatchSMS(ctx context.Context, message SMSMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	message, err := message.normalized()
	if err != nil {
		return err
	}
	d.logger.InfoContext(ctx, "sandbox transactional SMS accepted",
		"recipient", "***"+message.To[len(message.To)-3:],
		"purpose", string(message.Purpose),
		"characters", utf8.RuneCountInString(message.Body),
	)
	return nil
}
