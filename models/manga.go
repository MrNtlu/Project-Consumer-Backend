package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"context"
	"fmt"

	p "github.com/gobeam/mongo-go-pagination"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MangaModel struct {
	Collection *mongo.Collection
}

func NewMangaModel(mongoDB *db.MongoDB) *MangaModel {
	return &MangaModel{
		Collection: mongoDB.Database.Collection("mangas"),
	}
}

const (
	mangaSearchLimit     = 50
	mangaPaginationLimit = 40
)

//TODO Endpoints
// Increase premium threshhold for comics and manga
// Sort/Filter status, demographics, themes, genres,
// Currently Publishing
// Top Rated
// Popularity
// Extract and save to mobile -> status, genres, demographics and themes

func (mangaModel *MangaModel) GetPreviewCurrentlyPublishingManga() ([]responses.PreviewManga, error) {
	match := bson.M{
		"status": "Publishing",
	}

	opts := options.Find().SetSort(bson.M{"mal_score": -1}).SetLimit(mangaPaginationLimit)

	cursor, err := mangaModel.Collection.Find(context.TODO(), match, opts)

	var results []responses.PreviewManga
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode preview publishing manga: ", err)

		return nil, fmt.Errorf("Failed to decode preview publishing manga.")
	}

	return results, nil
}

func (mangaModel *MangaModel) GetPreviewTopManga() ([]responses.PreviewManga, error) {
	filter := bson.D{}
	opts := options.Find().SetSort(bson.M{"mal_score": -1}).SetLimit(mangaPaginationLimit)

	cursor, err := mangaModel.Collection.Find(context.TODO(), filter, opts)

	var results []responses.PreviewManga
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode preview top manga: ", err)

		return nil, fmt.Errorf("Failed to decode preview top manga.")
	}

	return results, nil
}

func (mangaModel *MangaModel) GetPreviewPopularManga() ([]responses.PreviewManga, error) {
	addFields := bson.M{"$addFields": bson.M{
		"popularity": bson.M{
			"$multiply": bson.A{
				"$mal_score", "$mal_scored_by",
			},
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"popularity": -1,
	}}

	limit := bson.D{{"$limit", mangaPaginationLimit}}

	cursor, err := mangaModel.Collection.Aggregate(context.TODO(), bson.A{
		addFields, sort, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate popular preview anime: ", err)

		return nil, fmt.Errorf("Failed to aggregate preview popular animes.")
	}

	var results []responses.PreviewManga
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode preview top manga: ", err)

		return nil, fmt.Errorf("Failed to decode preview top manga.")
	}

	return results, nil
}

func (mangaModel *MangaModel) GetMangaBySortAndFilter(data requests.SortFilterManga) ([]responses.Manga, p.PaginationData, error) {
	addFields := bson.M{"$addFields": bson.M{
		"popularity": bson.M{
			"$multiply": bson.A{
				"$mal_score", "$mal_scored_by",
			},
		},
	}}

	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "top":
		sortType = "mal_score"
		sortOrder = -1
	case "popularity":
		sortType = "popularity"
		sortOrder = -1
	case "new":
		sortType = "published.from"
		sortOrder = -1
	case "old":
		sortType = "published.from"
		sortOrder = 1
	}

	matchFields := bson.M{}
	if data.Status != nil || data.Genres != nil || data.Demographics != nil || data.Themes != nil {

		if data.Status != nil {

			var status string
			switch *data.Status {
			case "publishing":
				status = "Publishing"
			case "discontinued":
				status = "Discontinued"
			case "finished":
				status = "Finished"
			}

			matchFields["status"] = status
		}

		if data.Genres != nil {
			matchFields["genres.name"] = bson.M{
				"$in": bson.A{data.Genres},
			}
		}

		if data.Demographics != nil {
			matchFields["demographics.name"] = bson.M{
				"$in": bson.A{data.Demographics},
			}
		}

		if data.Themes != nil {
			matchFields["themes.name"] = bson.M{
				"$in": bson.A{data.Themes},
			}
		}
	}

	match := bson.M{"$match": matchFields}

	paginatedData, err := p.New(mangaModel.Collection).Context(context.TODO()).Limit(mangaPaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Aggregate(match, addFields)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
			"match":   match,
		}).Error("failed to aggregate manga by sort and filter: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get manga by selected filters.")
	}

	var mangas []responses.Manga
	for _, raw := range paginatedData.Data {
		var manga *responses.Manga
		if marshalErr := bson.Unmarshal(raw, &manga); marshalErr == nil {
			mangas = append(mangas, *manga)
		}
	}

	return mangas, paginatedData.Pagination, nil
}
