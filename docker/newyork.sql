DROP USER IF EXISTS 'newyork'@'%';
CREATE USER 'newyork'@'%' IDENTIFIED BY 'newyork';
GRANT ALL PRIVILEGES ON *.* TO 'newyork'@'%';
FLUSH PRIVILEGES;

DROP USER IF EXISTS 'gotham'@'%';
CREATE USER 'gotham'@'%' IDENTIFIED BY 'gotham';
GRANT ALL PRIVILEGES ON *.* TO 'gotham'@'%';
FLUSH PRIVILEGES;


DROP DATABASE IF EXISTS newyork;
CREATE DATABASE newyork;
USE newyork;

DROP TABLE IF EXISTS person;
DROP TABLE IF EXISTS ego;

CREATE TABLE ego (
  id   INT(11)      NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,

  PRIMARY KEY (id)
)
  ENGINE = InnoDB
  DEFAULT CHARACTER SET = utf8;

CREATE TABLE person (
  id     INT(11)      NOT NULL AUTO_INCREMENT,
  ego_id INT(11)               DEFAULT NULL,
  first  VARCHAR(255) NOT NULL DEFAULT '',
  middle VARCHAR(255) NOT NULL DEFAULT '',
  last   VARCHAR(255) NOT NULL DEFAULT '',

  PRIMARY KEY (id),
  UNIQUE KEY (first, middle, last),
  FOREIGN KEY (ego_id) REFERENCES ego (id)

)
  ENGINE = InnoDB
  DEFAULT CHARACTER SET = utf8;


INSERT INTO ego (name, id) VALUES
  ('Iron Man', 1),
  ('Spider-Man', 2),
  ('Daredevil', 3),
  ('Captain America', 4),
  ('Doctor Strange', 5),
  ('Punisher', 6),
  ('Professor X', 7),
  ('Phoenix', 8),
  ('Mister Fantasitc', 9),
  ('Invisible Woman', 10),
  ('Human Torch', 11),
  ('Thing', 12),
  ('Jessica Jones', 13),
  ('Luke Cage',14);

INSERT INTO person (first, middle, last, ego_id) VALUES
  ('Tony', '', 'Stark', 1),
  ('Peter', '', 'Parker', 2),
  ('Steve', '', 'Rogers', 3),
  ('Matt', '', 'Murdock', 4),
  ('Stephen', 'Vincent', 'Strange', 5),
  ('Frank', '', 'Castle', 6),
  ('Charles', 'Francis', 'Xavier', 7),
  ('Jean', '', 'Grey', 8),
  ('Reed', '', 'Richards', 9),
  ('Sue', '', 'Storm', 10),
  ('Johnny', '', 'Storm', 11),
  ('Ben', '', 'Grimm', 12),
  ('Jessica', 'Campbell', 'Jones', 13),
  ('Carl', '', 'Lucas', 14);
