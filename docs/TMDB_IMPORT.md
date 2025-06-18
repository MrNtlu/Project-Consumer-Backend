# TMDB Import Documentation

## Overview
The TMDB import feature allows users to import their watchlists and rated movies/TV shows from The Movie Database (TMDB) into the application.

## How It Works

### Authentication
- **No authentication required** - Uses TMDB's public API
- Users only need to provide their **TMDB username**
- The system accesses public profile data and lists

### Data Sources
The import fetches data from:
1. **User's Watchlist** - Movies and TV shows marked "want to watch"
2. **User's Rated Items** - Movies and TV shows the user has rated
3. **User's Favorite Items** - Movies and TV shows marked as favorites

### Import Process
1. **Username Validation** - Verifies the TMDB username exists
2. **Data Fetching** - Retrieves watchlist, ratings, and favorites via TMDB API
3. **Content Matching** - Maps TMDB IDs to internal database entries
4. **Status Mapping** - Converts TMDB data to application statuses:
   - Watchlist items → `active` status
   - Rated items → `finished` status (with rating)
   - Favorites → `finished` status (high rating if no rating exists)
5. **Bulk Import** - Efficiently imports all matched entries

### API Endpoints Used
- `GET /3/account/{account_id}/watchlist/movies` - User's movie watchlist
- `GET /3/account/{account_id}/watchlist/tv` - User's TV watchlist
- `GET /3/account/{account_id}/rated/movies` - User's rated movies
- `GET /3/account/{account_id}/rated/tv` - User's rated TV shows
- `GET /3/account/{account_id}/favorite/movies` - User's favorite movies
- `GET /3/account/{account_id}/favorite/tv` - User's favorite TV shows

## Usage

### Request Format
```json
{
  "tmdb_username": "your_tmdb_username"
}
```

### Response Format
```json
{
  "imported_count": 45,
  "skipped_count": 12,
  "error_count": 3,
  "message": "Import completed: 45 imported, 12 skipped, 3 errors",
  "imported_titles": ["The Matrix", "Breaking Bad", "..."],
  "skipped_titles": ["Inception", "Game of Thrones", "..."]
}
```

## Requirements

### Environment Variables
```bash
TMDB_API_KEY=your_tmdb_api_key_here
```

### TMDB API Key Setup
1. Create account at https://www.themoviedb.org/
2. Go to Settings → API
3. Request API key (free)
4. Add to environment variables

### User Requirements
- **TMDB account** with public profile
- **Username** (not email) - visible in TMDB profile URL
- **Public lists** - private lists cannot be accessed

## Limitations

### Rate Limits
- **40 requests per 10 seconds** (TMDB API limit)
- Import handles rate limiting automatically with delays

### Content Availability
- Only imports content that exists in our database
- Content not in our database is counted as "error"
- TMDB content must have matching TMDB ID in our database

### Profile Privacy
- User profile must be public
- Private watchlists/ratings cannot be imported
- User can make profile public temporarily for import

## Error Handling

### Common Errors
- **"Username not found"** - Invalid TMDB username
- **"Profile is private"** - User profile/lists are private
- **"API key not configured"** - Missing TMDB_API_KEY environment variable
- **"Rate limit exceeded"** - Too many requests (handled automatically)

### Troubleshooting
1. **Verify username** - Check TMDB profile URL: `https://www.themoviedb.org/u/USERNAME`
2. **Check privacy settings** - Ensure profile and lists are public
3. **Wait and retry** - If rate limited, wait a few minutes
4. **Contact support** - For persistent issues with valid usernames

## Technical Details

### Data Processing
- **Bulk operations** for optimal performance
- **Memory-efficient** mapping of TMDB IDs
- **Duplicate prevention** - Updates existing entries instead of creating duplicates
- **Comprehensive logging** for debugging and monitoring

### Status Mapping Logic
```
TMDB Watchlist → active
TMDB Rated (>= 7.0) → finished (with rating)
TMDB Rated (< 7.0) → finished (with rating)
TMDB Favorites → finished (rating: 9.0 if no existing rating)
```

### Performance Optimizations
- **Parallel API calls** where possible
- **Bulk database operations**
- **Efficient memory usage** with maps
- **Rate limit handling** with exponential backoff 