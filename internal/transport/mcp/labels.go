package mcpserver

import (
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// other stands in a line or a span for any value outside the list it was held
// to.
const other = "other"

// familyChatGPT is the family of the chat host that starts its sign-in from a
// challenge in a failed result, and not from a 401 in the middle of a call.
const familyChatGPT = "chatgpt"

// supportedVersions are the protocol versions the library speaks, newest first.
var supportedVersions = mcp.SupportedProtocolVersions()

// toolLabel is the name of a tool this endpoint defines, or other. A call may
// name any tool at all, and a name a caller chose is not written down.
func (b *boundary) toolLabel(name string) string {
	if _, defined := b.tools[name]; defined {
		return name
	}
	return other
}

// protocolLabel is a protocol version the library speaks, or other.
func protocolLabel(version string) string {
	if slices.Contains(supportedVersions, version) {
		return version
	}
	return other
}

// clientFamily is which chat host a call came from, as one word of a closed
// list: the name a client gives itself is its own to choose, and it is read for
// a word or two it can be told by, never written down. A call without a name is
// unknown; one with a name none of the words fits is other.
func clientFamily(info *mcp.Implementation) string {
	if info == nil || info.Name == "" {
		return "unknown"
	}
	name := strings.ToLower(info.Name)
	switch {
	case strings.Contains(name, "claude"), strings.Contains(name, "anthropic"):
		return "claude"
	case strings.Contains(name, "chatgpt"), strings.Contains(name, "openai"):
		return familyChatGPT
	case strings.Contains(name, "inspector"):
		return "inspector"
	}
	return other
}

// methodLabel is a method of the protocol a server answers, or other.
func methodLabel(method string) string {
	switch method {
	case "initialize", "server/discover", "ping",
		"tools/list", "tools/call",
		"resources/list", "resources/read", "resources/templates/list",
		"prompts/list", "prompts/get",
		"notifications/initialized", "notifications/cancelled":
		return method
	}
	return other
}

// trapLabel is a trap of the catalog, or other. The trap of an answer comes
// from a task the checks accepted, so it is one of the catalog's; what reaches
// a span or a line is held to the list all the same.
func (s *Service) trapLabel(id string) string {
	if s.content.HasTrap(id) {
		return id
	}
	return other
}

// topicLabel is a topic of the catalog, or other. The topic of an answer is
// read from the profile, which a person can edit, and a line names only topics
// the catalog has.
func (s *Service) topicLabel(id string) string {
	if s.content.HasTopic(id) {
		return id
	}
	return other
}
