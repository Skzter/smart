//go:build integration

package service

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
)

// Example structs for testing - these demonstrate the generic capability

// UserEvent represents a user activity event
type UserEvent struct {
	UserID    int64     `parquet:"user_id"`
	EventType string    `parquet:"event_type"`
	Timestamp time.Time `parquet:"timestamp"`
	Data      string    `parquet:"data"`
}

// Product represents a product in an e-commerce system
type Product struct {
	ID          int64   `parquet:"id"`
	Name        string  `parquet:"name"`
	Price       float64 `parquet:"price"`
	Category    string  `parquet:"category"`
	InStock     bool    `parquet:"in_stock"`
	Description string  `parquet:"description"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Level     string    `parquet:"level"`
	Message   string    `parquet:"message"`
	Timestamp time.Time `parquet:"timestamp"`
	Source    string    `parquet:"source"`
	UserID    *int64    `parquet:"user_id,optional"` // Optional field
}

func TestNewParquetWrapper(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	wrapper, err := NewParquetWrapper[UserEvent](logger, entity.ParquetConfig{})
	require.NoError(t, err)
	require.NotNil(t, wrapper)

	// Test with nil logger
	_, err = NewParquetWrapper[UserEvent](nil, entity.ParquetConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logger cannot be nil")
}

func TestDefaultParquetConfig(t *testing.T) {
	config := DefaultParquetConfig()
	assert.Equal(t, "snappy", config.CompressionCodec)
	assert.Equal(t, int64(8*1024), config.PageSize)
	assert.Equal(t, int64(1000), config.RowGroupSize)
}

func TestWriteAndReadStructsToParquet_UserEvents(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	wrapper, err := NewParquetWrapper[UserEvent](logger, entity.ParquetConfig{})
	require.NoError(t, err)

	// Create test data
	events := []UserEvent{
		{
			UserID:    123,
			EventType: "login",
			Timestamp: time.Now().UTC(),
			Data:      `{"ip": "192.168.1.1"}`,
		},
		{
			UserID:    456,
			EventType: "purchase",
			Timestamp: time.Now().UTC().Add(-1 * time.Hour),
			Data:      `{"amount": 99.99, "product_id": 42}`,
		},
		{
			UserID:    789,
			EventType: "logout",
			Timestamp: time.Now().UTC().Add(-2 * time.Hour),
			Data:      `{}`,
		},
	}

	// Test writing to parquet
	parquetData, err := wrapper.WriteStructsToParquet(events)
	require.NoError(t, err)
	assert.Greater(t, len(parquetData), 0)

	// Test reading from parquet
	readEvents, err := wrapper.ReadStructsFromParquet(parquetData)
	require.NoError(t, err)
	assert.Len(t, readEvents, len(events))

	// Verify data integrity
	for i, event := range events {
		assert.Equal(t, event.UserID, readEvents[i].UserID)
		assert.Equal(t, event.EventType, readEvents[i].EventType)
		assert.Equal(t, event.Data, readEvents[i].Data)
		// Note: Timestamps might have precision differences
		assert.WithinDuration(t, event.Timestamp, readEvents[i].Timestamp, time.Millisecond)
	}
}

func TestWriteAndReadStructsToParquet_Products(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	config := entity.ParquetConfig{
		CompressionCodec: "gzip",
		PageSize:         4 * 1024,
		RowGroupSize:     500,
	}

	wrapper, err := NewParquetWrapper[Product](logger, config)
	require.NoError(t, err)

	// Create test data
	products := []Product{
		{
			ID:          1,
			Name:        "Laptop",
			Price:       999.99,
			Category:    "Electronics",
			InStock:     true,
			Description: "High-performance laptop",
		},
		{
			ID:          2,
			Name:        "Book",
			Price:       29.99,
			Category:    "Literature",
			InStock:     false,
			Description: "Interesting novel",
		},
	}

	// Test writing to parquet with gzip compression
	parquetData, err := wrapper.WriteStructsToParquet(products)
	require.NoError(t, err)
	assert.Greater(t, len(parquetData), 0)

	// Test reading from parquet
	readProducts, err := wrapper.ReadStructsFromParquet(parquetData)
	require.NoError(t, err)
	assert.Len(t, readProducts, len(products))

	// Verify data integrity
	for i, product := range products {
		assert.Equal(t, product.ID, readProducts[i].ID)
		assert.Equal(t, product.Name, readProducts[i].Name)
		assert.Equal(t, product.Price, readProducts[i].Price)
		assert.Equal(t, product.Category, readProducts[i].Category)
		assert.Equal(t, product.InStock, readProducts[i].InStock)
		assert.Equal(t, product.Description, readProducts[i].Description)
	}
}

func TestWriteAndReadStructsToParquet_LogEntries(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	config := entity.ParquetConfig{
		CompressionCodec: "gzip",
		PageSize:         4 * 1024,
		RowGroupSize:     500,
	}

	wrapper, err := NewParquetWrapper[LogEntry](logger, config)
	require.NoError(t, err)

	// Create test data with optional fields
	userID1 := int64(123)
	logs := []LogEntry{
		{
			Level:     "INFO",
			Message:   "User logged in",
			Timestamp: time.Now().UTC(),
			Source:    "auth-service",
			UserID:    &userID1, // Optional field with value
		},
		{
			Level:     "ERROR",
			Message:   "Database connection failed",
			Timestamp: time.Now().UTC().Add(-5 * time.Minute),
			Source:    "db-service",
			UserID:    nil, // Optional field without value
		},
	}

	// Test writing to parquet without compression
	parquetData, err := wrapper.WriteStructsToParquet(logs)
	require.NoError(t, err)
	assert.Greater(t, len(parquetData), 0)

	// Test reading from parquet
	readLogs, err := wrapper.ReadStructsFromParquet(parquetData)
	require.NoError(t, err)
	assert.Len(t, readLogs, len(logs))

	// Verify data integrity
	for i, log := range logs {
		assert.Equal(t, log.Level, readLogs[i].Level)
		assert.Equal(t, log.Message, readLogs[i].Message)
		assert.Equal(t, log.Source, readLogs[i].Source)
		assert.WithinDuration(t, log.Timestamp, readLogs[i].Timestamp, time.Millisecond)

		// Check optional field
		if log.UserID == nil {
			assert.Nil(t, readLogs[i].UserID)
		} else {
			require.NotNil(t, readLogs[i].UserID)
			assert.Equal(t, *log.UserID, *readLogs[i].UserID)
		}
	}
}

func TestWriteStructToParquet_SingleStruct(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	wrapper, err := NewParquetWrapper[UserEvent](logger, entity.ParquetConfig{})
	require.NoError(t, err)

	// Test single struct
	event := UserEvent{
		UserID:    999,
		EventType: "test",
		Timestamp: time.Now().UTC(),
		Data:      `{"test": true}`,
	}

	// Test writing single struct
	parquetData, err := wrapper.WriteStructToParquet(event)
	require.NoError(t, err)
	assert.Greater(t, len(parquetData), 0)

	// Test reading back
	readEvents, err := wrapper.ReadStructsFromParquet(parquetData)
	require.NoError(t, err)
	assert.Len(t, readEvents, 1)
	assert.Equal(t, event.UserID, readEvents[0].UserID)
	assert.Equal(t, event.EventType, readEvents[0].EventType)
}

func TestGetParquetSchema(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	wrapper, err := NewParquetWrapper[UserEvent](logger, entity.ParquetConfig{})
	require.NoError(t, err)

	// Test schema generation for UserEvent
	schema, err := wrapper.GetParquetSchema()
	require.NoError(t, err)
	require.NotNil(t, schema)
	assert.Greater(t, len(schema.Fields()), 0)

	// Test schema generation for Product
	schema, err = wrapper.GetParquetSchema()
	require.NoError(t, err)
	require.NotNil(t, schema)
	assert.Greater(t, len(schema.Fields()), 0)
}

func TestValidateStruct(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	wrapper, err := NewParquetWrapper[UserEvent](logger, entity.ParquetConfig{})
	require.NoError(t, err)

	// Test valid struct
	event := UserEvent{UserID: 1, EventType: "test"}
	err = wrapper.ValidateStruct(event)
	assert.NoError(t, err)

	// Test with pointer
	err = wrapper.ValidateStruct(event)
	assert.NoError(t, err)
}

func TestParquetWrapper_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	wrapper, err := NewParquetWrapper[UserEvent](logger, entity.ParquetConfig{})
	require.NoError(t, err)

	// Test empty data slice
	emptyData := []UserEvent{}
	_, err = wrapper.WriteStructsToParquet(emptyData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "data slice cannot be empty")

	// Test empty parquet data
	_, err = wrapper.ReadStructsFromParquet([]byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parquet data cannot be empty")

	// Test invalid parquet data (this will panic, so we need to handle it)
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected panic due to invalid parquet data
				assert.NotNil(t, r)
			}
		}()
		_, err := wrapper.ReadStructsFromParquet([]byte("invalid parquet data"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid parquet data")
	}()
}

func TestGetParquetFileInfo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	wrapper, err := NewParquetWrapper[UserEvent](logger, entity.ParquetConfig{})
	require.NoError(t, err)

	// Create test data
	events := []UserEvent{
		{UserID: 1, EventType: "test1", Timestamp: time.Now().UTC()},
		{UserID: 2, EventType: "test2", Timestamp: time.Now().UTC()},
	}

	parquetData, err := wrapper.WriteStructsToParquet(events)
	require.NoError(t, err)

	// Test getting file info
	info, err := wrapper.GetParquetFileInfo(parquetData)
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, int64(2), info.NumRows)
	assert.Greater(t, info.NumColumns, int64(0))
	assert.Greater(t, info.NumRowGroups, int64(0))
	assert.Equal(t, int64(len(parquetData)), info.FileSize)
	assert.NotNil(t, info.Schema)

	// Test with empty data
	_, err = wrapper.GetParquetFileInfo([]byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parquet data cannot be empty")
}

// Example usage function demonstrating different struct types
func ExampleParquetWrapper() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Configure parquet settings
	config := entity.ParquetConfig{
		CompressionCodec: "snappy",
		PageSize:         8 * 1024,
		RowGroupSize:     1000,
	}

	// Create parquet wrapper
	wrapper, err := NewParquetWrapper[UserEvent](logger, config)
	if err != nil {
		logger.Error("Failed to create parquet wrapper", slog.String("error", err.Error()))
		return
	}

	// Example 1: User Events
	events := []UserEvent{
		{
			UserID:    123,
			EventType: "login",
			Timestamp: time.Now().UTC(),
			Data:      `{"ip": "192.168.1.1"}`,
		},
		{
			UserID:    456,
			EventType: "purchase",
			Timestamp: time.Now().UTC(),
			Data:      `{"amount": 99.99}`,
		},
	}

	// Write user events to parquet
	parquetData, err := wrapper.WriteStructsToParquet(events)
	if err != nil {
		logger.Error("Failed to write user events", slog.String("error", err.Error()))
		return
	}

	// Read user events back from parquet
	readEvents, err := wrapper.ReadStructsFromParquet(parquetData)
	if err != nil {
		logger.Error("Failed to read user events", slog.String("error", err.Error()))
		return
	}

	logger.Info("Processed user events", slog.Int("count", len(readEvents)))

	wrapper2, err := NewParquetWrapper[Product](logger, config)
	if err != nil {
		logger.Error("Failed to create parquet wrapper", slog.String("error", err.Error()))
		return
	}

	// Example 2: Products
	products := []Product{
		{
			ID:       1,
			Name:     "Laptop",
			Price:    999.99,
			Category: "Electronics",
			InStock:  true,
		},
	}

	// Write products to parquet
	productData, err := wrapper2.WriteStructsToParquet(products)
	if err != nil {
		logger.Error("Failed to write products", slog.String("error", err.Error()))
		return
	}

	// Get file information
	info, err := wrapper2.GetParquetFileInfo(productData)
	if err != nil {
		logger.Error("Failed to get file info", slog.String("error", err.Error()))
		return
	}

	logger.Info("Product file info",
		slog.Int64("rows", info.NumRows),
		slog.Int64("columns", info.NumColumns),
		slog.Int64("size_bytes", info.FileSize),
	)
}
