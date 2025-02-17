package mongo

import (
	"GoNews/pkg/storage"
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	client *mongo.Client
	dbName string
}

func New(conf Config) (*Store, error) {
	s := Store{
		dbName: conf.DBName,
	}

	opt := conf.Options()
	client, err := mongo.Connect(context.Background(), opt)
	if err != nil {
		return nil, err
	}

	s.client = client

	// Create the posts collection in advance if not exist.
	collExists, err := collectionExists(s.client.Database(s.dbName), "posts")
	if err != nil {
		return nil, err
	}
	if !collExists {
		err = s.client.Database(s.dbName).CreateCollection(context.Background(), "posts")
		if err != nil {
			return nil, err
		}
	}

	// Create the unique index on ID field.
	err = s.CreateUniqueIndexOnID()
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Store) Ping() error {
	return s.client.Ping(context.Background(), nil)
}

func (s *Store) Close() {
	s.client.Disconnect(context.Background())
}

func (s *Store) AddPost(post storage.Post) error {
	collection := s.client.Database(s.dbName).Collection("posts")
	_, err := collection.InsertOne(context.Background(), post)
	if err != nil {
		log.Errorf("error adding post: %v", err)
		return err
	}

	log.Infof("post ID:%v added successfully", post.ID)
	return nil
}

func (s *Store) Posts() ([]storage.Post, error) {
	collection := s.client.Database(s.dbName).Collection("posts")
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Errorf("error requesting posts: %v", err)
		return nil, err
	}
	defer cur.Close(context.Background())

	var posts []storage.Post
	for cur.Next(context.Background()) {
		var p storage.Post
		err := cur.Decode(&p)
		if err != nil {
			log.Errorf("error requesting posts: %v", err)
			return nil, err
		}
		posts = append(posts, p)
	}

	log.Infof("retrieved %d posts", len(posts))
	return posts, cur.Err()
}

func (s *Store) UpdatePost(post storage.Post) error {
	collection := s.client.Database(s.dbName).Collection("posts")
	filter := bson.D{{Key: "id", Value: post.ID}}
	update := bson.D{{Key: "$set", Value: bson.M{
		"Title":       post.Title,
		"Content":     post.Content,
		"AuthorID":    post.AuthorID,
		"AuthorName":  post.AuthorName,
		"CreatedAt":   post.CreatedAt,
		"PublishedAt": post.PublishedAt,
	}}}
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Errorf("error updating post: %v", err)
		return err
	}

	if result.ModifiedCount == 0 {
		log.Errorf("error updating post: post with ID %v not found", post.ID)
		return storage.ErrEntryNotExist
	}

	log.Infof("post ID:%v updated successfully", post.ID)
	return nil
}

func (s *Store) DeletePost(post storage.Post) error {
	collection := s.client.Database(s.dbName).Collection("posts")
	filter := bson.D{{Key: "id", Value: post.ID}}
	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Errorf("error deleting post: %v", err)
		return err
	}

	if result.DeletedCount == 0 {
		log.Errorf("error deleting post: post with ID %v not found", post.ID)
		return storage.ErrEntryNotExist
	}

	log.Infof("post ID:%v deleted successfully", post.ID)
	return nil
}

// CreateUniqueIndexOnID creates a unique index on the id field if not exists.
func (s *Store) CreateUniqueIndexOnID() error {
	collection := s.client.Database(s.dbName).Collection("posts")
	cur, err := collection.Indexes().List(context.Background())
	if err != nil {
		return err
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var index bson.M
		err = cur.Decode(&index)
		if err != nil {
			return err
		}
		// Mongo automatically names the index on the "id" as "id_1".
		if index["name"] == "id_1" {
			return nil
		}
	}
	if err := cur.Err(); err != nil {
		return err
	}

	_, err = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	return nil
}

func collectionExists(db *mongo.Database, collName string) (bool, error) {
	names, err := db.ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		return false, fmt.Errorf("failed to list collection names: %w", err)
	}

	for _, name := range names {
		if name == collName {
			return true, nil
		}
	}

	return false, nil
}
