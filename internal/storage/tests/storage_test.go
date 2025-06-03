package tests

import (
	"context"
	"github.com/azaliaz/quote-service/internal/storage"
	"github.com/azaliaz/quote-service/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log/slog"
	"strconv"
	"testing"
	"time"
)

type QuoteRepositoryTestSuite struct {
	container *postgres.PostgresContainer
	suite.Suite

	dbConfig storage.Config
	db       *storage.DB
	repo     storage.QuoteStorage
}

func (s *QuoteRepositoryTestSuite) SetupSuite() {
	ctx := context.Background()
	s.dbConfig = s.setupPostgres(ctx)

	logger := slog.Default()
	db := storage.NewDB(&s.dbConfig, logger)
	require.NoError(s.T(), db.Init())
	s.db = db
	s.repo = db
}

func (s *QuoteRepositoryTestSuite) SetupTest() {
	ctx := context.Background()
	conn, err := s.db.Pool().Acquire(ctx)
	require.NoError(s.T(), err)
	defer conn.Release()

	_, err = conn.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	require.NoError(s.T(), err)
	require.NoError(s.T(), migrations.PostgresMigrate(s.dbConfig.UrlPostgres()))
}

func (s *QuoteRepositoryTestSuite) TearDownTest() {
	require.NoError(s.T(), migrations.PostgresMigrateDown(s.dbConfig.UrlPostgres()))
}

func (s *QuoteRepositoryTestSuite) setupPostgres(ctx context.Context) storage.Config {
	cfg := storage.Config{
		Host:             "",
		DbName:           "test-db",
		User:             "user",
		Password:         "1",
		MaxOpenConns:     10,
		ConnIdleLifetime: 60 * time.Second,
		ConnMaxLifetime:  60 * time.Minute,
	}
	pgContainer, err := postgres.Run(ctx,
		"postgres:14-alpine", // образ задаётся явно вторым аргументом
		postgres.WithDatabase(cfg.DbName),
		postgres.WithUsername(cfg.User),
		postgres.WithPassword(cfg.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)

	require.NoError(s.T(), err)
	s.container = pgContainer

	host, err := pgContainer.Host(ctx)
	require.NoError(s.T(), err)
	cfg.Host = host
	ports, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(s.T(), err)
	cfg.Host += ":" + strconv.Itoa(ports.Int())

	s.dbConfig = cfg
	return cfg
}

func (s *QuoteRepositoryTestSuite) TearDownSuite() {
	s.db.Stop()

	if err := s.container.Stop(context.Background(), nil); err != nil {
		s.T().Logf("failed to stop container: %v", err)
	}
}

func (s *QuoteRepositoryTestSuite) TestAddQuoteAndGetAllQuotes() {
	ctx := context.Background()
	quote := &storage.Quote{
		Author: "Confucius",
		Quote:  "Life is simple, but we insist on making it complicated.",
	}
	id, err := s.repo.AddQuote(ctx, quote)
	require.NoError(s.T(), err)
	assert.Greater(s.T(), id, int64(0))

	quotes, err := s.repo.GetAllQuotes(ctx)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), quotes)
	found := false
	for _, q := range quotes {
		if q.ID == id {
			found = true
			assert.Equal(s.T(), quote.Author, q.Author)
			assert.Equal(s.T(), quote.Quote, q.Quote)
		}
	}
	assert.True(s.T(), found)
}

func (s *QuoteRepositoryTestSuite) TestGetRandomQuote() {
	ctx := context.Background()
	_, err := s.repo.AddQuote(ctx, &storage.Quote{Author: "A", Quote: "A quote"})
	require.NoError(s.T(), err)
	_, err = s.repo.AddQuote(ctx, &storage.Quote{Author: "B", Quote: "B quote"})
	require.NoError(s.T(), err)

	q, err := s.repo.GetRandomQuote(ctx)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), q)
	assert.Contains(s.T(), []string{"A", "B"}, q.Author)
}

func (s *QuoteRepositoryTestSuite) TestGetQuotesByAuthor() {
	ctx := context.Background()
	_, err := s.repo.AddQuote(ctx, &storage.Quote{Author: "AuthorX", Quote: "Quote 1"})
	require.NoError(s.T(), err)
	_, err = s.repo.AddQuote(ctx, &storage.Quote{Author: "Other", Quote: "Quote 2"})
	require.NoError(s.T(), err)

	quotes, err := s.repo.GetQuotesByAuthor(ctx, "AuthorX")
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), quotes)
	for _, q := range quotes {
		assert.Equal(s.T(), "AuthorX", q.Author)
	}
}

func (s *QuoteRepositoryTestSuite) TestDeleteQuote() {
	ctx := context.Background()
	id, err := s.repo.AddQuote(ctx, &storage.Quote{Author: "ToDelete", Quote: "Delete me"})
	require.NoError(s.T(), err)

	err = s.repo.DeleteQuote(ctx, id)
	require.NoError(s.T(), err)

	quotes, err := s.repo.GetAllQuotes(ctx)
	require.NoError(s.T(), err)
	for _, q := range quotes {
		assert.NotEqual(s.T(), id, q.ID)
	}

	err = s.repo.DeleteQuote(ctx, id)
	assert.Error(s.T(), err)
}

func (s *QuoteRepositoryTestSuite) TestGetAllQuotes_Empty() {
	ctx := context.Background()
	quotes, err := s.repo.GetAllQuotes(ctx)
	require.NoError(s.T(), err)
	assert.Empty(s.T(), quotes)
}

func (s *QuoteRepositoryTestSuite) TestGetRandomQuote_Empty() {
	ctx := context.Background()
	q, err := s.repo.GetRandomQuote(ctx)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), q)
}

func (s *QuoteRepositoryTestSuite) TestGetQuotesByAuthor_Empty() {
	ctx := context.Background()
	quotes, err := s.repo.GetQuotesByAuthor(ctx, "NonExistingAuthor")
	require.NoError(s.T(), err)
	assert.Empty(s.T(), quotes)
}

func TestQuoteRepositorySuite(t *testing.T) {
	suite.Run(t, new(QuoteRepositoryTestSuite))
}
