-- +migrate Up
INSERT INTO `role` (id, name, display_name) VALUES (1, "admin", "Administrator");
INSERT INTO `role` (id, name, display_name) VALUES (2, "moderator", "Moderator");
INSERT INTO `role` (id, name, display_name) VALUES (3, "user", "User");
INSERT INTO `language` VALUES (1,  "deutsch");
INSERT INTO `user` (id, title, firstname, surname, email, password, role_id, preferred_language_id) VALUES (1, "Admin", "Admin", "Admin", "admin@learningbay24.de", '$2a$10$QAH6TEUDRno9twPglVPTn.h0xmXOofA/2HPkJPVDWI54rnLnhFIjq', 1, 1);

-- +migrate Down
DELETE FROM `role` WHERE id = 1;
DELETE FROM `role` WHERE id = 2;
DELETE FROM `role` WHERE id = 3;
DELETE FROM `language` WHERE id = 1;
DELETE FROM `user` WHERE id = 1;
