# Trakt.tv Import Documentation

## Overview
The Trakt.tv import feature allows users to import their movie and TV show tracking data from Trakt.tv into the application using Trakt's comprehensive REST API.

## How It Works

### Authentication
- **No user authentication required** - Uses Trakt's public API
- Users only need to provide their **Trakt username**
- The system accesses public profile data and lists

### Data Sources
The import fetches data from:
1. **Watched Movies** - Movies marked as watched with ratings
2. **Watched TV Shows** - TV shows and episodes watched
3. **Watchlist** - Movies and shows planned to watch
4. **Collection** - Movies and shows in user's collection
5. **Ratings** - User ratings for movies and shows

### Import Process
1. **Username Validation** - Verifies the Trakt username exists and profile is public
2. **Data Fetching** - Retrieves comprehensive viewing data via Trakt API
3. **Content Matching** - Maps Trakt IDs/IMDB IDs to internal database entries
4. **Status Mapping** - Converts Trakt data to application statuses:
   - Watched items → `finished` status (with rating if available)
   - Watchlist items → `active` status
   - Collection items → `finished` status
   - Currently watching → `active` status
5. **Progress Import** - Imports watch progress and ratings
6. **Bulk Import** - Efficiently imports all matched entries

### API Endpoints Used
- `GET /users/{username}/watched/movies` - User's watched movies
- `GET /users/{username}/watched/shows` - User's watched TV shows
- `GET /users/{username}/watchlist` - User's watchlist
- `GET /users/{username}/collection/movies` - User's movie collection
- `GET /users/{username}/collection/shows` - User's TV show collection
- `GET /users/{username}/ratings` - User's ratings

## Usage

### Request Format
```json
{
  "trakt_username": "your_trakt_username"
}
```

### Response Format
```json
{
  "imported_count": 89,
  "skipped_count": 15,
  "error_count": 6,
  "message": "Import completed: 89 imported, 15 skipped, 6 errors",
  "imported_titles": ["The Matrix", "Breaking Bad", "Inception", "..."],
  "skipped_titles": ["Game of Thrones", "The Office", "..."]
}
```

## Requirements

### Environment Variables
```bash
TRAKT_CLIENT_ID=your_trakt_client_id_here
```

### Trakt API Setup
1. Create account at https://trakt.tv/
2. Go to https://trakt.tv/oauth/applications/new
3. Create new application to get Client ID
4. Add Client ID to environment variables

### User Requirements
- **Trakt.tv account** with public profile
- **Username** (not email) - visible in Trakt profile URL
- **Public viewing data** - private profiles cannot be imported

## Limitations

### Rate Limits
- **1000 requests per 5 minutes** for GET requests
- **200 requests per 5 minutes** for POST/PUT/DELETE requests
- Import handles rate limiting automatically with delays

### Content Availability
- Only imports content that exists in our database
- Content not in our database is counted as "error"
- Uses IMDB IDs and Trakt IDs for matching

### Profile Privacy
- User profile must be public
- Private viewing data cannot be imported
- User can make profile public temporarily for import

## Error Handling

### Common Errors
- **"Username not found"** - Invalid Trakt username
- **"Profile is private"** - User profile is private
- **"API key not configured"** - Missing TRAKT_CLIENT_ID environment variable
- **"Rate limit exceeded"** - Too many requests (handled automatically)

### Troubleshooting
1. **Verify username** - Check Trakt profile URL: `https://trakt.tv/users/USERNAME`
2. **Check privacy settings** - Ensure profile is public in Trakt settings
3. **Wait and retry** - If rate limited, wait a few minutes
4. **Contact support** - For persistent issues with valid usernames

## Technical Details

### Data Processing
- **Multiple API endpoints** for comprehensive data
- **Bulk operations** for optimal database performance
- **Memory-efficient** mapping of Trakt/IMDB IDs
- **Duplicate prevention** - Updates existing entries instead of creating duplicates
- **Comprehensive logging** for debugging and monitoring

### Status Mapping Logic
```
Trakt Data Type → Application Status
Watched         → finished
Watchlist       → active  
Collection      → finished
Currently Watching → active
```

### Rating Conversion
- **Trakt 10-point scale** → **Application 10-point scale** (direct mapping)
- **No rating** → No score in application
- **Ratings imported** with watch data

### Progress Tracking
- **Movies** → Marked as finished when watched
- **TV Shows** → Progress based on episodes watched
- **Watch dates** → Imported when available
- **Rewatch tracking** → Multiple watches recorded

### Performance Optimizations
- **Parallel API calls** for different data types
- **Bulk database operations**
- **Efficient memory usage** with maps
- **Rate limit handling** with exponential backoff
- **Concurrent processing** of movies and TV shows

## Advanced Features

### Multiple ID Matching
- **Trakt IDs** - Primary identifier
- **IMDB IDs** - Fallback matching
- **TMDB IDs** - Additional matching option
- **TVDB IDs** - For TV shows

### Watch History
- **Complete watch history** imported
- **Multiple watches** of same content tracked
- **Watch dates** preserved when available
- **Season/episode progress** for TV shows

### Collection Management
- **Collection status** imported separately from watched status
- **Physical/digital** collection data
- **Collection dates** when available

### Rating Synchronization
- **Dedicated ratings import** from ratings endpoint
- **Rating dates** when available
- **Ratings without watches** handled appropriately

## Data Types Supported

### Movies
- Watched movies with dates and ratings
- Movie watchlist
- Movie collection
- Movie ratings

### TV Shows
- Watched shows with episode progress
- Show watchlist
- Show collection
- Show and episode ratings
- Season-level progress tracking 