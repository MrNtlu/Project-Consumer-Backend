package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"app/utils"
	"context"
	"fmt"
	"strconv"

	p "github.com/gobeam/mongo-go-pagination"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TVModel struct {
	Collection *mongo.Collection
}

func NewTVModel(mongoDB *db.MongoDB) *TVModel {
	return &TVModel{
		Collection: mongoDB.Database.Collection("tv-series"),
	}
}

const (
	tvSeriesUpcomingPaginationLimit = 10
	tvSeriesPaginationLimit         = 15
)

/* TODO Endpoints
* [x] Get upcoming tv series by popularity etc.
* [x] Get tv series by release date, popularity, genre etc. (sort & filter)
* [ ] Get tv series details
* [x] Get top tv series by every decade 1980's 1990's etc.
* [x] Get top tv series by every genre
 */

func (tvModel *TVModel) GetUpcomingTVSeries(data requests.SortUpcoming) ([]responses.TVSeries, p.PaginationData, error) {
	var (
		sortType        string
		sortOrder       int8
		hasAirDateOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "tmdb_popularity"
		sortOrder = -1
		hasAirDateOrder = -1
	case "soon":
		sortType = "first_air_date"
		sortOrder = 1
		hasAirDateOrder = -1
	case "later":
		sortType = "first_air_date"
		sortOrder = -1
		hasAirDateOrder = 1
	}

	match := bson.M{"$match": bson.M{
		"status": "In Production",
	}}

	addFields := bson.M{"$addFields": bson.M{
		"has_air_date": bson.M{
			"$or": bson.A{
				bson.M{
					"$ne": bson.A{"$release_date", ""},
				},
				bson.M{
					"$ne": bson.A{"$release_date", nil},
				},
			},
		},
	}}

	paginatedData, err := p.New(tvModel.Collection).Context(context.TODO()).Limit(tvSeriesUpcomingPaginationLimit).
		Page(data.Page).Sort("has_air_date", hasAirDateOrder).Sort(sortType, sortOrder).Sort("_id", 1).Aggregate(match, addFields)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate upcoming tv series: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get upcoming tv series.")
	}

	var upcomingTVSeries []responses.TVSeries
	for _, raw := range paginatedData.Data {
		var tvSeries *responses.TVSeries
		if marshalErr := bson.Unmarshal(raw, &tvSeries); marshalErr == nil {
			upcomingTVSeries = append(upcomingTVSeries, *tvSeries)
		}
	}

	return upcomingTVSeries, paginatedData.Pagination, nil
}

//TODO Get the latest season and sort by that.
func (tvModel *TVModel) GetUpcomingSeasonTVSeries(data requests.SortUpcoming) ([]responses.TVSeries, p.PaginationData, error) {
	var (
		sortType        string
		sortOrder       int8
		hasAirDateOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "tmdb_popularity"
		sortOrder = -1
		hasAirDateOrder = -1
	case "soon":
		sortType = "release_date"
		sortOrder = 1
		hasAirDateOrder = -1
	case "later":
		sortType = "release_date"
		sortOrder = -1
		hasAirDateOrder = 1
	}

	match := bson.M{"$match": bson.M{
		"seasons.air_date": bson.M{
			"$gte": utils.GetCurrentDate(),
		},
	}}

	addFields := bson.M{"$addFields": bson.M{
		"has_release_date": bson.M{
			"$or": bson.A{
				bson.M{
					"$ne": bson.A{"$release_date", ""},
				},
				bson.M{
					"$ne": bson.A{"$release_date", nil},
				},
			},
		},
	}}

	paginatedData, err := p.New(tvModel.Collection).Context(context.TODO()).Limit(tvSeriesUpcomingPaginationLimit).
		Page(data.Page).Sort("has_air_date", hasAirDateOrder).Sort(sortType, sortOrder).Sort("_id", 1).Aggregate(match, addFields)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate upcoming tv series: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get upcoming tv series.")
	}

	var upcomingTVSeries []responses.TVSeries
	for _, raw := range paginatedData.Data {
		var tvSeries *responses.TVSeries
		if marshalErr := bson.Unmarshal(raw, &tvSeries); marshalErr == nil {
			upcomingTVSeries = append(upcomingTVSeries, *tvSeries)
		}
	}

	return upcomingTVSeries, paginatedData.Pagination, nil
}

func (tvModel *TVModel) GetTVSeriesBySortAndFilter(data requests.SortFilterTVSeries) ([]responses.TVSeries, p.PaginationData, error) {
	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "tmdb_popularity"
		sortOrder = -1
	case "soon":
		sortType = "first_air_date"
		sortOrder = 1
	case "later":
		sortType = "first_air_date"
		sortOrder = -1
	}

	match := bson.M{}
	if data.Status != nil || data.Genres != nil || data.ProductionCompanies != nil ||
		data.FirstAirDateFrom != nil || data.NumSeason != nil {

		if data.Status != nil {
			var status string
			switch *data.Status {
			case "production":
				status = "In Production"
			case "airing":
				status = "Returning Series"
			case "ended":
				status = "Ended"
			}

			match["status"] = status
		}

		if data.Genres != nil {
			match["genres.name"] = bson.M{
				"$in": bson.A{data.Genres},
			}
		}

		if data.ProductionCompanies != nil {
			match["production_companies.name"] = bson.M{
				"$in": bson.A{data.ProductionCompanies},
			}
		}

		if data.FirstAirDateFrom != nil {
			if data.FirstAirDateTo != nil {
				match["first_air_date"] = bson.M{
					"$gte": strconv.Itoa(*data.FirstAirDateFrom),
					"$lt":  strconv.Itoa(*data.FirstAirDateTo),
				}
			} else {
				match["first_air_date"] = bson.M{
					"$gte": strconv.Itoa(*data.FirstAirDateFrom),
				}
			}
		}

		if data.NumSeason != nil {
			match["total_seasons"] = bson.M{
				"$gte": *data.NumSeason,
			}
		}
	}

	var tvSeries []responses.TVSeries
	paginatedData, err := p.New(tvModel.Collection).Context(context.TODO()).Limit(tvSeriesPaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Filter(match).Decode(&tvSeries).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate tv series by sort and filter: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get tv series by selected filters.")
	}

	return tvSeries, paginatedData.Pagination, nil
}

func (tvModel *TVModel) GetTVSeriesDetails(data requests.ID) (responses.TVSeries, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)

	result := tvModel.Collection.FindOne(context.TODO(), bson.M{
		"_id": objectID,
	})

	var tvSeries responses.TVSeries
	if err := result.Decode(&tvSeries); err != nil {
		logrus.WithFields(logrus.Fields{
			"game_id": data.ID,
		}).Error("failed to find tv series details by id: ", err)

		return responses.TVSeries{}, fmt.Errorf("Failed to find tv series by id.")
	}

	return tvSeries, nil
}
