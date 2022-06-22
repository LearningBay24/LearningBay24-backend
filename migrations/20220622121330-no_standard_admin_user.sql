-- +migrate Up
DELETE FROM `user` WHERE id = 1;

-- +migrate Down
INSERT INTO `user` (id, title, firstname, surname, email, password, role_id, preferred_language_id) VALUES (1, "Admin", "Admin", "Admin", "admin@learningbay24.de", '$2a$10$QAH6TEUDRno9twPglVPTn.h0xmXOofA/2HPkJPVDWI54rnLnhFIjq', 1, 1);
