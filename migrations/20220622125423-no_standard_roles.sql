-- +migrate Up
DELETE FROM `role` WHERE id = 1;
DELETE FROM `role` WHERE id = 2;
DELETE FROM `role` WHERE id = 3;
DELETE FROM `language` WHERE id = 1;

-- +migrate Down
INSERT INTO `role` (id, name, display_name) VALUES (1, "admin", "Administrator");
INSERT INTO `role` (id, name, display_name) VALUES (2, "moderator", "Moderator");
INSERT INTO `role` (id, name, display_name) VALUES (3, "user", "User");
INSERT INTO `language` VALUES (1,  "deutsch");
