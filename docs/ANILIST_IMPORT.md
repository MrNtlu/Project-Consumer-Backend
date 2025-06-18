# AniList Import Documentation

## Overview
The AniList import feature allows users to import their anime and manga lists from AniList into the application using AniList's powerful GraphQL API.

## How It Works

### Authentication
- **No authentication required** - Uses AniList's public GraphQL API
- Users only need to provide their **AniList username**
- The system accesses public profile data and media lists

### Data Sources
The import fetches data from:
1. **Anime List** - All anime entries (watching, completed, planning, etc.)
2. **Manga List** - All manga entries (reading, completed, planning, etc.)
3. **User Ratings** - Scores given to anime/manga
4. **Progress Tracking** - Episodes watched, chapters read

### Import Process
1. **Username Validation** - Verifies the AniList username exists
2. **GraphQL Queries** - Fetches comprehensive media lists via AniList API
3. **Content Matching** - Maps AniList IDs to internal database entries
4. **Status Mapping** - Converts AniList statuses to application statuses:
   - `CURRENT` → `active`
   - `COMPLETED` → `finished`
   - `PLANNING` → `active`
   - `PAUSED` → `active`
   - `DROPPED` → `dropped`
   - `REPEATING` → `finished`
5. **Progress Import** - Imports watch/read progress and ratings
6. **Bulk Import** - Efficiently imports all matched entries

### GraphQL Queries Used
```graphql
query ($username: String) {
  User(name: $username) {
    id
    mediaListOptions {
      animeList { ... }
      mangaList { ... }
    }
  }
  
  MediaListCollection(userName: $username, type: ANIME) {
    lists {
      entries {
        id
        status
        score
        progress
        progressVolumes
        repeat
        media {
          id
          title { romaji english }
          episodes
          chapters
        }
      }
    }
  }
}
```

## Usage

### Request Format
```json
{
  "anilist_username": "your_anilist_username"
}
```

### Response Format
```json
{
  "imported_count": 150,
  "skipped_count": 25,
  "error_count": 8,
  "message": "Import completed: 150 imported, 25 skipped, 8 errors",
  "imported_titles": ["Attack on Titan", "Death Note", "One Piece", "..."],
  "skipped_titles": ["Naruto", "Bleach", "..."]
}
```

## Requirements

### User Requirements
- **AniList account** with public profile
- **Username** (not email) - visible in AniList profile URL
- **Public media lists** - private lists cannot be accessed

### No API Key Required
- AniList's GraphQL API is public and doesn't require authentication for public data
- No environment variables needed for basic functionality

## Limitations

### Rate Limits
- **90 requests per minute** (AniList API limit)
- Import handles rate limiting automatically with delays
- Large lists may take several minutes to import

### Content Availability
- Only imports content that exists in our database
- Content not in our database is counted as "error"
- AniList content must have matching AniList ID in our database

### Profile Privacy
- User profile must be public
- Private media lists cannot be imported
- User can make profile public temporarily for import

## Error Handling

### Common Errors
- **"Username not found"** - Invalid AniList username
- **"Profile is private"** - User profile/lists are private
- **"Rate limit exceeded"** - Too many requests (handled automatically)
- **"GraphQL error"** - API response issues

### Troubleshooting
1. **Verify username** - Check AniList profile URL: `https://anilist.co/user/USERNAME`
2. **Check privacy settings** - Ensure profile and media lists are public
3. **Wait and retry** - If rate limited, wait a few minutes
4. **Contact support** - For persistent issues with valid usernames

## Technical Details

### Data Processing
- **GraphQL optimization** for efficient data fetching
- **Bulk operations** for optimal database performance
- **Memory-efficient** mapping of AniList IDs
- **Duplicate prevention** - Updates existing entries instead of creating duplicates
- **Comprehensive logging** for debugging and monitoring

### Status Mapping Logic
```
AniList Status → Application Status
CURRENT        → active
COMPLETED      → finished
PLANNING       → active
PAUSED         → active
DROPPED        → dropped
REPEATING      → finished
```

### Score Conversion
- **AniList 10-point scale** → **Application 10-point scale**
- **AniList 100-point scale** → **Application 10-point scale** (divided by 10)
- **AniList 5-star scale** → **Application 10-point scale** (multiplied by 2)

### Progress Tracking
- **Episodes watched** → Stored in application
- **Chapters read** → Stored for manga entries
- **Rewatch count** → Times finished tracking
- **Start/End dates** → Imported when available

### Performance Optimizations
- **Single GraphQL query** for all user data
- **Bulk database operations**
- **Efficient memory usage** with maps
- **Rate limit handling** with exponential backoff
- **Concurrent processing** of anime and manga lists

## Advanced Features

### Custom Lists Support
- Imports from custom AniList lists if public
- Maintains list organization where possible
- Maps custom statuses to closest application equivalent

### Score Format Detection
- Automatically detects user's preferred scoring system
- Converts scores appropriately to application format
- Handles users with no scoring system

### Progress Synchronization
- Imports current episode/chapter progress
- Tracks completion dates
- Preserves rewatch/reread counts 