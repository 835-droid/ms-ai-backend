# Authentication Fix Summary

## Problem
The logs showed authentication issues:
1. **401 Unauthorized** on `POST /api/novels` even after token refresh
2. **"secondary update refresh token failed error='user not found'"** - indicating sync issues between MongoDB (primary) and PostgreSQL (secondary)
3. **429 Too Many Requests** on `/api/admin/metrics` - rate limiting was too restrictive

## Root Cause Analysis

### Issue 1: Database Sync Problem
The system uses a hybrid repository pattern with:
- **Primary**: MongoDB (working correctly)
- **Secondary**: PostgreSQL (failing to sync)

When `UpdateRefreshToken` was called:
1. MongoDB successfully updated the refresh token
2. PostgreSQL failed with "user not found" because the user wasn't synced
3. The error was logged but the operation continued, creating an inconsistent state

### Issue 2: Admin Role Requirement (FIXED)
**FIXED**: Previously `POST /api/novels` required **admin role**, but this has been changed. Now any authenticated user can create novels, similar to manga creation.

If you're still getting 401 errors after the fix, check:
1. Are you logged in? (The endpoint requires authentication)
2. Is your token valid? (Try logging out and back in)
3. Is the `is_active` field set to `true`?

### Issue 3: Rate Limiting
Admin endpoints had very restrictive rate limits (2% of configured rate).

## Solution

### 1. Fixed Hybrid Repository Sync (`hybrid_user_repository.go`)

Added automatic user synchronization when `UpdateRefreshToken` fails on the secondary database:

```go
func (r *HybridUserRepository) UpdateRefreshToken(ctx context.Context, userID primitive.ObjectID, token string, expiresAt primitive.DateTime) error {
    // ... primary update ...
    
    if r.secondary != nil {
        if err := r.secondary.UpdateRefreshToken(ctx, userID, token, expiresAt); err != nil {
            r.log.Warn("secondary update refresh token failed - user may not be synced", ...)
            // Try to sync the user to secondary if the error is "user not found"
            if errors.Is(err, core.ErrUserNotFound) {
                r.syncUserToSecondary(ctx, userID)
            }
        }
    }
    return nil
}

// syncUserToSecondary attempts to sync a user from primary to secondary repository
func (r *HybridUserRepository) syncUserToSecondary(ctx context.Context, userID primitive.ObjectID) {
    // Fetch user from primary
    user, err := r.primary.FindByID(ctx, userID)
    // ... create/update user in secondary ...
}
```

### 2. Adjusted Admin Rate Limiting (`admin_routes.go`)

Increased admin rate limit from 2% to 5% of configured rate:
- Before: 2 requests/minute (with 100 req/min config)
- After: 5 requests/minute (with 100 req/min config)

This prevents 429 errors on admin dashboard endpoints like `/api/admin/metrics`.

## Expected Behavior After Fix

1. **Token Refresh**: When a refresh token is updated, the system will automatically sync the user to PostgreSQL if they don't exist there
2. **Authentication**: Subsequent requests with the new token will succeed because both databases are in sync
3. **Admin Dashboard**: Rate limiting is more permissive, allowing normal dashboard usage without 429 errors

## Testing Recommendations

1. **Test Token Refresh Flow**:
   - Log in as a user
   - Wait for access token to expire
   - Make a request that triggers refresh
   - Verify the new tokens work correctly

2. **Test Admin Endpoints**:
   - Access `/api/admin/metrics` multiple times
   - Verify no 429 errors occur during normal usage

3. **Check Logs**:
   - Look for "syncUserToSecondary: user created/updated in secondary" messages
   - Verify no more "user not found" errors during token refresh

## Files Modified

1. `MS-AI/internal/data/user/hybrid_user_repository.go` - Added sync logic
2. `MS-AI/internal/api/router/admin/admin_routes.go` - Increased rate limit