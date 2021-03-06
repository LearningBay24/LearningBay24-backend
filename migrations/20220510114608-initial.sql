-- +migrate Up
CREATE TABLE `appointment` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `date` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT 'The date the appointment should be.',
  `location` varchar(256) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'Represents where the appointment will be held. For example, this could be a room number for an offline appointment or a URL for an online appointment.',
  `online` tinyint(4) NOT NULL COMMENT 'Whether the appointment is held online or offline.',
  `course_id` int(11) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_appointment_course1_idx` (`course_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `certificate` (
  `id` char(36) COLLATE utf8_unicode_ci NOT NULL COMMENT 'UUID generated before insertion into the database.\nNeeds to be a randomly generated String of a specific format in order to avoid getting any desired certificate.',
  `user_id` int(11) NOT NULL,
  `linked_course_id` int(11) NOT NULL,
  `exam_id` int(11) DEFAULT NULL COMMENT 'The exam (if any) that was passed in order to gain this certificate.',
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_certificate_user_idx` (`user_id`),
  KEY `fk_certificate_course1_idx` (`linked_course_id`),
  KEY `fk_certificate_exam1_idx` (`exam_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `course` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8_unicode_ci NOT NULL,
  `description` varchar(512) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'The detailed description of this course.',
  `enroll_key` varchar(45) COLLATE utf8_unicode_ci NOT NULL,
  `forum_id` int(11) NOT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  KEY `fk_course_forum1_idx` (`forum_id`)
) ENGINE=InnoDB AUTO_INCREMENT=41 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `course_has_files` (
  `course_id` int(11) NOT NULL,
  `file_id` int(11) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`course_id`,`file_id`),
  KEY `fk_course_has_files_file1_idx` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `course_requires_certificate` (
  `course_id` int(11) NOT NULL,
  `certificate_id` char(36) COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`course_id`,`certificate_id`),
  KEY `fk_certificate_has_course_course1_idx` (`course_id`),
  KEY `fk_certificate_has_course_certificate1_idx` (`certificate_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `directory` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(128) COLLATE utf8_unicode_ci NOT NULL COMMENT 'The displayed name of the directory.',
  `course_id` int(11) NOT NULL COMMENT 'The course this directory is displayed in.',
  `visible_from` timestamp NOT NULL DEFAULT current_timestamp() COMMENT 'At which date the folder will be visible to enrolled users.',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT 'The date this directory has been created.',
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_directory_course1_idx` (`course_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci COMMENT='A directory in the listing of materials of a course containing materials.';

CREATE TABLE `directory_has_files` (
  `directory_id` int(11) NOT NULL,
  `file_id` int(11) NOT NULL,
  PRIMARY KEY (`directory_id`,`file_id`),
  KEY `fk_directory_has_files_directory1_idx` (`directory_id`),
  KEY `fk_directory_has_files_file1_idx` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `exam` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(128) COLLATE utf8_unicode_ci NOT NULL COMMENT 'Name of the exam.',
  `description` varchar(512) COLLATE utf8_unicode_ci NOT NULL COMMENT 'Description of the exam.',
  `date` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT 'Date the exam will be held.',
  `duration` int(11) NOT NULL COMMENT 'How long the exam will be in seconds.',
  `online` tinyint(4) NOT NULL COMMENT 'Whether the exam is an online or offline one.',
  `location` varchar(256) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'Location where the exam takes place.',
  `course_id` int(11) NOT NULL COMMENT 'The course this exam is part of.',
  `creator_id` int(11) NOT NULL COMMENT 'Creator of the exam.',
  `graded` tinyint(4) NOT NULL DEFAULT 0 COMMENT 'Whether all the submissions for this exam have been graded.',
  `register_deadline` timestamp NULL DEFAULT NULL COMMENT 'When the deadline to register is, if there is one.',
  `deregister_deadline` timestamp NULL DEFAULT NULL COMMENT 'When the deadline to deregister is, if there is one.',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  KEY `fk_exam_course1_idx` (`course_id`),
  KEY `fk_exam_user1_idx` (`creator_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `exam_has_files` (
  `exam_id` int(11) NOT NULL,
  `file_id` int(11) NOT NULL,
  PRIMARY KEY (`exam_id`,`file_id`),
  KEY `fk_exam_has_files_file1_idx` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `field_of_study` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(45) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'Name of the field of study.',
  `semesters` varchar(45) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'Amount of semesters this field of study has.',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci COMMENT='The current field of study of the user.';

CREATE TABLE `field_of_study_has_course` (
  `field_of_study_id` int(11) NOT NULL,
  `course_id` int(11) NOT NULL,
  `semester` int(11) NOT NULL COMMENT 'The semester the course is supposed to take place in in the field of study with the id field_of_study_id.',
  PRIMARY KEY (`field_of_study_id`,`course_id`),
  KEY `fk_field_of_study_has_course_course1_idx` (`course_id`),
  KEY `fk_field_of_study_has_course_field_of_study1_idx` (`field_of_study_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `file` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8_unicode_ci NOT NULL COMMENT 'Displayed name of the file.',
  `uri` varchar(256) COLLATE utf8_unicode_ci NOT NULL COMMENT 'Local or remote file. Stored as an URI.',
  `local` tinyint(4) NOT NULL COMMENT 'Wether the file is a local or remote one.',
  `uploader_id` int(11) NOT NULL COMMENT 'User that uploaded this file.',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT 'When the file was created.',
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_file_user1_idx` (`uploader_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci COMMENT='A local or remote file.\nLocal files do not store a hash as it is easily generated. Files with the same hash should be combined as one entry.';

CREATE TABLE `forum` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(45) COLLATE utf8_unicode_ci NOT NULL COMMENT 'The name given to the forum by the course administrator.',
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=39 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci COMMENT='Admin of the forum is it''s course''s admin.';

CREATE TABLE `forum_entry` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `subject` varchar(64) COLLATE utf8_unicode_ci NOT NULL COMMENT 'The subject of the forum entry.',
  `content` varchar(4096) COLLATE utf8_unicode_ci NOT NULL COMMENT 'The content of the forum entry.',
  `in_reply_to` int(11) DEFAULT NULL COMMENT 'New posts have a value of `NULL`, whereas replies to a top-level-post refer to the top-level-post with this field.',
  `author_id` int(11) NOT NULL COMMENT 'The author that created this entry.',
  `forum_id` int(11) NOT NULL COMMENT 'The forum this entry belongs to.',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT 'When this entry was created.',
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_forum_entry_forum1_idx` (`forum_id`),
  KEY `fk_forum_entry_user1_idx` (`author_id`),
  KEY `fk_forum_entry_forum_entry1_idx` (`in_reply_to`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `graduation_level` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `graduation_level` varchar(128) COLLATE utf8_unicode_ci NOT NULL,
  `level` int(11) NOT NULL COMMENT 'Level (or "rank") of the graduation compared to others. Ranks with the a similar "meaning" should get the same level.',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `language` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '2-letter ISO 639-1 language code.',
  `name` varchar(64) COLLATE utf8_unicode_ci NOT NULL COMMENT 'Display name of the language.',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `notification` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(64) COLLATE utf8_unicode_ci NOT NULL COMMENT 'The title of the notification.',
  `body` varchar(128) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'The body of the notification.',
  `url` varchar(256) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'A URL that links the notification to the proper page.',
  `user_to_id` int(11) NOT NULL COMMENT 'The user that received this notification.',
  `time_read` timestamp NULL DEFAULT NULL COMMENT 'The time this notification was read.',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT 'The time this notification was created.',
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_notification_user1_idx` (`user_to_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `role` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(45) COLLATE utf8_unicode_ci NOT NULL COMMENT 'The name of the role.',
  `display_name` varchar(45) COLLATE utf8_unicode_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `submission` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(45) COLLATE utf8_unicode_ci NOT NULL COMMENT 'Name of the submission.',
  `deadline` timestamp NULL DEFAULT NULL COMMENT 'Deadline for submitting solutions for this submission.',
  `course_id` int(11) NOT NULL COMMENT 'The course this submission is from.',
  `max_filesize` int(11) NOT NULL DEFAULT 5 COMMENT 'Maximum Filesize in MB.',
  `visible_from` timestamp NOT NULL DEFAULT current_timestamp() COMMENT 'At which date the submission will be visible to enrolled users.',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT 'When the submission was created.',
  `updated_at` timestamp NULL DEFAULT NULL,
  `graded_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  KEY `fk_submission_course1_idx` (`course_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci COMMENT='Submission which a course administrator can create for enrolled users to upload files to for grading.';

CREATE TABLE `submission_has_files` (
  `submission_id` int(11) NOT NULL,
  `file_id` int(11) NOT NULL,
  PRIMARY KEY (`submission_id`,`file_id`),
  KEY `fk_submission_has_files_file1_idx` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(64) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'The title this user has.',
  `firstname` varchar(32) COLLATE utf8_unicode_ci NOT NULL,
  `surname` varchar(64) COLLATE utf8_unicode_ci NOT NULL,
  `email` varchar(256) COLLATE utf8_unicode_ci NOT NULL,
  `password` binary(60) NOT NULL COMMENT 'The password stored as a bcrypt hash.',
  `role_id` int(11) NOT NULL COMMENT 'The role this user has.',
  `graduation_level` int(11) DEFAULT NULL COMMENT 'What prior graduation level the user has.',
  `semester` int(11) DEFAULT NULL COMMENT 'The current semester of the user. This can be NULL as the user doesn''t have to be a "student" (or similar).',
  `phone_number` varchar(45) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'Phone numbers are only stored with numbers - the rest is done in the application.',
  `residence` varchar(256) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'General place of residency.',
  `profile_picture` int(11) DEFAULT NULL COMMENT 'Profile picture this user has created.',
  `biography` varchar(512) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'Something the user writes about themself.',
  `preferred_language_id` int(11) NOT NULL COMMENT 'Preferred language of the user.',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  UNIQUE KEY `phone_number_UNIQUE` (`phone_number`),
  KEY `fk_user_graduation_level1_idx` (`graduation_level`),
  KEY `fk_user_language1_idx` (`preferred_language_id`),
  KEY `fk_user_file1_idx` (`profile_picture`),
  KEY `fk_user_role1_idx` (`role_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `user_has_course` (
  `user_id` int(11) NOT NULL,
  `course_id` int(11) NOT NULL,
  `role_id` int(11) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT 'The time the user enrolled in the given course.',
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`user_id`,`course_id`),
  KEY `fk_user_has_course_course1_idx` (`course_id`),
  KEY `fk_user_has_course_user1_idx` (`user_id`),
  KEY `fk_user_has_course_role1_idx` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `user_has_exam` (
  `user_id` int(11) NOT NULL,
  `exam_id` int(11) NOT NULL,
  `attended` tinyint(4) NOT NULL DEFAULT 0 COMMENT 'Whether the user has attended the exam or not.',
  `grade` int(11) DEFAULT NULL,
  `passed` tinyint(4) DEFAULT NULL COMMENT 'If the user that attended the exam passed it or not.',
  `feedback` varchar(512) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'The feedback given to the user about their solution to the exam.',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT 'When the user registered for the exam.',
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT 'When the user deregistered from the exam.',
  PRIMARY KEY (`user_id`,`exam_id`),
  KEY `fk_user_has_exam_exam1_idx` (`exam_id`),
  KEY `fk_user_has_exam_user1_idx` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `user_has_field_of_study` (
  `user_id` int(11) NOT NULL,
  `field_of_study_id` int(11) NOT NULL,
  PRIMARY KEY (`user_id`,`field_of_study_id`),
  KEY `fk_user_has_field_of_study_field_of_study1_idx` (`field_of_study_id`),
  KEY `fk_user_has_field_of_study_user1_idx` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `user_submission` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'Name of the submission.',
  `submitter_id` int(11) NOT NULL COMMENT 'The user that submitted this solution.',
  `submission_id` int(11) NOT NULL COMMENT 'The submission this user is submitting their solution to.',
  `grade` int(11) DEFAULT NULL COMMENT 'The grade of the user''s solution.',
  `ignores_submission_deadline` tinyint(4) NOT NULL DEFAULT 0 COMMENT 'Whether the user is allowed to submit their solutions after the deadline defined in the submission is over.',
  `submission_time` timestamp NULL DEFAULT NULL COMMENT 'When the user submitted their solution.',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT 'When this user_submission was created.',
  `deleted_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_submission_file_user1_idx` (`submitter_id`),
  KEY `fk_user_submission_submission1_idx` (`submission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci COMMENT='A submission of a user. An entry in this table can exist without the user actually submitting any files in case that the user is allowed to submit the files later than the original due date in the submission.';

CREATE TABLE `user_submission_has_files` (
  `user_submission_id` int(11) NOT NULL,
  `file_id` int(11) NOT NULL,
  PRIMARY KEY (`user_submission_id`,`file_id`),
  KEY `fk_user_submission_has_files_file1_idx` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

ALTER TABLE `appointment`
	ADD CONSTRAINT `fk_appointment_course1` FOREIGN KEY (`course_id`) REFERENCES `course` (`id`);

ALTER TABLE `certificate`
	ADD CONSTRAINT `fk_certificate_course1` FOREIGN KEY (`linked_course_id`) REFERENCES `course` (`id`),
	ADD CONSTRAINT `fk_certificate_exam1` FOREIGN KEY (`exam_id`) REFERENCES `exam` (`id`),
	ADD CONSTRAINT `fk_certificate_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`);

ALTER TABLE `course`
	ADD CONSTRAINT `fk_course_forum1` FOREIGN KEY (`forum_id`) REFERENCES `forum` (`id`);

ALTER TABLE `course_has_files`
	ADD CONSTRAINT `fk_course_has_files_course1` FOREIGN KEY (`course_id`) REFERENCES `course` (`id`),
	ADD CONSTRAINT `fk_course_has_files_file1` FOREIGN KEY (`file_id`) REFERENCES `file` (`id`);

ALTER TABLE `course_requires_certificate`
	ADD CONSTRAINT `fk_certificate_has_course_certificate1` FOREIGN KEY (`certificate_id`) REFERENCES `certificate` (`id`),
	ADD CONSTRAINT `fk_certificate_has_course_course1` FOREIGN KEY (`course_id`) REFERENCES `course` (`id`);

ALTER TABLE `directory`
	ADD CONSTRAINT `fk_directory_course1` FOREIGN KEY (`course_id`) REFERENCES `course` (`id`);

ALTER TABLE `directory_has_files`
	ADD CONSTRAINT `fk_directory_has_files_directory1` FOREIGN KEY (`directory_id`) REFERENCES `directory` (`id`),
	ADD CONSTRAINT `fk_directory_has_files_file1` FOREIGN KEY (`file_id`) REFERENCES `file` (`id`);

ALTER TABLE `exam`
	ADD CONSTRAINT `fk_exam_course1` FOREIGN KEY (`course_id`) REFERENCES `course` (`id`),
	ADD CONSTRAINT `fk_exam_user1` FOREIGN KEY (`creator_id`) REFERENCES `user` (`id`);

ALTER TABLE `exam_has_files`
	ADD CONSTRAINT `fk_exam_has_files_exam1` FOREIGN KEY (`exam_id`) REFERENCES `exam` (`id`),
	ADD CONSTRAINT `fk_exam_has_files_file1` FOREIGN KEY (`file_id`) REFERENCES `file` (`id`);

ALTER TABLE `field_of_study_has_course`
	ADD CONSTRAINT `fk_field_of_study_has_course_course1` FOREIGN KEY (`course_id`) REFERENCES `course` (`id`),
	ADD CONSTRAINT `fk_field_of_study_has_course_field_of_study1` FOREIGN KEY (`field_of_study_id`) REFERENCES `field_of_study` (`id`);

ALTER TABLE `file`
	ADD CONSTRAINT `fk_file_user1` FOREIGN KEY (`uploader_id`) REFERENCES `user` (`id`);

ALTER TABLE `forum_entry`
	ADD CONSTRAINT `fk_forum_entry_forum1` FOREIGN KEY (`forum_id`) REFERENCES `forum` (`id`),
	ADD CONSTRAINT `fk_forum_entry_forum_entry1` FOREIGN KEY (`in_reply_to`) REFERENCES `forum_entry` (`id`),
	ADD CONSTRAINT `fk_forum_entry_user1` FOREIGN KEY (`author_id`) REFERENCES `user` (`id`);

ALTER TABLE `notification`
	ADD CONSTRAINT `fk_notification_user1` FOREIGN KEY (`user_to_id`) REFERENCES `user` (`id`);

ALTER TABLE `submission`
	ADD CONSTRAINT `fk_submission_course1` FOREIGN KEY (`course_id`) REFERENCES `course` (`id`);

ALTER TABLE `submission_has_files`
	ADD CONSTRAINT `fk_submission_has_files_file1` FOREIGN KEY (`file_id`) REFERENCES `file` (`id`),
	ADD CONSTRAINT `fk_submission_has_files_submission1` FOREIGN KEY (`submission_id`) REFERENCES `submission` (`id`);

ALTER TABLE `user`
	ADD CONSTRAINT `fk_user_file1` FOREIGN KEY (`profile_picture`) REFERENCES `file` (`id`),
	ADD CONSTRAINT `fk_user_graduation_level1` FOREIGN KEY (`graduation_level`) REFERENCES `graduation_level` (`id`),
	ADD CONSTRAINT `fk_user_language1` FOREIGN KEY (`preferred_language_id`) REFERENCES `language` (`id`),
	ADD CONSTRAINT `fk_user_role1` FOREIGN KEY (`role_id`) REFERENCES `role` (`id`);

ALTER TABLE `user_has_course`
	ADD CONSTRAINT `fk_user_has_course_course1` FOREIGN KEY (`course_id`) REFERENCES `course` (`id`),
	ADD CONSTRAINT `fk_user_has_course_role1` FOREIGN KEY (`role_id`) REFERENCES `role` (`id`),
	ADD CONSTRAINT `fk_user_has_course_user1` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`);

ALTER TABLE `user_has_exam`
	ADD CONSTRAINT `fk_user_has_exam_exam1` FOREIGN KEY (`exam_id`) REFERENCES `exam` (`id`),
	ADD CONSTRAINT `fk_user_has_exam_user1` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`);

ALTER TABLE `user_has_field_of_study`
	ADD CONSTRAINT `fk_user_has_field_of_study_field_of_study1` FOREIGN KEY (`field_of_study_id`) REFERENCES `field_of_study` (`id`),
	ADD CONSTRAINT `fk_user_has_field_of_study_user1` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`);

ALTER TABLE `user_submission`
	ADD CONSTRAINT `fk_submission_file_user1` FOREIGN KEY (`submitter_id`) REFERENCES `user` (`id`),
	ADD CONSTRAINT `fk_user_submission_submission1` FOREIGN KEY (`submission_id`) REFERENCES `submission` (`id`);

ALTER TABLE `user_submission_has_files`
	ADD CONSTRAINT `fk_submission_has_files_user_submission1` FOREIGN KEY (`user_submission_id`) REFERENCES `user_submission` (`id`),
	ADD CONSTRAINT `fk_user_submission_has_files_file1` FOREIGN KEY (`file_id`) REFERENCES `file` (`id`);

-- +migrate Down
-- we don't care about foreign key checks, everything should be deleted if this migration is applied anyway
SET FOREIGN_KEY_CHECKS=OFF;
DROP TABLE `appointment`;
DROP TABLE `certificate`;
DROP TABLE `course`;
DROP TABLE `course_has_files`;
DROP TABLE `course_requires_certificate`;
DROP TABLE `directory`;
DROP TABLE `directory_has_files`;
DROP TABLE `exam`;
DROP TABLE `exam_has_files`;
DROP TABLE `field_of_study`;
DROP TABLE `field_of_study_has_course`;
DROP TABLE `file`;
DROP TABLE `forum`;
DROP TABLE `forum_entry`;
DROP TABLE `graduation_level`;
DROP TABLE `language`;
DROP TABLE `notification`;
DROP TABLE `role`;
DROP TABLE `submission`;
DROP TABLE `submission_has_files`;
DROP TABLE `user`;
DROP TABLE `user_has_course`;
DROP TABLE `user_has_exam`;
DROP TABLE `user_has_field_of_study`;
DROP TABLE `user_submission`;
DROP TABLE `user_submission_has_files`;
-- turn on foreign key checks again for good measure
SET FOREIGN_KEY_CHECKS=ON;
