package oauthserver

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit/ratelimittest"
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

// renewingGoogle renews a grant with an access token good for as long as a
// case says, from the moment the case says it is.
type renewingGoogle struct {
	googleStandIn
	now      time.Time
	lifetime time.Duration
}

func (g renewingGoogle) Refresh(_ context.Context, refreshToken string) (googleauth.Renewal, error) {
	return googleauth.Renewal{AccessToken: "a-renewed-access-token", Expiry: g.now.Add(g.lifetime), RefreshToken: refreshToken}, nil
}

// issuingAt is the part of a server that issues a host's tokens, at testDay,
// over the ring given and with the Google given.
func issuingAt(t testing.TB, ring *seal.KeyRing, google googleauth.SignIn) *tokens {
	t.Helper()

	return &tokens{
		issuer: testIssuer, resource: testIssuer + "/mcp", scope: "mcp",
		clients: &clients{registrations: ring.For(seal.PurposeClient), issuer: testIssuer},
		codes:   ring.For(seal.PurposeCode), access: ring.For(seal.PurposeAccess), refresh: ring.For(seal.PurposeRefresh),
		google: google, renewals: ratelimittest.Roomy(t), events: &signInLog{logger: zap.NewNop()},
		now: func() time.Time { return testDay },
	}
}

// aSession is a session of a parent who signed in some time before testDay,
// holding a Google access token with some time left at testDay.
func aSession(user, client string, signedInAgo, googleLeft time.Duration) *session {
	return &session{
		user: user, client: client, resource: testIssuer + "/mcp", scope: "mcp",
		signedInAt: testDay.Add(-signedInAgo),
		google: googleGrant{
			accessToken: "an-access-token", accessExpiry: testDay.Add(googleLeft), refreshToken: "a-refresh-token",
		},
	}
}

func TestIssuedTokensEndInTime(t *testing.T) {
	t.Parallel()

	ring := ringOf(t, 'k')
	now := testDay.Unix()
	properties := gopter.NewProperties(nil)

	// The resource takes a token for a skew past its end, by a clock that may
	// run a skew behind the one that issued it, and the tool it lets in goes on
	// calling Drive for a request's span. All of that fits before the Google
	// token inside ends, or Drive answers the last requests of a token as it
	// answers access taken back.
	properties.Property("an access token ends within fifteen minutes, and early enough that a request it lets in reaches Drive before the Google token inside it ends", prop.ForAll(
		func(googleLeft, renewedFor int64) bool {
			grants := issuingAt(t, ring, renewingGoogle{now: testDay, lifetime: time.Duration(renewedFor) * time.Second})
			given, err := grants.issue(t.Context(), aSession("a-user", "a-client", 0, time.Duration(googleLeft)*time.Second))
			if err != nil {
				return errors.Is(err, errShortGrant)
			}
			var access accessGrant
			if grants.openAs(grants.access, given.access, &access) != nil {
				return false
			}
			return access.ExpiresAt > now && access.ExpiresAt <= now+int64(accessLifetime/time.Second) &&
				access.ExpiresAt+int64((2*clockSkew+requestSpan)/time.Second) <= access.GoogleAccessExpiry &&
				given.expiresIn == access.ExpiresAt-now
		},
		gen.Int64Range(-600, 1200), gen.Int64Range(0, 600),
	))

	properties.Property("no token outlives the ninety days of its sign-in, and a refresh token lasts thirty at most", prop.ForAll(
		func(signedInAgo int64) bool {
			grants := issuingAt(t, ring, googleStandIn{})
			signedIn := time.Duration(signedInAgo) * time.Second
			sessionEnd := testDay.Add(-signedIn).Unix() + int64(sessionLifetime/time.Second)
			given, err := grants.issue(t.Context(), aSession("a-user", "a-client", signedIn, time.Hour))
			if err != nil {
				return errors.Is(err, errSessionEnded) && sessionEnd <= now
			}
			var access accessGrant
			var refresh refreshGrant
			if grants.openAs(grants.access, given.access, &access) != nil || grants.openAs(grants.refresh, given.refresh, &refresh) != nil {
				return false
			}
			return access.ExpiresAt <= sessionEnd && refresh.ExpiresAt <= sessionEnd &&
				refresh.ExpiresAt <= now+int64(refreshLifetime/time.Second)
		},
		gen.Int64Range(0, 100*24*3600),
	))

	properties.TestingRun(t)
}

func TestIssuedTokensCarryTheirSession(t *testing.T) {
	t.Parallel()

	ring := ringOf(t, 'k')
	properties := gopter.NewProperties(nil)

	properties.Property("the tokens handed out carry the session they were issued for", prop.ForAll(
		func(user, client, country string, signedInAgo int64) bool {
			grants := issuingAt(t, ring, googleStandIn{})
			issuedFor := aSession(user, client, time.Duration(signedInAgo)*time.Second, time.Hour)
			issuedFor.country = country
			given, err := grants.issue(t.Context(), issuedFor)
			if err != nil {
				return false
			}
			var access accessGrant
			var refresh refreshGrant
			if grants.openAs(grants.access, given.access, &access) != nil || grants.openAs(grants.refresh, given.refresh, &refresh) != nil {
				return false
			}
			carried := sessionOf(&refresh)
			return access.User == user && access.Country == country && access.Client == client &&
				access.GoogleAccessToken == issuedFor.google.accessToken &&
				carried.user == user && carried.country == country && carried.client == client &&
				carried.resource == issuedFor.resource && carried.scope == issuedFor.scope &&
				carried.signedInAt.Equal(issuedFor.signedInAt) &&
				carried.google.accessToken == issuedFor.google.accessToken &&
				carried.google.accessExpiry.Equal(issuedFor.google.accessExpiry) &&
				carried.google.refreshToken == issuedFor.google.refreshToken
		},
		gen.RegexMatch(`[A-Za-z0-9_-]{16}`), gen.RegexMatch(`[A-Za-z0-9_-]{43}`), gen.OneConstOf("", "NZ", "US"), gen.Int64Range(0, 80*24*3600),
	))

	properties.TestingRun(t)
}
