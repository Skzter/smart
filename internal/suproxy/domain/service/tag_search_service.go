package service

/*
func FindKeysByTag(ctx, tag) → list of keys or error:
    // 1. Get all Parquet file names (S3 keys) from the bucket
    keys := s3.ListParquetFiles()

    // 2. Create an empty list for hits
    matchingKeys := []

    // 3. For every key:
    for each key in keys:
        // 3a. Load file metadata (e.g. tags)
        data, metadata := s3.DownloadParquetFile(key)

        // 3b. If an error occurs during download → skip
        if error → continue

        // 3c. If the tag you are looking for appears in the key → add
        if tag is part of key:
            matchingKeys.add(key)

    // 4. Return all keys found
    return matchingKeys
*/
