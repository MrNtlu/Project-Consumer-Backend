package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Achievement struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Title       string             `bson:"title" json:"title"`
	ImageURL    string             `bson:"image_url" json:"image_url"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGO_ATLAS_URI")))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(context.TODO())

	// Get achievements collection
	db := client.Database("project-consumer") // Replace with your database name if different
	collection := db.Collection("achievements")

	// Define achievements to add
	achievements := []Achievement{
		{
			Title:       "First Steps",
			ImageURL:    "https://example.com/images/first-steps.png",
			Description: "Complete your first activity on Watchlistfy",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Reviewer",
			ImageURL:    "https://example.com/images/reviewer.png",
			Description: "Write your first review",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Critic",
			ImageURL:    "https://example.com/images/critic.png",
			Description: "Write 10 reviews",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Silver Screen Scout",
			ImageURL:    "https://example.com/images/silver-screen-scout.png",
			Description: "Watched 10 different movies. You're getting a taste of cinema!",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Reel Explorer",
			ImageURL:    "https://example.com/images/reel-explorer.png",
			Description: "Watched 25 different movies. Your watchlist is looking impressive!",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Cinema Sage",
			ImageURL:    "https://example.com/images/cinema-sage.png",
			Description: "Watched 50 different movies. The silver screen bows to your wisdom. 🎬",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Pilot Hunter",
			ImageURL:    "https://example.com/images/pilot-hunter.png",
			Description: "Watched 10 different TV series. You're tuning into the multiverse!",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Seasoned Binger",
			ImageURL:    "https://example.com/images/seasoned-binger.png",
			Description: "Watched 25 different series. Cliffhangers can't stop you.",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Episodic Legend",
			ImageURL:    "https://example.com/images/episodic-legend.png",
			Description: "Watched 50 different series. You've conquered the binge realm! 📺",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Rookie Adventurer",
			ImageURL:    "https://example.com/images/rookie-adventurer.png",
			Description: "Played 10 different games. The quest begins!",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Pixel Challenger",
			ImageURL:    "https://example.com/images/pixel-challenger.png",
			Description: "Played 25 different games. Your XP bar is growing.",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Digital Conqueror",
			ImageURL:    "https://example.com/images/digital-conqueror.png",
			Description: "Played 50 different games. Boss battles fear you! 🎮",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Opening Act",
			ImageURL:    "https://example.com/images/shonen-starter.png",
			Description: "Watched 10 different anime series. You've taken your first step on the otaku path.",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Story Arc Wanderer",
			ImageURL:    "https://example.com/images/seinen-seeker.png",
			Description: "Watched 25 different anime series. A true anime explorer.",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Final Episode Veteran",
			ImageURL:    "https://example.com/images/anime-archon.png",
			Description: "Watched 50 different anime series. You are one with the anime world! 🌸",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Devoted Soul",
			ImageURL:    "https://example.com/images/devoted-soul.png",
			Description: "Finished the same content 10 times. Whether comfort or obsession, you meant it. 🔁",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Future Watcher",
			ImageURL:    "https://example.com/images/future-watcher.png",
			Description: "Added 10 items to Watch Later. So many stories await you.",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Content Collector",
			ImageURL:    "https://example.com/images/content-collector.png",
			Description: "Added 25 items to Watch Later. Planning your viewing like a strategist!",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Title:       "Archiver of Anticipation",
			ImageURL:    "https://example.com/images/archiver-of-anticipation.png",
			Description: "Added 50 items to Watch Later. You've curated a treasure trove of tales! 📜",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}

	// Convert to interface slice for insertion
	var docs []interface{}
	for _, achievement := range achievements {
		docs = append(docs, achievement)
	}

	// Insert achievements
	result, err := collection.InsertMany(context.TODO(), docs)
	if err != nil {
		log.Fatal("Failed to insert achievements:", err)
	}

	fmt.Printf("Successfully inserted %d achievements!\n", len(result.InsertedIDs))
	fmt.Println("Achievement IDs:")
	for i, id := range result.InsertedIDs {
		fmt.Printf("- %s: %v\n", achievements[i].Title, id)
	}
}
