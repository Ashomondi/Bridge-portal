-- ==============================================================================
-- 000003_r1_academics.down.sql
-- Rollback Release 1 CBC Academics & Classroom Operations
-- ==============================================================================

DROP TABLE IF EXISTS released_marks_snapshots CASCADE;
DROP TABLE IF EXISTS result_releases CASCADE;
DROP TABLE IF EXISTS report_comments CASCADE;
DROP TABLE IF EXISTS student_marks CASCADE;
DROP TABLE IF EXISTS attendance_daily CASCADE;
DROP TABLE IF EXISTS calendar_events CASCADE;
DROP TABLE IF EXISTS timetable_slots CASCADE;
DROP TABLE IF EXISTS grading_scale_versions CASCADE;
DROP TABLE IF EXISTS subject_teacher_assignments CASCADE;
DROP TABLE IF EXISTS subjects CASCADE;
