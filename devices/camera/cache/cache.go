package cache

import (
	"container/list"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	fileCacheDir = "./camera-images" // Directory where files will be stored
)

// Cache struct encapsulates the memory cache, file cache, and index tracking
type Cache struct {
	memoryCache     map[uint32][]byte // Memory cache (FIFO)
	memoryCacheList *list.List        // List to track the order of keys in memory cache
	memoryCacheLock sync.RWMutex      // Lock for memory cache
	fileLocks       []sync.RWMutex    // Array of lock buckets
	maxMemoryFiles  uint32            // Maximum files in memory cache
	maxFiles        uint32            // Maximum files in disk cache
	maxBuckets      uint32            // Number of lock buckets
	currentIndex    uint32            // Tracks the current file index (accessed atomically)
}

// New creates and initializes a new Cache instance
func New(maxMemoryFiles, maxFiles uint32) *Cache {
	var c Cache

	c.maxMemoryFiles = maxMemoryFiles
	c.maxFiles = maxFiles
	c.maxBuckets = (maxFiles / 50) + 1

	// Initialize the file lock buckets
	c.fileLocks = make([]sync.RWMutex, c.maxBuckets)
	for i := uint32(0); i < c.maxBuckets; i++ {
		c.fileLocks[i] = sync.RWMutex{}
	}

	c.memoryCache = make(map[uint32][]byte)
	c.memoryCacheList = list.New()

	return &c
}

// Preload reads files from the disk and loads them into memory
func (c *Cache) Preload() error {

	// Create file cache directory if it doesn't exist
	if _, err := os.Stat(fileCacheDir); os.IsNotExist(err) {
		err := os.MkdirAll(fileCacheDir, 0755)
		if err != nil {
			return fmt.Errorf("Error creating cache directory: %v", err)
		}
	}

	// List all files in the fileCacheDir
	files, err := ioutil.ReadDir(fileCacheDir)
	if err != nil {
		return fmt.Errorf("Error reading directory %s: %v", fileCacheDir, err)
	}

	// Filter out the files and get the indices along with their modified times
	type fileInfo struct {
		index    uint32
		modTime  time.Time
		filename string
	}
	var fileData []fileInfo
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".jpg" {
			var index uint32
			_, err := fmt.Sscanf(file.Name(), "%d.jpg", &index)
			if err == nil {
				fileData = append(fileData, fileInfo{
					index:    index,
					modTime:  file.ModTime(),
					filename: fmt.Sprintf("%s/%s", fileCacheDir, file.Name()),
				})
			}
		}
	}

	// Sort files by modification time (newest first)
	sort.Slice(fileData, func(i, j int) bool {
		return fileData[i].modTime.After(fileData[j].modTime)
	})

	// Set currentIndex to the index of the most recent file
	if len(fileData) > 0 {
		atomic.StoreUint32(&c.currentIndex, fileData[0].index)
	} else {
		atomic.StoreUint32(&c.currentIndex, 0)
	}

	// Add files to memory cache in the correct order (newest first), but no more than c.maxMemoryFiles
	for i, file := range fileData {
		if uint32(i) >= c.maxMemoryFiles {
			break // Stop after adding c.maxMemoryFiles files to memory cache
		}
		jpeg, err := os.ReadFile(file.filename)
		if err != nil {
			return err
		}
		c.addToMemoryCache(file.index, jpeg)
	}

	return nil
}

// GetJpeg retrieves the jpeg file from the cache, along with the previous and next file indices
func (c *Cache) GetJpeg(index uint32) ([]byte, uint32, uint32, error) {

	var curr uint32 = atomic.LoadUint32(&c.currentIndex)
	var prev, next uint32

	// If 0 is passed as the index, use the currentIndex
	if index == 0 || index == curr {
		index = curr
		prev = c.calculatePreviousIndex(index)
		next = 0
	} else {
		prev = c.calculatePreviousIndex(index)
		next = c.calculateNextIndex(index)
	}

	// If index is still 0, we're waiting for first image
	if index == 0 {
		return nil, prev, next, fmt.Errorf("Waiting on first image...")
	}

	// Try to get from memory cache
	c.memoryCacheLock.RLock()
	if jpeg, found := c.memoryCache[index]; found {
		c.memoryCacheLock.RUnlock()
		return jpeg, prev, next, nil
	}
	c.memoryCacheLock.RUnlock()

	// Read file from disc
	c.lockFile(index)
	filename := fmt.Sprintf("%s/%d.jpg", fileCacheDir, index)
	jpeg, err := os.ReadFile(filename)
	if err != nil {
		c.unlockFile(index)
		if errors.Is(err, os.ErrNotExist) {
			err = errors.New("Oops, no previous image")
		}
		return nil, prev, next, err
	}
	c.unlockFile(index)

	// Save file in cache
	c.addToMemoryCache(index, jpeg)

	return jpeg, prev, next, nil
}

func (c *Cache) calculatePreviousIndex(index uint32) uint32 {
	// If the index is 1, the previous index is c.maxFiles (wrap around)
	if index == 1 {
		return c.maxFiles
	}
	// Otherwise, just subtract 1
	return index - 1
}

func (c *Cache) calculateNextIndex(index uint32) uint32 {
	// If the index is c.maxFiles, the next index is 1 (wrap around)
	if index == c.maxFiles {
		return 1
	}
	// Otherwise, just add 1
	return index + 1
}

func (c *Cache) lockFile(index uint32) {
	bucketIndex := index % c.maxBuckets
	c.fileLocks[bucketIndex].Lock()
}

func (c *Cache) unlockFile(index uint32) {
	bucketIndex := index % c.maxBuckets
	c.fileLocks[bucketIndex].Unlock()
}

func (c *Cache) addToMemoryCache(index uint32, jpeg []byte) {

	c.memoryCacheLock.Lock()

	if uint32(len(c.memoryCache)) >= c.maxMemoryFiles {
		// Remove the oldest memory cache item if full
		oldestElement := c.memoryCacheList.Front()
		if oldestElement != nil {
			// Remove the oldest element from map and list
			delete(c.memoryCache, oldestElement.Value.(uint32))
			c.memoryCacheList.Remove(oldestElement)
		}
	}

	c.memoryCache[index] = jpeg
	c.memoryCacheList.PushBack(index)

	c.memoryCacheLock.Unlock()
}

func (c *Cache) SaveJpeg(jpeg []byte) error {

	current := atomic.LoadUint32(&c.currentIndex)

	next := current + 1
	if next > c.maxFiles {
		next = 1
	}

	c.lockFile(next)

	filename := fmt.Sprintf("%s/%d.jpg", fileCacheDir, next)
	err := os.WriteFile(filename, jpeg, 0644)
	if err != nil {
		c.unlockFile(next)
		return fmt.Errorf("failed to write JPEG file: %v", err)
	}

	c.unlockFile(next)

	c.addToMemoryCache(next, jpeg)

	atomic.StoreUint32(&c.currentIndex, next)

	return nil
}
