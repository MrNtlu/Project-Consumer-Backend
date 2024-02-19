package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"context"
	"fmt"
	"strconv"

	p "github.com/gobeam/mongo-go-pagination"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//lint:file-ignore ST1005 Ignore all

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
	if err != nil {
		logrus.Error("failed to find preview publishing manga: ", err)

		return nil, fmt.Errorf("Failed to find preview publishing manga.")
	}

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
	if err != nil {
		logrus.Error("failed to find preview top manga: ", err)

		return nil, fmt.Errorf("Failed to find preview top manga.")
	}

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

func (mangaModel *MangaModel) GetMangaDetails(data requests.ID) (responses.Manga, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)
	malID, _ := strconv.Atoi(data.ID)

	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"_id": objectID,
			},
			bson.M{
				"mal_id": malID,
			},
		},
	}}

	set := bson.M{"$set": bson.M{
		"manga_id": bson.M{
			"$toString": "$_id",
		},
	}}

	unwindRelations := bson.M{"$unwind": bson.M{
		"path":                       "$relations",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	unwindSource := bson.M{"$unwind": bson.M{
		"path":                       "$relations.source",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	setRelation := bson.M{"$set": bson.M{
		"relations.mal_id": "$relations.source.mal_id",
		"relations.type":   "$relations.source.type",
	}}

	animeRelationLookup := bson.M{"$lookup": bson.M{
		"from": "animes",
		"let": bson.M{
			"mal_id":   "$relations.mal_id",
			"relation": "$relations.relation",
			"type":     "$relations.type",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$eq": bson.A{"$mal_id", "$$mal_id"},
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"_id": 1,
					"anime_id": bson.M{
						"$toString": "$_id",
					},
					"mal_id":         1,
					"title_en":       1,
					"title_original": 1,
					"image_url":      1,
					"relation":       "$$relation",
					"type":           "$$type",
				},
			},
		},
		"as": "anime_relation",
	}}

	mangaRelationLookup := bson.M{"$lookup": bson.M{
		"from": "mangas",
		"let": bson.M{
			"mal_id":   "$relations.mal_id",
			"relation": "$relations.relation",
			"type":     "$relations.type",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$eq": bson.A{"$mal_id", "$$mal_id"},
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"_id": 1,
					"anime_id": bson.M{
						"$toString": "$_id",
					},
					"mal_id":         1,
					"title_en":       1,
					"title_original": 1,
					"image_url":      1,
					"relation":       "$$relation",
					"type":           "$$type",
				},
			},
		},
		"as": "manga_relation",
	}}

	unwindAnimeRelation := bson.M{"$unwind": bson.M{
		"path":                       "$anime_relation",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	unwindMangaRelation := bson.M{"$unwind": bson.M{
		"path":                       "$manga_relation",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	setAnimeRelation := bson.M{"$set": bson.M{
		"anime_relation": bson.M{
			"$ifNull": bson.A{"$anime_relation", false},
		},
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$_id",
		"title_original": bson.M{
			"$first": "$title_original",
		},
		"title_en": bson.M{
			"$first": "$title_en",
		},
		"title_jp": bson.M{
			"$first": "$title_jp",
		},
		"description": bson.M{
			"$first": "$description",
		},
		"image_url": bson.M{
			"$first": "$image_url",
		},
		"mal_id": bson.M{
			"$first": "$mal_id",
		},
		"mal_score": bson.M{
			"$first": "$mal_score",
		},
		"mal_scored_by": bson.M{
			"$first": "$mal_scored_by",
		},
		"type": bson.M{
			"$first": "$type",
		},
		"chapters": bson.M{
			"$first": "$chapters",
		},
		"volumes": bson.M{
			"$first": "$volumes",
		},
		"status": bson.M{
			"$first": "$status",
		},
		"is_publishing": bson.M{
			"$first": "$is_publishing",
		},
		"published": bson.M{
			"$first": "$published",
		},
		"recommendations": bson.M{
			"$first": "$recommendations",
		},
		"serializations": bson.M{
			"$first": "$serializations",
		},
		"genres": bson.M{
			"$first": "$genres",
		},
		"themes": bson.M{
			"$first": "$themes",
		},
		"demographics": bson.M{
			"$first": "$demographics",
		},
		"relations": bson.M{
			"$addToSet": bson.M{
				"$cond": bson.A{
					bson.M{
						"$eq": bson.A{"$anime_relation", false},
					},
					"$manga_relation",
					"$anime_relation",
				},
			},
		},
		"characters": bson.M{
			"$first": "$characters",
		},
		"anime_list": bson.M{
			"$first": "$anime_list",
		},
		"watch_later": bson.M{
			"$first": "$watch_later",
		},
	}}

	cursor, err := mangaModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, unwindRelations, unwindSource, setRelation,
		animeRelationLookup, mangaRelationLookup, unwindAnimeRelation,
		unwindMangaRelation, group, setAnimeRelation,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id": data.ID,
		}).Error("failed to aggregate manga details: ", err)

		return responses.Manga{}, fmt.Errorf("Failed to aggregate manga details.")
	}

	var mangaDetails []responses.Manga
	if err = cursor.All(context.TODO(), &mangaDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"id": data.ID,
		}).Error("failed to decode anime details: ", err)

		return responses.Manga{}, fmt.Errorf("Failed to decode manga details.")
	}

	if len(mangaDetails) > 0 {
		return mangaDetails[0], nil
	}

	return responses.Manga{}, nil
}

func (mangaModel *MangaModel) GetMangaDetailsWithWatchList(data requests.ID, uuid string) (responses.MangaDetails, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)
	malID, _ := strconv.Atoi(data.ID)

	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"_id": objectID,
			},
			bson.M{
				"mal_id": malID,
			},
		},
	}}

	set := bson.M{"$set": bson.M{
		"manga_id": bson.M{
			"$toString": "$_id",
		},
	}}

	lookup := bson.M{"$lookup": bson.M{
		"from": "manga-lists",
		"let": bson.M{
			"uuid":     uuid,
			"manga_id": "$manga_id",
			"mal_id":   "$mal_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{
								"$or": bson.A{
									bson.M{"$eq": bson.A{"$manga_id", "$$manga_id"}},
									bson.M{"$eq": bson.A{"$manga_mal_id", "$$manga_id"}},
								},
							},
							bson.M{"$eq": bson.A{"$user_id", "$$uuid"}},
						},
					},
				},
			},
		},
		"as": "manga_list",
	}}

	unwindWatchList := bson.M{"$unwind": bson.M{
		"path":                       "$manga_list",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	lookupWatchLater := bson.M{"$lookup": bson.M{
		"from": "consume-laters",
		"let": bson.M{
			"uuid":     uuid,
			"manga_id": "$manga_id",
			"mal_id":   "$mal_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{
								"$or": bson.A{
									bson.M{"$eq": bson.A{"$content_id", "$$manga_id"}},
									bson.M{"$eq": bson.A{"$content_external_id", "$$mal_id"}},
								},
							},
							bson.M{"$eq": bson.A{"content_type", "manga"}},
							bson.M{"$eq": bson.A{"$user_id", "$$uuid"}},
						},
					},
				},
			},
		},
		"as": "watch_later",
	}}

	unwindWatchLater := bson.M{"$unwind": bson.M{
		"path":                       "$watch_later",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	unwindRelations := bson.M{"$unwind": bson.M{
		"path":                       "$relations",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	unwindSource := bson.M{"$unwind": bson.M{
		"path":                       "$relations.source",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	setRelation := bson.M{"$set": bson.M{
		"relations.mal_id": "$relations.source.mal_id",
		"relations.type":   "$relations.source.type",
	}}

	animeRelationLookup := bson.M{"$lookup": bson.M{
		"from": "animes",
		"let": bson.M{
			"mal_id":   "$relations.mal_id",
			"relation": "$relations.relation",
			"type":     "$relations.type",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$eq": bson.A{"$mal_id", "$$mal_id"},
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"_id": 1,
					"anime_id": bson.M{
						"$toString": "$_id",
					},
					"mal_id":         1,
					"title_en":       1,
					"title_original": 1,
					"image_url":      1,
					"relation":       "$$relation",
					"type":           "$$type",
				},
			},
		},
		"as": "anime_relation",
	}}

	mangaRelationLookup := bson.M{"$lookup": bson.M{
		"from": "mangas",
		"let": bson.M{
			"mal_id":   "$relations.mal_id",
			"relation": "$relations.relation",
			"type":     "$relations.type",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$eq": bson.A{"$mal_id", "$$mal_id"},
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"_id": 1,
					"anime_id": bson.M{
						"$toString": "$_id",
					},
					"mal_id":         1,
					"title_en":       1,
					"title_original": 1,
					"image_url":      1,
					"relation":       "$$relation",
					"type":           "$$type",
				},
			},
		},
		"as": "manga_relation",
	}}

	unwindAnimeRelation := bson.M{"$unwind": bson.M{
		"path":                       "$anime_relation",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	unwindMangaRelation := bson.M{"$unwind": bson.M{
		"path":                       "$manga_relation",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	setAnimeRelation := bson.M{"$set": bson.M{
		"anime_relation": bson.M{
			"$ifNull": bson.A{"$anime_relation", false},
		},
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$_id",
		"title_original": bson.M{
			"$first": "$title_original",
		},
		"title_en": bson.M{
			"$first": "$title_en",
		},
		"title_jp": bson.M{
			"$first": "$title_jp",
		},
		"description": bson.M{
			"$first": "$description",
		},
		"image_url": bson.M{
			"$first": "$image_url",
		},
		"mal_id": bson.M{
			"$first": "$mal_id",
		},
		"mal_score": bson.M{
			"$first": "$mal_score",
		},
		"mal_scored_by": bson.M{
			"$first": "$mal_scored_by",
		},
		"type": bson.M{
			"$first": "$type",
		},
		"chapters": bson.M{
			"$first": "$chapters",
		},
		"volumes": bson.M{
			"$first": "$volumes",
		},
		"status": bson.M{
			"$first": "$status",
		},
		"is_publishing": bson.M{
			"$first": "$is_publishing",
		},
		"published": bson.M{
			"$first": "$published",
		},
		"recommendations": bson.M{
			"$first": "$recommendations",
		},
		"serializations": bson.M{
			"$first": "$serializations",
		},
		"genres": bson.M{
			"$first": "$genres",
		},
		"themes": bson.M{
			"$first": "$themes",
		},
		"demographics": bson.M{
			"$first": "$demographics",
		},
		"relations": bson.M{
			"$addToSet": bson.M{
				"$cond": bson.A{
					bson.M{
						"$eq": bson.A{"$anime_relation", false},
					},
					"$manga_relation",
					"$anime_relation",
				},
			},
		},
		"characters": bson.M{
			"$first": "$characters",
		},
		"anime_list": bson.M{
			"$first": "$anime_list",
		},
		"watch_later": bson.M{
			"$first": "$watch_later",
		},
	}}

	cursor, err := mangaModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, lookup, unwindWatchList, lookupWatchLater, unwindWatchLater,
		unwindRelations, unwindSource, setRelation,
		animeRelationLookup, mangaRelationLookup, unwindAnimeRelation,
		unwindMangaRelation, group, setAnimeRelation,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id": data.ID,
		}).Error("failed to aggregate manga details: ", err)

		return responses.MangaDetails{}, fmt.Errorf("Failed to aggregate manga details.")
	}

	var mangaDetails []responses.MangaDetails
	if err = cursor.All(context.TODO(), &mangaDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"id": data.ID,
		}).Error("failed to decode anime details: ", err)

		return responses.MangaDetails{}, fmt.Errorf("Failed to decode manga details.")
	}

	if len(mangaDetails) > 0 {
		return mangaDetails[0], nil
	}

	return responses.MangaDetails{}, nil
}
