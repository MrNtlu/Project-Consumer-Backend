# Upstash Redis Deployment Guide

## Overview
Your search caching system has been optimized for Upstash Redis with the following specifications:
- **Storage**: 100MB free tier
- **Commands**: 10,000 daily commands 
- **Performance**: 1,000 commands/sec
- **Security**: SSL/TLS enabled

## Environment Configuration

### Set Redis URL
Add this environment variable to your Heroku app:

```bash
heroku config:set REDIS_URL="<PLACEHOLDER>"
```

### Verify Configuration
```bash
heroku config:get REDIS_URL
```

## Cache Configuration

### Optimized TTL Settings
- **Embedding Cache**: 5 days (reduced from 7 to save commands)
- **Search Cache**: 90 minutes (reduced from 2 hours)
- **Cache Stats**: 24 hours

### Memory Allocation Strategy
- **Embeddings**: ~70MB (primary cache)
- **Search Results**: ~30MB (secondary cache)

## Key Optimizations

### 1. Command Efficiency
- **Pipeline Operations**: Batch commands to reduce API calls
- **Async Caching**: Non-blocking cache writes
- **Smart Statistics**: Efficient hit/miss tracking

### 2. Memory Management
- **Compact Keys**: Short prefixes (`emb:`, `search:`)
- **Chunked Deletions**: Batch operations for cache clearing
- **Lua Scripts**: Server-side key counting

### 3. Error Handling
- **Graceful Degradation**: System works without cache
- **Connection Resilience**: Automatic reconnection
- **Timeout Management**: Prevents hanging requests

## Monitoring Endpoints

### Cache Statistics
```
GET /api/v1/search/cache/stats
```

Response includes:
- Hit/miss ratios
- Key counts
- Memory usage
- TTL settings

### Cache Management
```
DELETE /api/v1/search/cache
```

Clears all cached data efficiently.

## Daily Command Usage Estimation

### Typical Usage Breakdown
- **Cache Hits**: ~60% of requests (1 command each)
- **Cache Misses**: ~40% of requests (2 commands each - read + write)
- **Statistics**: ~50 commands/day
- **Monitoring**: ~20 commands/day

### Usage Calculation
For 1000 daily search requests:
- Cache hits: 600 × 1 = 600 commands
- Cache misses: 400 × 2 = 800 commands  
- Overhead: 70 commands
- **Total**: ~1,470 commands/day

This fits comfortably within the 10,000 daily limit.

## Performance Benefits

### Expected Improvements
- **Embedding Cache Hit**: ~95% faster (no OpenAI API call)
- **Search Cache Hit**: ~80% faster (no Pinecone + MongoDB queries)
- **Overall Response Time**: 200-500ms reduction on cache hits

### Cache Hit Rates
- **Embeddings**: 85-95% (queries repeat frequently)
- **Search Results**: 60-80% (depends on user behavior)

## Troubleshooting

### Connection Issues
1. Verify `REDIS_URL` environment variable
2. Check Upstash dashboard for connection status
3. Review application logs for Redis errors

### Memory Issues
1. Monitor cache statistics endpoint
2. Clear cache if approaching 100MB limit
3. Consider reducing TTL values if needed

### Command Limit Issues
1. Monitor daily command usage in Upstash dashboard
2. Optimize cache hit rates
3. Consider upgrading to paid tier if needed

## Deployment Steps

1. **Set Environment Variable**:
   ```bash
   heroku config:set REDIS_URL="your-upstash-url"
   ```

2. **Deploy Application**:
   ```bash
   git add .
   git commit -m "Add Upstash Redis caching"
   git push heroku master
   ```

3. **Verify Connection**:
   ```bash
   heroku logs --tail
   ```
   Look for "✅ Redis client ready" message

4. **Test Cache**:
   ```bash
   curl https://your-app.herokuapp.com/api/v1/search/cache/stats
   ```

## Success Indicators

✅ **Redis Connection**: "Redis client ready" in logs  
✅ **Cache Working**: Statistics endpoint returns data  
✅ **Performance**: Faster response times on repeated queries  
✅ **Memory Usage**: Under 100MB in Upstash dashboard  
✅ **Command Usage**: Under 10,000 daily commands  

Your caching system is now optimized for production use with Upstash Redis! 