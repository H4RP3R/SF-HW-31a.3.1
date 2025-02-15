package mongo

import (
	"GoNews/pkg/storage"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
		return err
	}

	return nil
}

func (s *Store) Posts() ([]storage.Post, error) {
	collection := s.client.Database(s.dbName).Collection("posts")
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	var posts []storage.Post
	for cur.Next(context.Background()) {
		var p storage.Post
		err := cur.Decode(&p)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, cur.Err()
}

func (s *Store) UpdatePost(post storage.Post) error {
	collection := s.client.Database(s.dbName).Collection("posts")
	filter := bson.D{{Key: "id", Value: post.ID}}
	update := bson.D{{Key: "$set", Value: bson.M{
		"title":       post.Title,
		"content":     post.Content,
		"AuthorID":    post.AuthorID,
		"AuthorName":  post.AuthorName,
		"CreatedAt":   post.CreatedAt,
		"PublishedAt": post.PublishedAt,
	}}}
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return storage.ErrEntryNotExist
	}

	return nil
}

func (s *Store) DeletePost(post storage.Post) error {
	collection := s.client.Database(s.dbName).Collection("posts")
	filter := bson.D{{Key: "id", Value: post.ID}}
	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return storage.ErrEntryNotExist
	}

	return nil
}
