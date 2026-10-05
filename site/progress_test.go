package site_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The progress a page draws is written out by hand, but its ranks are the
// service's: as many as there are, and the grades marked under them matched
// with the ranks the service matches them with. A sample that drifted from
// the rule would show a parent a match the service never makes.
func TestTheSampleProgressMarksTheGradesAsTheServiceDoes(t *testing.T) {
	t.Parallel()

	file, err := os.ReadFile("data.json")
	if err != nil {
		t.Fatalf("ReadFile(data.json) error = %v, want the site's data", err)
	}
	type grades struct {
		GradeLevel string `json:"grade_level"`
		FirstRank  int    `json:"first_rank"`
		LastRank   int    `json:"last_rank"`
	}
	var data struct {
		Progress struct {
			Overall struct {
				Ranks  int      `json:"ranks"`
				Grades []grades `json:"grades"`
			} `json:"overall"`
		} `json:"progress"`
	}
	if err := json.Unmarshal(file, &data); err != nil {
		t.Fatalf("Unmarshal(data.json) error = %v, want nil", err)
	}

	overall := data.Progress.Overall
	if overall.Ranks != rating.Ranks {
		t.Errorf("the sample's ranks = %d, want %d", overall.Ranks, rating.Ranks)
	}
	levels := rating.GradeLevels()
	if len(overall.Grades) != len(levels) {
		t.Fatalf("the sample marks %d runs of ranks, want one for each of the %d levels", len(overall.Grades), len(levels))
	}
	for i, level := range levels {
		want := grades{GradeLevel: string(level), FirstRank: level.FirstRank(), LastRank: level.LastRank()}
		if overall.Grades[i] != want {
			t.Errorf("the sample's grades[%d] = %+v, want %+v", i, overall.Grades[i], want)
		}
	}
}
