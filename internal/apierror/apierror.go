// Package apierror defines the shape of an HTTP error response.
//
// One shape for every error the service's HTTP surface answers on its own
// account — a path nobody declared, a method a path does not take, a fault — a
// stable machine-readable code and a sentence meant to be relayed to a person.
// Neither ever carries an internal error message, a stack or anything about
// the child. Where a protocol prescribes a shape of its own, the protocol's is
// the one answered with: the sign-in answers as OAuth says, with an error and
// its description, and the MCP endpoint with an error of JSON-RPC.
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
	// CodeInternal answers a request the service failed on: a 500.
	CodeInternal = "INTERNAL_ERROR"
	// CodeNotFound answers a path the service does not serve: a 404.
	CodeNotFound = "NOT_FOUND"
	// CodeMethodNotAllowed answers a path asked with a method it does not
	// take: a 405, with the methods it does take in its Allow header.
	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
)
