package manga

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Manga chapter repository operations are implemented below.

// MongoRepository implements the Repository interface using MongoDB
type MongoRepository struct {
	db  *mongo.Database
	col *mongo.Collection
	log *zerolog.Logger
}

// NewMongoRepository creates a new MongoRepository instance
func NewMongoRepository(db *mongo.Database, log *zerolog.Logger) *MongoRepository {
	if log == nil {
		l := zerolog.Nop()
		log = &l
	}
	return &MongoRepository{
		db:  db,
		col: db.Collection("manga"),
		log: log,
	}
}

func (r *MongoRepository) Create(manga *Manga) error {
	if manga.ID.IsZero() {
		manga.ID = primitive.NewObjectID()
	}

	_, err := r.col.InsertOne(context.Background(), manga)
	if err != nil {
		r.log.Error().
			Str("title", manga.Title).
			Err(err).
			Msg("failed to create manga")
		return err
	}

	return nil
}

func (r *MongoRepository) Update(manga *Manga) error {
	filter := bson.M{"_id": manga.ID}
	_, err := r.col.ReplaceOne(context.Background(), filter, manga)
	if err != nil {
		r.log.Error().
			Str("id", manga.ID.Hex()).
			Str("title", manga.Title).
			Err(err).
			Msg("failed to update manga")
		return err
	}

	return nil
}

func (r *MongoRepository) Delete(id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.col.DeleteOne(context.Background(), filter)
	if err != nil {
		r.log.Error().
			Str("id", id.Hex()).
			Err(err).
			Msg("failed to delete manga")
		return err
	}

	return nil
}

func (r *MongoRepository) FindByID(id primitive.ObjectID) (*Manga, error) {
	var manga Manga
	filter := bson.M{"_id": id}
	err := r.col.FindOne(context.Background(), filter).Decode(&manga)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		r.log.Error().
			Str("id", id.Hex()).
			Err(err).
			Msg("failed to find manga by id")
		return nil, err
	}

	return &manga, nil
}

func (r *MongoRepository) FindAll(page, limit int) ([]*Manga, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := r.col.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		r.log.Error().
			Int("page", page).
			Int("limit", limit).
			Err(err).
			Msg("failed to find manga")
		return nil, err
	}
	defer cursor.Close(context.Background())

	var mangas []*Manga
	if err := cursor.All(context.Background(), &mangas); err != nil {
		r.log.Error().
			Int("page", page).
			Int("limit", limit).
			Err(err).
			Msg("failed to decode manga results")
		return nil, err
	}

	return mangas, nil
}

func (r *MongoRepository) Search(query string, page, limit int) ([]*Manga, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"createdAt": -1})

	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"alternateTitles": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	cursor, err := r.col.Find(context.Background(), filter, opts)
	if err != nil {
		r.log.Error().
			Str("query", query).
			Int("page", page).
			Int("limit", limit).
			Err(err).
			Msg("failed to search manga")
		return nil, err
	}
	defer cursor.Close(context.Background())

	var mangas []*Manga
	if err := cursor.All(context.Background(), &mangas); err != nil {
		r.log.Error().
			Str("query", query).
			Int("page", page).
			Int("limit", limit).
			Err(err).
			Msg("failed to decode manga search results")
		return nil, err
	}

	return mangas, nil
}

func (r *MongoRepository) CreateMangaChapter(mangaID primitive.ObjectID, chapter *MangaChapter) error {
	filter := bson.M{"_id": mangaID}
	update := bson.M{
		"$push": bson.M{
			"chapters": chapter,
		},
		"$set": bson.M{
			"updatedAt": chapter.CreatedAt,
		},
	}

	_, err := r.col.UpdateOne(context.Background(), filter, update)
	if err != nil {
		r.log.Error().
			Str("mangaId", mangaID.Hex()).
			Int("chapterNumber", chapter.Number).
			Err(err).
			Msg("failed to add chapter")
		return err
	}

	return nil
}

func (r *MongoRepository) UpdateMangaChapter(mangaID primitive.ObjectID, chapter *MangaChapter) error {
	filter := bson.M{
		"_id":             mangaID,
		"chapters.number": chapter.Number,
	}
	update := bson.M{
		"$set": bson.M{
			"chapters.$": chapter,
			"updatedAt":  chapter.UpdatedAt,
		},
	}

	_, err := r.col.UpdateOne(context.Background(), filter, update)
	if err != nil {
		r.log.Error().
			Str("mangaId", mangaID.Hex()).
			Int("chapterNumber", chapter.Number).
			Err(err).
			Msg("failed to update chapter")
		return err
	}

	return nil
}

func (r *MongoRepository) DeleteMangaChapter(mangaID primitive.ObjectID, chapterNumber int) error {
	filter := bson.M{"_id": mangaID}
	update := bson.M{
		"$pull": bson.M{
			"chapters": bson.M{
				"number": chapterNumber,
			},
		},
		"$set": bson.M{
			"updatedAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	_, err := r.col.UpdateOne(context.Background(), filter, update)
	if err != nil {
		r.log.Error().
			Str("mangaId", mangaID.Hex()).
			Int("chapterNumber", chapterNumber).
			Err(err).
			Msg("failed to delete chapter")
		return err
	}

	return nil
}
