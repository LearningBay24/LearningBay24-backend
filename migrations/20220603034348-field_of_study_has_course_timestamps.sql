-- +migrate Up
ALTER TABLE field_of_study_has_course
ADD `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
ADD `updated_at` timestamp NULL DEFAULT NULL,
ADD `deleted_at` timestamp NULL DEFAULT NULL;


-- +migrate Down
ALTER TABLE field_of_study_has_course
DROP COLUMN  `created_at`,
DROP COLUMN  `updated_at`,
DROP COLUMN  `deleted_at`;