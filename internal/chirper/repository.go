package chirper

import (
	"fmt"

	"os"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// Repository provides methods for accessing data stored in a database.
type Repository interface {
	CreateChirp(chirps Chirp) error
	GetChirps(tags []string) ([]Chirp, error)
	CountChirps(startingDate, endingDate string, tags []string) (int, error)
	EnsureSchema()
}

// PostgresProvider implements Repository interface.
// It's a provider for PostgreSQL and uses postgres-specific keywords
// in SQL queries.
type PostgresProvider struct {
	DB *sqlx.DB
}

// NewPostgresProvider is a constructor for the PostgresProvider.
// It allows to pass raw configuration string.
func NewPostgresProvider(config string) (*PostgresProvider, error) {
	db, err := sqlx.Open("postgres", config)
	if err != nil {
		return nil, err
	}
	PostgresProvider := &PostgresProvider{DB: db}
	return PostgresProvider, nil
}

// NewDefaultPostgresProvider is a constructor for PostgresProvider.
// The configuration is based on environment variables.
// As stated in README - in the future the configuration should be abstracted
// with a dedicated logic.
func NewDefaultPostgresProvider() (*PostgresProvider, error) {
	config := fmt.Sprintf(
		"host=%s user=%s password=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	db, err := sqlx.Open(
		"postgres",
		config,
	)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "connection to the database could not be established")
	}

	PostgresProvider := &PostgresProvider{DB: db}
	return PostgresProvider, nil
}

// EnsureSchema is a helper that ensures the core tables are created in a database.
func (p PostgresProvider) EnsureSchema() {
	p.DB.MustExec(DBSchema)
}

// GetChirps fetches chirps from the database.
// It takes a slice of tags as a parameter for searching.
func (p PostgresProvider) GetChirps(tags []string) ([]Chirp, error) {
	chirps := []Chirp{}
	query, args, err := sqlx.In(
		`WITH p_chirps AS (SELECT chirps.id as cid, chirps.message as message,
		users.username as author FROM chirps
		JOIN chirps_tags ON chirps.id = chirps_tags.chirp_id
		JOIN tags ON chirps_tags.tag_id = tags.id
		JOIN users ON users.id = chirps.author_id
		WHERE tags.name IN (?)
		GROUP BY cid, message, author)
		SELECT p_chirps.cid as cid, p_chirps.message as message,
		p_chirps.author as author, array_agg(tags.name) as tags
		FROM p_chirps JOIN chirps_tags ON chirps_tags.chirp_id = p_chirps.cid
		JOIN tags ON tags.id = chirps_tags.tag_id
		GROUP BY cid, message, author;`,
		tags,
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	query = p.DB.Rebind(query)

	err = p.DB.Select(&chirps, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	return chirps, nil
}

// CreateChirp creates a new chirp using a Chirp object as an input.
// There were many shortcuts taken in this function. In the future it should
// base on the proper Rebind syntax and use database-agnostic '?' syntax.
func (p PostgresProvider) CreateChirp(chirp Chirp) error {
	tx, err := p.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	query := `with p_users as (
			INSERT INTO users (username)
			VALUES ($1)
			ON CONFLICT (username) DO UPDATE SET username=EXCLUDED.username
			RETURNING id), p_chirps as (
			INSERT INTO chirps (author_id, message)
			SELECT id, $2 FROM p_users
			RETURNING id)`

	argCounter := 3
	additionalArgsList := []interface{}{}
	additionalArgsList = append(additionalArgsList, chirp.Author, chirp.Message)

	for i, tag := range chirp.Tags {
		if i == len(chirp.Tags)-1 {
			query += fmt.Sprintf(`, tag%d as (INSERT INTO tags (name)
				 VALUES ($%d)
				 ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name
				 RETURNING id)

						INSERT INTO chirps_tags (chirp_id, tag_id)
						SELECT p_chirps.id, tag%d.id FROM p_chirps, tag%d
						RETURNING id`, i, argCounter, i, i)
			additionalArgsList = append(additionalArgsList, tag)
			argCounter++
		} else {
			query += fmt.Sprintf(`, tag%d as (INSERT INTO tags (name)
				 VALUES ($%d)
				 ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name
				 RETURNING id), ct%d as (
						INSERT INTO chirps_tags (chirp_id, tag_id)
						SELECT p_chirps.id, tag%d.id FROM p_chirps, tag%d
						RETURNING id)`, i, argCounter, i, i, i)
			additionalArgsList = append(additionalArgsList, tag)
			argCounter++
		}
	}

	_, err = tx.Exec(
		query,
		additionalArgsList...,
	)

	if err != nil {
		return errors.Wrap(err, "failed to exec transaction")
	}

	return tx.Commit()
}

// CountChirps returns count of the chirps created between the given dates.
// Two of the first arguments are expected to be date strings in format "YYYY-MM-DD".
// The third parameter is a slice of tags for searching.
func (p PostgresProvider) CountChirps(startingDate, endingDate string, tags []string) (int, error) {
	var recordsCount []int
	query, args, err := sqlx.In(
		`WITH res as (
			SELECT chirps.id as cid, chirps.message as message,
			array_agg(tags.name) as tags,
			users.username as author FROM chirps
			JOIN chirps_tags ON chirps.id = chirps_tags.chirp_id
			JOIN tags ON chirps_tags.tag_id = tags.id
			JOIN users ON users.id = chirps.author_id
			WHERE name IN (?)
			AND chirps.created_at >= ? AND chirps.created_at < ?
			GROUP BY cid, message, author
		) SELECT COUNT(*) FROM res;`,
		tags,
		startingDate,
		endingDate,
	)

	if err != nil {
		return 0, errors.Wrap(err, "failed to build query")
	}

	query = p.DB.Rebind(query)

	err = p.DB.Select(&recordsCount, query, args...)
	if err != nil {
		return 0, errors.Wrap(err, "failed to execute query")
	}

	return recordsCount[0], nil
}
