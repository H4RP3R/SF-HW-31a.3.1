package postgres

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"GoNews/pkg/storage"
)

func postgresConf() Config {
	conf := Config{
		User:     "postgres",
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Host:     "localhost",
		Port:     "5433",
		DBName:   "gonews",
	}

	return conf
}

func storageConnect() (*Store, error) {
	conf := postgresConf()
	db, err := New(conf.ConString())
	if err != nil {
		return nil, storage.ErrConnectDB
	}

	err = db.Ping()
	if err != nil {
		return nil, storage.ErrDBNotResponding
	}

	return db, nil
}

// truncatePosts restores the original state of DB for further testing.
func truncatePosts(db *Store) error {
	_, err := db.db.Exec(context.Background(), "TRUNCATE TABLE posts")
	if err != nil {
		return err
	}

	_, err = db.db.Exec(context.Background(), "ALTER SEQUENCE posts_id_seq RESTART WITH 1")
	if err != nil {
		return nil
	}

	return nil
}

func TestStore_AddPost(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := truncatePosts(db)
		if err != nil {
			t.Errorf("unexpected error clearing posts table: %v", err)
		}

		db.Close()
	})

	for _, tp := range storage.TestPosts {
		err := db.AddPost(tp)
		if err != nil {
			t.Fatalf("unexpected error adding post: %v", err)
		}
	}

	var postCnt int
	err = db.db.QueryRow(context.Background(), `
		SELECT COUNT(id) FROM posts
	`).Scan(&postCnt)
	if err != nil {
		t.Fatalf("unexpected error counting post in DB: %v", err)
	}
	if postCnt != len(storage.TestPosts) {
		t.Errorf("expected %d posts in DB, but got %d", len(storage.TestPosts), postCnt)
	}

}

func TestStore_Posts(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := truncatePosts(db)
		if err != nil {
			t.Errorf("unexpected error clearing posts table: %v", err)
		}

		db.Close()
	})

	for _, tp := range storage.TestPosts {
		err := db.AddPost(tp)
		if err != nil {
			t.Fatalf("unexpected error adding post: %v", err)
		}
	}

	var postCnt int
	err = db.db.QueryRow(context.Background(), `
		SELECT COUNT(id) FROM posts
	`).Scan(&postCnt)
	if err != nil {
		t.Fatalf("unexpected error counting post in DB: %v", err)
	}
	if postCnt != len(storage.TestPosts) {
		t.Fatalf("posts in DB %d, should be %d, aborting test", postCnt, len(storage.TestPosts))
	}

	posts, err := db.Posts()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	postCnt = len(posts)
	if postCnt != len(storage.TestPosts) {
		t.Errorf("expected %d posts, but got %d posts", len(storage.TestPosts), postCnt)
	}
	if !reflect.DeepEqual(posts, storage.TestPosts) {
		t.Errorf("posts do not match expected posts. Expected: %+v, Got: %+v", storage.TestPosts, posts)
	}
}

func TestStore_UpdatePost_postExists(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := truncatePosts(db)
		if err != nil {
			t.Errorf("unexpected error clearing posts table: %v", err)
		}

		db.Close()
	})

	for _, tp := range storage.TestPosts {
		err := db.AddPost(tp)
		if err != nil {
			t.Fatalf("unexpected error adding post: %v", err)
		}
	}

	targetPost := storage.TestPosts[0]
	targetPost.Title = "Updated title"
	targetPost.Content = "Updated content"
	targetPost.AuthorID = 3
	targetPost.AuthorName = "Travis"
	targetPost.CreatedAt = time.Now().Unix()
	targetPost.PublishedAt = time.Now().Add(time.Hour).Unix()

	err = db.UpdatePost(targetPost)
	if err != nil {
		t.Errorf("unexpected error updating post: %v", err)
	}

	var updatedPost storage.Post
	err = db.db.QueryRow(context.Background(), `
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
		WHERE p.id = $1
	`, targetPost.ID).Scan(
		&updatedPost.ID,
		&updatedPost.Title,
		&updatedPost.Content,
		&updatedPost.AuthorID,
		&updatedPost.AuthorName,
		&updatedPost.CreatedAt,
		&updatedPost.PublishedAt,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(updatedPost, targetPost) {
		t.Errorf("updated post do not match target post. Expected: %+v, Got: %+v", targetPost, updatedPost)
	}
}

func TestStore_UpdatePost_postNotExist(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := truncatePosts(db)
		if err != nil {
			t.Errorf("unexpected error clearing posts table: %v", err)
		}

		db.Close()
	})

	for _, tp := range storage.TestPosts {
		err := db.AddPost(tp)
		if err != nil {
			t.Fatalf("unexpected error adding post: %v", err)
		}
	}

	targetPost := storage.Post{}
	targetPost.ID = 999999
	targetPost.Title = "Updated title"
	targetPost.Content = "Updated content"
	targetPost.AuthorID = 3
	targetPost.AuthorName = "Travis"
	targetPost.CreatedAt = time.Now().Unix()
	targetPost.PublishedAt = time.Now().Add(time.Hour).Unix()

	err = db.UpdatePost(targetPost)
	if !errors.Is(err, storage.ErrEntryNotExist) {
		t.Errorf("expected error %v, got error %v", storage.ErrEntryNotExist, err)
	}
}

func TestStore_DeletePost_postExists(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := truncatePosts(db)
		if err != nil {
			t.Errorf("unexpected error clearing posts table: %v", err)
		}

		db.Close()
	})

	for _, tp := range storage.TestPosts {
		err := db.AddPost(tp)
		if err != nil {
			t.Fatalf("unexpected error adding post: %v", err)
		}
	}

	var postCnt int
	err = db.db.QueryRow(context.Background(), `
		SELECT COUNT(id) FROM posts
	`).Scan(&postCnt)
	if err != nil {
		t.Fatalf("unexpected error counting post in DB: %v", err)
	}
	if postCnt != len(storage.TestPosts) {
		t.Fatalf("posts in DB %d, should be %d, aborting test", postCnt, len(storage.TestPosts))
	}

	for _, post := range storage.TestPosts {
		err := db.DeletePost(post)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		postsRemain, err := db.Posts()
		if err != nil {
			t.Fatalf("unexpected error retrieving posts: %v", err)
		}
		for _, remPost := range postsRemain {
			if remPost.ID == post.ID {
				t.Errorf("post ID:%v wasn't deleted from DB", post.ID)
			}
		}
	}

	err = db.db.QueryRow(context.Background(), `
		SELECT COUNT(id) FROM posts
	`).Scan(&postCnt)
	if err != nil {
		t.Fatalf("unexpected error counting post in DB: %v", err)
	}
	if postCnt > 0 {
		t.Errorf("DB should be empty. Posts in DB %d", postCnt)
	}
}

func TestStore_DeletePost_postNotExist(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := truncatePosts(db)
		if err != nil {
			t.Errorf("unexpected error clearing posts table: %v", err)
		}

		db.Close()
	})

	for _, tp := range storage.TestPosts {
		err := db.AddPost(tp)
		if err != nil {
			t.Fatalf("unexpected error adding post: %v", err)
		}
	}

	nonExistentPost := storage.Post{ID: 999999}
	err = db.DeletePost(nonExistentPost)
	if !errors.Is(err, storage.ErrEntryNotExist) {
		t.Errorf("expected error %v, got error %v", storage.ErrEntryNotExist, err)
	}
}

func init() {
	log.SetOutput(io.Discard)
}
