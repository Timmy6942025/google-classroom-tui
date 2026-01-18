// Package cache provides file-based caching for API responses.
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Cache provides file-based caching for API responses.
type Cache struct {
	directory     string
	coursesTTL    time.Duration
	courseworkTTL time.Duration
}

// Configuration holds cache configuration.
type Configuration struct {
	Enabled       bool
	CoursesTTL    time.Duration
	CourseworkTTL time.Duration
	Directory     string
}

// DefaultConfiguration returns the default cache configuration.
func DefaultConfiguration() *Configuration {
	homeDir, _ := os.UserHomeDir()
	return &Configuration{
		Enabled:       true,
		CoursesTTL:    5 * time.Minute,
		CourseworkTTL: 1 * time.Hour,
		Directory:     filepath.Join(homeDir, ".cache", "google-classroom"),
	}
}

// CacheEntry represents a cached entry.
type CacheEntry struct {
	Data      json.RawMessage `json:"data"`
	CachedAt  time.Time       `json:"cached_at"`
	ExpiresAt time.Time       `json:"expires_at"`
}

// NewCache creates a new cache instance.
func NewCache(cfg *Configuration) (*Cache, error) {
	if cfg == nil {
		cfg = DefaultConfiguration()
	}

	// Ensure directory exists
	if err := os.MkdirAll(cfg.Directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{
		directory:     cfg.Directory,
		coursesTTL:    cfg.CoursesTTL,
		courseworkTTL: cfg.CourseworkTTL,
	}, nil
}

// Get retrieves a cached value.
func (c *Cache) Get(key string) (*CacheEntry, error) {
	path := c.getPath(key)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse cache entry: %w", err)
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		// Clean up expired entry
		os.Remove(path)
		return nil, nil // Cache miss (expired)
	}

	return &entry, nil
}

// Set stores a value in the cache.
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	path := c.getPath(key)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Marshal data
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Create entry
	now := time.Now()
	entry := CacheEntry{
		Data:      jsonData,
		CachedAt:  now,
		ExpiresAt: now.Add(ttl),
	}

	// Write to file
	jsonBytes, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	if err := os.WriteFile(path, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// Delete removes a cached value.
func (c *Cache) Delete(key string) error {
	path := c.getPath(key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache: %w", err)
	}
	return nil
}

// Clear removes all cached values.
func (c *Cache) Clear() error {
	entries, err := os.ReadDir(c.directory)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(c.directory, entry.Name())
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to delete %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// Stats returns cache statistics.
type CacheStats struct {
	TotalEntries   int
	ValidEntries   int
	ExpiredEntries int
	TotalSize      int64
}

// GetStats returns cache statistics.
func (c *Cache) GetStats() (*CacheStats, error) {
	stats := &CacheStats{}

	entries, err := os.ReadDir(c.directory)
	if err != nil {
		if os.IsNotExist(err) {
			return stats, nil
		}
		return nil, fmt.Errorf("failed to read cache directory: %w", err)
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		stats.TotalEntries++

		path := filepath.Join(c.directory, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		info, _ := os.Stat(path)
		stats.TotalSize += info.Size()

		var cacheEntry CacheEntry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			continue
		}

		if now.After(cacheEntry.ExpiresAt) {
			stats.ExpiredEntries++
		} else {
			stats.ValidEntries++
		}
	}

	return stats, nil
}

// GenerateKey generates a cache key from endpoint and parameters.
func GenerateKey(endpoint string, params map[string]string) string {
	var parts []string
	parts = append(parts, endpoint)

	for key, value := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(parts, "&")
}

// getPath returns the file path for a cache key.
func (c *Cache) getPath(key string) string {
	// Sanitize key for file system
	safeKey := strings.ReplaceAll(key, "/", "_")
	safeKey = strings.ReplaceAll(safeKey, ":", "_")
	safeKey = strings.ReplaceAll(safeKey, " ", "_")
	return filepath.Join(c.directory, safeKey+".json")
}

// GetCoursesTTL returns the TTL for courses.
func (c *Cache) GetCoursesTTL() time.Duration {
	return c.coursesTTL
}

// GetCourseworkTTL returns the TTL for coursework.
func (c *Cache) GetCourseworkTTL() time.Duration {
	return c.courseworkTTL
}
