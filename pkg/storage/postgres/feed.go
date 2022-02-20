package postgres

import (
	"database/sql"
	"fmt"
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
	"github.com/themisir/myfeed/pkg/updating"
)

type feedRepository struct {
	c                *Connection
	addFeedStmt      *sql.Stmt
	getUserFeedsStmt *sql.Stmt
	getFeedStmt      *sql.Stmt
	removeFeedStmt   *sql.Stmt
	updateFeedStmt   *sql.Stmt
}

const (
	addFeedQuery           = `INSERT INTO feeds (name, user_id, is_public) VALUES (?, ?, ?) RETURNING id`
	getUserFeedsQuery      = `SELECT (id, name, user_id, is_public) FROM feeds WHERE user_id = ?`
	getFeedQuery           = `SELECT (id, name, user_id, is_public) FROM feeds WHERE id = ?`
	removeFeedQuery        = `DELETE FROM feeds WHERE id = ?`
	updateFeedQuery        = `UPDATE feeds SET name = ?, is_public = ? WHERE id = ?`
	removeFeedSourcesQuery = `DELETE FROM feed_source WHERE feed_id = ?`
)

func newFeedRepository(c *Connection) (r *feedRepository, err error) {
	r = &feedRepository{c: c}
	err = c.Batch().
		Prepare(addFeedQuery, &r.addFeedStmt).
		Prepare(getUserFeedsQuery, &r.getUserFeedsStmt).
		Prepare(getFeedQuery, &r.getFeedStmt).
		Prepare(removeFeedQuery, &r.removeFeedStmt).
		Prepare(updateFeedQuery, &r.updateFeedStmt).
		Exec()
	return
}

func (r *feedRepository) AddFeed(data adding.FeedData) (adding.Feed, error) {
	var id int
	err := r.addFeedStmt.QueryRow(data.Name, data.UserId, data.IsPublic).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &feed{
		id:       id,
		name:     data.Name,
		userId:   data.UserId,
		isPublic: data.IsPublic,
	}, nil
}

func (r *feedRepository) GetUserFeeds(userId string) ([]listing.Feed, error) {
	rows, err := r.getUserFeedsStmt.Query(userId)
	if err != nil {
		return nil, err
	}
	var result []listing.Feed
	for rows.Next() {
		var f feed
		err := rows.Scan(&f.id, &f.name, &f.userId, &f.isPublic)
		if err != nil {
			return nil, err
		}
		result = append(result, &f)
	}
	return result, nil
}

func (r *feedRepository) GetFeed(feedId int) (listing.Feed, error) {
	var f feed
	err := r.getFeedStmt.QueryRow(feedId).Scan(&f.id, &f.name, &f.userId, &f.isPublic)
	return &f, err
}

func (r *feedRepository) RemoveFeed(feedId int) error {
	_, err := r.removeFeedStmt.Exec(feedId)
	return err
}

func (r *feedRepository) UpdateFeed(feedId int, data updating.Feed) error {
	_, err := r.updateFeedStmt.Exec(feedId, data.Name, data.IsPublic)
	return err
}

func (r *feedRepository) UpdateFeedSources(feedId int, sourceIds ...int) error {
	// build insert query
	insertQuery := `INSERT INTO feed_source (feed_id, source_id) VALUES `
	for i, sourceId := range sourceIds {
		var prefix string
		if i > 0 {
			prefix = " ,"
		}
		insertQuery += fmt.Sprintf("%s(%v, %v)", prefix, feedId, sourceId)
	}

	// Apply updates
	tx, err := r.c.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(removeFeedSourcesQuery, feedId); err != nil {
		return err
	}
	if _, err := tx.Exec(insertQuery); err != nil {
		return err
	}
	return tx.Commit()
}

type feed struct {
	id       int
	name     string
	userId   string
	isPublic bool
}

func (f *feed) Id() int {
	return f.id
}

func (f *feed) Name() string {
	return f.name
}

func (f *feed) UserId() string {
	return f.userId
}

func (f *feed) IsPublic() bool {
	return f.isPublic
}
