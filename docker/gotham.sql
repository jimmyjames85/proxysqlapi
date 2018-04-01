DROP USER IF EXISTS 'gotham'@'%';
CREATE USER 'gotham'@'%' IDENTIFIED BY 'gotham';
GRANT ALL PRIVILEGES ON *.* TO 'gotham'@'%';
FLUSH PRIVILEGES;

DROP USER IF EXISTS 'newyork'@'%';
CREATE USER 'newyork'@'%' IDENTIFIED BY 'newyork';
GRANT ALL PRIVILEGES ON *.* TO 'newyork'@'%';
FLUSH PRIVILEGES;

DROP DATABASE IF EXISTS gotham;
CREATE DATABASE gotham;
USE gotham;

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
  ('Deadpool', 1),
  ('Harley Quinn', 2),
  ('Hugo Strange', 3),
  ('Killer Croc', 4),
  ('Mad Hatter', 5),
  ('Mr. Freeze', 6),
  ('Penguin', 7),
  ('Poison Ivy', 8),
  ('The Riddler', 9),
  ('Ra''s al Ghul', 10),
  ('Scarecrow', 11),
  ('Solomon Grundy', 12),
  ('Two-Face', 13),
  ('Ventriloquist', 14),
  ('Victor Zsasz', 15),
  ('Clock King', 16),
  ('Black Mask', 17),
  ('Catwoman', 18);

INSERT INTO person (first, middle, last, ego_id) VALUES
  ('Floyd', '', 'Lawton', 1),
  ('Harleen', 'Frances', 'Queinzel', 2),
  ('Hugo', '', 'Strange', 3),
  ('Waylon', '', 'Jones', 4),
  ('Jervis', '', 'Tetch', 5),
  ('Victor', '', 'Dr. Fries', 6),
  ('Oswald', 'Chesterfield', 'Cobblepot', 7),
  ('Pamela', 'Lilian', 'Isley', 8),
  ('Edward', '', 'Nigma', 9),
  ('Henri', '', 'Ducard', 10),
  ('Jonathan', '', 'Dr. Crane', 11),
  ('Cryus', '', 'Gold', 12),
  ('Harvey', '', 'Dent', 13),
  ('Arnold', '', 'Wesker', 14),
  ('Victor', '', 'Zsasz', 15),
  ('William', '', 'Tockman', 16),
  ('Roman', '', 'Sionis', 17),
  ('Selina', '', 'Kyle', 18);
