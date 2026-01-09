package repository

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	domainRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

const prefixGroup = "group"

// groupStorageRepository implements the GroupStorageRepository interface
// and encapsulates logic for S3 and Parquet operations.
type groupStorageRepository struct {
	s3Wrapper           service.S3StorageWrapper
	groupParquetWrapper service.ParquetFileWrapper[entity.Group]
	logger              *slog.Logger
	tracer              trace.Tracer
}

// NewGroupStorageRepository creates a new repository for Group entities.
// Returns the repository or an error.
func NewGroupStorageRepository(
	logger *slog.Logger,
	s3Wrapper service.S3StorageWrapper,
	groupParquetWrapper service.ParquetFileWrapper[entity.Group],
	tracer trace.Tracer,
) (domainRepo.GroupStorageRepository, error) {
	if err := assert.NotNil(logger, s3Wrapper, groupParquetWrapper, tracer); err != nil {
		return nil, err
	}
	return &groupStorageRepository{
		s3Wrapper:           s3Wrapper,
		groupParquetWrapper: groupParquetWrapper,
		logger:              logger,
		tracer:              tracer,
	}, nil
}

// Create stores the provided Group object in the underlying storage system.
// The storage key is generated from the group's Id, so duplicate entities will be overwritten.
// Returns an error if unsuccessful, or nil otherwise.
func (r *groupStorageRepository) Create(ctx context.Context, obj *entity.Group) error {
	if err := assert.NotNil(ctx, obj); err != nil {
		return err
	}

	ctx, span := r.tracer.Start(ctx, "groupStorageRepository.Create")
	defer span.End()
	span.SetAttributes(
		attribute.String("group.id", obj.Id),
		attribute.String("group.name", obj.Name),
		attribute.String("group.createdBy", obj.CreatedBy),
	)

	groupParquet, err := r.groupParquetWrapper.WriteStructToParquet(ctx, *obj)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to serialize group object")
		return err
	}

	key := generateKey(obj.Id)
	err = r.s3Wrapper.UploadParquetFile(ctx, key, groupParquet, map[string]string{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to upload group parquet")
		return err
	}

	r.logger.Debug("create: object successfully created",
		slog.String("key", key),
	)

	span.AddEvent("group created", trace.WithAttributes(
		attribute.String("group.key", key),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

// Read retrieves a Group object from storage by its groupId.
// Returns a pointer to the Group or an error if not found or read fails.
func (r *groupStorageRepository) Read(ctx context.Context, groupId string) (*entity.Group, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	if err := assert.StringNotEmpty(groupId); err != nil {
		return nil, err
	}

	ctx, span := r.tracer.Start(ctx, "groupStorageRepository.Read")
	defer span.End()
	span.SetAttributes(
		attribute.String("group.id", groupId),
	)

	key := generateKey(groupId)
	parquetData, _, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to download group parquet")
		return nil, fmt.Errorf("%w: %w", sharedErrors.ErrGroupNotFound, err)
	}

	groups, err := r.groupParquetWrapper.ReadStructsFromParquet(ctx, parquetData)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to deserialize group parquet")
		return nil, err
	}

	if len(groups) != 1 {
		err := fmt.Errorf("expected 1 group, got %d", len(groups))
		span.RecordError(err)
		span.SetStatus(codes.Error, "incorrect number of groups in parquet file")
		return nil, err
	}

	r.logger.Debug("read: object successfully loaded",
		slog.String("key", key),
	)

	span.AddEvent("group loaded", trace.WithAttributes(
		attribute.String("group.key", key),
		attribute.String("group.id", groups[0].Id),
	))
	span.SetStatus(codes.Ok, "")

	return &groups[0], nil
}

// Update updates an existing Group object in storage.
// This is essentially the same as Create since we overwrite by key.
// Returns an error if unsuccessful, or nil otherwise.
func (r *groupStorageRepository) Update(ctx context.Context, obj *entity.Group) error {
	if err := assert.NotNil(ctx, obj); err != nil {
		return err
	}

	ctx, span := r.tracer.Start(ctx, "groupStorageRepository.Update")
	defer span.End()
	span.SetAttributes(
		attribute.String("group.id", obj.Id),
		attribute.String("group.name", obj.Name),
	)

	// Verify the group exists before updating
	_, err := r.Read(ctx, obj.Id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "group not found for update")
		return fmt.Errorf("cannot update non-existent group: %w", err)
	}

	groupParquet, err := r.groupParquetWrapper.WriteStructToParquet(ctx, *obj)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to serialize group object")
		return err
	}

	key := generateKey(obj.Id)
	err = r.s3Wrapper.UploadParquetFile(ctx, key, groupParquet, map[string]string{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to upload group parquet")
		return err
	}

	r.logger.Debug("update: object successfully updated",
		slog.String("key", key),
	)

	span.AddEvent("group updated", trace.WithAttributes(
		attribute.String("group.key", key),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

// Delete removes a Group object from storage by its groupId.
// Returns an error if unsuccessful, or nil otherwise.
func (r *groupStorageRepository) Delete(ctx context.Context, groupId string) error {
	if err := assert.NotNil(ctx); err != nil {
		return err
	}
	if err := assert.StringNotEmpty(groupId); err != nil {
		return err
	}

	ctx, span := r.tracer.Start(ctx, "groupStorageRepository.Delete")
	defer span.End()
	span.SetAttributes(
		attribute.String("group.id", groupId),
	)

	key := generateKey(groupId)
	if err := r.s3Wrapper.DeleteParquetFile(ctx, key); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete group parquet")
		return err
	}

	r.logger.Debug("delete: object successfully deleted",
		slog.String("key", key),
	)

	span.AddEvent("group deleted", trace.WithAttributes(
		attribute.String("group.key", key),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

// ListAll retrieves all Groups from storage.
// Returns a slice of Group objects or an error.
func (r *groupStorageRepository) ListAll(ctx context.Context) ([]*entity.Group, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}

	ctx, span := r.tracer.Start(ctx, "groupStorageRepository.ListAll")
	defer span.End()

	keys, err := r.s3Wrapper.ListParquetFiles(ctx, prefixGroup)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to list group parquet files")
		return nil, fmt.Errorf("failed to list group parquet files: %w", err)
	}
	if len(keys) == 0 {
		err := fmt.Errorf("no group files found in storage")
		span.RecordError(err)
		span.SetStatus(codes.Error, "no groups found")
		return nil, err
	}

	result := make([]*entity.Group, 0, len(keys))
	for _, key := range keys {
		parquetData, _, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
		if err != nil {
			r.logger.Error("ListAll: failed to download file", "key", key, "error", err)
			continue
		}
		groups, err := r.groupParquetWrapper.ReadStructsFromParquet(ctx, parquetData)
		if err != nil {
			r.logger.Error("ListAll: failed to read parquet data", "key", key, "error", err)
			continue
		}
		if len(groups) != 1 {
			r.logger.Error("ListAll: incorrect number of structs in parquet file", "key", key, "number", len(groups))
			continue
		}
		result = append(result, &groups[0])
	}

	r.logger.Debug("ListAll: finished loading groups", slog.Int("count", len(result)))
	span.AddEvent("groups loaded", trace.WithAttributes(
		attribute.Int("group.count", len(result)),
	))
	span.SetStatus(codes.Ok, "")

	return result, nil
}

func generateKey(groupId string) string {
	return fmt.Sprintf("%s/%s.parquet", prefixGroup, groupId)
}
