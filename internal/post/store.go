package post

import (
	"database/sql"
	"errors"
)

// ErrNotFound is returned when an article id does not exist.
var ErrNotFound = errors.New("post not found")

// ListMaxLimit caps how many rows a single list request may return.
const ListMaxLimit = 100

// MySQLRepository implements Repository on top of a *sql.DB.
type MySQLRepository struct {
	DB *sql.DB
}

const cols = "`Id`, `Title`, `Content`, `Category`, `Created_date`, `Updated_date`, `Status`"

func scanArticle(sc interface{ Scan(dest ...any) error }) (*Article, error) {
	var a Article
	err := sc.Scan(&a.Id, &a.Title, &a.Content, &a.Category,
		&a.CreatedDate, &a.UpdatedDate, &a.Status)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func mapRows(rows *sql.Rows) ([]Article, error) {
	defer rows.Close()
	out := []Article{}
	for rows.Next() {
		a, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

// Create inserts an article and returns the stored row.
func (s *MySQLRepository) Create(a *Article) (*Article, error) {
	res, err := s.DB.Exec(
		"INSERT INTO posts (`Title`, `Content`, `Category`, `Status`) VALUES (?, ?, ?, ?)",
		a.Title, a.Content, a.Category, a.Status)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.Get(id)
}

// List returns one page of rows ordered newest first, optionally filtered by
// status, plus the total row count matching the filter (before pagination).
func (s *MySQLRepository) List(limit, offset int, status string) ([]Article, int64, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > ListMaxLimit {
		limit = ListMaxLimit
	}
	if offset < 0 {
		offset = 0
	}

	where := ""
	args := []any{}
	if status != "" {
		where = "WHERE `Status` = ?"
		args = append(args, status)
	}

	var total int64
	if err := s.DB.QueryRow("SELECT COUNT(*) FROM posts "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.DB.Query(
		"SELECT "+cols+" FROM posts "+where+" ORDER BY `Id` DESC LIMIT ? OFFSET ?",
		append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	list, err := mapRows(rows)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Get returns a single article by id.
func (s *MySQLRepository) Get(id int64) (*Article, error) {
	row := s.DB.QueryRow("SELECT "+cols+" FROM posts WHERE `Id` = ?", id)
	a, err := scanArticle(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

// Update performs a full update (including status); Updated_date refreshes via ON UPDATE.
func (s *MySQLRepository) Update(a *Article) (*Article, error) {
	res, err := s.DB.Exec(
		"UPDATE posts SET `Title` = ?, `Content` = ?, `Category` = ?, `Status` = ? WHERE `Id` = ?",
		a.Title, a.Content, a.Category, a.Status, a.Id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return s.Get(a.Id)
}

// Trash soft-deletes an article by flipping its status to thrash.
func (s *MySQLRepository) Trash(id int64) error {
	res, err := s.DB.Exec("UPDATE posts SET `Status` = ? WHERE `Id` = ?", StatusTrash, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

var _ Repository = (*MySQLRepository)(nil)
