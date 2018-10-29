CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS "user" (
	id 			bigserial 	NOT NULL PRIMARY KEY,
	nickname 	CITEXT,
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
	FOREIGN KEY (u_id) REFERENCES "user" (id)
);

CREATE TABLE IF NOT EXISTS "thread" (
	id 			bigserial NOT NULL PRIMARY KEY,
	slug 		CITEXT UNIQUE,
	created 	timestamp WITH TIME ZONE NOT NULL,
	title 		text NOT NULL,
	message 	text NOT NULL,
	u_id 		bigserial NOT NULL,
	f_id 		bigserial NOT NULL,
	FOREIGN KEY (u_id) REFERENCES "user" (id),
	FOREIGN KEY (f_id) REFERENCES "forum" (id)
);

CREATE TABLE IF NOT EXISTS "post" (
	id 			bigserial NOT NULL PRIMARY KEY,
	created 	timestamp WITH TIME ZONE NOT NULL,
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
