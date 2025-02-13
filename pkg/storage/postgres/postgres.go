package postgres

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"GoNews/pkg/storage"
)

type Store struct {
	db *pgxpool.Pool
}

func New(constr string) (*Store, error) {
	db, err := pgxpool.Connect(context.Background(), constr)
	if err != nil {
		return nil, err
	}
	s := Store{
		db: db,
	}
	return &s, nil
}

func (s *Store) Ping() error {
	return s.db.Ping(context.Background())
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) AddPost(post storage.Post) error {
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO posts (author_id, title, content, created_at, published_at)
		VALUES ($1, $2, $3, $4, $5)
	`,
		post.AuthorID,
		post.Title,
		post.Content,
		post.CreatedAt,
		post.PublishedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Posts() ([]storage.Post, error) {
	rows, err := s.db.Query(context.Background(), `
		SELECT
			p.id,
			p.title,
			p.content,
			p.author_id,
			a.name,
			p.created_at,
			p.published_at
		FROM posts AS p
		JOIN authors AS a
		ON p.author_id = a.id
	`)
	if err != nil {
		return nil, err
	}

	var posts []storage.Post
	for rows.Next() {
		var p storage.Post
		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&p.AuthorID,
			&p.AuthorName,
			&p.CreatedAt,
			&p.PublishedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

func (s *Store) UpdatePost(p storage.Post) error {
	result, err := s.db.Exec(context.Background(), `
		UPDATE posts
		SET
			title = $2,
			content = $3,
			author_id = $4,
			created_at = $5,
			published_at = $6
		WHERE id = $1
	`,
		p.ID,
		p.Title,
		p.Content,
		p.AuthorID,
		p.CreatedAt,
		p.PublishedAt,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return storage.ErrEntryNotExist
	}

	return nil
}

func (s *Store) DeletePost(post storage.Post) error {
	result, err := s.db.Exec(context.Background(), `
		DELETE FROM posts
		WHERE id = $1
	`,
		post.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return storage.ErrEntryNotExist
	}

	return nil
}
