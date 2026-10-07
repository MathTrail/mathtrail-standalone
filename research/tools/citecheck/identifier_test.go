package main

import "testing"

func TestDOIFromID(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ id, want string }{
		{"10.1016/j.compedu.2011.02.003", "10.1016/j.compedu.2011.02.003"},
		{"doi:10.1016/j.compedu.2011.02.003", "10.1016/j.compedu.2011.02.003"},
		{"https://doi.org/10.1016/J.COMPEDU.2011.02.003", "10.1016/J.COMPEDU.2011.02.003"},
		{"arXiv:2503.16460v2", "10.48550/arXiv.2503.16460"},
		{"2402.15861", "10.48550/arXiv.2402.15861"},
		{" 10.48550/arXiv.2402.15861 ", "10.48550/arXiv.2402.15861"},
		{"http://doi.org/10.5281/zenodo.3554625", "10.5281/zenodo.3554625"},
		{"https://dx.doi.org/10.1016/j.compedu.2016.03.017", "10.1016/j.compedu.2016.03.017"},
		{"https://arxiv.org/abs/2404.18796v2", "10.48550/arXiv.2404.18796"},
		{"https://arxiv.org/pdf/2404.18796.pdf", "10.48550/arXiv.2404.18796"},
	} {
		t.Run(tt.id, func(t *testing.T) {
			t.Parallel()
			if got := DOIFromID(tt.id); got != tt.want {
				t.Errorf("DOIFromID(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
