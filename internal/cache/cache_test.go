package cache

import (
	"fmt"
	"testing"
	"time"
)

// TestNewCache tests creating a new cache.
func TestNewCache(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Configuration{
		Enabled:       true,
		CoursesTTL:    5 * time.Minute,
		CourseworkTTL: 1 * time.Hour,
		Directory:     tmpDir,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	if cache == nil {
		t.Fatal("Cache is nil")
	}
}

// TestCacheSetAndGet tests setting and getting cached values.
func TestCacheSetAndGet(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Configuration{
		Enabled:       true,
		CoursesTTL:    5 * time.Minute,
		CourseworkTTL: 1 * time.Hour,
		Directory:     tmpDir,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	testData := map[string]interface{}{
		"id":    "123",
		"name":  "Test Course",
		"count": 42,
	}

	// Set a value
	err = cache.Set("test_key", testData, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set cache value: %v", err)
	}

	// Get the value
	entry, err := cache.Get("test_key")
	if err != nil {
		t.Fatalf("Failed to get cache value: %v", err)
	}

	if entry == nil {
		t.Fatal("Cache entry is nil")
	}
}

// TestCacheGetMiss tests getting a non-existent value.
func TestCacheGetMiss(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Configuration{
		Enabled:       true,
		CoursesTTL:    5 * time.Minute,
		CourseworkTTL: 1 * time.Hour,
		Directory:     tmpDir,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Get a non-existent value
	entry, err := cache.Get("non_existent_key")
	if err != nil {
		t.Fatalf("Failed to get cache value: %v", err)
	}

	if entry != nil {
		t.Error("Expected nil entry for non-existent key")
	}
}

// TestCacheExpiration tests cache expiration.
func TestCacheExpiration(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Configuration{
		Enabled:       true,
		CoursesTTL:    1 * time.Second, // Very short TTL
		CourseworkTTL: 1 * time.Hour,
		Directory:     tmpDir,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	testData := map[string]interface{}{
		"id": "123",
	}

	// Set a value with very short TTL
	err = cache.Set("expiring_key", testData, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to set cache value: %v", err)
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Get the expired value
	entry, err := cache.Get("expiring_key")
	if err != nil {
		t.Fatalf("Failed to get cache value: %v", err)
	}

	if entry != nil {
		t.Error("Expected nil entry for expired key")
	}
}

// TestCacheDelete tests deleting cached values.
func TestCacheDelete(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Configuration{
		Enabled:       true,
		CoursesTTL:    5 * time.Minute,
		CourseworkTTL: 1 * time.Hour,
		Directory:     tmpDir,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	testData := map[string]interface{}{
		"id": "123",
	}

	// Set a value
	err = cache.Set("delete_key", testData, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set cache value: %v", err)
	}

	// Delete the value
	err = cache.Delete("delete_key")
	if err != nil {
		t.Fatalf("Failed to delete cache value: %v", err)
	}

	// Verify it's gone
	entry, err := cache.Get("delete_key")
	if err != nil {
		t.Fatalf("Failed to get cache value: %v", err)
	}

	if entry != nil {
		t.Error("Expected nil entry after deletion")
	}
}

// TestCacheClear tests clearing all cached values.
func TestCacheClear(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Configuration{
		Enabled:       true,
		CoursesTTL:    5 * time.Minute,
		CourseworkTTL: 1 * time.Hour,
		Directory:     tmpDir,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set multiple values
	for i := 0; i < 5; i++ {
		testData := map[string]interface{}{
			"id": i,
		}
		err = cache.Set(fmt.Sprintf("key_%d", i), testData, 5*time.Minute)
		if err != nil {
			t.Fatalf("Failed to set cache value: %v", err)
		}
	}

	// Clear all
	err = cache.Clear()
	if err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	// Verify all are gone
	for i := 0; i < 5; i++ {
		entry, err := cache.Get(fmt.Sprintf("key_%d", i))
		if err != nil {
			t.Fatalf("Failed to get cache value: %v", err)
		}
		if entry != nil {
			t.Errorf("Expected nil entry for key_%d after clear", i)
		}
	}
}

// TestCacheStats tests getting cache statistics.
func TestCacheStats(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Configuration{
		Enabled:       true,
		CoursesTTL:    5 * time.Minute,
		CourseworkTTL: 1 * time.Hour,
		Directory:     tmpDir,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set some values
	for i := 0; i < 3; i++ {
		testData := map[string]interface{}{
			"id": i,
		}
		err = cache.Set(fmt.Sprintf("stats_key_%d", i), testData, 5*time.Minute)
		if err != nil {
			t.Fatalf("Failed to set cache value: %v", err)
		}
	}

	// Get stats
	stats, err := cache.GetStats()
	if err != nil {
		t.Fatalf("Failed to get cache stats: %v", err)
	}

	if stats.TotalEntries != 3 {
		t.Errorf("Expected 3 total entries, got %d", stats.TotalEntries)
	}

	if stats.ValidEntries != 3 {
		t.Errorf("Expected 3 valid entries, got %d", stats.ValidEntries)
	}
}

// TestGenerateKey tests generating cache keys.
func TestGenerateKey(t *testing.T) {
	params := map[string]string{
		"courseId": "123",
		"userId":   "456",
	}

	key := GenerateKey("courses", params)

	// Key should contain endpoint and parameters
	if len(key) == 0 {
		t.Error("Generated key is empty")
	}
}
