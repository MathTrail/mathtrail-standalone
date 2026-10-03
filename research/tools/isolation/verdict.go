package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Offer is what one request put before a model: the model it asked for, if it
// names one, and the tools the model could call.
type Offer struct {
	Model string
	Tools []string
}

// Verdict sums the requests up: one line each, and whether the client is
// isolated.
type Verdict struct {
	Lines    []string
	Summary  string
	Isolated bool
}

// Judge decides whether the client is isolated: it asked a model for at least
// one reply, every request could be read, and no request of any kind offered a
// tool.
func Judge(requests []Request) Verdict {
	var v Verdict
	replies, offering, unreadable := 0, 0, 0
	for _, request := range requests {
		line := fmt.Sprintf("%s %s: ", request.Method, request.Target)
		offer, reply, err := OfferOf(request)
		switch {
		case err != nil:
			unreadable++
			line += "cannot be read: " + err.Error()
		case reply:
			replies++
			line += fmt.Sprintf("model %s, %s", orUnnamed(offer.Model), toolList(offer.Tools))
		default:
			line += "asks for no reply, " + toolList(offer.Tools)
		}
		if len(offer.Tools) > 0 {
			offering++
		}
		v.Lines = append(v.Lines, line)
	}
	switch {
	case offering > 0:
		v.Summary = fmt.Sprintf("not isolated: %d of %s offered the model tools", offering, count(len(requests), "request"))
	case unreadable > 0:
		v.Summary = fmt.Sprintf("not shown: %d of %s could not be read, and any of them may have offered tools", unreadable, count(len(requests), "request"))
	case replies == 0:
		v.Summary = fmt.Sprintf("not shown: none of %s asked for a reply, so what a model would be offered was never seen", count(len(requests), "request"))
	default:
		v.Isolated = true
		v.Summary = fmt.Sprintf("isolated: %s for a reply, and no request offered a tool", count(replies, "request"))
	}
	return v
}

func toolList(tools []string) string {
	if len(tools) == 0 {
		return "no tools"
	}
	return fmt.Sprintf("%s: %s", count(len(tools), "tool"), strings.Join(tools, ", "))
}

// OfferOf reads what a request put before a model. It knows the three shapes
// of a request for a reply: Anthropic's Messages API ("messages"), OpenAI's
// Responses API ("input") and the Gemini API ("contents"), the last also as
// Google's Code Assist wraps it (under "request"). reply is false for any
// other request, whose tools, if it has any, are read all the same. A request
// with no body offers nothing; a body that is not a JSON object cannot be
// read, and err says why.
//
// Tools are looked for everywhere in the body, not only where each API puts
// them: a client may declare them in a place of its own — Codex sends some
// models theirs inside an item of "input" — and a check that looked only in
// the expected place would pass a client that moved them.
func OfferOf(request Request) (offer Offer, reply bool, err error) {
	if request.Encoding != "" {
		return Offer{}, false, fmt.Errorf("the body is encoded as %s", request.Encoding)
	}
	if len(request.Body) == 0 {
		return Offer{}, false, nil
	}
	var body map[string]any
	if json.Unmarshal(request.Body, &body) != nil || body == nil {
		return Offer{}, false, errors.New("the body is not a JSON object")
	}
	offer.Model = modelOf(body, request.Target)
	offer.Tools = toolsIn(body)
	if wrapped, ok := body["request"].(map[string]any); ok && wrapped["contents"] != nil {
		return offer, true, nil
	}
	for _, key := range []string{"messages", "input", "contents"} {
		if body[key] != nil {
			return offer, true, nil
		}
	}
	return offer, false, nil
}

// modelOf is the model a request names: in its body, or, for the Gemini API,
// in its path (".../models/<model>:generateContent").
func modelOf(body map[string]any, target string) string {
	if model, ok := body["model"].(string); ok && model != "" {
		return model
	}
	_, after, found := strings.Cut(target, "/models/")
	if !found {
		return ""
	}
	model, _, _ := strings.Cut(after, ":")
	model, _, _ = strings.Cut(model, "?")
	return model
}

// toolLists are the keys whose value lists tools: every API's "tools", the
// Gemini API's function declarations, and the MCP servers Anthropic's API can
// be handed, each of which brings its tools with it.
var toolLists = []string{"tools", "functionDeclarations", "function_declarations", "mcp_servers"}

// labels are the keys that name or describe an entry that holds tools, rather
// than offering one.
var labels = []string{"type", "name", "description"}

// toolsIn names every tool listed anywhere in a value, at any depth.
func toolsIn(value any) []string {
	var names []string
	switch value := value.(type) {
	case map[string]any:
		for _, key := range slices.Sorted(maps.Keys(value)) {
			if slices.Contains(toolLists, key) {
				names = append(names, listed(value[key], "")...)
				continue
			}
			names = append(names, toolsIn(value[key])...)
		}
	case []any:
		for _, item := range value {
			names = append(names, toolsIn(item)...)
		}
	}
	return names
}

// listed names the tools a tool-list key holds, whatever shape it has: each
// entry of a list, each key of an object, and anything else as one tool. Null
// and an empty list or object hold none.
func listed(value any, prefix string) []string {
	var names []string
	switch value := value.(type) {
	case nil:
	case []any:
		for _, tool := range value {
			names = append(names, toolNames(tool, prefix)...)
		}
	case map[string]any:
		for _, key := range slices.Sorted(maps.Keys(value)) {
			names = append(names, prefix+key)
		}
	default:
		names = append(names, prefix+"(unnamed)")
	}
	return names
}

// toolNames names the tools one entry of a list offers. An entry that lists
// tools of its own — a namespace, or a Gemini Tool with its functions — offers
// those, plus every key that neither lists tools nor labels the entry: a
// Gemini Tool may declare built-in tools, such as a search, beside its
// functions. Any other entry is a tool, named by its "name", or by its "type"
// when it has none (a tool the service itself runs, such as a web search), or
// else by its keys (a Gemini built-in tool).
func toolNames(tool any, prefix string) []string {
	entry, ok := tool.(map[string]any)
	if !ok {
		name, _ := tool.(string)
		return []string{prefix + orUnnamed(name)}
	}
	name := entryName(entry)
	holder := false
	var held []string
	for _, key := range toolLists {
		if value, found := entry[key]; found {
			holder = true
			held = append(held, listed(value, qualified(prefix, name, key))...)
		}
	}
	if !holder {
		return []string{prefix + name}
	}
	for _, key := range slices.Sorted(maps.Keys(entry)) {
		if !slices.Contains(toolLists, key) && !slices.Contains(labels, key) {
			held = append(held, prefix+key)
		}
	}
	return held
}

// qualified is the prefix for the tools an entry lists: the entry's name, but
// for a Gemini Tool's functions, whose entry has no name worth repeating.
func qualified(prefix, name, key string) string {
	if key == "functionDeclarations" || key == "function_declarations" {
		return prefix
	}
	return prefix + name + "."
}

func entryName(entry map[string]any) string {
	for _, key := range []string{"name", "type"} {
		if s, ok := entry[key].(string); ok && s != "" {
			return s
		}
	}
	keys := slices.Sorted(maps.Keys(entry))
	return orUnnamed(strings.Join(keys, "+"))
}

func orUnnamed(name string) string {
	if name == "" {
		return "(unnamed)"
	}
	return name
}

func count(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
