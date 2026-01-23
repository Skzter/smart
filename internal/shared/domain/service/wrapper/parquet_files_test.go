//go:build integration

package service

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

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
	tracer := otel.Tracer("test")
	testCases := []struct {
		name        string
		logger      *slog.Logger
		config      entity.ParquetConfig
		expectError bool
		err         error
	}{
		{
			name:        "Valid logger and config",
			logger:      slog.New(slog.DiscardHandler),
			config:      entity.ParquetConfig{},
			expectError: false,
		},
		{
			name:        "Nil logger",
			logger:      nil,
			config:      entity.ParquetConfig{},
			expectError: true,
			err:         ErrNilLogger,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrapper, err := NewParquetWrapper[UserEvent](tc.logger, tc.config, tracer)
			if tc.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.err)
				assert.Nil(t, wrapper)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, wrapper)
			}
		})
	}
}

func TestDefaultParquetConfig(t *testing.T) {
	config := DefaultParquetConfig()
	assert.Equal(t, "snappy", config.CompressionCodec)
	assert.Equal(t, int64(8*1024), config.PageSize)
	assert.Equal(t, int64(1000), config.RowGroupSize)
}

func runWriteReadTest[T any](t *testing.T, ctx context.Context, logger *slog.Logger, config entity.ParquetConfig, data []T, assertFunc func(t *testing.T, original, read []T), tracer trace.Tracer) {
	t.Helper()

	wrapper, err := NewParquetWrapper[T](logger, config, tracer)
	require.NoError(t, err)

	parquetData, err := wrapper.WriteStructsToParquet(ctx, data)
	require.NoError(t, err)
	assert.Greater(t, len(parquetData), 0)

	readData, err := wrapper.ReadStructsFromParquet(ctx, parquetData)
	require.NoError(t, err)
	assert.Len(t, readData, len(data))

	if assertFunc != nil {
		assertFunc(t, data, readData)
	}
}

func TestParquetWrapper_WriteAndRead(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	ctx := context.Background()

	t.Run("UserEvents", func(t *testing.T) {
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
		runWriteReadTest(t, ctx, logger, entity.ParquetConfig{}, events, func(t *testing.T, original, read []UserEvent) {
			for i, event := range original {
				assert.Equal(t, event.UserID, read[i].UserID)
				assert.Equal(t, event.EventType, read[i].EventType)
				assert.Equal(t, event.Data, read[i].Data)
				assert.WithinDuration(t, event.Timestamp, read[i].Timestamp, time.Millisecond)
			}
		}, tracer)
	})

	t.Run("Products", func(t *testing.T) {
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
		config := entity.ParquetConfig{
			CompressionCodec: "gzip",
			PageSize:         4 * 1024,
			RowGroupSize:     500,
		}
		runWriteReadTest(t, ctx, logger, config, products, func(t *testing.T, original, read []Product) {
			for i, product := range original {
				assert.Equal(t, product.ID, read[i].ID)
				assert.Equal(t, product.Name, read[i].Name)
				assert.Equal(t, product.Price, read[i].Price)
				assert.Equal(t, product.Category, read[i].Category)
				assert.Equal(t, product.InStock, read[i].InStock)
				assert.Equal(t, product.Description, read[i].Description)
			}
		}, tracer)
	})

	t.Run("LogEntries", func(t *testing.T) {
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
		config := entity.ParquetConfig{
			CompressionCodec: "gzip",
			PageSize:         4 * 1024,
			RowGroupSize:     500,
		}
		runWriteReadTest(t, ctx, logger, config, logs, func(t *testing.T, original, read []LogEntry) {
			for i, log := range original {
				assert.Equal(t, log.Level, read[i].Level)
				assert.Equal(t, log.Message, read[i].Message)
				assert.Equal(t, log.Source, read[i].Source)
				assert.WithinDuration(t, log.Timestamp, read[i].Timestamp, time.Millisecond)

				if log.UserID == nil {
					assert.Nil(t, read[i].UserID)
				} else {
					require.NotNil(t, read[i].UserID)
					assert.Equal(t, *log.UserID, *read[i].UserID)
				}
			}
		}, tracer)
	})
}

func TestWriteStructToParquetSingleStruct(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	ctx := context.Background()

	testCases := []struct {
		name       string
		event      UserEvent
		expectID   int64
		expectType string
	}{
		{
			name: "Single UserEvent",
			event: UserEvent{
				UserID:    999,
				EventType: "test",
				Timestamp: time.Now().UTC(),
				Data:      `{"test": true}`,
			},
			expectID:   999,
			expectType: "test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrapper, err := NewParquetWrapper[UserEvent](logger, entity.ParquetConfig{}, tracer)
			require.NoError(t, err)

			// Test writing single struct
			parquetData, err := wrapper.WriteStructToParquet(ctx, tc.event)
			require.NoError(t, err)
			assert.Greater(t, len(parquetData), 0)

			// Test reading back
			readEvents, err := wrapper.ReadStructsFromParquet(ctx, parquetData)
			require.NoError(t, err)
			assert.Len(t, readEvents, 1)
			assert.Equal(t, tc.expectID, readEvents[0].UserID)
			assert.Equal(t, tc.expectType, readEvents[0].EventType)
		})
	}
}

func TestGetParquetSchemaUserEvents(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	ctx := context.Background()
	wrapper, err := NewParquetWrapper[UserEvent](logger, entity.ParquetConfig{}, tracer)
	require.NoError(t, err)

	// Test schema generation for UserEvent
	schema, err := wrapper.GetParquetSchema(ctx)
	require.NoError(t, err)
	require.NotNil(t, schema)
	assert.Greater(t, len(schema.Fields()), 0)

	// Test schema generation for Product
	schema, err = wrapper.GetParquetSchema(ctx)
	require.NoError(t, err)
	require.NotNil(t, schema)
	assert.Greater(t, len(schema.Fields()), 0)
}

func TestValidateStructUserEvents(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	ctx := context.Background()
	wrapper, err := NewParquetWrapper[UserEvent](logger, entity.ParquetConfig{}, tracer)
	require.NoError(t, err)

	// Test valid struct
	event := UserEvent{UserID: 1, EventType: "test"}
	err = wrapper.ValidateStruct(ctx, event)
	assert.NoError(t, err)

	// Test with pointer
	err = wrapper.ValidateStruct(ctx, event)
	assert.NoError(t, err)
}

func TestParquetWrapperErrorHandling(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	ctx := context.Background()
	wrapper, err := NewParquetWrapper[UserEvent](logger, entity.ParquetConfig{}, tracer)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		action        func() error
		expectError   bool
		expectedErr   error
		expectPanic   bool
		recoveryCheck func(t *testing.T, r interface{})
	}{
		{
			name: "Write empty data slice",
			action: func() error {
				_, err := wrapper.WriteStructsToParquet(ctx, []UserEvent{})
				return err
			},
			expectError: true,
			expectedErr: ErrEmptyData,
		},
		{
			name: "Read empty parquet data",
			action: func() error {
				_, err := wrapper.ReadStructsFromParquet(ctx, []byte{})
				return err
			},
			expectError: true,
			expectedErr: ErrEmptyParquetData,
		},
		{
			name: "Read invalid parquet data",
			action: func() error {
				_, err := wrapper.ReadStructsFromParquet(ctx, []byte("invalid parquet data"))
				return err // this error is not actually returned due to panic
			},
			expectPanic: true,
			recoveryCheck: func(t *testing.T, r interface{}) {
				assert.NotNil(t, r, "Expected a panic for invalid parquet data")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPanic {
				defer func() {
					if r := recover(); r != nil {
						if tc.recoveryCheck != nil {
							tc.recoveryCheck(t, r)
						}
					} else {
						t.Error("Expected a panic but did not get one")
					}
				}()
			}
			err := tc.action()
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					assert.ErrorIs(t, err, tc.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetParquetFileInfoUserEvents(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	ctx := context.Background()
	wrapper, err := NewParquetWrapper[UserEvent](logger, entity.ParquetConfig{}, tracer)
	require.NoError(t, err)

	// Create test data
	events := []UserEvent{
		{UserID: 1, EventType: "test1", Timestamp: time.Now().UTC()},
		{UserID: 2, EventType: "test2", Timestamp: time.Now().UTC()},
	}

	parquetData, err := wrapper.WriteStructsToParquet(ctx, events)
	require.NoError(t, err)

	// Test getting file info
	info, err := wrapper.GetParquetFileInfo(ctx, parquetData)
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, int64(2), info.NumRows)
	assert.Greater(t, info.NumColumns, int64(0))
	assert.Greater(t, info.NumRowGroups, int64(0))
	assert.Equal(t, int64(len(parquetData)), info.FileSize)
	assert.NotNil(t, info.Schema)

	// Test with empty data
	_, err = wrapper.GetParquetFileInfo(ctx, []byte{})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyParquetData)
}

// Example usage function demonstrating different struct types
func ExampleParquetWrapper() {
	// Initialize logger
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	ctx := context.Background()

	// Configure parquet settings
	config := entity.ParquetConfig{
		CompressionCodec: "snappy",
		PageSize:         8 * 1024,
		RowGroupSize:     1000,
	}

	// Create parquet wrapper
	wrapper, err := NewParquetWrapper[UserEvent](logger, config, tracer)
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
	parquetData, err := wrapper.WriteStructsToParquet(ctx, events)
	if err != nil {
		logger.Error("Failed to write user events", slog.String("error", err.Error()))
		return
	}

	// Read user events back from parquet
	readEvents, err := wrapper.ReadStructsFromParquet(ctx, parquetData)
	if err != nil {
		logger.Error("Failed to read user events", slog.String("error", err.Error()))
		return
	}

	logger.Info("Processed user events", slog.Int("count", len(readEvents)))

	wrapper2, err := NewParquetWrapper[Product](logger, config, tracer)
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
	productData, err := wrapper2.WriteStructsToParquet(ctx, products)
	if err != nil {
		logger.Error("Failed to write products", slog.String("error", err.Error()))
		return
	}

	// Get file information
	info, err := wrapper2.GetParquetFileInfo(ctx, productData)
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
