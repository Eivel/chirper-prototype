package chirper

import "github.com/lib/pq"

type Chirp struct {
	ID      int            `db:"cid"`
	Message string         `db:"message"`
	Tags    pq.StringArray `db:"tags"`
	Author  string         `db:"author"`
}

type User struct {
	Username string `db:"username"`
}

type Tag struct {
	Name string `db:"name"`
}

var DBSchema = `
CREATE TABLE IF NOT EXISTS users (
	id SERIAL NOT NULL PRIMARY KEY,
	username VARCHAR(65) NOT NULL UNIQUE,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chirps (
	id SERIAL NOT NULL PRIMARY KEY,
	message VARCHAR(255) NOT NULL,
	author_id int REFERENCES users(id),
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tags (
	id SERIAL NOT NULL PRIMARY KEY,
	name VARCHAR(65) UNIQUE NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chirps_tags (
	id SERIAL NOT NULL PRIMARY KEY,
	chirp_id int REFERENCES chirps(id),
	tag_id int REFERENCES tags(id),
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);`
