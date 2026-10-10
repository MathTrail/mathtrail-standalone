// Package plugin_test holds the package ChatGPT's directory takes, in this
// folder, to what the directory accepts and to what the listing promises: the
// limits of every text, the addresses it leads to, the icons, the review's
// cases, and no credential and no price anywhere in it.
package plugin_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// The addresses the listing leads to: the site's pages, and the service.
const (
	site   = "https://mathtrail.app"
	server = "https://mcp.mathtrail.app/mcp"
)

// modelTools are the tools the chat's model calls, as the service registers
// them; the card's own three are not among them, since no prompt triggers one.
var modelTools = []string{
	"get_profile", "save_profile", "get_progress",
	"next_task", "get_package", "prepare_task", "submit_task", "submit_answer", "show_result",
}

// manifest is plugin.json as the Agent Plugins format and OpenAI's extension
// of it write it. It is read refusing any field it does not name, so a field
// misspelled is found here rather than at the upload.
type manifest struct {
	Schema      string   `json:"$schema"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      author   `json:"author"`
	Homepage    string   `json:"homepage"`
	Repository  string   `json:"repository"`
	License     string   `json:"license"`
	Keywords    []string `json:"keywords"`
	Extensions  struct {
		OpenAI struct {
			Interface   listing     `json:"interface"`
			Review      review      `json:"review"`
			Publication publication `json:"publication"`
		} `json:"com.openai"`
	} `json:"extensions"`
}

type author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	URL   string `json:"url"`
}

type listing struct {
	DisplayName       string   `json:"displayName"`
	ShortDescription  string   `json:"shortDescription"`
	LongDescription   string   `json:"longDescription"`
	DeveloperName     string   `json:"developerName"`
	Category          string   `json:"category"`
	Capabilities      []string `json:"capabilities"`
	WebsiteURL        string   `json:"websiteURL"`
	SupportURL        string   `json:"supportURL"`
	PrivacyPolicyURL  string   `json:"privacyPolicyURL"`
	TermsOfServiceURL string   `json:"termsOfServiceURL"`
	DefaultPrompt     []string `json:"defaultPrompt"`
	ComposerIcon      string   `json:"composerIcon"`
	ComposerIconDark  string   `json:"composerIconDark"`
	Logo              string   `json:"logo"`
	LogoDark          string   `json:"logoDark"`
}

type review struct {
	TestCases struct {
		Positive []testCase `json:"positive"`
		Negative []testCase `json:"negative"`
	} `json:"test_cases"`
	Commerce            *bool  `json:"commerce"`
	CommerceDescription string `json:"commerce_description"`
}

type testCase struct {
	Description      string `json:"description"`
	Prompt           string `json:"prompt"`
	ToolsTriggered   string `json:"tools_triggered"`
	ExpectedBehavior string `json:"expected_behavior"`
}

type publication struct {
	ReleaseNotes string                 `json:"release_notes"`
	Translations map[string]translation `json:"translations"`
}

type translation struct {
	Subtitle    string `json:"subtitle"`
	Description string `json:"description"`
}

// mcpConfig is mcp.json: the one server the package connects.
type mcpConfig struct {
	Schema     string                       `json:"$schema"`
	MCPServers map[string]map[string]string `json:"mcpServers"`
}

func strictly(t *testing.T, name string, into any) {
	t.Helper()

	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v, want nil", name, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil {
		t.Fatalf("decode %s: %v, want a manifest of the known fields alone", name, err)
	}
}

func readManifest(t *testing.T) manifest {
	t.Helper()

	var m manifest
	strictly(t, "plugin.json", &m)
	return m
}

// claudeFields are the texts of Claude's listing, each in a block of
// docs/listing.md whose info string names it, which is what is pasted.
func claudeFields(t *testing.T) map[string]string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "docs", "listing.md"))
	if err != nil {
		t.Fatalf("ReadFile(docs/listing.md) error = %v, want nil", err)
	}
	fields := map[string]string{}
	for _, found := range regexp.MustCompile("(?s)```text ([a-z-]+)\n(.*?)\n```").FindAllStringSubmatch(string(data), -1) {
		if _, again := fields[found[1]]; again {
			t.Errorf("docs/listing.md holds the block %s twice, want it once", found[1])
		}
		fields[found[1]] = found[2]
	}
	for _, name := range []string{"claude-name", "claude-one-liner", "claude-description", "claude-categories", "company"} {
		if _, ok := fields[name]; !ok {
			t.Fatalf("docs/listing.md holds no block %s, want one", name)
		}
	}
	return fields
}

// limit is a text the directories hold to a number of characters.
type limit struct {
	field   string
	text    string
	most    int
	oneLine bool
}

func limitsOf(m *manifest, claude map[string]string) []limit {
	face := m.Extensions.OpenAI.Interface
	limits := []limit{
		{"description", m.Description, 1024, false},
		{"author.name", m.Author.Name, 120, true},
		{"interface.displayName", face.DisplayName, 30, true},
		{"interface.shortDescription", face.ShortDescription, 30, true},
		{"interface.longDescription", face.LongDescription, 4000, false},
		{"interface.developerName", face.DeveloperName, 80, true},
		{"publication.release_notes", m.Extensions.OpenAI.Publication.ReleaseNotes, 4000, false},
		{"Claude's name", claude["claude-name"], 100, true},
		{"Claude's one-liner", claude["claude-one-liner"], 200, true},
		{"Claude's description", claude["claude-description"], 2000, false},
	}
	for i, capability := range face.Capabilities {
		limits = append(limits, limit{fmt.Sprintf("interface.capabilities[%d]", i), capability, 120, true})
	}
	for i, prompt := range face.DefaultPrompt {
		limits = append(limits, limit{fmt.Sprintf("interface.defaultPrompt[%d]", i), prompt, 128, true})
	}
	for _, locale := range slices.Sorted(maps.Keys(m.Extensions.OpenAI.Publication.Translations)) {
		translated := m.Extensions.OpenAI.Publication.Translations[locale]
		limits = append(limits,
			limit{"translations." + locale + ".subtitle", translated.Subtitle, 30, true},
			limit{"translations." + locale + ".description", translated.Description, 4000, false})
	}
	for i, positive := range m.Extensions.OpenAI.Review.TestCases.Positive {
		limits = append(limits, limit{fmt.Sprintf("test_cases.positive[%d].description", i), positive.Description, 4000, false})
	}
	return limits
}

// Every text of both listings is held to the directory's limit, counted in
// characters; run with -v, the test prints each count beside its limit.
func TestTheListingTextsAreWithinTheirLimits(t *testing.T) {
	t.Parallel()

	m := readManifest(t)
	claude := claudeFields(t)
	for _, l := range limitsOf(&m, claude) {
		count := utf8.RuneCountInString(l.text)
		t.Logf("%s: %d of %d", l.field, count, l.most)
		if strings.TrimSpace(l.text) == "" {
			t.Errorf("%s is empty, want a text", l.field)
		}
		if count > l.most {
			t.Errorf("%s has %d characters, want at most %d", l.field, count, l.most)
		}
		if l.oneLine && strings.Contains(l.text, "\n") {
			t.Errorf("%s holds a line break, want one line", l.field)
		}
	}
	if n := len(m.Extensions.OpenAI.Interface.Capabilities); n > 20 {
		t.Errorf("interface.capabilities has %d labels, want at most 20", n)
	}
	if n := len(m.Extensions.OpenAI.Interface.DefaultPrompt); n < 1 || n > 3 {
		t.Errorf("interface.defaultPrompt has %d prompts, want 1 to 3", n)
	}
	categories := strings.Split(claude["claude-categories"], "\n")
	if len(categories) > 5 || slices.ContainsFunc(categories, func(category string) bool { return strings.TrimSpace(category) == "" }) {
		t.Errorf("Claude's categories are %q, want 1 to 5 of them, one a line", categories)
	}
}

// walk visits every key and every text of a decoded JSON document, with
// where each is: a key with no text, a text with no key.
func walk(value any, at string, visit func(at, key string, text *string)) {
	switch v := value.(type) {
	case map[string]any:
		for key, inner := range v {
			visit(at, key, nil)
			walk(inner, at+"."+key, visit)
		}
	case []any:
		for i, inner := range v {
			walk(inner, fmt.Sprintf("%s[%d]", at, i), visit)
		}
	case string:
		visit(at, "", &v)
	}
}

func readAny(t *testing.T, name string) any {
	t.Helper()

	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v, want nil", name, err)
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("Unmarshal(%s) error = %v, want nil", name, err)
	}
	return value
}

// The directory refuses control characters, the line and paragraph
// separators and the invisible marks that format text, wherever they are.
func TestTheListingTextsHoldOnlyCharactersADirectoryTakes(t *testing.T) {
	t.Parallel()

	walk(readAny(t, "plugin.json"), "plugin.json", func(at, _ string, text *string) {
		if text == nil {
			return
		}
		for _, r := range *text {
			if (unicode.IsControl(r) && r != '\n') || r == '\u2028' || r == '\u2029' || unicode.Is(unicode.Cf, r) {
				t.Errorf("%s holds the character %U, want none of control, separator or format", at, r)
			}
		}
	})
}

// priceWords are what the directory's rules forbid a listing to sell by:
// pricing, subscriptions, free trials, discounts and promotions, in English
// and in Russian. A word is told by the letters around it rather than by \b,
// which knows only the Latin letters as letters.
var priceWords = regexp.MustCompile(`(?i)(?:^|\P{L})(free|prices?|priced|pricing|subscri\p{L}*|trials?|discounts?|promotions?|paid|pay|payments?|бесплатн\p{L}*|цен[аыуе]|ценой|стоимост\p{L}*|подписк\p{L}*|скидк\p{L}*|акци[яиюей]|оплат\p{L}*|платн\p{L}*)(?:\P{L}|$)`)

// A listing sells nothing and names no protocol after the name; the
// commerce declaration, which says that nothing is sold, is not a listing.
func TestTheListingTextsSellNothing(t *testing.T) {
	t.Parallel()

	m := readManifest(t)
	face := m.Extensions.OpenAI.Interface
	texts := map[string]string{
		"description":                m.Description,
		"interface.shortDescription": face.ShortDescription,
		"interface.longDescription":  face.LongDescription,
		"publication.release_notes":  m.Extensions.OpenAI.Publication.ReleaseNotes,
	}
	for i, text := range face.Capabilities {
		texts[fmt.Sprintf("interface.capabilities[%d]", i)] = text
	}
	for i, text := range face.DefaultPrompt {
		texts[fmt.Sprintf("interface.defaultPrompt[%d]", i)] = text
	}
	for locale, translated := range m.Extensions.OpenAI.Publication.Translations {
		texts["translations."+locale+".subtitle"] = translated.Subtitle
		texts["translations."+locale+".description"] = translated.Description
	}
	for name, text := range claudeFields(t) {
		texts["docs/listing.md "+name] = text
	}
	afterName := regexp.MustCompile(`(?i)mathtrail\s+(mcp|plugin)`)
	for _, field := range slices.Sorted(maps.Keys(texts)) {
		if found := priceWords.FindStringSubmatch(texts[field]); found != nil {
			t.Errorf("%s says %q, want no price, subscription, trial, discount or promotion", field, found[1])
		}
		if found := afterName.FindString(texts[field]); found != "" {
			t.Errorf("%s says %q, want no MCP or Plugin after the name", field, found)
		}
	}
	switch commerce := m.Extensions.OpenAI.Review.Commerce; {
	case commerce == nil:
		t.Error("review.commerce is not given, want false")
	case *commerce:
		t.Error("review.commerce = true, want false")
	}
}

// A package carries no credential: ChatGPT refuses test credentials and
// reviewer instructions in a ZIP, and the reviewer's password goes into the
// dashboard's form alone. No key names one, and no text speaks of one, as a
// password pasted into a case's expected behaviour would.
func TestThePackageHoldsNoCredentials(t *testing.T) {
	t.Parallel()

	secretKey := regexp.MustCompile(`(?i)password|secret|token|credential|reviewer_instructions`)
	secretText := regexp.MustCompile(`(?i)(?:^|\P{L})(password|passwd|secret|api[ _-]?key|bearer|пароль|пароля|паролем)(?:\P{L}|$)`)
	for _, name := range []string{"plugin.json", "mcp.json"} {
		walk(readAny(t, name), name, func(at, key string, text *string) {
			if key != "" && secretKey.MatchString(key) {
				t.Errorf("%s has the key %q, want no credential in the package", at, key)
			}
			if text != nil && secretText.MatchString(*text) {
				t.Errorf("%s speaks of %q, want no credential in the package", at, secretText.FindStringSubmatch(*text)[1])
			}
		})
	}
}

// The package is the format it names: its schema, a stable name, a version
// raised with every ZIP, the product's own name and the directory's category.
func TestThePackageIsWhatItNamesItself(t *testing.T) {
	t.Parallel()

	m := readManifest(t)
	if want := "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json"; m.Schema != want {
		t.Errorf("$schema = %q, want %q", m.Schema, want)
	}
	if !regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(m.Name) {
		t.Errorf("name = %q, want lowercase letters, digits and single hyphens", m.Name)
	}
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(m.Version) {
		t.Errorf("version = %q, want a semantic version", m.Version)
	}
	face := m.Extensions.OpenAI.Interface
	if face.DisplayName != "MathTrail" {
		t.Errorf("interface.displayName = %q, want MathTrail", face.DisplayName)
	}
	if want := "Education & Research"; face.Category != want {
		t.Errorf("interface.category = %q, want %q", face.Category, want)
	}
}

// The listing leads to the site's pages and the package to the service; each
// page the listing names is one the site publishes, and every address is
// HTTPS with no credentials in it.
func TestThePackageLeadsWhereMathTrailIs(t *testing.T) {
	t.Parallel()

	m := readManifest(t)
	face := m.Extensions.OpenAI.Interface
	pages := map[string]struct{ got, file string }{
		site + "/":            {face.WebsiteURL, "index.yaml"},
		site + "/en/help/":    {face.SupportURL, "help.md"},
		site + "/en/privacy/": {face.PrivacyPolicyURL, "privacy.md"},
		site + "/en/terms/":   {face.TermsOfServiceURL, "terms.md"},
	}
	for want, page := range pages {
		if page.got != want {
			t.Errorf("a listing address is %q, want %q", page.got, want)
		}
		if _, err := os.Stat(filepath.Join("..", "site", "content", "en", page.file)); err != nil {
			t.Errorf("the site has no page for %s: %v", want, err)
		}
	}
	for _, address := range []string{m.Homepage, m.Author.URL, face.WebsiteURL, face.SupportURL, face.PrivacyPolicyURL, face.TermsOfServiceURL} {
		parsed, err := url.Parse(address)
		if err != nil || parsed.Scheme != "https" || parsed.User != nil {
			t.Errorf("%q is not an HTTPS address without credentials", address)
		}
	}
}

// The package connects one server, the service, over streamable HTTP, and
// says nothing else about it.
func TestThePackageConnectsTheServiceAlone(t *testing.T) {
	t.Parallel()

	var config mcpConfig
	strictly(t, "mcp.json", &config)
	if want := "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json"; config.Schema != want {
		t.Errorf("mcp.json $schema = %q, want %q", config.Schema, want)
	}
	want := map[string]map[string]string{"mathtrail": {"type": "streamable-http", "url": server}}
	if !maps.EqualFunc(config.MCPServers, want, maps.Equal) {
		t.Errorf("mcp.json servers = %v, want %v", config.MCPServers, want)
	}
}

// The first review asks for exactly five positive and three negative cases;
// a positive one names the prompt, the tools it triggers, every one a tool
// the model calls, and what is seen.
func TestTheReviewHasFivePositiveAndThreeNegativeCases(t *testing.T) {
	t.Parallel()

	cases := readManifest(t).Extensions.OpenAI.Review.TestCases
	if len(cases.Positive) != 5 || len(cases.Negative) != 3 {
		t.Fatalf("the review has %d positive and %d negative cases, want 5 and 3", len(cases.Positive), len(cases.Negative))
	}
	for i, c := range cases.Positive {
		checkPositiveCase(t, i, &c)
	}
	for i, c := range cases.Negative {
		if c.Description == "" || c.Prompt == "" || c.ToolsTriggered != "" || c.ExpectedBehavior != "" {
			t.Errorf("negative case %d = %+v, want a description and a prompt alone", i, c)
		}
	}
}

// checkPositiveCase holds a positive case to what the review reads: the
// prompt, the tools it triggers, each one the model calls, and what is seen.
func checkPositiveCase(t *testing.T, i int, c *testCase) {
	t.Helper()

	if c.Description == "" || c.Prompt == "" || c.ToolsTriggered == "" || c.ExpectedBehavior == "" {
		t.Errorf("positive case %d = %+v, want a description, a prompt, its tools and what is seen", i, *c)
	}
	for tool := range strings.SplitSeq(c.ToolsTriggered, ",") {
		if !slices.Contains(modelTools, strings.TrimSpace(tool)) {
			t.Errorf("positive case %d triggers %q, want one of %v", i, strings.TrimSpace(tool), modelTools)
		}
	}
}

// The starter prompts are different from one another and hold no @mention
// of an app, as the directory asks.
func TestTheStarterPromptsAreDistinctAndHoldNoMention(t *testing.T) {
	t.Parallel()

	prompts := readManifest(t).Extensions.OpenAI.Interface.DefaultPrompt
	for i, prompt := range prompts {
		if strings.Contains(prompt, "@") || slices.Index(prompts, prompt) != i {
			t.Errorf("starter prompt %q is repeated or holds an @mention, want neither", prompt)
		}
	}
}

// Both directories take a square picture: one of 512 to 2048 pixels under
// 2 MB, the other of at least 48 pixels under 5 MiB. Each icon's corners are
// clear, since the tile rounds them off, and its middle is the logo itself.
func TestTheIconsAreSquarePicturesWithClearCorners(t *testing.T) {
	t.Parallel()

	face := readManifest(t).Extensions.OpenAI.Interface
	named := map[string]bool{}
	for _, path := range []string{face.Logo, face.LogoDark, face.ComposerIcon, face.ComposerIconDark} {
		if !strings.HasPrefix(path, "./assets/") {
			t.Errorf("an icon's path is %q, want one under ./assets/", path)
			continue
		}
		named[strings.TrimPrefix(path, "./")] = true
	}
	kept, err := filepath.Glob(filepath.Join("assets", "*"))
	if err != nil {
		t.Fatalf("Glob(assets/*) error = %v, want nil", err)
	}
	if got, want := slices.Sorted(slices.Values(kept)), slices.Sorted(maps.Keys(named)); !slices.Equal(got, want) {
		t.Errorf("assets/ holds %v, want exactly the icons the manifest names, %v", got, want)
	}
	for _, path := range slices.Sorted(maps.Keys(named)) {
		checkIcon(t, path)
	}
}

// checkIcon holds one icon to what both directories take.
func checkIcon(t *testing.T, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("ReadFile(%s) error = %v, want nil", path, err)
		return
	}
	if len(data) >= 2_000_000 {
		t.Errorf("%s has %d bytes, want under 2,000,000", path, len(data))
	}
	picture, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("png.Decode(%s) error = %v, want a PNG", path, err)
		return
	}
	bounds := picture.Bounds()
	if bounds.Dx() != bounds.Dy() || bounds.Dx() < 512 || bounds.Dx() > 2048 {
		t.Errorf("%s is %dx%d, want a square of 512 to 2048 pixels", path, bounds.Dx(), bounds.Dy())
	}
	corners := []image.Point{
		bounds.Min, {bounds.Max.X - 1, bounds.Min.Y},
		{bounds.Min.X, bounds.Max.Y - 1}, bounds.Max.Sub(image.Pt(1, 1)),
	}
	for _, corner := range corners {
		if _, _, _, alpha := picture.At(corner.X, corner.Y).RGBA(); alpha != 0 {
			t.Errorf("%s's corner %v has alpha %d, want it clear", path, corner, alpha)
		}
	}
	middle := bounds.Min.Add(bounds.Size().Div(2))
	if _, _, _, alpha := picture.At(middle.X, middle.Y).RGBA(); alpha != 0xffff {
		t.Errorf("%s's middle has alpha %d, want it opaque", path, alpha)
	}
}

// The developer is one name: the manifest's author, the listing's developer
// and the company Claude's form is given are the same, so that replacing the
// placeholder in one place and not another is found here.
func TestTheDeveloperIsNamedTheSameEverywhere(t *testing.T) {
	t.Parallel()

	m := readManifest(t)
	company := claudeFields(t)["company"]
	if m.Author.Name != company || m.Extensions.OpenAI.Interface.DeveloperName != company {
		t.Errorf("author.name = %q, interface.developerName = %q, Claude's company = %q, want one name",
			m.Author.Name, m.Extensions.OpenAI.Interface.DeveloperName, company)
	}
}
