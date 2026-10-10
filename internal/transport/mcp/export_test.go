package mcpserver

import "time"

// HoldForNews sets how long a card's question waits for news, so that a test
// need not wait the hold a deployment keeps. It is called before the service
// serves anything.
func HoldForNews(s *Service, hold time.Duration) { s.hold = hold }

// PauseForAnswer sets how long show_result waits before it reads the profile
// again for an answer not yet in it, so that a test need not wait the pause a
// deployment keeps. It is called before the service serves anything.
func PauseForAnswer(s *Service, pause time.Duration) { s.answerPause = pause }
