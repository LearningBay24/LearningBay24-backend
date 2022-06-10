-- +migrate Up
ALTER TABLE `user_has_exam` ADD `file_id` int(11) NULL;
ALTER TABLE `user_has_exam` ADD CONSTRAINT `fk_user_has_exam_file1` FOREIGN KEY (`file_id`) REFERENCES `file` (`id`);

-- +migrate Down
ALTER TABLE `user_has_exam` DROP CONSTRAINT `fk_user_has_exam_file1`;
ALTER TABLE `user_has_exam` DROP COLUMN `file_id`;
