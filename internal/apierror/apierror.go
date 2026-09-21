// Package apierror defines the shape of an HTTP error response.
//
// One shape for every endpoint: a stable machine-readable code and a sentence
// meant to be relayed to a person. Neither ever carries an internal error
// message, a stack or anything about the child.
package apierror

// Response is a structured HTTP error.
type Response struct {
	// Code is a stable identifier a client can branch on.
	Code string `json:"code"`
	// Message is a short sentence describing what happened.
	Message string `json:"message"`
}

// The codes the transport itself produces. Everything a tool refuses has its
// own vocabulary and does not travel through here.
const (
	CodeInternal         = "INTERNAL_ERROR"
	CodeNotFound         = "NOT_FOUND"
	CodeBadRequest       = "INVALID_REQUEST"
	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
)
