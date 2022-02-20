package postgres

import (
	"database/sql"
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
	"github.com/themisir/myfeed/pkg/updating"
)

type sourceRepository struct {
	c                   *Connection
	addSourceStmt       *sql.Stmt
	getSourceStmt       *sql.Stmt
	getSourcesStmt      *sql.Stmt
	getFeedSourcesStmt  *sql.Stmt
	findSourceByUrlStmt *sql.Stmt
	removeSourceStmt    *sql.Stmt
	updateSourceStmt    *sql.Stmt
}

const (
	addSourceQuery       = `INSERT INTO sources (title, url) VALUES (?, ?) RETURNING id`
	getSourceQuery       = `SELECT (id, title, url) FROM sources WHERE id = ?`
	getSourcesQuery      = `SELECT (id, title, url) FROM sources`
	getFeedSourcesQuery  = `SELECT (id, title, url) FROM sources JOIN source_feed sf ON sf.source_id = sources.id WHERE sf.feed_id = ?`
	findSourceByUrlQuery = `SELECT (id, title, url) FROM sources WHERE url = ?`
	removeSource         = `DELETE FROM sources WHERE id = ?`
	updateSource         = `UPDATE sources SET title = ? WHERE id = ?`
)

func newSourceRepository(c *Connection) (r *sourceRepository, err error) {
	r = &sourceRepository{c: c}
	err = c.Batch().
		Prepare(addSourceQuery, &r.addSourceStmt).
		Prepare(getSourceQuery, &r.getSourceStmt).
		Prepare(getSourcesQuery, &r.getSourcesStmt).
		Prepare(getFeedSourcesQuery, &r.getFeedSourcesStmt).
		Prepare(findSourceByUrlQuery, &r.findSourceByUrlStmt).
		Prepare(removeSource, &r.removeSourceStmt).
		Prepare(updateSource, &r.updateSourceStmt).
		Exec()
	return
}

func (r *sourceRepository) AddSource(data adding.SourceData) (adding.Source, error) {
	var id int
	err := r.addSourceStmt.QueryRow(data.Title, data.Url).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &source{
		id:    id,
		title: data.Title,
		url:   data.Url,
	}, nil
}

func (r *sourceRepository) scanRow(row *sql.Row) (listing.Source, error) {
	var s source
	err := row.Scan(&s.id, &s.title, &s.url)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *sourceRepository) scanRows(rows *sql.Rows) ([]listing.Source, error) {
	var sources []listing.Source
	for rows.Next() {
		var s source
		if err := rows.Scan(&s.id, &s.title, &s.url); err != nil {
			return nil, err
		}
		sources = append(sources, &s)
	}
	return sources, nil
}

func (r *sourceRepository) GetSource(sourceId int) (listing.Source, error) {
	return r.scanRow(r.getSourceStmt.QueryRow(sourceId))
}

func (r *sourceRepository) GetSources() ([]listing.Source, error) {
	rows, err := r.getSourcesStmt.Query()
	if err != nil {
		return nil, err
	}
	return r.scanRows(rows)
}

func (r *sourceRepository) GetFeedSources(feedId int) ([]listing.Source, error) {
	rows, err := r.getFeedSourcesStmt.Query(feedId)
	if err != nil {
		return nil, err
	}
	return r.scanRows(rows)
}

func (r *sourceRepository) FindSourceByUrl(url string) (listing.Source, error) {
	return r.scanRow(r.findSourceByUrlStmt.QueryRow(url))
}

func (r *sourceRepository) RemoveSource(sourceId int) (err error) {
	_, err = r.removeSourceStmt.Exec(sourceId)
	return
}

func (r *sourceRepository) UpdateSource(sourceId int, data updating.Source) (err error) {
	_, err = r.updateSourceStmt.Exec(data.Title, sourceId)
	return
}

type source struct {
	id    int
	title string
	url   string
}

func (s *source) Id() int {
	return s.id
}

func (s *source) Title() string {
	return s.title
}

func (s *source) Url() string {
	return s.url
}
