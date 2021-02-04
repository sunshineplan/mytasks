-- Initialize the database.

CREATE TABLE user (
  id INT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(20) UNIQUE NOT NULL,
  password VARCHAR(120) DEFAULT '123456',
  uid VARCHAR(24) UNIQUE NOT NULL DEFAULT ''
);

CREATE TABLE list (
  id INT PRIMARY KEY AUTO_INCREMENT,
  list VARCHAR(15) NOT NULL,
  user_id INT NOT NULL
);

CREATE TABLE list_seq (
  user_id INT NOT NULL,
  list_id INT NOT NULL,
  seq INT NOT NULL
);

CREATE TABLE task (
  id INT PRIMARY KEY AUTO_INCREMENT,
  task VARCHAR(1000) NOT NULL,
  created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  list_id INT NOT NULL
);

CREATE TABLE task_seq (
  list_id INT NOT NULL,
  task_id INT NOT NULL,
  seq INT NOT NULL
);

CREATE TABLE completed (
  id INT PRIMARY KEY AUTO_INCREMENT,
  task VARCHAR(1000) NOT NULL,
  created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  list_id INT NOT NULL
);

CREATE VIEW lists AS
  SELECT list.id, list.user_id, list, seq,
  (SELECT COUNT(task) FROM task WHERE list_id = list.id) incomplete,
  (SELECT COUNT(task) FROM completed WHERE list_id = list.id) completed
  FROM list
  LEFT JOIN list_seq ON list.user_id = list_seq.user_id AND list.id = list_seq.list_id
  ORDER BY seq;

CREATE VIEW tasks AS
  SELECT user_id, task.id task_id, task, task.list_id, list, created, seq
  FROM task LEFT JOIN list ON task.list_id = list.id
  LEFT JOIN task_seq ON task.list_id = task_seq.list_id AND task.id = task_seq.task_id
  ORDER BY seq DESC;

CREATE VIEW completeds AS
  SELECT user_id, completed.id task_id, task, completed.list_id, list, created
  FROM completed LEFT JOIN list ON completed.list_id = list.id
  ORDER BY created DESC;

DELIMITER ;;
CREATE PROCEDURE delete_list(lid INT)
BEGIN
  DECLARE EXIT HANDLER FOR SQLEXCEPTION
  BEGIN ROLLBACK; RESIGNAL; END;
  START TRANSACTION;
  DELETE FROM list WHERE id = lid;
  DELETE FROM task WHERE list_id = lid;
  DELETE FROM completed WHERE list_id = lid;
  COMMIT;
END;;

CREATE PROCEDURE complete_task(task_id INT)
BEGIN
  DECLARE EXIT HANDLER FOR SQLEXCEPTION
  BEGIN ROLLBACK; RESIGNAL; END;
  START TRANSACTION;
  INSERT INTO completed (task, list_id)
  SELECT task, list_id FROM task WHERE id = task_id;
  DELETE FROM task WHERE id = task_id;
  COMMIT;
  SELECT LAST_INSERT_ID();
END;;

CREATE PROCEDURE revert_completed(task_id INT)
BEGIN
  DECLARE EXIT HANDLER FOR SQLEXCEPTION
  BEGIN ROLLBACK; RESIGNAL; END;
  START TRANSACTION;
  INSERT INTO task (task, list_id)
  SELECT task, list_id FROM completed WHERE id = task_id;
  DELETE FROM completed WHERE id = task_id;
  COMMIT;
  SELECT LAST_INSERT_ID();
END;;

CREATE PROCEDURE list_reorder(uid INT, new_id INT, old_id INT)
BEGIN
  DECLARE EXIT HANDLER FOR SQLEXCEPTION
  BEGIN ROLLBACK; RESIGNAL; END;
  START TRANSACTION;
  SET @new_seq := (SELECT seq FROM list_seq WHERE list_id = new_id);
  SET @old_seq := (SELECT seq FROM list_seq WHERE list_id = old_id);
  IF @old_seq > @new_seq
  THEN UPDATE list_seq SET seq = seq + 1 WHERE seq >= @new_seq AND seq < @old_seq AND user_id = uid;
  ELSE UPDATE list_seq SET seq = seq - 1 WHERE seq > @old_seq AND seq <= @new_seq AND user_id = uid;
  END IF;
  UPDATE list_seq SET seq = @new_seq WHERE list_id = old_id;
  COMMIT;
END;;

CREATE PROCEDURE task_reorder(lid INT, new_id INT, old_id INT)
BEGIN
  DECLARE EXIT HANDLER FOR SQLEXCEPTION
  BEGIN ROLLBACK; RESIGNAL; END;
  START TRANSACTION;
  SET @new_seq := (SELECT seq FROM task_seq WHERE task_id = new_id);
  SET @old_seq := (SELECT seq FROM task_seq WHERE task_id = old_id);
  IF @old_seq > @new_seq
  THEN UPDATE task_seq SET seq = seq + 1 WHERE seq >= @new_seq AND seq < @old_seq AND list_id = lid;
  ELSE UPDATE task_seq SET seq = seq - 1 WHERE seq > @old_seq AND seq <= @new_seq AND list_id = lid;
  END IF;
  UPDATE task_seq SET seq = @new_seq WHERE task_id = old_id;
  COMMIT;
END;;

CREATE TRIGGER add_user AFTER INSERT ON user
FOR EACH ROW BEGIN
  INSERT INTO list (user_id, list)
  VALUES (new.id, 'My Tasks');
  INSERT INTO task (list_id, task)
  VALUES (LAST_INSERT_ID(), 'Welcome to use mytasks!');
END;;

CREATE TRIGGER add_list_seq AFTER INSERT ON list
FOR EACH ROW BEGIN
  SET @seq := (SELECT IFNULL(MAX(seq)+1, 1) FROM list_seq WHERE user_id = new.user_id);
  INSERT INTO list_seq (user_id, list_id, seq)
  VALUES (new.user_id, new.id, @seq);
END;;

CREATE TRIGGER add_task_seq AFTER INSERT ON task
FOR EACH ROW BEGIN
  SET @seq := (SELECT IFNULL(MAX(seq)+1, 1) FROM task_seq WHERE list_id = new.list_id);
  INSERT INTO task_seq (list_id, task_id, seq)
  VALUES (new.list_id, new.id, @seq);
END;;

CREATE TRIGGER reorder_list AFTER DELETE ON list
FOR EACH ROW BEGIN
  SET @seq := (SELECT seq FROM list_seq WHERE user_id = old.user_id AND list_id = old.id);
  DELETE FROM list_seq
  WHERE user_id = old.user_id AND seq = @seq;
  UPDATE list_seq SET seq = seq-1
  WHERE user_id = old.user_id AND seq > @seq;
END;;

CREATE TRIGGER reorder_task AFTER DELETE ON task
FOR EACH ROW BEGIN
  SET @seq := (SELECT seq FROM task_seq WHERE list_id = old.list_id AND task_id = old.id);
  DELETE FROM task_seq
  WHERE list_id = old.list_id AND seq = @seq;
  UPDATE task_seq SET seq = seq-1
  WHERE list_id = old.list_id AND seq > @seq;
END;;
DELIMITER ;

SET SESSION sql_mode = 'NO_AUTO_VALUE_ON_ZERO';
INSERT INTO user (id, username) VALUES (0, 'guest');
