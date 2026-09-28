package seal

import "testing"

// The subkey user identifiers are derived with is no purpose's: a purpose
// named like it would seal with the very key identifiers are made from.
func TestTheUserIDSubkeyIsNoPurposes(t *testing.T) {
	t.Parallel()

	for purpose := range labels {
		if subkeyInfo+string(purpose) == userIDInfo {
			t.Errorf("the purpose %q derives its subkey from %q, the context of user identifiers", purpose, userIDInfo)
		}
	}
}
