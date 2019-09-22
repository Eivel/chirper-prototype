package chirper

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/testfixtures.v2"
	"gotest.tools/assert"
)

type Suite struct {
	suite.Suite

	DB         *sqlx.DB
	mock       sqlmock.Sqlmock
	repository Repository
	fixtures   *testfixtures.Context
}

func (s *Suite) SetupSuite() {
	godotenv.Load()
	config := fmt.Sprintf(
		"host=%s user=%s password=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("TEST_DB_HOST"),
		os.Getenv("TEST_DB_USER"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_PORT"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_DB_SSL_MODE"),
	)

	provider, err := NewPostgresProvider(config)
	require.NoError(s.T(), err)

	provider.DB.Exec(DBSchema)

	dbFixtures, err := sql.Open("postgres", config)
	require.NoError(s.T(), err)

	fixtures, err := testfixtures.NewFolder(dbFixtures, &testfixtures.PostgreSQL{}, "fixtures")

	s.fixtures = fixtures
	s.DB = provider.DB
	s.repository = provider
}

func (s *Suite) BeforeTest(_, _ string) {
	s.clearDB()
}

func (s *Suite) AfterTest(_, _ string) {
	s.clearDB()
}

func TestInit(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) Test_repository_CreateChirps() {
	chirp := Chirp{
		Message: "test_message",
		Tags:    []string{"tag1", "tag2"},
		Author:  "test_user",
	}

	err := s.repository.CreateChirp(chirp)
	require.NoError(s.T(), err)

	resultChirps := []Chirp{}
	err = s.DB.Select(
		&resultChirps,
		`SELECT chirps.id as cid,
			chirps.message as message,
			array_agg(tags.name) as tags,
			users.username as author
		FROM chirps
		JOIN chirps_tags ON chirps.id = chirps_tags.chirp_id
		JOIN tags ON chirps_tags.tag_id = tags.id
		JOIN users ON users.id = chirps.author_id
		GROUP BY cid, message, author;`,
	)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), len(resultChirps), 1, "there should be one chirp")

	assert.Equal(s.T(), resultChirps[0].Message, "test_message", "chirp should contain correct message")
	assert.Equal(s.T(), resultChirps[0].Author, "test_user", "chirp should contain correct author")

	sort.Strings(resultChirps[0].Tags)

	assert.Equal(s.T(), reflect.DeepEqual(resultChirps[0].Tags, pq.StringArray{"tag1", "tag2"}), true, "chirp should contain correct tags")
}

func (s *Suite) Test_repository_GetChirps() {
	s.loadFixtures()

	resultChirps, err := s.repository.GetChirps([]string{"tag1"})
	require.NoError(s.T(), err)

	assert.Equal(s.T(), len(resultChirps), 1, "there should be one chirp")

	assert.Equal(s.T(), resultChirps[0].Message, "test_message", "chirp should contain correct message")
	assert.Equal(s.T(), resultChirps[0].Author, "test_user", "chirp should contain correct author")

	sort.Strings(resultChirps[0].Tags)

	assert.Equal(s.T(), reflect.DeepEqual(resultChirps[0].Tags, pq.StringArray{"tag1", "tag2"}), true, "chirp should contain correct tags")
}

func (s *Suite) Test_repository_CountChirps() {
	s.loadFixtures()

	chirpsCount, err := s.repository.CountChirps("2016-01-01", "2016-01-02", []string{"tag1"})
	require.NoError(s.T(), err)

	assert.Equal(s.T(), chirpsCount, 1, "there should be one chirp")
}

func (s *Suite) clearDB() {
	tx := s.DB.MustBegin()
	tx.MustExec(`DELETE FROM chirps_tags;`)
	tx.MustExec(`DELETE FROM chirps;`)
	tx.MustExec(`DELETE FROM tags;`)
	tx.MustExec(`DELETE FROM users;`)
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (s *Suite) loadFixtures() {
	err := s.fixtures.Load()
	if err != nil {
		log.Fatal(errors.Wrap(err, "could not load fixtures"))
	}
}
