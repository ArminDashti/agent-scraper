package store

import (
	"context"
	"embed"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

const settingTargetAPIURL = "target_api_url"

//go:embed migrations/*.sql
var migrationFS embed.FS

type Store struct {
	pool *pgxpool.Pool
}

type User struct {
	ID           int64
	Username     string
	PasswordHash string
}

type Source struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	WebsiteURL string    `json:"websiteUrl"`
	IsEnabled  bool      `json:"isEnabled"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type JobRun struct {
	ID             int64      `json:"id"`
	Status         string     `json:"status"`
	StartedAt      time.Time  `json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	ErrorMessage   *string    `json:"errorMessage"`
	ExtractedCount int        `json:"extractedCount"`
}

type ExpenseRow struct {
	ID        int64     `json:"id"`
	Shop      string    `json:"shop"`
	Item      string    `json:"item"`
	Expense   string    `json:"expense"`
	SourceID  *int64    `json:"sourceId"`
	JobRunID  *int64    `json:"jobRunId"`
	CreatedAt time.Time `json:"createdAt"`
}

type ForwardingState struct {
	TargetAPIURL     string     `json:"targetApiUrl"`
	LastStatus       string     `json:"lastStatus"`
	LastResponseCode *int       `json:"lastResponseCode"`
	LastError        string     `json:"lastError"`
	LastAt           *time.Time `json:"lastAt"`
}

func Connect(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	s := &Store{pool: pool}
	if err := s.Migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) Migrate(ctx context.Context) error {
	sqlBytes, err := migrationFS.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, string(sqlBytes))
	return err
}

func (s *Store) CountUsers(ctx context.Context) (int64, error) {
	var n int64
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash string) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO users (username, password_hash) VALUES ($1, $2)`, username, passwordHash)
	return err
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `SELECT id, username, password_hash FROM users WHERE username = $1`, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *Store) ListSources(ctx context.Context) ([]Source, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, website_url, is_enabled, created_at, updated_at
		FROM sources
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSources(rows)
}

func (s *Store) ListEnabledSources(ctx context.Context) ([]Source, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, website_url, is_enabled, created_at, updated_at
		FROM sources
		WHERE is_enabled = TRUE
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSources(rows)
}

func scanSources(rows pgx.Rows) ([]Source, error) {
	out := make([]Source, 0)
	for rows.Next() {
		var src Source
		if err := rows.Scan(&src.ID, &src.Name, &src.WebsiteURL, &src.IsEnabled, &src.CreatedAt, &src.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, src)
	}
	return out, rows.Err()
}

func (s *Store) CreateSource(ctx context.Context, name, websiteURL string, isEnabled bool) (*Source, error) {
	var src Source
	err := s.pool.QueryRow(ctx, `
		INSERT INTO sources (name, website_url, is_enabled)
		VALUES ($1, $2, $3)
		RETURNING id, name, website_url, is_enabled, created_at, updated_at
	`, name, websiteURL, isEnabled).Scan(&src.ID, &src.Name, &src.WebsiteURL, &src.IsEnabled, &src.CreatedAt, &src.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &src, nil
}

func (s *Store) UpdateSource(ctx context.Context, id int64, name, websiteURL string, isEnabled bool) (*Source, error) {
	var src Source
	err := s.pool.QueryRow(ctx, `
		UPDATE sources
		SET name = $2, website_url = $3, is_enabled = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, website_url, is_enabled, created_at, updated_at
	`, id, name, websiteURL, isEnabled).Scan(&src.ID, &src.Name, &src.WebsiteURL, &src.IsEnabled, &src.CreatedAt, &src.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &src, nil
}

func (s *Store) DeleteSource(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM sources WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListJobRuns(ctx context.Context) ([]JobRun, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, status, started_at, finished_at, error_message, extracted_count
		FROM job_runs
		ORDER BY id DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]JobRun, 0)
	for rows.Next() {
		var run JobRun
		if err := rows.Scan(&run.ID, &run.Status, &run.StartedAt, &run.FinishedAt, &run.ErrorMessage, &run.ExtractedCount); err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, rows.Err()
}

func (s *Store) CreateJobRun(ctx context.Context) (*JobRun, error) {
	var run JobRun
	err := s.pool.QueryRow(ctx, `
		INSERT INTO job_runs (status)
		VALUES ('running')
		RETURNING id, status, started_at, finished_at, error_message, extracted_count
	`).Scan(&run.ID, &run.Status, &run.StartedAt, &run.FinishedAt, &run.ErrorMessage, &run.ExtractedCount)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (s *Store) FinishJobRun(ctx context.Context, id int64, status string, extractedCount int, errorMessage string) error {
	var msg any
	if errorMessage != "" {
		msg = errorMessage
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE job_runs
		SET status = $2, extracted_count = $3, error_message = $4, finished_at = NOW()
		WHERE id = $1
	`, id, status, extractedCount, msg)
	return err
}

func (s *Store) ListExpenses(ctx context.Context) ([]ExpenseRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, shop, item, expense::text, source_id, job_run_id, created_at
		FROM expenses
		ORDER BY id DESC
		LIMIT 500
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ExpenseRow, 0)
	for rows.Next() {
		var row ExpenseRow
		if err := rows.Scan(&row.ID, &row.Shop, &row.Item, &row.Expense, &row.SourceID, &row.JobRunID, &row.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *Store) InsertExpense(ctx context.Context, shop, item, expenseAmount string, sourceID, jobRunID *int64) (*ExpenseRow, error) {
	var row ExpenseRow
	err := s.pool.QueryRow(ctx, `
		INSERT INTO expenses (shop, item, expense, source_id, job_run_id)
		VALUES ($1, $2, $3::numeric, $4, $5)
		RETURNING id, shop, item, expense::text, source_id, job_run_id, created_at
	`, shop, item, expenseAmount, sourceID, jobRunID).
		Scan(&row.ID, &row.Shop, &row.Item, &row.Expense, &row.SourceID, &row.JobRunID, &row.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Store) GetSetting(ctx context.Context, key, fallback string) (string, error) {
	var value string
	err := s.pool.QueryRow(ctx, `SELECT value FROM app_settings WHERE key = $1`, key).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fallback, nil
		}
		return "", err
	}
	return value, nil
}

func (s *Store) SetSetting(ctx context.Context, key, value string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO app_settings (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, key, value)
	return err
}

func (s *Store) TargetAPIURL(ctx context.Context, envFallback string) (string, error) {
	return s.GetSetting(ctx, settingTargetAPIURL, envFallback)
}

func (s *Store) SetTargetAPIURL(ctx context.Context, url string) error {
	return s.SetSetting(ctx, settingTargetAPIURL, url)
}

func (s *Store) InsertDelivery(ctx context.Context, expenseID int64, targetURL, status string, responseCode *int, errorMessage string) error {
	var msg any
	if errorMessage != "" {
		msg = errorMessage
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO outbound_deliveries (expense_id, target_url, status, response_code, error_message)
		VALUES ($1, $2, $3, $4, $5)
	`, expenseID, targetURL, status, responseCode, msg)
	return err
}

func (s *Store) LatestDelivery(ctx context.Context) (status string, responseCode *int, errorMessage string, createdAt *time.Time, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT status, response_code, COALESCE(error_message, ''), created_at
		FROM outbound_deliveries
		ORDER BY id DESC
		LIMIT 1
	`).Scan(&status, &responseCode, &errorMessage, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, "", nil, nil
	}
	return status, responseCode, errorMessage, createdAt, err
}

func (s *Store) ForwardingState(ctx context.Context, envFallback string) (*ForwardingState, error) {
	target, err := s.TargetAPIURL(ctx, envFallback)
	if err != nil {
		return nil, err
	}
	status, code, message, at, err := s.LatestDelivery(ctx)
	if err != nil {
		return nil, err
	}
	return &ForwardingState{
		TargetAPIURL:     target,
		LastStatus:       status,
		LastResponseCode: code,
		LastError:        message,
		LastAt:           at,
	}, nil
}
