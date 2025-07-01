package entity

// S3Config holds configuration for S3 client
type S3Config struct {
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	Endpoint  string // For S3-compatible services like MinIO
}
