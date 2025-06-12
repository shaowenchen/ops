package log

import (
	"os"
	"sync"
	"testing"
	"time"
)

// BenchmarkAsyncLogger tests the performance of async logger
func BenchmarkAsyncLogger(b *testing.B) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "benchmark-log")
	if err != nil {
		b.Fatal("Failed to create temp file:", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Redirect output to the temporary file
	originalStd := Std
	originalFile := File
	Std = tmpFile
	File = tmpFile
	defer func() {
		Std = originalStd
		File = originalFile
	}()

	// Test the async logger (default)
	b.Run("AsyncLogger", func(b *testing.B) {
		logger := NewLogger().SetStd().SetFile().Build()
		defer logger.Close()

		b.ResetTimer()
		benchmarkLogger(b, logger)
	})

	// Test the sync logger for comparison
	b.Run("SyncLogger", func(b *testing.B) {
		logger := NewLogger().SetStd().SetFile().SetSync().Build()
		defer logger.Close()

		b.ResetTimer()
		benchmarkLogger(b, logger)
	})
}

// benchmarkLogger performs the actual benchmark
func benchmarkLogger(b *testing.B, logger *Logger) {
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			logger.Infof("This is benchmark log message #%d", counter)
			counter++
		}
	})
}

// TestHighConcurrencyLogging tests the logger under high concurrency
func TestHighConcurrencyLogging(t *testing.T) {
	// Redirect output to discard
	originalStd := Std
	originalFile := File
	Std = nil
	File = nil
	defer func() {
		Std = originalStd
		File = originalFile
	}()

	// Create loggers
	asyncLogger := NewLogger().Build() // Default is async
	syncLogger := NewLogger().SetSync().Build()

	defer asyncLogger.Close()

	const (
		numGoroutines     = 100
		numLogsPerRoutine = 1000
	)

	runTest := func(logger *Logger, name string) time.Duration {
		var wg sync.WaitGroup
		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()
				for j := 0; j < numLogsPerRoutine; j++ {
					logger.Infof("Goroutine %d: Log message %d", routineID, j)
				}
			}(i)
		}

		wg.Wait()
		if name == "AsyncLogger" {
			// Ensure all async logs are processed
			asyncLogger.Close()
		}
		elapsed := time.Since(start)

		t.Logf("%s completed %d logs in %v (%.2f logs/sec)",
			name,
			numGoroutines*numLogsPerRoutine,
			elapsed,
			float64(numGoroutines*numLogsPerRoutine)/elapsed.Seconds())

		return elapsed
	}

	// Run tests
	syncTime := runTest(syncLogger, "SyncLogger")

	// Reset async logger for fair comparison
	asyncLogger = NewLogger().Build()
	asyncTime := runTest(asyncLogger, "AsyncLogger")

	speedup := float64(syncTime) / float64(asyncTime)
	t.Logf("Async logger is %.2fx faster than sync logger", speedup)
}

// TestChannelCapacity tests what happens when the channel reaches capacity
func TestChannelCapacity(t *testing.T) {
	// Redirect output to discard to focus on channel behavior
	originalStd := Std
	originalFile := File
	devNull, _ := os.Open("/dev/null")
	Std = devNull
	File = devNull
	defer func() {
		Std = originalStd
		File = originalFile
		devNull.Close()
	}()

	// Create a logger with a smaller channel for easier testing
	smallChannelLogger := &Logger{
		Level:    LevelInfo,
		Flag:     0,
		Instant:  true,
		async:    true,
		msgChan:  make(chan logMessage, 10), // Small channel capacity
		stopChan: make(chan struct{}),
	}
	smallChannelLogger.SetStd().Build()
	defer smallChannelLogger.Close()

	// Create a channel blocker by pausing the processing goroutine
	pauseChan := make(chan struct{})
	go func() {
		// Block the channel consumer
		<-pauseChan
	}()

	// Fill the channel to capacity and then some more
	for i := 0; i < 20; i++ {
		if i == 10 {
			t.Log("Channel should be full now, logs should be written synchronously")
		}
		// Log will be sync after channel is full
		smallChannelLogger.Infof("Test message %d", i)
	}

	// Release the processing goroutine
	close(pauseChan)

	// Let async processing catch up
	time.Sleep(100 * time.Millisecond)

	t.Log("Channel capacity test completed")
}
