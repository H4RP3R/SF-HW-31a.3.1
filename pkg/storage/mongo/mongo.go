package mongo

import (
	"GoNews/pkg/storage"
	"context"

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

	// Create the posts collection in advance.
	err = s.client.Database(s.dbName).CreateCollection(context.Background(), "posts")
	if err != nil {
		return nil, err
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
		"title":       post.Title,
		"content":     post.Content,
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

func (s *Store) CreateUniqueIndexOnID() error {
	collection := s.client.Database(s.dbName).Collection("posts")
	_, err := collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return err
}
