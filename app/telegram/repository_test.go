package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/storage"
	"context"
	"path/filepath"
	"testing"
	"time"

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

func TestMessageChainFromAnyChunkOfSplitReply(t *testing.T) {
	rep := newTestMessageRepository(t)
	saveTestMessage(t, rep, 1, 1, 0, "question", "user")
	chunks := []*tgbotapi.Message{
		{MessageID: 2, Chat: &tgbotapi.Chat{ID: 1}, ReplyToMessage: &tgbotapi.Message{MessageID: 1}, Text: "first"},
		{MessageID: 3, Chat: &tgbotapi.Chat{ID: 1}, ReplyToMessage: &tgbotapi.Message{MessageID: 1}, Text: "second"},
		{MessageID: 4, Chat: &tgbotapi.Chat{ID: 1}, ReplyToMessage: &tgbotapi.Message{MessageID: 1}, Text: "third"},
	}
	if err := rep.SaveMessages(context.Background(), chunks, "**full** answer", "assistant"); err != nil {
		t.Fatalf("failed to save chunks: %v", err)
	}
	for _, chunkId := range []int64{2, 3, 4} {
		chain, err := rep.GetMessageChain(context.Background(), 1, chunkId, 30)
		if err != nil {
			t.Fatalf("chunk %d: failed to get chain: %v", chunkId, err)
		}
		if len(chain) != 2 || chain[0].Text != "question" || chain[1].Text != "**full** answer" {
			texts := make([]string, 0, len(chain))
			for _, m := range chain {
				texts = append(texts, m.Text)
			}
			t.Errorf("chunk %d: expected [question, **full** answer], got %q", chunkId, texts)
		}
	}
	if chunks[0].Text != "first" {
		t.Errorf("sent messages should not be mutated")
	}
}

func TestPhotoMessageIsStoredWithCaptionAndImage(t *testing.T) {
	rep := newTestMessageRepository(t)
	photos := []tgbotapi.PhotoSize{{FileID: "small"}, {FileID: "large"}}
	for _, m := range []*tgbotapi.Message{
		{MessageID: 1, Chat: &tgbotapi.Chat{ID: 1}, Photo: photos, Caption: "what is it?"},
		{MessageID: 2, Chat: &tgbotapi.Chat{ID: 1}, Photo: photos, ReplyToMessage: &tgbotapi.Message{MessageID: 1}},
	} {
		if err := rep.SaveMessage(context.Background(), m, "user"); err != nil {
			t.Fatal(err)
		}
	}
	chain, err := rep.GetMessageChain(context.Background(), 1, 2, 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(chain) != 2 {
		t.Fatalf("photo without caption should stay in the chain, got %d messages", len(chain))
	}
	if chain[0].Text != "what is it?" || chain[0].ImageFileId == nil || *chain[0].ImageFileId != "large" {
		t.Errorf("expected caption and largest photo, got %q, %v", chain[0].Text, chain[0].ImageFileId)
	}
}

func TestImplicitReplyContinuesLatestConversation(t *testing.T) {
	rep := newTestMessageRepository(t)
	bot := &GPTBot{msgService: &MessageDbService{rep: rep}, botOptions: BotOptions{MaxConversationDepth: config.MaxMessageContextDepth}}
	ctx := context.Background()
	const chatId = 1
	newMessage := func(id int) *tgbotapi.Message {
		return &tgbotapi.Message{MessageID: id, Chat: &tgbotapi.Chat{ID: chatId}, Text: "next"}
	}

	if got := bot.withImplicitReply(ctx, newMessage(1)); got.ReplyToMessage != nil {
		t.Errorf("first message in the chat should start a conversation")
	}

	saveTestMessage(t, rep, chatId, 1, 0, "question", "user")
	saveTestMessage(t, rep, chatId, 2, 1, "answer", "assistant")
	saveTestMessage(t, rep, chatId, 3, 2, "follow-up", "user")
	saveTestMessage(t, rep, chatId, 4, 3, "second answer", "assistant")
	saveTestMessage(t, rep, 2, 10, 0, "other chat answer", "assistant")

	got := bot.withImplicitReply(ctx, newMessage(5))
	if got.ReplyToMessage == nil || got.ReplyToMessage.MessageID != 4 {
		t.Fatalf("expected implicit reply to the latest answer 4, got %+v", got.ReplyToMessage)
	}
	chain, err := bot.GetConversationChain(ctx, got)
	if err != nil || len(chain) != 4 {
		t.Errorf("expected the whole conversation in the chain, got %d messages, %v", len(chain), err)
	}

	explicit := newMessage(6)
	explicit.ReplyToMessage = &tgbotapi.Message{MessageID: 2}
	if got := bot.withImplicitReply(ctx, explicit); got.ReplyToMessage.MessageID != 2 {
		t.Errorf("explicit reply should be kept, got %d", got.ReplyToMessage.MessageID)
	}

	if err := rep.ResetConversation(ctx, chatId, 7); err != nil {
		t.Fatal(err)
	}
	if got := bot.withImplicitReply(ctx, newMessage(8)); got.ReplyToMessage != nil {
		t.Errorf("message after /new should start a conversation, got reply to %d", got.ReplyToMessage.MessageID)
	}
	saveTestMessage(t, rep, chatId, 9, 8, "answer after reset", "assistant")
	if got := bot.withImplicitReply(ctx, newMessage(10)); got.ReplyToMessage == nil || got.ReplyToMessage.MessageID != 9 {
		t.Errorf("expected to continue the conversation started after /new")
	}
}

func TestImplicitReplyExpiresAfterTimeout(t *testing.T) {
	rep := newTestMessageRepository(t)
	bot := &GPTBot{msgService: &MessageDbService{rep: rep}, botOptions: BotOptions{MaxConversationDepth: config.MaxMessageContextDepth}}
	saveTestMessage(t, rep, 1, 1, 0, "question", "user")
	saveTestMessage(t, rep, 1, 2, 1, "answer", "assistant")
	expired := time.Now().Add(-config.ConversationTimeout - time.Minute).Unix()
	if _, err := rep.db.Exec("update message set created_at = $1", expired); err != nil {
		t.Fatal(err)
	}
	msg := &tgbotapi.Message{MessageID: 3, Chat: &tgbotapi.Chat{ID: 1}, Text: "next"}
	if got := bot.withImplicitReply(context.Background(), msg); got.ReplyToMessage != nil {
		t.Errorf("expired conversation should not be continued")
	}
}
