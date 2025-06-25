# S3 and Parquet Wrappers

This package provides comprehensive wrappers for both Amazon S3 operations and Parquet file handling. It includes functionality for uploading, downloading, listing, and managing parquet files in S3, as well as generic parquet file operations that work with any Go struct.

## Components

1. **S3Wrapper** - Handle S3 operations for parquet files
2. **ParquetWrapper** - Generic parquet file operations with Go structs
3. **ParquetS3Wrapper** - Combined S3 and Parquet operations for seamless file handling

## Features

### S3 Operations
- **Upload parquet files** with metadata support
- **Download parquet files** with metadata retrieval
- **List parquet files** with optional prefix filtering
- **Check file existence** and get file size
- **Delete parquet files**
- **Automatic .parquet extension handling**
- **Support for S3-compatible services** (like MinIO)
- **AWS credential chain support**

### Parquet Operations
- **Generic struct serialization** to parquet format
- **Generic struct deserialization** from parquet format
- **Multiple compression options** (Snappy, Gzip, None)
- **Schema generation** from Go structs
- **Struct validation** for parquet compatibility
- **File metadata extraction**
- **Support for optional fields** and complex data types
- **Batch processing** for large datasets

### General Features
- **Comprehensive logging** with structured logs
- **Input validation** and error handling
- **Type safety** with Go generics
- **Thread-safe operations**

## Installation

The required dependencies are already included in the project:
- `github.com/aws/aws-sdk-go-v2` - AWS SDK for S3 operations
- `github.com/aws/aws-sdk-go-v2/config` - AWS configuration loading
- `github.com/aws/aws-sdk-go-v2/service/s3` - S3 service client
- `github.com/parquet-go/parquet-go` - Parquet file format support
- `github.com/parquet-go/parquet-go/compress/snappy` - Snappy compression
- `github.com/parquet-go/parquet-go/compress/gzip` - Gzip compression

## Configuration

### ParquetConfig Structure

```go
type ParquetConfig struct {
    CompressionCodec string // Compression codec ("snappy", "gzip", "none")
    PageSize         int64  // Page size in bytes (default: 8KB)
    RowGroupSize     int64  // Row group size in rows (default: 1000)
}
```

#### Default Configuration

```go
config := service.DefaultParquetConfig()
// Returns:
// ParquetConfig{
//     CompressionCodec: "snappy",
//     PageSize:         8 * 1024, // 8KB
//     RowGroupSize:     1000,     // 1000 rows per group
// }
```

### S3Config Structure

```go
type S3Config struct {
    Region    string // AWS region (default: "us-east-1")
    Bucket    string // S3 bucket name (required)
    AccessKey string // AWS access key (optional)
    SecretKey string // AWS secret key (optional)
    Endpoint  string // Custom endpoint for S3-compatible services (optional)
}
```

### Configuration Options

1. **Using AWS Credential Chain** (recommended for production):
```go
config := S3Config{
    Region: "us-east-1",
    Bucket: "my-parquet-bucket",
}
```

2. **Using explicit credentials**:
```go
config := S3Config{
    Region:    "us-east-1",
    Bucket:    "my-parquet-bucket",
    AccessKey: "AKIAIOSFODNN7EXAMPLE",
    SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
}
```

3. **Using S3-compatible services (MinIO, etc.)**:
```go
config := S3Config{
    Region:    "us-east-1",
    Bucket:    "my-parquet-bucket",
    AccessKey: "minioadmin",
    SecretKey: "minioadmin",
    Endpoint:  "http://localhost:9000",
}
```

## Usage Examples

### Struct Definition for Parquet

First, define your structs with parquet tags:

```go
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

// LogEntry with optional field
type LogEntry struct {
    Level     string    `parquet:"level"`
    Message   string    `parquet:"message"`
    Timestamp time.Time `parquet:"timestamp"`
    Source    string    `parquet:"source"`
    UserID    *int64    `parquet:"user_id,optional"` // Optional field
}
```

### Basic Parquet Operations

```go
package main

import (
    "log/slog"
    "os"
    "time"
    
    "your-project/internal/shared/domain/service/wrapper"
)

func main() {
    // Initialize logger
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    
    // Create parquet wrapper
    parquetWrapper, err := service.NewParquetWrapper(logger)
    if err != nil {
        logger.Error("Failed to create parquet wrapper", slog.String("error", err.Error()))
        return
    }
    
    // Configure parquet settings
    config := service.ParquetConfig{
        CompressionCodec: "snappy",
        PageSize:         8 * 1024,
        RowGroupSize:     1000,
    }
    
    // Example data
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
    
    // Write structs to parquet format
    parquetData, err := service.WriteStructsToParquet(parquetWrapper, events, config)
    if err != nil {
        logger.Error("Failed to write parquet", slog.String("error", err.Error()))
        return
    }
    
    logger.Info("Parquet data created", slog.Int("size_bytes", len(parquetData)))
    
    // Read structs back from parquet format
    readEvents, err := service.ReadStructsFromParquet[UserEvent](parquetWrapper, parquetData)
    if err != nil {
        logger.Error("Failed to read parquet", slog.String("error", err.Error()))
        return
    }
    
    logger.Info("Events read from parquet", slog.Int("count", len(readEvents)))
}
```

### Combined S3 and Parquet Operations

```go
func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    
    // Configure S3
    s3Config := service.S3Config{
        Region: "us-east-1",
        Bucket: "my-data-bucket",
    }
    
    // Create combined wrapper
    wrapper, err := service.NewParquetS3Wrapper(logger, s3Config)
    if err != nil {
        logger.Error("Failed to create wrapper", slog.String("error", err.Error()))
        return
    }
    
    ctx := context.Background()
    
    // Your data
    products := []Product{
        {ID: 1, Name: "Laptop", Price: 999.99, Category: "Electronics", InStock: true},
        {ID: 2, Name: "Book", Price: 29.99, Category: "Literature", InStock: false},
    }
    
    parquetConfig := service.DefaultParquetConfig()
    metadata := map[string]string{
        "source": "product-catalog",
        "version": "1.0",
    }
    
    // Upload structs as parquet to S3
    err = service.UploadStructsAsParquet(wrapper, ctx, "products/2024/catalog", products, parquetConfig, metadata)
    if err != nil {
        logger.Error("Upload failed", slog.String("error", err.Error()))
        return
    }
    
    // Download and convert back to structs
    downloadedProducts, fileMetadata, err := service.DownloadStructsFromParquet[Product](wrapper, ctx, "products/2024/catalog.parquet")
    if err != nil {
        logger.Error("Download failed", slog.String("error", err.Error()))
        return
    }
    
    logger.Info("Products downloaded",
        slog.Int("count", len(downloadedProducts)),
        slog.Any("metadata", fileMetadata),
    )
}
```

### Advanced Parquet Operations

```go
func advancedParquetOperations() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    wrapper, _ := service.NewParquetWrapper(logger)
    
    // 1. Get schema for a struct type
    schema, err := service.GetParquetSchema[UserEvent](wrapper)
    if err != nil {
        logger.Error("Failed to get schema", slog.String("error", err.Error()))
        return
    }
    
    logger.Info("Schema generated", slog.Int("fields", len(schema.Fields())))
    
    // 2. Validate struct before serialization
    event := UserEvent{UserID: 1, EventType: "test"}
    err = service.ValidateStruct(wrapper, event)
    if err != nil {
        logger.Error("Struct validation failed", slog.String("error", err.Error()))
        return
    }
    
    // 3. Write single struct (convenience method)
    config := service.DefaultParquetConfig()
    parquetData, err := service.WriteStructToParquet(wrapper, event, config)
    if err != nil {
        logger.Error("Failed to write single struct", slog.String("error", err.Error()))
        return
    }
    
    // 4. Get file information without full deserialization
    info, err := wrapper.GetParquetFileInfo(parquetData)
    if err != nil {
        logger.Error("Failed to get file info", slog.String("error", err.Error()))
        return
    }
    
    logger.Info("File info",
        slog.Int64("rows", info.NumRows),
        slog.Int64("columns", info.NumColumns),
        slog.Int64("size_bytes", info.FileSize),
        slog.String("compression", info.Compression),
    )
    
    // 5. Different compression options
    configs := []service.ParquetConfig{
        {CompressionCodec: "snappy", PageSize: 8192, RowGroupSize: 1000},
        {CompressionCodec: "gzip", PageSize: 4096, RowGroupSize: 500},
        {CompressionCodec: "none", PageSize: 16384, RowGroupSize: 2000},
    }
    
    events := []UserEvent{{UserID: 1, EventType: "test", Timestamp: time.Now()}}
    
    for _, cfg := range configs {
        data, err := service.WriteStructsToParquet(wrapper, events, cfg)
        if err != nil {
            continue
        }
        logger.Info("Compression test",
            slog.String("codec", cfg.CompressionCodec),
            slog.Int("size_bytes", len(data)),
        )
    }
}
```

### Basic Setup

```go
package main

import (
    "context"
    "log/slog"
    "os"
    
    "your-project/internal/shared/domain/service/wrapper"
)

func main() {
    // Initialize logger
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    
    // Configure S3 wrapper
    config := service.S3Config{
        Region: "us-east-1",
        Bucket: "my-parquet-bucket",
    }
    
    // Create S3 wrapper
    s3Wrapper, err := service.NewS3Wrapper(logger, config)
    if err != nil {
        logger.Error("Failed to create S3 wrapper", slog.String("error", err.Error()))
        return
    }
    
    ctx := context.Background()
    
    // Use the wrapper...
}
```

### Upload a Parquet File

```go
// Prepare parquet data (typically from a parquet library)
parquetData := []byte("your parquet file data")

// Optional metadata
metadata := map[string]string{
    "source":      "data-pipeline",
    "created-by":  "user-123",
    "data-format": "parquet",
    "schema":      "user-events-v1",
}

// Upload file (automatically adds .parquet extension if missing)
err := s3Wrapper.UploadParquetFile(ctx, "data/2024/01/user-events", parquetData, metadata)
if err != nil {
    logger.Error("Upload failed", slog.String("error", err.Error()))
    return
}
```

### Download a Parquet File

```go
// Download file
data, metadata, err := s3Wrapper.DownloadParquetFile(ctx, "data/2024/01/user-events.parquet")
if err != nil {
    logger.Error("Download failed", slog.String("error", err.Error()))
    return
}

logger.Info("File downloaded", 
    slog.Int("size", len(data)),
    slog.Any("metadata", metadata),
)

// Process the parquet data...
```

### List Parquet Files

```go
// List all parquet files with prefix
files, err := s3Wrapper.ListParquetFiles(ctx, "data/2024/01/")
if err != nil {
    logger.Error("Failed to list files", slog.String("error", err.Error()))
    return
}

for _, file := range files {
    logger.Info("Found parquet file", slog.String("key", file))
}
```

### Check File Existence and Size

```go
// Check if file exists
exists, err := s3Wrapper.FileExists(ctx, "data/2024/01/user-events.parquet")
if err != nil {
    logger.Error("Failed to check existence", slog.String("error", err.Error()))
    return
}

if exists {
    // Get file size
    size, err := s3Wrapper.GetFileSize(ctx, "data/2024/01/user-events.parquet")
    if err != nil {
        logger.Error("Failed to get size", slog.String("error", err.Error()))
        return
    }
    
    logger.Info("File info", 
        slog.Bool("exists", exists),
        slog.Int64("size_bytes", size),
    )
}
```

### Delete a Parquet File

```go
err := s3Wrapper.DeleteParquetFile(ctx, "data/2024/01/user-events.parquet")
if err != nil {
    logger.Error("Failed to delete file", slog.String("error", err.Error()))
    return
}

logger.Info("File deleted successfully")
```

## Environment Variables

The wrapper supports standard AWS environment variables when using the credential chain:

- `AWS_REGION` - AWS region
- `AWS_ACCESS_KEY_ID` - AWS access key
- `AWS_SECRET_ACCESS_KEY` - AWS secret key
- `AWS_SESSION_TOKEN` - AWS session token (for temporary credentials)
- `AWS_PROFILE` - AWS profile name

## Error Handling

The wrapper provides comprehensive error handling:

- **Input validation errors** - for empty keys, nil contexts, missing bucket names
- **AWS service errors** - wrapped with additional context
- **File operation errors** - with detailed logging

All errors are wrapped with context information to help with debugging.

## Logging

The wrapper uses structured logging with the following log levels:

- **INFO** - Successful operations, initialization
- **ERROR** - Failed operations with error details

Log entries include relevant context such as:
- Bucket name
- File key
- File size
- Operation type
- Error details

## Testing

Run the tests with:

```bash
go test ./internal/shared/domain/service/wrapper/ -v
```

The tests include:
- Configuration validation
- Input validation
- Error handling
- Extension handling (.parquet auto-append)

## Best Practices

1. **Use structured logging** - All operations are logged with context
2. **Handle errors appropriately** - Check all return values
3. **Use context with timeouts** - For long-running operations
4. **Validate inputs** - The wrapper validates inputs but additional validation is recommended
5. **Use prefixes for organization** - Organize files with prefixes like `data/year/month/`
6. **Include metadata** - Add relevant metadata for better file management
7. **Use AWS credential chain** - Avoid hardcoding credentials in production

## Thread Safety

The S3Wrapper is thread-safe and can be used concurrently from multiple goroutines. The underlying AWS SDK client handles concurrent requests safely. 