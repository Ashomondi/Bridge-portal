package internalapi

import (
	"context"
	"database/sql"
	"os"
	"testing"
)

func TestSQLReleasedMarksUseLockedScale(t *testing.T) {
	url := os.Getenv("DEV5_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("disposable PostgreSQL database and registered driver required")
	}
	app, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	admin, err := sql.Open("pgx", os.Getenv("DEV5_TEST_ADMIN_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	school := "00000000-0000-4000-8000-000000000001"
	student := "00000000-0000-4000-8000-000000000100"
	term := "00000000-0000-4000-8000-000000000040"
	if _, err := admin.Exec(`
 DELETE FROM released_marks_snapshots; DELETE FROM result_releases; DELETE FROM subjects; DELETE FROM grading_scale_versions;
 INSERT INTO subjects(id,school_id,grade_level_id,code,name) VALUES ('00000000-0000-4000-8000-000000000601','00000000-0000-4000-8000-000000000001','00000000-0000-4000-8000-000000000020','SYN','Synthetic subject');
 INSERT INTO grading_scale_versions(id,school_id,version_number,ee_min,me_min,ae_min,is_active) VALUES ('00000000-0000-4000-8000-000000000602','00000000-0000-4000-8000-000000000001',1,80,60,40,false),('00000000-0000-4000-8000-000000000603','00000000-0000-4000-8000-000000000001',2,60,50,30,true);
 INSERT INTO result_releases(id,school_id,class_id,term_id,grading_scale_version_id) VALUES ('00000000-0000-4000-8000-000000000604','00000000-0000-4000-8000-000000000001','00000000-0000-4000-8000-000000000030','00000000-0000-4000-8000-000000000040','00000000-0000-4000-8000-000000000602');
 INSERT INTO released_marks_snapshots(school_id,result_release_id,student_id,term_id,subject_id,score,cbc_level) VALUES ('00000000-0000-4000-8000-000000000001','00000000-0000-4000-8000-000000000604','00000000-0000-4000-8000-000000000100','00000000-0000-4000-8000-000000000040','00000000-0000-4000-8000-000000000601',65,'ME');
 `); err != nil {
		t.Fatal(err)
	}
	store := NewSQLStore(app)
	for _, tc := range []struct {
		status, scale, grade string
		released, allowed    bool
	}{
		{"draft", "602", "", false, false}, {"released", "602", "", false, false},
		{"released", "602", "ME", true, true}, {"reopened", "602", "ME", true, true},
		{"rereleased", "603", "EE", true, true},
	} {
		if _, err := admin.Exec(`UPDATE result_releases SET status=$1, grading_scale_version_id=$2, released_at=CASE WHEN $3 THEN NOW() ELSE NULL END`, tc.status, "00000000-0000-4000-8000-000000000"+tc.scale, tc.released); err != nil {
			t.Fatal(err)
		}
		err := store.WithTx(context.Background(), school, func(tx Store) error {
			summary, err := tx.GetMarksSummary(context.Background(), school, student, term)
			if tc.allowed {
				if err != nil || summary.MeanGrade != tc.grade || summary.Subjects[0].CBCLevel != "ME" {
					t.Fatalf("locked scale %s/%s: %+v %v", tc.status, tc.scale, summary, err)
				}
			} else if err == nil {
				t.Fatal("unreleased marks exposed")
			}
			if _, err := tx.GetMarksSummary(context.Background(), school, "00000000-0000-4000-8000-000000000200", term); err == nil {
				t.Fatal("cross-school marks exposed")
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
