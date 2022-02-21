package postgres

import (
	"database/sql"
	"fmt"
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
	"github.com/themisir/myfeed/pkg/updating"
	"time"
)

type postRepository struct {
	c                        *Connection
	addPostStmt              *sql.Stmt
	getSourcePostsStmt       *sql.Stmt
	getFeedPostsStmt         *sql.Stmt
	removeSourcePostStmt     *sql.Stmt
	removeAllSourcePostsStmt *sql.Stmt
	updateSourcePostStmt     *sql.Stmt
}

const (
	addPostQuery              = `INSERT INTO posts (source_id, title, description, url, published_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	getSourcePostsQuery       = `SELECT id, title, description, url, published_at, updated_at FROM posts WHERE source_id = $1 ORDER BY published_at DESC, created_at DESC`
	getFeedPostsQuery         = `SELECT p.id, p.title, p.description, p.url, p.published_at, p.updated_at, s.id, s.title, s.url FROM posts p JOIN sources s ON s.id = p.source_id JOIN feed_source fs ON fs.source_id = p.source_id WHERE fs.feed_id = $1 ORDER BY published_at DESC, created_at DESC`
	removeSourcePostQuery     = `DELETE FROM posts WHERE source_id = $1 AND id = $2`
	removeAllSourcePostsQuery = `DELETE FROM posts WHERE source_id = $1`
	updateSourcePostQuery     = `UPDATE posts SET title = $1, description = $2, url = $3, published_at = $4, updated_at = $5 WHERE source_id = $6 AND id = $7`
)

func newPostRepository(c *Connection) (r *postRepository, err error) {
	r = &postRepository{c: c}
	err = c.Batch().
		Prepare(addPostQuery, &r.addPostStmt).
		Prepare(getSourcePostsQuery, &r.getSourcePostsStmt).
		Prepare(getFeedPostsQuery, &r.getFeedPostsStmt).
		Prepare(removeSourcePostQuery, &r.removeSourcePostStmt).
		Prepare(removeAllSourcePostsQuery, &r.removeAllSourcePostsStmt).
		Prepare(updateSourcePostQuery, &r.updateSourcePostStmt).
		Exec()
	return
}

func (r *postRepository) AddPost(data adding.PostData) (adding.Post, error) {
	var id int
	err := r.addPostStmt.QueryRow(data.SourceId, data.Title, data.Description, data.Url, data.PublishedAt, data.UpdatedAt).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &post{
		id:          id,
		title:       data.Title,
		description: data.Description,
		url:         data.Url,
		publishedAt: data.PublishedAt,
		updatedAt:   data.UpdatedAt,
	}, nil
}

func (r *postRepository) AddManyPosts(items ...adding.PostData) error {
	var query string
	params := make([]interface{}, 6*len(items))
	for i, item := range items {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($%v, $%v, $%v, $%v, $%v, $%v)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6)
		params := params[i*6:]
		params[0] = item.SourceId
		params[1] = item.Title
		params[2] = item.Description
		params[3] = item.Url
		params[4] = item.PublishedAt
		params[5] = item.UpdatedAt
	}
	query = fmt.Sprintf(`INSERT INTO posts (source_id, title, description, url, published_at, updated_at) VALUES %s`, query)
	_, err := r.c.db.Exec(query, params...)
	return err
}

func (r *postRepository) GetSourcePosts(sourceId int) ([]listing.Post, error) {
	rows, err := r.getSourcePostsStmt.Query(sourceId)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var result []listing.Post
	for rows.Next() {
		var p post
		err := rows.Scan(&p.id, &p.title, &p.description, &p.url, &p.publishedAt, &p.updatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, &p)
	}
	return result, nil
}

func (r *postRepository) GetFeedPosts(feedId int) ([]listing.SourcePost, error) {
	rows, err := r.getFeedPostsStmt.Query(feedId)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var result []listing.SourcePost
	for rows.Next() {
		var p sourcePost
		err := rows.Scan(&p.id, &p.title, &p.description, &p.url, &p.publishedAt, &p.updatedAt, &p.source.id, &p.source.title, &p.source.url)
		if err != nil {
			return nil, err
		}
		result = append(result, &p)
	}
	return result, nil
}

func (r *postRepository) RemoveSourcePost(sourceId int, postId int) error {
	_, err := r.removeSourcePostStmt.Exec(sourceId, postId)
	return err
}

func (r *postRepository) RemoveAllSourcePosts(sourceId int) error {
	_, err := r.removeAllSourcePostsStmt.Exec(sourceId)
	return err
}

func (r *postRepository) UpdateSourcePost(sourceId int, postId int, data updating.Post) error {
	_, err := r.updateSourcePostStmt.Exec(data.Title, data.Description, data.Url, data.PublishedAt, data.UpdatedAt, sourceId, postId)
	return err
}

type post struct {
	id          int
	title       string
	description string
	url         string
	publishedAt *time.Time
	updatedAt   *time.Time
}

func (p *post) Id() int {
	return p.id
}

func (p *post) Title() string {
	return p.title
}

func (p *post) Description() string {
	return p.description
}

func (p *post) Url() string {
	return p.url
}

func (p *post) PublishedAt() *time.Time {
	return p.publishedAt
}

func (p *post) UpdatedAt() *time.Time {
	return p.updatedAt
}

type sourcePost struct {
	post
	source source
}

func (p *sourcePost) Source() listing.Source {
	return &p.source
}
