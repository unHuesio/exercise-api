package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"gym-api/m/models"

	"cloud.google.com/go/storage"
)

type GoogleCloudExerciseStore struct {
	bucket string
	object string
	client *storage.Client
}

func NewGoogleCloudExerciseStore(bucket, object string) (*GoogleCloudExerciseStore, error) {
	if strings.TrimSpace(bucket) == "" {
		return nil, errors.New("bucket name is required")
	}
	if strings.TrimSpace(object) == "" {
		return nil, errors.New("object name is required")
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("create storage client: %w", err)
	}

	return &GoogleCloudExerciseStore{bucket: bucket, object: object, client: client}, nil
}

func (s *GoogleCloudExerciseStore) List(ctx context.Context, filters map[string]string) ([]models.Exercise, error) {
	items, err := s.readAll(ctx)
	if err != nil {
		return nil, err
	}
	return filterExercises(items, filters), nil
}

func (s *GoogleCloudExerciseStore) GetByID(ctx context.Context, id string) (models.Exercise, error) {
	var zero models.Exercise
	items, err := s.readAll(ctx)
	if err != nil {
		return zero, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return zero, ErrNotFound
}

func (s *GoogleCloudExerciseStore) Create(ctx context.Context, exercise models.Exercise) (string, error) {
	items, err := s.readAll(ctx)
	if err != nil {
		return "", err
	}
	if exercise.ID == "" {
		exercise.ID = fmt.Sprintf("%03d", len(items)+1)
	}
	for _, item := range items {
		if item.ID == exercise.ID {
			return "", fmt.Errorf("exercise %s already exists", exercise.ID)
		}
	}
	items = append(items, exercise)
	if err := s.writeAll(ctx, items); err != nil {
		return "", err
	}
	return exercise.ID, nil
}

func (s *GoogleCloudExerciseStore) Update(ctx context.Context, id string, exercise models.Exercise) error {
	items, err := s.readAll(ctx)
	if err != nil {
		return err
	}
	for i, item := range items {
		if item.ID == id {
			exercise.ID = id
			items[i] = exercise
			return s.writeAll(ctx, items)
		}
	}
	return ErrNotFound
}

func (s *GoogleCloudExerciseStore) Delete(ctx context.Context, id string) error {
	items, err := s.readAll(ctx)
	if err != nil {
		return err
	}
	for i, item := range items {
		if item.ID == id {
			items = append(items[:i], items[i+1:]...)
			return s.writeAll(ctx, items)
		}
	}
	return ErrNotFound
}

func (s *GoogleCloudExerciseStore) Health(ctx context.Context) error {
	if s == nil || s.client == nil {
		return errors.New("storage client is not initialized")
	}
	_, err := s.readAll(ctx)
	if err != nil {
		return fmt.Errorf("storage health check failed: %w", err)
	}
	return nil
}

func (s *GoogleCloudExerciseStore) BackendName() string {
	return "gcs"
}

func (s *GoogleCloudExerciseStore) readAll(ctx context.Context) ([]models.Exercise, error) {
	if s.client == nil {
		return nil, errors.New("storage client is not initialized")
	}

	reader, err := s.client.Bucket(s.bucket).Object(s.object).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("open bucket object: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read bucket object: %w", err)
	}
	return parseExerciseJSON(data)
}

func (s *GoogleCloudExerciseStore) writeAll(ctx context.Context, exercises []models.Exercise) error {
	if s.client == nil {
		return errors.New("storage client is not initialized")
	}

	data, err := json.MarshalIndent(exercises, "", "  ")
	if err != nil {
		return err
	}

	writer := s.client.Bucket(s.bucket).Object(s.object).NewWriter(ctx)
	writer.ContentType = "application/json"
	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write bucket object: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close bucket object: %w", err)
	}
	return nil
}

func parseExerciseJSON(data []byte) ([]models.Exercise, error) {
	if len(data) == 0 {
		return []models.Exercise{}, nil
	}
	var exercises []models.Exercise
	if err := json.Unmarshal(data, &exercises); err != nil {
		return nil, err
	}
	return exercises, nil
}

func filterExercises(items []models.Exercise, filters map[string]string) []models.Exercise {
	focus := strings.TrimSpace(filters["focus"])
	typeFilter := strings.TrimSpace(filters["type"])
	muscle := strings.TrimSpace(filters["muscle"])

	filtered := make([]models.Exercise, 0, len(items))
	for _, exercise := range items {
		if focus != "" && !strings.EqualFold(exercise.Focus, focus) {
			continue
		}
		if typeFilter != "" && !strings.EqualFold(exercise.Type, typeFilter) {
			continue
		}
		if muscle != "" {
			match := strings.Contains(strings.ToLower(exercise.PrimaryMuscles), strings.ToLower(muscle)) ||
				strings.Contains(strings.ToLower(exercise.SecondaryMuscles), strings.ToLower(muscle))
			if !match {
				continue
			}
		}
		filtered = append(filtered, exercise)
	}
	return filtered
}
