package seal

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

const (
	// userIDInfo is the context string the subkey of user identifiers is
	// derived from. It names no purpose, so the subkey can never be the one a
	// purpose seals with.
	userIDInfo = subkeyInfo + "user-id"
	// userIDLength is how many base64url characters an identifier keeps: 96
	// bits, far past any number of families that could ever sign in.
	userIDLength = 16
)

// UserID is the identifier an account is known by in this service, derived
// from the account's own identifier at the provider it signed in with, which
// the caller names with the provider in front — "google-sub:" and Google's
// subject. The provider's identifier stands for a real person wherever that
// provider signs them in, so it is kept nowhere and written nowhere; this one
// says nothing about it.
//
// It is a keyed digest under a subkey of the current key alone. The same
// account gets the same identifier for as long as that key is current, and a
// different one once a rotation has made another key current: whoever holds
// an identifier keeps it, and the next sign-in is given the new one.
func (r *KeyRing) UserID(account string) string {
	return userIDUnder(r.current.userID, account)
}

// UserIDs are every identifier an account may be known by while the ring's
// keys open what they sealed: the one the current key derives, first, and
// during a rotation the one the previous key derived, which a sign-in made
// before the rotation still carries in its tokens. Once that key leaves the
// ring, its tokens open no more, and its identifier is nobody's.
func (r *KeyRing) UserIDs(account string) []string {
	ids := []string{r.UserID(account)}
	if r.previous != nil {
		ids = append(ids, userIDUnder(r.previous.userID, account))
	}
	return ids
}

// userIDUnder is an account's identifier under one key's subkey of user
// identifiers.
func userIDUnder(subkey []byte, account string) string {
	mac := hmac.New(sha256.New, subkey)
	mac.Write([]byte(account))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))[:userIDLength]
}
