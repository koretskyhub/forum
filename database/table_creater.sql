CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS "user" (
	id 			bigserial 	NOT NULL PRIMARY KEY,
	nickname 	CITEXT collate "C" NOT NULL,
	about 		text NOT NULL,
	email 		CITEXT NOT NULL UNIQUE,
	fullname 	text NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_nickname ON "user" (nickname);

CREATE TABLE IF NOT EXISTS "forum" (
	id 			bigserial NOT NULL PRIMARY KEY,
	slug 		CITEXT NOT NULL UNIQUE,
	title 		text NOT NULL,
	u_id 		bigserial NOT NULL,
	u_nickname	CITEXT	collate "C" NOT NULL,
	posts		BIGINT	DEFAULT 0,
	threads		INTEGER DEFAULT 0,
	FOREIGN KEY (u_id) REFERENCES "user" (id)
);

CREATE TABLE IF NOT EXISTS "thread" (
	id 			bigserial NOT NULL PRIMARY KEY,
	slug 		CITEXT UNIQUE,
	created 	timestamp (6) WITH TIME ZONE NOT NULL,
	title 		text NOT NULL,
	message 	text NOT NULL,
	u_id 		bigserial NOT NULL,
	f_id 		bigserial NOT NULL,
	FOREIGN KEY (u_id) REFERENCES "user" (id),
	FOREIGN KEY (f_id) REFERENCES "forum" (id)
);

CREATE TABLE IF NOT EXISTS "post" (
	id 			bigserial NOT NULL PRIMARY KEY,
	created 	timestamp (6) WITH TIME ZONE NOT NULL,
	is_edited 	bool NOT NULL DEFAULT FALSE,
	message 	text NOT NULL,
	path 		integer[] NOT NULL,
	u_id 		bigserial NOT NULL,
	t_id 		bigserial NOT NULL,
	FOREIGN KEY (u_id) REFERENCES "user" (id),
	FOREIGN KEY (t_id) REFERENCES "thread" (id)
);

CREATE TABLE IF NOT EXISTS "vote" (
	id 		bigserial NOT NULL PRIMARY KEY,
	voice 		bool NOT NULL,
	t_id		bigserial NOT NULL,
	u_id		bigserial NOT NULL,
	FOREIGN KEY (t_id) REFERENCES "thread" (id),
	FOREIGN KEY (u_id) REFERENCES "user" (id),
	UNIQUE (t_id, u_id)
);

CREATE TABLE IF NOT EXISTS "forum_users" (
	u_nickname 	CITEXT collate "C" NOT NULL,
	f_slug		 	CITEXT NOT NULL,
	FOREIGN KEY (u_nickname) REFERENCES "user" (nickname),
	FOREIGN KEY (f_slug) REFERENCES "forum" (slug),
	CONSTRAINT forum_users_pk PRIMARY KEY (u_nickname, f_slug) 
);

CREATE OR REPLACE FUNCTION forum_inc_threads()
  RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  UPDATE forum
  SET threads = threads + 1
  WHERE forum.id = new.f_id;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS forum_threads_num
ON thread;

CREATE TRIGGER forum_threads_num
  BEFORE INSERT
  ON thread
  FOR EACH ROW
EXECUTE PROCEDURE forum_inc_threads();

CREATE INDEX idx_user_by_forum_user ON public."user" 		USING btree (nickname);
CREATE INDEX idx_forum_stat_user ON public."user" 			USING btree (id, nickname);
CREATE INDEX idx_forum_stat_thread ON public.thread 		USING btree (id, f_id);
CREATE INDEX idx_user_by_forum_thread ON public.thread 	USING btree (f_id, id);
CREATE INDEX idx_user_by_forum_thread2 ON public.thread USING btree (f_id, u_id);
CREATE INDEX idx_forum_stat_post ON public.post 				USING btree (t_id, id);
CREATE INDEX idx_forum_stat_forum ON public.forum 			USING btree (slug, id, u_id, title);