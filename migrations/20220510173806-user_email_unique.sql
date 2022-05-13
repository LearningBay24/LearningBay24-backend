-- +migrate Up
ALTER TABLE `user` ADD CONSTRAINT UC_user_email UNIQUE (`email`);

-- +migrate Down
ALTER TABLE `user` DROP CONSTRAINT UC_user_email;
