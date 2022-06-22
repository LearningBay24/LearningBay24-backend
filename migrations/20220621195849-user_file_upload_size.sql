-- +migrate Up
ALTER TABLE `user` ADD `uploaded_bytes` int(64) DEFAULT 0 NOT NULL;

-- +migrate Down
ALTER TABLE `user` DROP COLUMN `uploaded_bytes`;
