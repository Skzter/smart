package repository

// s3StorageRepository implements StorageRepository for AWS S3 storage.
type s3StorageRepository struct {
	// wrapper service
}

func NewS3StorageRepository() StorageRepository {
	return &s3StorageRepository{}
}

func (s3 *s3StorageRepository) Create() {
	// UploadParquetFile(ctx, key, data, metadata)

	// Fragen für Johannes:
	// soll key hereingegeben werden oder hier erzeugt werden
	// wenn hereingegeben validieren das es den noch nicht gab
}

func (s3 *s3StorageRepository) Read(key string) {
	// DownloadParquetFile(ctx, key)

	// key erst validieren mit Exists, dann lesen
}

func (s3 *s3StorageRepository) Update(key string) error {
	// UploadParquetFile(ctx, key, data, metadata)
	// with existing key, overwrites the old file

	// key erst validieren mit Exists, dann überschreiben
	return nil
}

func (s3 *s3StorageRepository) Delete(key string) error {
	// DeleteParquetFile(ctx, key)

	// key erst validieren mit Exists, dann löschen
	return nil
}
