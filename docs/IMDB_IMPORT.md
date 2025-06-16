# IMDB Import Feature

The IMDB import feature allows users to automatically import their watchlists from IMDB into the application. This feature supports both user watchlists and public lists.

## How It Works

The system can extract watchlist data from IMDB using two methods:

### 1. User Watchlist Import

- **URL Format**: `https://www.imdb.com/user/ur[USER_ID]/watchlist`
- **Example**: `https://www.imdb.com/user/ur1000000/watchlist`
- **User ID**: The numeric ID that appears in your IMDB profile URL (e.g., `ur1000000`)

### 2. Public List Import

- **URL Format**: `https://www.imdb.com/list/ls[LIST_ID]/`
- **Example**: `https://www.imdb.com/list/ls021970607/`
- **List ID**: The alphanumeric ID of any public IMDB list (e.g., `ls021970607`)

## API Endpoint

**POST** `/api/v1/import/imdb`

### Headers

```
Authorization: Bearer <your_jwt_token>
Content-Type: application/json
```

### Request Body

```json
{
  "imdb_user_id": "ur1000000",     // Optional: For user watchlists
  "imdb_list_id": "ls021970607"    // Optional: For public lists
}
```

**Note**: Provide either `imdb_user_id` OR `imdb_list_id`, not both.

### Response

```json
{
  "message": "IMDB import completed successfully.",
  "data": {
    "imported_count": 25,
    "skipped_count": 5,
    "error_count": 2,
    "message": "Import completed: 25 imported, 5 skipped, 2 errors"
  }
}
```

## How to Find Your IMDB User ID

1. Go to your IMDB profile page
2. Look at the URL: `https://www.imdb.com/user/ur1234567/`
3. Your User ID is the part after `/user/` (e.g., `ur1234567`)

## How to Find a List ID

1. Navigate to any public IMDB list
2. Look at the URL: `https://www.imdb.com/list/ls021970607/`
3. The List ID is the part after `/list/` (e.g., `ls021970607`)

## What Gets Imported

The system imports:

- **Movies** and **TV Series** from your watchlist
- **User ratings** (if available)
- **Titles** and **years**
- All items are added with status "planning" by default

## Requirements

- **Public Lists**: Your watchlist must be public for the import to work
- **Database Match**: Only content that exists in our database will be imported

## Import Behavior

- **New Items**: Added to your list with "planning" status
- **Existing Items**: Updated with new rating information (if available)
- **Missing Content**: Items not in our database are skipped and counted as errors
- **Bulk Processing**: All imports are processed efficiently in batches

## Error Handling

The system handles various error scenarios:

- Private or non-existent lists
- Network timeouts
- Invalid IMDB IDs
- Content not found in database

## Privacy & Security

- No IMDB credentials are stored
- Only public watchlist data is accessed
- All data is processed securely and stored according to our privacy policy

## Limitations

- Only works with public IMDB lists/watchlists
- Requires content to exist in our database
- Rate limited to prevent abuse

## Support

If you encounter issues with IMDB import:

1. Ensure your watchlist is public
2. Verify your User ID or List ID is correct
3. Contact support if problems persist
