package main

import (
	"regexp"
	"strings"
)

var newArXivID = regexp.MustCompile(`^\d{4}\.\d{4,5}(v\d+)?$`)

// DOIFromID reads what a person types to name a work: a DOI, with or without
// "doi:" or a resolver's address (doi.org, dx.doi.org), or an arXiv id, with or
// without "arXiv:", or the address of its abstract or PDF.
func DOIFromID(id string) string {
	id = strings.TrimSpace(id)
	for _, prefix := range []string{"https://doi.org/", "http://doi.org/", "https://dx.doi.org/", "http://dx.doi.org/", "doi:"} {
		if rest, found := strings.CutPrefix(strings.ToLower(id), prefix); found {
			return id[len(id)-len(rest):]
		}
	}
	lower := strings.ToLower(id)
	for _, prefix := range []string{"https://arxiv.org/abs/", "http://arxiv.org/abs/", "https://arxiv.org/pdf/", "http://arxiv.org/pdf/"} {
		if rest, found := strings.CutPrefix(lower, prefix); found {
			return arXivDOIPrefix + arXivVersion.ReplaceAllString(strings.TrimSuffix(rest, ".pdf"), "")
		}
	}
	if rest, found := strings.CutPrefix(lower, "arxiv:"); found {
		return arXivDOIPrefix + arXivVersion.ReplaceAllString(rest, "")
	}
	if newArXivID.MatchString(id) {
		return arXivDOIPrefix + arXivVersion.ReplaceAllString(id, "")
	}
	return id
}
