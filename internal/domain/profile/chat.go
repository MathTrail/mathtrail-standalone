package profile

// Chat is what the chat a task would go to is known to show. A task solved
// from the picture holds its facts on the card alone, and no words carry them,
// so it goes only to a chat known to draw the card. The zero value is the
// chat not known to: taken wrongly for one, a chat loses only the tasks solved
// from the picture, while a chat taken wrongly for one that draws the card
// would be given a task nobody can see.
type Chat int

const (
	// ChatOfWords is a chat not known to draw the card: one whose client
	// declared that it shows the words alone, or declared nothing at all.
	ChatOfWords Chat = iota
	// ChatWithCards is a chat known to draw the card a task is shown on.
	ChatWithCards
)

// ShowsPictures reports whether a task solved from the picture may go to the
// chat.
func (c Chat) ShowsPictures() bool { return c == ChatWithCards }
