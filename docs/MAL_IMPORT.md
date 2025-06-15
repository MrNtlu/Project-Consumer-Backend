# MyAnimeList Import Feature

This **premium-only** feature allows users to import their anime lists from MyAnimeList (MAL) into Watchlistfy.

## Premium Requirement

⭐ **This feature is only available to premium users.** Users must have an active premium subscription to use the MAL import functionality.

## How it works

The import feature uses MyAnimeList's public JSON endpoint to fetch user anime lists. This endpoint doesn't require authentication and can access public anime lists. The import process has been optimized for speed using bulk database operations.

## API Endpoint

**POST** `/api/v1/import/mal`

### Authentication

Requires Bearer token authentication.

### Request Body

```json
{
  "username": "your_mal_username"
}
```

### Response

```json
{
  "message": "MyAnimeList import completed successfully.",
  "data": {
    "imported_count": 150,
    "skipped_count": 5,
    "error_count": 2,
    "message": "Import completed: 150 imported, 5 skipped, 2 errors"
  }
}
```

## How to use

1. **Make sure your MAL list is public**: The user's MyAnimeList anime list must be set to public for the import to work.

2. **Send a POST request** to the import endpoint with your MAL username:

   ```bash
   curl -X POST "https://your-api-domain.com/api/v1/import/mal" \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"username": "your_mal_username"}'
   ```

3. **Check the response** for import statistics:
   - `imported_count`: Number of anime successfully imported
   - `skipped_count`: Number of anime skipped (already in your list)
   - `error_count`: Number of anime that failed to import

## Status Mapping

The import feature maps MyAnimeList statuses to Watchlistfy statuses:

| MAL Status | Watchlistfy Status |
|------------|-------------------|
| Currently Watching (1) | active |
| Completed (2) | finished |
| On-Hold (3) | active |
| Dropped (4) | dropped |
| Plan to Watch (6) | active |

## What gets imported

- Anime title and MAL ID
- Watch status
- Number of watched episodes
- User score (if rated)
- Times finished (defaults to 1 for completed anime)

## Limitations

1. **Public lists only**: Private MAL lists cannot be imported
2. **Anime must exist in database**: Only anime that already exist in the Watchlistfy database will be imported
3. **No duplicates**: Anime already in your Watchlistfy list will be skipped
4. **Rate limiting**: The import respects MAL's rate limits

## Error Handling

Common errors and solutions:

- **"user not found or anime list is private"**: Make sure the username is correct and the anime list is set to public
- **"Failed to fetch anime list"**: Network issues or MAL is temporarily unavailable
- **High error count**: Many anime in your MAL list might not exist in the Watchlistfy database yet

## Performance Optimizations

The import has been heavily optimized for speed:

- **Bulk Database Operations**: Uses MongoDB's `InsertMany` for batch inserts
- **In-Memory Lookups**: Pre-loads existing entries and anime MAL IDs into memory maps
- **Single Query Processing**: Eliminates individual database calls per anime entry
- **Efficient Filtering**: Skips duplicates and missing anime without database queries

**Expected Performance:**

- Small lists (1-100 entries): ~5-15 seconds
- Medium lists (100-500 entries): ~15-45 seconds
- Large lists (500-1000+ entries): ~45-120 seconds

## Notes

- The import process is now significantly faster than before (previously took 4+ minutes)
- Anime not found in the Watchlistfy database will be logged and skipped
- You can run the import multiple times - existing entries will be skipped
