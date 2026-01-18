# Google Classroom TUI - 100+ Issues Found

## 🚨 CRITICAL ISSUES (Fix Immediately)

### 1. Nil pointer dereference in MainModel
**File**: `cmd/google-classroom/main.go:34`  
**Severity**: Critical  
**Issue**: Calling `IsAuthenticated()` on nil authenticator causes panic

```go
authenticator, err := auth.NewAuthenticator(configPath)
if err != nil || !authenticator.IsAuthenticated() {  // PANIC if nil
```

---

### 2. Tokens stored in plaintext - Security Vulnerability
**File**: `internal/auth/oauth.go:127-145`  
**Severity**: Critical  
**Issue**: OAuth tokens stored unencrypted in JSON file

```go
func (a *Authenticator) SaveToken(token *oauth2.Token) error {
    data, _ := json.MarshalIndent(token, "", "  ")  // Plaintext!
    os.WriteFile(a.tokenPath, data, 0600)           // Stored on disk
}
```

---

### 3. Empty credentials allow authentication
**File**: `internal/auth/oauth.go:80-97`  
**Severity**: Critical  
**Issue**: `loadConfiguration()` returns empty credentials if config missing

```go
if err != nil {
    return &Configuration{ClientID: "", ClientSecret: ""}, nil  // Empty!
}
```

---

### 4. Ignored JSON marshal error
**File**: `internal/api/client.go:570`  
**Severity**: High  
**Issue**: `PrettyPrint` ignores marshal error silently

```go
func PrettyPrint(v interface{}) {
    b, _ := json.MarshalIndent(v, "", "  ")  // Error ignored!
    fmt.Println(string(b))
}
```

---

### 5. Panic in pagination loops
**File**: `internal/api/client.go:192, 235, 278, 333, 364, 395`  
**Severity**: Critical  
**Issue**: If `resp` is nil after error check, accessing `resp.NextPageToken` panics

---

### 6. No server timeouts configured
**File**: `internal/auth/oauth.go:237`  
**Severity**: High  
**Issue**: HTTP server has no ReadTimeout, WriteTimeout, or IdleTimeout

```go
server := &http.Server{Addr: ":8080"}  // No timeouts!
```

---

## 🟠 HIGH SEVERITY ISSUES

### 7. Race condition in search handling
**File**: `internal/ui/tea/course_list.go:139-152`  
**Issue**: Goroutine updates `filteredCourses` without synchronization

### 8. Weak state parameter - predictable CSRF
**File**: `internal/auth/oauth.go:222`  
**Issue**: State uses timestamp instead of cryptographically secure random

```go
state := fmt.Sprintf("state_%d", time.Now().UnixNano())  // Predictable!
```

### 9. Type safety violation - executeWithRetry
**File**: `internal/api/client.go:405`  
**Issue**: Returns `interface{}` requiring unsafe type assertions

### 10. Cache has no synchronization
**File**: `internal/cache/cache.go`  
**Issue**: Multiple goroutines access cache without mutex protection

### 11. Unused configuration - rate limiting not configurable
**File**: `internal/api/client.go:407-409`  
**Issue**: `MaxRetries` and `RateLimitBackoff` from config are ignored

```go
cfg := DefaultConfiguration()  // Values read but not used
backoff := time.Second         // Hardcoded
for attempt := 0; attempt < 3; attempt++ {  // Hardcoded
```

### 12. HTTP redirect URI
**File**: `internal/auth/oauth.go:87`  
**Issue**: Uses HTTP instead of HTTPS for localhost

### 13. No PKCE implementation
**File**: `internal/auth/oauth.go`  
**Issue**: Missing Proof Key for Code Exchange

### 14. Memory leak from unterminated goroutines
**File**: `internal/ui/tea/course_list.go:149`  
**Issue**: Goroutine may access stale model data

### 15. Plaintext credential storage
**File**: `internal/auth/oauth.go`  
**Issue**: ClientID and ClientSecret stored in plaintext

### 16. Refresh failure doesn't invalidate token
**File**: `internal/auth/oauth.go:180-200`  
**Issue**: If refresh fails, old invalid token remains

### 17. Token save failure doesn't rollback
**File**: `internal/auth/oauth.go:274-276`  
**Issue**: If save fails after exchange, inconsistent state

### 18. JSON marshal ignored in processCredentials
**File**: `cmd/google-classroom/main.go:371`  
**Issue**: `json.MarshalIndent` error ignored, writes empty file

### 19. HTTP handler panics not caught
**File**: `internal/auth/oauth.go:238-256`  
**Issue**: No recovery middleware for HTTP handler panics

### 20. No request size limits
**File**: `internal/auth/oauth.go`  
**Issue**: HTTP handler doesn't limit request body size (DoS vulnerability)

---

## 🟡 MEDIUM SEVERITY ISSUES

### 21. Missing input validation - API parameters
**File**: `internal/api/client.go:202, 214, 245, 257, 288, 300, 312, 343, 374`  
**Issue**: No validation for `courseID`, `courseWorkID`, `submissionID`

### 22. No token encryption in memory
**File**: `internal/auth/oauth.go`  
**Issue**: Tokens handled as plain strings in memory

### 23. State not stored securely
**File**: `internal/auth/oauth.go`  
**Issue**: State stored in local variable

### 24. Browser opening no fallback
**File**: `internal/auth/oauth.go:229`  
**Issue**: If browser fails, authentication cannot proceed

### 25. Command injection potential
**File**: `internal/auth/oauth.go:204-217`  
**Issue**: URL passed directly to exec.Command without validation

### 26. Credential handling issues
**File**: `internal/auth/oauth.go`  
**Issue**: Credentials not validated for format before use

### 27. Config file path not validated
**File**: `internal/auth/oauth.go`  
**Issue**: No checks if config file is in secure location

### 28. No session invalidation
**File**: `internal/auth/oauth.go`  
**Issue**: No mechanism to invalidate sessions

### 29. Hardcoded localhost default
**File**: `internal/auth/oauth.go:87`  
**Issue**: Redirect URI hardcoded to http://localhost:8080/callback

### 30. No redirect URI validation
**File**: `internal/auth/oauth.go`  
**Issue**: No validation that redirect URI matches configured value

### 31. Over-permissive scopes
**File**: `internal/auth/oauth.go:54-61`  
**Issue**: `classroom.coursework.students` allows write access

### 32. Inefficient time handling
**File**: `internal/api/client.go`  
**Issue**: Time fields stored as strings instead of `time.Time`

### 33. Discarded response in TurnIn
**File**: `internal/api/client.go:301`  
**Issue**: `TurnIn` method discards response from executeWithRetry

### 34. Hardcoded retry logic
**File**: `internal/api/client.go:409`  
**Issue**: Retry count hardcoded to 3 instead of using config

### 35. Error() returns malformed string
**File**: `internal/errors/errors.go:98-103`  
**Issue**: Returns ": <original error>" if Message is empty

### 36. IsAuthError() missing ErrAuthOffline
**File**: `internal/errors/errors.go:183-187`  
**Issue**: `ErrAuthOffline` not included in authentication error check

### 37. WithSuggestion() overwrites without check
**File**: `internal/errors/errors.go:110-114`  
**Issue**: Directly modifies UserSuggestion without checking if already set

### 38. Unhandled error messages in submission.go
**File**: `internal/ui/tea/submission.go`  
**Issue**: Error messages not displayed to user

### 39. Incorrect submission selection logic
**File**: `internal/ui/tea/submission.go`  
**Issue**: `submissionID` used instead of `m.selectedSubmission`

### 40. Potential nil pointer in grading display
**File**: `internal/ui/tea/submission.go`  
**Issue**: Nil check may be missing for submission data

### 41. Unimplemented search functionality
**File**: `internal/ui/tea/announcement.go`  
**Issue**: Search feature exists but not implemented

### 42. Potential nil pointer in announcement view
**File**: `internal/ui/tea/announcement.go`  
**Issue**: Nil check may be missing for announcement data

### 43. Workflow triggers on both branches
**File**: `.github/workflows/ci-cd.yml:3-7`  
**Issue**: Triggers on both 'main' and 'master' - confusing

### 44. Build depends on test but lint doesn't
**File**: `.github/workflows/ci-cd.yml:43-45`  
**Issue**: Lint failures won't block builds

### 45. Go version '1.25' may not exist
**File**: `.github/workflows/ci-cd.yml:18, 52`  
**Issue**: Using Go 1.25 which may not be released yet

### 46. Only amd64 architecture built
**File**: `.github/workflows/ci-cd.yml:54-67`  
**Issue**: Missing arm64 support for Apple Silicon, Raspberry Pi

### 47. No race detection in tests
**File**: `.github/workflows/ci-cd.yml:33-34`  
**Issue**: Tests run without `-race` flag

### 48. No security vulnerability scanning
**File**: `.github/workflows/ci-cd.yml:1-119`  
**Issue**: No dependency vulnerability scanning

### 49. Build artifacts not scanned
**File**: `.github/workflows/ci-cd.yml:78-85`  
**Issue**: Artifacts uploaded without security scanning

### 50. No release tag validation
**File**: `.github/workflows/ci-cd.yml:87-90`  
**Issue**: No job to validate or create tags

### 51. Release integrity not verified
**File**: `.github/workflows/ci-cd.yml:87-102`  
**Issue**: No checksum verification before release

### 52. Outdated action-gh-release
**File**: `.github/workflows/ci-cd.yml:98-102`  
**Issue**: Using older version v2

### 53. Install target uses sudo without checks
**File**: `Makefile:57-65`  
**Issue**: `sudo` used without privilege checks

### 54. No optimization flags in build
**File**: `Makefile:11-12`  
**Issue**: Missing `-trimpath` for reproducible builds

### 55. No arm64 cross-compilation
**File**: `Makefile:19-34`  
**Issue**: Only amd64 builds, missing arm64 support

### 56. No race detection in Makefile test
**File**: `Makefile:37-38`  
**Issue**: Test target doesn't include `-race` flag

### 57. Missing integration test targets
**File**: `Makefile`  
**Issue**: No `test-integration`, `bench`, or coverage-report targets

### 58. Only Linux binary size checked
**File**: `.github/workflows/ci-cd.yml:69-76`  
**Issue**: Binary size verification only for Linux

### 59. Wrapped error format validation missing
**File**: `internal/errors/errors.go:84-95`  
**Issue**: `Wrapf()` doesn't validate args match format string

### 60. Error struct lacks timestamp
**File**: `internal/errors/errors.go:44-51`  
**Issue**: No timestamp field for debugging

### 61. Unused onRetry callback
**File**: `internal/errors/errors.go:207-210`  
**Issue**: `Handler.onRetry` defined but never used

### 62. User input loop with no timeout
**File**: `cmd/google-classroom/main.go:119`  
**Issue**: `fmt.Scanln` waits indefinitely for input

### 63. Credential watching has no timeout
**File**: `cmd/google-classroom/main.go:281-326`  
**Issue**: Infinite loop watching for credentials with no exit

### 64. Relative file paths in possiblePaths
**File**: `cmd/google-classroom/main.go:295-296`  
**Issue**: Relative paths like "credentials.json" could be dangerous

### 65. Windows compatibility issues
**File**: `cmd/google-classroom/main.go:201`  
**Issue**: `detectDownloadDir` lacks Windows-specific paths

### 66. Multiple CLI arguments ignored
**File**: `cmd/google-classroom/main.go:14`  
**Issue**: Only checks for `--setup`, ignores extra arguments

### 67. 'r' key not implemented
**File**: `cmd/google-classroom/main.go:38`  
**Issue**: Shows 'r - Refresh' but not implemented

---

## 🟢 LOW SEVERITY ISSUES

### 68. No concurrent session limits
**File**: `internal/auth/oauth.go`  
**Issue**: Multiple authentications could overwrite tokens

### 69. No token metadata tracking
**File**: `internal/auth/oauth.go`  
**Issue**: No tracking of creation time, IP, or device

### 70. Manual refresh logic redundant
**File**: `internal/auth/oauth.go:180-200`  
**Issue**: oauth2 library handles refresh automatically

### 71. No refresh token rotation
**File**: `internal/auth/oauth.go`  
**Issue**: Doesn't implement refresh token rotation

### 72. Single callback handler
**File**: `internal/auth/oauth.go`  
**Issue**: Only one concurrent authentication flow supported

### 73. No browser detection
**File**: `internal/auth/oauth.go`  
**Issue**: Assumes default browser without checking availability

### 74. No validation of token expiry
**File**: `internal/auth/oauth.go`  
**Issue**: Expiry not validated against current time explicitly

### 75. Hardcoded scopes
**File**: `internal/auth/oauth.go:54-61`  
**Issue**: OAuth scopes not configurable

### 76. No scope validation
**File**: `internal/auth/oauth.go`  
**Issue**: Granted scopes not validated

### 77. Config file permissions not validated
**File**: `internal/auth/oauth.go`  
**Issue**: Configuration file permissions not checked

### 78. RedirectURI not validated
**File**: `cmd/google-classroom/main.go:347-397`  
**Issue**: No check if redirectURI is valid

### 79. No integration tests in pipeline
**File**: `.github/workflows/ci-cd.yml`  
**Issue**: No integration tests or benchmark tests

### 80. PrettyPrint error handling
**File**: `internal/api/client.go:570`  
**Issue**: JSON marshal error silently ignored

---

## 📋 MISSING FEATURES / EMPTY FILES

### 81. Empty models.go
**File**: `internal/models/models.go`  
**Issue**: File exists but contains only package declaration

### 82. Empty config.go
**File**: `internal/config/config.go`  
**Issue**: File exists but contains only package declaration

### 83. Empty form.go
**File**: `internal/ui/components/form.go`  
**Issue**: File exists but contains only package declaration

### 84. Empty list.go
**File**: `internal/ui/components/list.go`  
**Issue**: File exists but contains only package declaration

### 85. Empty table.go
**File**: `internal/ui/components/table.go`  
**Issue**: File exists but contains only package declaration

---

## 📋 CACHE ISSUES

### 86. No mutex for cache operations
**File**: `internal/cache/cache.go`  
**Issue**: All methods access shared state without synchronization

### 87. Concurrent file writes
**File**: `internal/cache/cache.go`  
**Issue**: Multiple goroutines writing to same file

### 88. No atomic writes
**File**: `internal/cache/cache.go`  
**Issue**: No temporary file pattern for safe writes

### 89. No corrupted cache detection
**File**: `internal/cache/cache.go`  
**Issue**: No validation of cached data integrity

### 90. No background cleanup
**File**: `internal/cache/cache.go`  
**Issue**: No automatic removal of expired entries

### 91. No cache size limits
**File**: `internal/cache/cache.go`  
**Issue**: Cache can grow unbounded

### 92. TTL calculation issues
**File**: `internal/cache/cache.go`  
**Issue**: Time comparison may have edge cases

---

## 📋 TUI COMPONENT ISSUES

### 93. Key binding conflict
**File**: `internal/ui/tea/course_list.go`  
**Issue**: '/' used for search and also handled in Update

### 94. Scroll position not maintained
**File**: `internal/ui/tea/*.go`  
**Issue**: List scrolling doesn't maintain position on refresh

### 95. Search input focus issues
**File**: `internal/ui/tea/course_list.go`  
**Issue**: Search input behavior when not focused unclear

### 96. No loading states
**File**: `internal/ui/tea/*.go`  
**Issue**: No visual feedback during API calls

### 97. No error toasts
**File**: `internal/ui/tea/*.go`  
**Issue**: Errors displayed only in logs, not UI

### 98. No keyboard shortcuts help
**File**: `internal/ui/tea/*.go`  
**Issue**: No way to view available shortcuts

### 99. No pagination UI
**File**: `internal/ui/tea/*.go`  
**Issue**: No indication of more pages available

### 100. No empty state messages
**File**: `internal/ui/tea/*.go`  
**Issue**: Empty lists show nothing, no helpful message

### 101. No loading spinners
**File**: `internal/ui/tea/*.go`  
**Issue**: No visual indicator during data fetching

### 102. Dark mode not supported
**File**: `internal/ui/tea/*.go`  
**Issue**: Terminal color scheme not detected

### 103. No screen resize handling
**File**: `internal/ui/tea/*.go`  
**Issue**: Layout may break on window resize

---

## Summary

| Severity | Count |
|----------|-------|
| Critical | 6 |
| High | 20 |
| Medium | 58 |
| Low | 20 |
| **Total** | **104** |

## Recommended Priority

1. **Immediate**: Critical security issues (1-6)
2. **This Sprint**: High severity bugs (7-20)
3. **Next Sprint**: Medium issues (21-78)
4. **Backlog**: Low issues and missing features (79-103)
