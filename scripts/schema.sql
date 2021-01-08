-- Initialize the database.

CREATE TABLE user (
  id INT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(20) UNIQUE NOT NULL,
  password VARCHAR(120) DEFAULT '123456',
  created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE list (
  id INT PRIMARY KEY AUTO_INCREMENT,
  list VARCHAR(15) NOT NULL,
  created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  user_id INT NOT NULL
);

CREATE TABLE task (
  id INT PRIMARY KEY AUTO_INCREMENT,
  task VARCHAR(40) NOT NULL,
  created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  list_id INT NOT NULL
);

CREATE TABLE seq (
  list_id INT NOT NULL,
  task_id INT NOT NULL,
  seq INT NOT NULL
);

CREATE VIEW tasks AS
  SELECT user_id, task.id task_id, task, task.list_id, list, seq
  FROM task LEFT JOIN list ON task.list_id = list.id
  LEFT JOIN seq ON task.list_id = seq.list_id AND task.id = seq.task_id
  ORDER BY seq;

CREATE VIEW lists AS
  SELECT list.id, list.user_id, list, COUNT(task) count
  FROM list LEFT JOIN task ON list.id = list_id
  GROUP BY list ORDER BY list;

DELIMITER ;;
CREATE TRIGGER add_user AFTER INSERT ON user
FOR EACH ROW BEGIN
    INSERT INTO list (user_id, list)
    VALUES (new.id, 'My Tasks');
    INSERT INTO task (list_id, task)
    VALUES (LAST_INSERT_ID(), 'Welcome to use mytasks!');
END;;

CREATE TRIGGER add_seq AFTER INSERT ON task
FOR EACH ROW BEGIN
    SET @seq := (SELECT IFNULL(MAX(seq)+1, 1) FROM seq WHERE list_id = new.list_id);
    INSERT INTO seq (list_id, task_id, seq)
    VALUES (new.list_id, new.id, @seq);
END;;

CREATE TRIGGER reorder AFTER DELETE ON task
FOR EACH ROW BEGIN
    SET @seq := (SELECT seq FROM seq WHERE list_id = old.list_id AND task_id = old.id);
    DELETE FROM seq
    WHERE list_id = old.list_id AND seq = @seq;
    UPDATE seq SET seq = seq-1
    WHERE list_id = old.list_id AND seq > @seq;
END;;
DELIMITER ;

SET SESSION sql_mode = 'NO_AUTO_VALUE_ON_ZERO';
INSERT INTO user (id, username) VALUES (0, 'guest');
