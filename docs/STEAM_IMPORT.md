# Steam Import Feature

This document explains how to set up and use the Steam import functionality to import user game libraries from Steam.

## Prerequisites

### 1. Steam API Key

- Visit: <https://steamcommunity.com/dev/apikey>
- Register for a Steam API key (requires Steam Guard Mobile Authenticator)
- Add the API key to your environment variables as `STEAM_API_KEY`

### 2. User Requirements

- Steam profile must be **public**
- Game details must be **public**
- Valid Steam ID (64-bit Steam ID)

## Setup

### Environment Variables

```bash
STEAM_API_KEY=your_steam_api_key_here
```

### Database Requirements

- Games collection must have `steam_appid` field populated
- This field maps Steam App IDs to your game database entries

## API Endpoint

### Import Steam Library

```json
POST /api/v1/import/steam
```

**Headers:**

```json
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body (Option 1 - Steam ID):**

```json
{
  "steam_id": "76561198000000000"
}
```

**Request Body (Option 2 - Steam Username):**

```json
{
  "steam_username": "your_steam_username"
}
```

**Response:**

```json
{
  "message": "Steam import completed successfully.",
  "data": {
    "imported_count": 45,
    "skipped_count": 12,
    "error_count": 3,
    "message": "Import completed: 45 imported, 12 skipped, 3 errors"
  }
}
```

## How It Works

### 1. Steam ID Resolution (if username provided)

- Converts Steam username to 64-bit Steam ID using Steam Web API
- Validates the username exists and is accessible

### 2. Profile Validation

- Validates Steam ID exists
- Checks if profile is public
- Verifies game details are accessible

### 3. Game Library Fetch

- Retrieves owned games using Steam Web API
- Includes playtime data and last played timestamps
- Fetches both paid and free games

### 4. Data Processing

- Maps Steam App IDs to database game entries
- Converts playtime from minutes to hours
- Determines game status based on playtime:
  - `planning`: Never played (0 minutes)
  - `active`: Any playtime > 0 minutes

### 5. Database Import

- Bulk inserts new game entries
- **Updates existing entries** with latest playtime data
- Logs detailed import statistics

## Data Imported

For each game, the following data is imported:

- **Game ID**: Mapped from Steam App ID to database game ID
- **Status**: Determined by playtime (planning/active)
- **Hours Played**: Converted from Steam's minute format
- **Times Finished**: Estimated based on playtime (20+ hours = 1 completion)
- **Timestamps**: Creation and update times

## Error Handling

### Common Errors

1. **"Steam API key not configured"**
   - Solution: Set `STEAM_API_KEY` environment variable

2. **"Steam profile is private"**
   - Solution: User must set profile and game details to public

3. **"Steam profile not found or invalid Steam ID"**
   - Solution: Verify the Steam ID is correct (64-bit format)

4. **"Steam import is a premium feature"**
   - Solution: User must have premium subscription

### Import Statistics

- **Imported**: Successfully added new games to user's list
- **Skipped**: Games that already existed and were updated with new playtime
- **Errors**: Games not found in database or other issues

## Finding Your Steam ID

### Method 1: Steam Profile URL

If your profile URL is: `https://steamcommunity.com/profiles/76561198000000000/`
Your Steam ID is: `76561198000000000`

### Method 2: Custom URL

If you have a custom URL like: `https://steamcommunity.com/id/username/`
Use online Steam ID converters to get your 64-bit Steam ID

### Method 3: Steam Client

1. Open Steam client
2. Go to View → Settings → Interface
3. Check "Display Steam URL address bar when available"
4. Visit your profile, the URL will show your Steam ID

## Rate Limits

Steam Web API has the following limits:

- **100,000 requests per day** per API key
- **1 request per second** recommended
- Import process respects these limits automatically

## Premium Feature

Steam import is a **premium feature** and requires:

- Active premium subscription
- Valid payment status
- Premium features enabled for the user account

## Troubleshooting

### Import Issues

1. **No games imported**: Check if games exist in your database with `steam_appid` field
2. **Partial import**: Some games may not be in your database yet
3. **API errors**: Verify Steam API key and network connectivity

### Profile Issues

1. **Private profile**: Set Steam profile to public
2. **Private game details**: Set game details to public in privacy settings
3. **Invalid Steam ID**: Use 64-bit Steam ID format

## Security Notes

- Steam API key should be kept secure
- API key has access to public Steam data only
- No sensitive user data is accessed
- Import only works with public profiles
