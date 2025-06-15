# Achievement Management Scripts

This directory contains scripts to help manage achievements in your Watchlistfy application.

## Add Achievements Script

The `add_achievements.go` script allows you to easily add achievements to your MongoDB database without using MongoDB Compass.

### Prerequisites

1. Make sure you have Go installed on your system
2. Ensure your `.env` file is properly configured with `MONGO_ATLAS_URI`
3. The script assumes your database name is "watchlistfy" - modify if different

### Usage

1. Navigate to the scripts directory:

   ```bash
   cd scripts
   ```

2. Run the script:

   ```bash
   go run add_achievements.go
   ```

### What the script does

The script will add the following achievements to your database:

- **First Steps**: Complete your first activity on Watchlistfy
- **Reviewer**: Write your first review
- **Critic**: Write 10 reviews
- **Silver Screen Scout**: Watched 10 different movies. You're getting a taste of cinema!
- **Reel Explorer**: Watched 25 different movies. Your watchlist is looking impressive!
- **Cinema Sage**: Watched 50 different movies. The silver screen bows to your wisdom. 🎬
- **Pilot Hunter**: Watched 10 different TV series. You're tuning into the multiverse!
- **Seasoned Binger**: Watched 25 different series. Cliffhangers can't stop you.
- **Episodic Legend**: Watched 50 different series. You've conquered the binge realm! 📺
- **Rookie Adventurer**: Played 10 different games. The quest begins!
- **Pixel Challenger**: Played 25 different games. Your XP bar is growing.
- **Digital Conqueror**: Played 50 different games. Boss battles fear you! 🎮
- **Opening Act**: Watched 10 different anime series. You've taken your first step on the otaku path.
- **Story Arc Wanderer**: Watched 25 different anime series. A true anime explorer.
- **Final Episode Veteran**: Watched 50 different anime series. You are one with the anime world! 🌸
- **Devoted Soul**: Finished the same content 10 times. Whether comfort or obsession, you meant it. 🔁
- **Future Watcher**: Added 10 items to Watch Later. So many stories await you.
- **Content Collector**: Added 25 items to Watch Later. Planning your viewing like a strategist!
- **Archiver of Anticipation**: Added 50 items to Watch Later. You've curated a treasure trove of tales! 📜

### Customizing achievements

To add your own achievements, modify the `achievements` slice in the script:

```go
achievements := []Achievement{
    {
        Title:       "Your Achievement Title",
        ImageURL:    "https://your-domain.com/images/achievement.png",
        Description: "Description of what the user needs to do",
        CreatedAt:   time.Now().UTC(),
        UpdatedAt:   time.Now().UTC(),
    },
    // Add more achievements here...
}
```

### Notes

- The script will create new achievements each time it's run - it doesn't check for duplicates
- Make sure to update the image URLs to point to your actual achievement images
- The database name is hardcoded as "watchlistfy" - change it in the script if your database has a different name
