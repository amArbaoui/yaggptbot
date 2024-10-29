package telegram

type Message struct {
	Id       int64
	Text     string
	RepyToId int64
	ChatId   int64
	Role     string
}
