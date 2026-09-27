package oauthserver

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// genRedirect produces an address a sign-in may send the parent back to.
func genRedirect() gopter.Gen {
	return gopter.CombineGens(
		gen.OneConstOf("https://", "http://localhost:", "http://127.0.0.1:"),
		gen.RegexMatch(`[a-z0-9]{1,20}`),
		gen.RegexMatch(`[a-zA-Z0-9_\-/]{0,40}`),
	).Map(func(parts []interface{}) string {
		scheme, _ := parts[0].(string)
		host, _ := parts[1].(string)
		path, _ := parts[2].(string)
		if scheme != "https://" {
			return fmt.Sprintf("%s%d/%s", scheme, 1024+len(host), path)
		}
		return scheme + host + ".example/" + path
	})
}

// genRegistration produces a registration a client could make: a few
// addresses, and a name that fits.
func genRegistration() gopter.Gen {
	return gopter.CombineGens(
		gen.SliceOfN(maxRedirectURIs, genRedirect()).SuchThat(func(uris []string) bool { return len(uris) > 0 }),
		gen.AnyString().SuchThat(func(name string) bool { return len([]rune(name)) <= maxClientName }),
		gen.Int64Range(0, 1<<40),
	).Map(func(parts []interface{}) registration {
		uris, _ := parts[0].([]string)
		name, _ := parts[1].(string)
		issuedAt, _ := parts[2].(int64)
		return registration{RedirectURIs: uris, ClientName: name, IssuedAt: issuedAt}
	})
}

func TestRegistrationsHoldTheirProperties(t *testing.T) {
	t.Parallel()

	ring := ringOf(t, 'k')
	known, _ := knownClients(t, ring, &documents{})
	properties := gopter.NewProperties(nil)

	properties.Property("a registration opens as itself", prop.ForAll(
		func(record registration) bool {
			clientID, err := known.register(record)
			if err != nil {
				return false
			}
			client, err := known.resolve(t.Context(), clientID)
			return err == nil && client.Registration == registrationDCR &&
				slices.Equal(client.RedirectURIs, record.RedirectURIs) && client.Name == record.ClientName
		},
		genRegistration(),
	))

	properties.Property("a registration made for one issuer never opens at another", prop.ForAll(
		func(record registration, other string) bool {
			elsewhere := &clients{
				registrations: known.registrations, issuer: other, documents: known.documents,
				events: known.events, now: known.now,
			}
			clientID, err := elsewhere.register(record)
			if err != nil {
				return false
			}
			_, err = known.resolve(t.Context(), clientID)
			return errors.Is(err, errUnknownClient)
		},
		genRegistration(), gen.AnyString().SuchThat(func(other string) bool { return other != testIssuer }),
	))

	properties.TestingRun(t)
}
