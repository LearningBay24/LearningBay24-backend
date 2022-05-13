-- +migrate Up
ALTER TABLE `field_of_study` MODIFY `semesters` int(2) DEFAULT 0;

-- +migrate Down
ALTER TABLE `field_of_study` MODIFY `semesters` varchar(45) COLLATE utf8_unicode_ci DEFAULT NULL;
