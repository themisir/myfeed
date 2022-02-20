package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/themisir/myfeed/pkg/models"
)

type Connection struct {
	db *sql.DB
}

func Connect(dataSource string) (*Connection, error) {
	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Connection{db}, nil
}

func (c *Connection) Users() (models.UserRepository, error) {
	return newUserRepository(c)
}

func (c *Connection) Sources() (models.SourceRepository, error) {
	return newSourceRepository(c)
}

func (c *Connection) Feeds() (models.FeedRepository, error) {
	return newFeedRepository(c)
}

func (c *Connection) Posts() (models.PostRepository, error) {
	return newPostRepository(c)
}

func (c *Connection) Close() error {
	return c.db.Close()
}

func (c *Connection) Batch() *Batch {
	return &Batch{
		db:         c.db,
		operations: []Operation{},
	}
}

type Operation interface {
	Exec(db *sql.DB) error
}

type Batch struct {
	db         *sql.DB
	operations []Operation
}

func (b *Batch) Add(op Operation) *Batch {
	b.operations = append(b.operations, op)
	return b
}

func (b *Batch) Exec() (err error) {
	for _, item := range b.operations {
		if err = item.Exec(b.db); err != nil {
			return
		}
	}
	return
}

func (b *Batch) Prepare(query string, stmt **sql.Stmt) *Batch {
	return b.Add(&prepare{query, stmt})
}

type prepare struct {
	query string
	stmt  **sql.Stmt
}

func (p *prepare) Exec(db *sql.DB) (err error) {
	*p.stmt, err = db.Prepare(p.query)
	if err != nil {
		err = fmt.Errorf("prepare query %s failed: %s", p.query, err)
	}
	return
}
