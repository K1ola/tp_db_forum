DROP TABLE IF EXISTS users, forum, post, thread, vote, forum_users CASCADE;

CREATE EXTENSION IF NOT EXISTS CITEXT;

-- USER
CREATE TABLE IF NOT EXISTS users (
  about       TEXT                NOT NULL,
  email       TEXT UNIQUE       NOT NULL,
  fullname    TEXT                NOT NULL,
  nickname    CITEXT UNIQUE   PRIMARY KEY
);

-- FORUM
CREATE TABLE IF NOT EXISTS forum (
  posts       BIGINT    DEFAULT 0,                
  slug        TEXT      UNIQUE ,  
  threads     INTEGER   DEFAULT 0,
  title       TEXT,
  nickname    CITEXT REFERENCES users (nickname)
);

-- THREAD
CREATE TABLE IF NOT EXISTS thread (
  author      TEXT NOT NULL,
  created     TIMESTAMPTZ DEFAULT now(),
  forum       TEXT,
  id          SERIAL NOT NULL,
  message     TEXT NOT NULL,
  slug        TEXT,
  title       TEXT NOT NULL,
  votes       INTEGER DEFAULT 0
);

-- POST
CREATE TABLE IF NOT EXISTS post (
  author      TEXT NOT NULL,
  created     TIMESTAMPTZ DEFAULT now(),
  forum       TEXT,
  id          BIGSERIAL NOT NULL,
  isEdited	  BOOLEAN DEFAULT false, 
  message     TEXT NOT NULL,
  parent	  BIGINT,
  thread      INTEGER
);
