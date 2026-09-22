package seal

// Purpose is what a sealed value is for. The list is closed: a purpose that is
// not here cannot be sealed at all, and a value sealed for one purpose never
// opens as another, because the purpose both picks the subkey and is
// authenticated alongside the ciphertext.
type Purpose string

const (
	// PurposeCode seals an authorization code.
	PurposeCode Purpose = "code"
	// PurposeAccess seals an access token.
	PurposeAccess Purpose = "access"
	// PurposeRefresh seals a refresh token.
	PurposeRefresh Purpose = "refresh"
	// PurposeClient seals a client registration into the identifier issued for it.
	PurposeClient Purpose = "client"
	// PurposeState seals the context of an authorization request in flight.
	PurposeState Purpose = "state"
	// PurposeConsent seals the record of which clients a parent has approved.
	PurposeConsent Purpose = "consent"
	// PurposeTaskAnswer seals the part of a task that gives its answer away.
	PurposeTaskAnswer Purpose = "task-answer"
)

// labels ties every purpose to the single character that carries it inside a
// sealed value. It is read in both directions and written in one, so a label
// and the purpose behind it cannot drift apart.
var labels = map[Purpose]string{
	PurposeCode:       "c",
	PurposeAccess:     "a",
	PurposeRefresh:    "r",
	PurposeClient:     "d",
	PurposeState:      "s",
	PurposeConsent:    "k",
	PurposeTaskAnswer: "t",
}

// purposesByLabel reads the same table the other way round.
var purposesByLabel = func() map[string]Purpose {
	byLabel := make(map[string]Purpose, len(labels))
	for purpose, label := range labels {
		byLabel[label] = purpose
	}
	return byLabel
}()
