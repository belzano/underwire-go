// Copyright (c) 2026 Guillaume Evrard
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package dal

import (
	"context"
	"errors"
	"fmt"
	"jij-service/configuration"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProfileDalMongo struct {
	collection *mongo.Collection
}

func NewProfileDalMongo(mongoConfiguration configuration.MongoConfiguration) *ProfileDalMongo {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const collectionName = "profile-components"
	client, err := mongo.Connect(options.Client().ApplyURI(mongoConfiguration.ClientUri))
	if err != nil {
		panic(err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		_ = client.Disconnect(ctx)
	}(client, ctx)

	if err := client.Ping(ctx, nil); err != nil {
		panic(err)
	}

	db := client.Database(mongoConfiguration.Database)
	collection := db.Collection(collectionName)

	return &ProfileDalMongo{
		collection: collection,
	}
}

func (r *ProfileDalMongo) Create(ctx context.Context, ProfileComponentEntity *ProfileComponentEntity) error {
	_, err := r.Retrieve(ctx, ProfileComponentEntity.Name, ProfileComponentEntity.UserID)
	if err == nil {
		return errors.New("already exists")
	}

	_, err = r.collection.InsertOne(ctx, ProfileComponentEntity)
	if err != nil {
		return fmt.Errorf("creation failure : %v", err)
	}
	return nil
}

func (r *ProfileDalMongo) Retrieve(ctx context.Context, nom, userID string) (*ProfileComponentEntity, error) {
	filter := bson.M{
		"nom":    nom,
		"userID": userID,
	}

	var ProfileComponentEntity ProfileComponentEntity
	err := r.collection.FindOne(ctx, filter).Decode(&ProfileComponentEntity)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("mot found")
		}
		return nil, fmt.Errorf("retrieval failure : %v", err)
	}
	return &ProfileComponentEntity, nil
}

func (r *ProfileDalMongo) Update(ctx context.Context, ProfileComponentEntity *ProfileComponentEntity) error {
	filter := bson.M{
		"nom":    ProfileComponentEntity.Name,
		"userID": ProfileComponentEntity.UserID,
	}

	update := bson.M{
		"$set": bson.M{
			"data": ProfileComponentEntity.Data,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("update failure : %v", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *ProfileDalMongo) Delete(ctx context.Context, nom, userID string) error {
	filter := bson.M{
		"nom":    nom,
		"userID": userID,
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("deletion failure : %v", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("not found")
	}
	return nil
}
