package telegram

import (
	"amArbaoui/yaggptbot/app/storage"
	"context"
	"path/filepath"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func newTestMessageRepository(t *testing.T) *MessageDbRepository {
	t.Helper()
	db, err := sqlx.Connect("sqlite", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	driver, err := sqlite.WithInstance(db.DB, &sqlite.Config{})
	if err != nil {
		t.Fatalf("failed to create migrate driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://../storage/migrations", "sqlite", driver)
	if err != nil {
		t.Fatalf("failed to init migrations: %v", err)
	}
	if err := m.Up(); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}
	key, err := storage.GenerateAESKey(32)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	return &MessageDbRepository{db: db, encryptService: storage.NewEncryptionService(storage.EncodeKeyToString(key))}
}

func saveTestMessage(t *testing.T, rep *MessageDbRepository, chatId int64, msgId int, replyTo int, text string, role string) {
	t.Helper()
	msg := &tgbotapi.Message{MessageID: msgId, Chat: &tgbotapi.Chat{ID: chatId}, Text: text}
	if replyTo > 0 {
		msg.ReplyToMessage = &tgbotapi.Message{MessageID: replyTo}
	}
	if err := rep.SaveMessage(context.Background(), msg, role); err != nil {
		t.Fatalf("failed to save message %d in chat %d: %v", msgId, chatId, err)
	}
}

func TestMessageChainIsScopedByChat(t *testing.T) {
	rep := newTestMessageRepository(t)
	const chatA, chatB = 100, 200

	// telegram message ids are unique only within a chat, so both chats reuse ids 1..3
	saveTestMessage(t, rep, chatA, 1, 0, "a: question", "user")
	saveTestMessage(t, rep, chatA, 2, 1, "a: answer", "assistant")
	saveTestMessage(t, rep, chatA, 3, 2, "a: follow-up", "user")
	saveTestMessage(t, rep, chatB, 1, 0, "b: question", "user")
	saveTestMessage(t, rep, chatB, 2, 1, "b: answer", "assistant")
	saveTestMessage(t, rep, chatB, 3, 2, "b: follow-up", "user")

	for _, tt := range []struct {
		chatId   int64
		expected []string
	}{
		{chatA, []string{"a: question", "a: answer", "a: follow-up"}},
		{chatB, []string{"b: question", "b: answer", "b: follow-up"}},
	} {
		chain, err := rep.GetMessageChain(context.Background(), tt.chatId, 3, 30)
		if err != nil {
			t.Fatalf("chat %d: failed to get chain: %v", tt.chatId, err)
		}
		if len(chain) != len(tt.expected) {
			t.Fatalf("chat %d: expected %d messages, got %d", tt.chatId, len(tt.expected), len(chain))
		}
		for i, m := range chain {
			if m.ChatId != tt.chatId || m.Text != tt.expected[i] {
				t.Errorf("chat %d, position %d: expected %q, got %q from chat %d", tt.chatId, i, tt.expected[i], m.Text, m.ChatId)
			}
		}

		msg, err := rep.GetMessageById(context.Background(), tt.chatId, 1)
		if err != nil {
			t.Fatalf("chat %d: failed to get message: %v", tt.chatId, err)
		}
		if msg.Text != tt.expected[0] {
			t.Errorf("chat %d: expected %q, got %q", tt.chatId, tt.expected[0], msg.Text)
		}
	}
}

func TestMessageChainRespectsDepth(t *testing.T) {
	rep := newTestMessageRepository(t)
	saveTestMessage(t, rep, 1, 1, 0, "first", "user")
	saveTestMessage(t, rep, 1, 2, 1, "second", "assistant")
	saveTestMessage(t, rep, 1, 3, 2, "third", "user")

	chain, err := rep.GetMessageChain(context.Background(), 1, 3, 2)
	if err != nil {
		t.Fatalf("failed to get chain: %v", err)
	}
	if len(chain) != 2 || chain[0].Text != "second" || chain[1].Text != "third" {
		t.Errorf("expected [second third], got %d messages", len(chain))
	}
}
