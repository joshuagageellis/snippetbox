package models

import (
	"database/sql"
	"errors"
	"time"
)

// Define a Snippet type to hold the data for an individual snippet. Notice how
// the fields of the struct correspond to the fields in our MySQL snippets
// table?
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// Define a SnippetModel type which wraps a sql.DB connection pool.
type SnippetModel struct {
	DB *sql.DB
}

// This will insert a new snippet into the database.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	// Use the LastInsertId() method on the result to get the ID of our
	// newly inserted record in the snippets table.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// The ID returned has the type int64, so we convert it to an int type
	// before returning.
	return int(id), nil
}

// This will return a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (Snippet, error) {
	var s Snippet

	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() AND id = ?`

	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		} else {
			return Snippet{}, err
		}
	}

	return s, nil
}

// This will return the # most recently created snippets.
func (m *SnippetModel) Latest(c int) ([]Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT ?`

	rows, err := m.DB.Query(stmt, c)
	if err != nil {
		return nil, err
	}

	// Ensure close.
	defer rows.Close()

	var snippets []Snippet

	for rows.Next() {
		// Create a pointer to a new zeroed Snippet struct.
		var s Snippet
		// Use rows.Scan() to copy the values from each field in the row to the
		// new Snippet object that we created. Again, the arguments to row.Scan()
		// must be pointers to the place you want to copy the data into, and the
		// number of arguments must be exactly the same as the number of
		// columns returned by your statement.
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		// Append it to the slice of snippets.
		snippets = append(snippets, s)
	}

	// When the rows.Next() loop has finished we call rows.Err() to retrieve any
	// error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}

// Create table if it does not exist.
func (m *SnippetModel) CreateSnippetTable() error {
	stmt := `
		CREATE TABLE IF NOT EXISTS snippets (
			id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
			title VARCHAR(100) NOT NULL,
			content TEXT NOT NULL,
			created DATETIME NOT NULL,
			expires DATETIME NOT NULL
		)
	`
	_, err := m.DB.Exec(stmt)
	return err
}

// Create index.
func (m *SnippetModel) CreateSnippetIndex() error {
	stmt := `CREATE INDEX idx_snippets_created ON snippets(created)`
	_, err := m.DB.Exec(stmt)
	return err
}

// Create sessions table.
func (m *SnippetModel) CreateSessionTable() error {
	stmt := `
		CREATE TABLE IF NOT EXISTS sessions (
			token CHAR(43) PRIMARY KEY,
			data BLOB NOT NULL,
			expiry TIMESTAMP(6) NOT NULL
		)
	`
	_, err := m.DB.Exec(stmt)
	return err
}

// Create sessions index.
func (m *SnippetModel) CreateSessionIndex() error {
	stmt := `CREATE INDEX sessions_expiry_idx ON sessions(expiry);`
	_, err := m.DB.Exec(stmt)
	return err
}

// Dev seed database.
func (m *SnippetModel) SeedDatabase() error {
	// Check if tables exist, if so return.
	stmt := `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name IN ('snippets', 'sessions')`
	var count int
	err := m.DB.QueryRow(stmt).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	// Setup tables and indexes.
	if err := m.CreateSnippetTable(); err != nil {
		return err
	}
	if err := m.CreateSnippetIndex(); err != nil {
		return err
	}
	if err := m.CreateSessionTable(); err != nil {
		return err
	}
	if err := m.CreateSessionIndex(); err != nil {
		return err
	}
	return nil
}
