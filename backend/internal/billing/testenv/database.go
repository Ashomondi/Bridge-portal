// Package testenv provides disposable synthetic PostgreSQL fixtures for owned
// modules. It is never used by the application or worker binaries.
package testenv

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Admin *pgxpool.Pool
	role  string
}

func New(t *testing.T, applyOwned bool) *DB {
	t.Helper()
	dsn := os.Getenv("DEV05_TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("DEV05_TEST_DATABASE_URL must be set in CI; PostgreSQL tests cannot be skipped")
		}
		t.Skip("DEV05_TEST_DATABASE_URL is required for real PostgreSQL verification")
	}
	ctx := context.Background()
	root, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	var random [12]byte
	if _, err := rand.Read(random[:]); err != nil {
		t.Fatal(err)
	}
	name := "dev05_" + hex.EncodeToString(random[:])
	ident := pgx.Identifier{name}.Sanitize()
	if _, err := root.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
		root.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := root.Exec(context.Background(), "DROP DATABASE "+ident+" WITH (FORCE)"); err != nil {
			t.Error(err)
		}
		if _, err := root.Exec(context.Background(), "DROP ROLE IF EXISTS "+ident); err != nil {
			t.Error(err)
		}
		root.Close()
	})
	if _, err := root.Exec(ctx, "CREATE ROLE "+ident+" NOLOGIN NOSUPERUSER NOBYPASSRLS"); err != nil {
		t.Fatal(err)
	}
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.Database = name
	config.MaxConns = 8
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	db := &DB{Admin: pool, role: ident}
	db.Apply(t, "000001_init_schema.up.sql")
	db.Apply(t, "000003_r1_academics.up.sql")
	if applyOwned {
		db.Apply(t, "000004_r1_billing_whatsapp_notifications.up.sql")
	}
	db.Grant(t)
	return db
}

func (d *DB) Apply(t *testing.T, filename string) {
	t.Helper()
	_, source, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(source), "..", "..", "..", "..", "migrations", filename)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Admin.Exec(context.Background(), string(body)); err != nil {
		t.Fatalf("migration %s: %v", filename, err)
	}
}

func (d *DB) Grant(t *testing.T) {
	t.Helper()
	if _, err := d.Admin.Exec(context.Background(), "GRANT USAGE ON SCHEMA public TO "+d.role+"; GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO "+d.role); err != nil {
		t.Fatal(err)
	}
}

func (d *DB) WithTenantTx(ctx context.Context, schoolID string, fn func(pgx.Tx) error) error {
	tx, err := d.Admin.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(ctx, "SET LOCAL ROLE "+d.role); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "SELECT set_config('app.school_id', $1, true)", schoolID); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type School struct {
	ID, UserID, MembershipID, YearID, TermID, GradeID, ClassID, StudentID string
}

func (d *DB) SeedSchool(t *testing.T, code string) School {
	t.Helper()
	ctx := context.Background()
	var s School
	queries := []struct {
		sql  string
		args []any
		out  *string
	}{
		{"INSERT INTO schools(code,name) VALUES($1,$2) RETURNING id", []any{code, "Ikonex Demo School " + code}, &s.ID},
		{"INSERT INTO users(phone_number,full_name) VALUES($1,'Synthetic Guardian') RETURNING id", []any{"+2547000000" + code}, &s.UserID},
	}
	for _, q := range queries {
		if err := d.Admin.QueryRow(ctx, q.sql, q.args...).Scan(q.out); err != nil {
			t.Fatal(err)
		}
	}
	for _, q := range []struct {
		sql  string
		args func() []any
		out  *string
	}{
		{"INSERT INTO memberships(school_id,user_id,role) VALUES($1,$2,'parent') RETURNING id", func() []any { return []any{s.ID, s.UserID} }, &s.MembershipID},
		{"INSERT INTO academic_years(school_id,year,is_current) VALUES($1,2026,true) RETURNING id", func() []any { return []any{s.ID} }, &s.YearID},
		{"INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current) VALUES($1,$2,1,'2026-01-05','2026-04-10',true) RETURNING id", func() []any { return []any{s.ID, s.YearID} }, &s.TermID},
		{"INSERT INTO grade_levels(school_id,number,phase,name) VALUES($1,7,'junior','Grade 7') RETURNING id", func() []any { return []any{s.ID} }, &s.GradeID},
		{"INSERT INTO classes(school_id,grade_level_id,stream_name) VALUES($1,$2,'Demo') RETURNING id", func() []any { return []any{s.ID, s.GradeID} }, &s.ClassID},
		{"INSERT INTO students(school_id,admission_number,full_name,dob,gender,current_class_id) VALUES($1,'DEMO-001','Synthetic Student','2013-01-01','female',$2) RETURNING id", func() []any { return []any{s.ID, s.ClassID} }, &s.StudentID},
	} {
		if err := d.Admin.QueryRow(ctx, q.sql, q.args()...).Scan(q.out); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := d.Admin.Exec(ctx, "INSERT INTO student_guardians(school_id,student_id,user_id,relationship) VALUES($1,$2,$3,'guardian')", s.ID, s.StudentID, s.UserID); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Admin.Exec(ctx, "INSERT INTO enrolments(school_id,student_id,academic_year_id,grade_level_id,class_id) VALUES($1,$2,$3,$4,$5)", s.ID, s.StudentID, s.YearID, s.GradeID, s.ClassID); err != nil {
		t.Fatal(err)
	}
	return s
}
