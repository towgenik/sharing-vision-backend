package post

import (
	"database/sql"
	"errors"
	"math"
)

// ErrNotFound is returned when an article id does not exist.
var ErrNotFound = errors.New("post not found")

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

// List returns all articles, optionally filtered by status.
func (s *MySQLRepository) List(status string) ([]Article, error) {
	if status == "" {
		rows, err := s.DB.Query("SELECT " + cols + " FROM posts ORDER BY `Id` DESC")
		if err != nil {
			return nil, err
		}
		return mapRows(rows)
	}
	rows, err := s.DB.Query("SELECT "+cols+" FROM posts WHERE `Status` = ? ORDER BY `Id` DESC", status)
	if err != nil {
		return nil, err
	}
	return mapRows(rows)
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

// Preview returns a page of published articles plus the total count.
func (s *MySQLRepository) Preview(page, perPage int) ([]Article, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	var total int64
	if err := s.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE `Status` = ?", StatusPublish).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.DB.Query(
		"SELECT "+cols+" FROM posts WHERE `Status` = ? ORDER BY `Updated_date` DESC, `Id` DESC LIMIT ? OFFSET ?",
		StatusPublish, perPage, (page-1)*perPage)
	if err != nil {
		return nil, 0, err
	}
	list, err := mapRows(rows)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// TotalPages yields the pagination bucket count.
func TotalPages(total int64, perPage int) int64 {
	if perPage < 1 {
		perPage = 1
	}
	return int64(math.Ceil(float64(total) / float64(perPage)))
}

var _ Repository = (*MySQLRepository)(nil)
