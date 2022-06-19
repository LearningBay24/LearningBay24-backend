-- +migrate Up
ALTER TABLE `appointment` ADD `duration` int(32) NOT NULL;

-- +migrate Down
ALTER TABLE `appointment` DROP COLUMN `duration`;
