package middleware

import (
	"net/url"
	"slices"
	"strings"

	"go.uber.org/zap"
)

// A queryValue is how much of a parameter's value a log line may carry: the
// words of its protocol, or the address of a client or a resource.
type queryValue struct {
	// words are the values its protocol defines, and the only ones written.
	// Anything else in their place is written as "other": a value is chosen by
	// whoever sends it as freely as a name is, and a word the protocol never
	// defined is where a caller's own text — a name, an email address — would
	// be. No words means an address.
	words []string
	// list is a value of several words separated by spaces, each held to the
	// words on its own, as a scope is.
	list bool
}

// otherWord stands in a log line for whatever a parameter held beyond its
// protocol's words: that something was there, and nothing of what it was.
const otherWord = "other"

// loggedQueryKeys are the query parameters a log line may carry: the ones the
// sign-in and its protocols define, which name a mechanism rather than a
// person, a secret or a place. Every other parameter is counted rather than
// written, its name included, because a name a caller chose can carry anything
// a value can. The sign-in carries codes, tokens and verifiers in query
// strings, a provider adds hints with an email address in them and the domain
// of a school, and a list of what to hide would let through whatever nobody
// thought to put on it. Compared without case.
//
// The words are the protocols' own: OAuth 2.0 and its extensions for PKCE,
// resource indicators and registration, OpenID Connect, and the scopes this
// service and Google's sign-in speak — ours is mcp, Google's are openid and
// drive.file, with the ones Google may add to an answer beside them. A scope
// this service starts to use is written as "other" until it is added here.
var loggedQueryKeys = map[string]queryValue{
	"client_id":             {},
	"code_challenge_method": {words: []string{"S256", "plain"}},
	"error": {words: []string{
		// RFC 6749
		"invalid_request", "unauthorized_client", "access_denied", "unsupported_response_type",
		"invalid_scope", "server_error", "temporarily_unavailable", "invalid_client", "invalid_grant",
		"unsupported_grant_type",
		// OpenID Connect
		"interaction_required", "login_required", "account_selection_required", "consent_required",
		"invalid_request_uri", "invalid_request_object", "request_not_supported",
		"request_uri_not_supported", "registration_not_supported",
		// RFC 8707, RFC 6750, RFC 7591
		"invalid_target", "invalid_token", "insufficient_scope", "invalid_redirect_uri",
		"invalid_client_metadata",
	}},
	"grant_type": {words: []string{
		"authorization_code", "refresh_token", "client_credentials",
		"urn:ietf:params:oauth:grant-type:device_code", "urn:ietf:params:oauth:grant-type:token-exchange",
		"urn:ietf:params:oauth:grant-type:jwt-bearer",
	}},
	"prompt":        {words: []string{"none", "login", "consent", "select_account", "create"}, list: true},
	"redirect_uri":  {},
	"resource":      {},
	"response_mode": {words: []string{"query", "fragment", "form_post"}},
	"response_type": {words: []string{"code", "token", "id_token", "none"}, list: true},
	"scope": {words: []string{
		"mcp", "openid", "profile", "email", "offline_access",
		"https://www.googleapis.com/auth/drive.file",
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	}, list: true},
}

// loggedValueLimit is the longest value a line carries. The addresses a client
// names itself by are short, and a longer one is something else, which is
// masked rather than let grow a line past what a log keeps.
const loggedValueLimit = 256

// queryFields are what a request's line says about its query: the parameters
// a log may carry, with what they carried, and how many others there were. A
// query that cannot be parsed is not read at all rather than guessed at.
//
// It always parses, and deliberately reads no name from the raw text: a name
// may be percent-encoded, "%73cope=a" is a scope to the handlers and
// "%63ode=a" a code, and parsing is the only reading of a query that agrees
// with theirs. What is written is therefore not the string that arrived: the
// parameters are sorted by name and encoded again.
func queryFields(raw string) []zap.Field {
	if raw == "" {
		return nil
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return []zap.Field{zap.String("query", "unparsable")}
	}

	kept, others := url.Values{}, 0
	for name, list := range values {
		kind, logged := loggedQueryKeys[strings.ToLower(name)]
		if !logged {
			others += len(list)
			continue
		}
		for _, value := range list {
			if value = loggedValue(kind, value); len(value) > loggedValueLimit {
				value = "masked"
			}
			kept.Add(name, value)
		}
	}

	var fields []zap.Field
	if len(kept) > 0 {
		fields = append(fields, zap.String("query", kept.Encode()))
	}
	if others > 0 {
		fields = append(fields, zap.Int("query_others", others))
	}
	return fields
}

// loggedValue is what a line may say of one value: an address without what
// can be hidden in it, or the words of its protocol, each once, with "other"
// once for all the rest.
func loggedValue(kind queryValue, value string) string {
	switch {
	case kind.words == nil:
		return addressOnly(value)
	case !kind.list:
		return wordOf(kind.words, value)
	}
	var said []string
	for _, word := range strings.Fields(value) {
		if word = wordOf(kind.words, word); !slices.Contains(said, word) {
			said = append(said, word)
		}
	}
	return strings.Join(said, " ")
}

// wordOf is a word as a log line writes it: itself when its protocol defines
// it, and "other" when it does not.
func wordOf(words []string, word string) string {
	if slices.Contains(words, word) {
		return word
	}
	return otherWord
}

// addressOnly is an address as its scheme and its host, which say which client
// or which resource it is, and nothing more. The path, the query, the fragment
// and a user are the client's own words, as free as any value and as able to
// carry a name. A value that is no absolute address is masked whole.
func addressOnly(value string) string {
	address, err := url.Parse(value)
	if err != nil || address.Scheme == "" || address.Host == "" {
		return "masked"
	}
	return address.Scheme + "://" + address.Host
}
