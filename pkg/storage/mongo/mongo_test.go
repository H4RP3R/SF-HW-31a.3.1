package mongo

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"GoNews/pkg/storage"
)

func mongoConf() Config {
	conf := Config{
		Host:   "localhost",
		Port:   "27018",
		DBName: "gonews",
	}

	return conf
}

func storageConnect() (*Store, error) {
	conf := mongoConf()
	db, err := New(conf)
	if err != nil {
		return nil, storage.ErrConnectDB
	}

	err = db.Ping()
	if err != nil {
		return nil, storage.ErrDBNotResponding
	}

	return db, nil
}

// restoreDB restores the original state of DB for further testing.
func restoreDB(db *Store) error {
	collection := db.client.Database(db.dbName).Collection("posts")
	return collection.Drop(context.Background())
}

func TestStore_AddPost(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := restoreDB(db)
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

	collection := db.client.Database(db.dbName).Collection("posts")
	postCnt, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("unexpected error counting post in DB: %v", err)
	}
	if postCnt != int64(len(storage.TestPosts)) {
		t.Errorf("expected %d posts in DB, but got %d", len(storage.TestPosts), postCnt)
	}
}

func TestStore_AddPost_duplicatedID(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := restoreDB(db)
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

	dupPost := storage.TestPosts[0]
	err = db.AddPost(dupPost)
	if !mongo.IsDuplicateKeyError(err) {
		t.Errorf("expected error inserting document with duplicate ID, got %v", err)
	}
}

func TestStore_Posts(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := restoreDB(db)
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

	collection := db.client.Database(db.dbName).Collection("posts")
	postCnt, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("unexpected error counting post in DB: %v", err)
	}
	if postCnt != int64(len(storage.TestPosts)) {
		t.Fatalf("expected %d posts in DB, but got %d", len(storage.TestPosts), postCnt)
	}

	posts, err := db.Posts()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	postCnt = int64(len(posts))
	if postCnt != int64(len(storage.TestPosts)) {
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
		err := restoreDB(db)
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

	collection := db.client.Database(db.dbName).Collection("posts")
	postCnt, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("unexpected error counting post in DB: %v", err)
	}
	if postCnt != int64(len(storage.TestPosts)) {
		t.Fatalf("expected %d posts in DB, but got %d", len(storage.TestPosts), postCnt)
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
	collection = db.client.Database(db.dbName).Collection("posts")
	filter := bson.D{{Key: "id", Value: targetPost.ID}}
	err = collection.FindOne(context.Background(), filter).Decode(&updatedPost)
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
		err := restoreDB(db)
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
		err := restoreDB(db)
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

	collection := db.client.Database(db.dbName).Collection("posts")
	postCnt, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("unexpected error counting post in DB: %v", err)
	}
	if postCnt != int64(len(storage.TestPosts)) {
		t.Fatalf("expected %d posts in DB, but got %d", len(storage.TestPosts), postCnt)
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

	collection = db.client.Database(db.dbName).Collection("posts")
	postCnt, err = collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("unexpected error counting post in DB: %v", err)
	}
	if postCnt > 0 {
		t.Errorf("expected %d posts in DB, but got %d", len(storage.TestPosts), postCnt)
	}
}

func TestStore_DeletePost_postNotExist(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := restoreDB(db)
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
	db, _ := storageConnect()
	restoreDB(db)
}
