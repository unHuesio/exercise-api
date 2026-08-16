package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gym-api/m/models"
)

var ErrNotFound = errors.New("exercise not found")

type ExerciseStore interface {
	List(ctx context.Context, filters map[string]string) ([]models.Exercise, error)
	GetByID(ctx context.Context, id string) (models.Exercise, error)
	Create(ctx context.Context, exercise models.Exercise) (string, error)
	Update(ctx context.Context, id string, exercise models.Exercise) error
	Delete(ctx context.Context, id string) error
	Health(ctx context.Context) error
	BackendName() string
}

type FileExerciseStore struct {
	path string
}

func NewFileExerciseStore(path string) (*FileExerciseStore, error) {
	if path == "" {
		return nil, errors.New("exercise storage path is required")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}

	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(path, []byte("[]\n"), 0o644); err != nil {
				return nil, fmt.Errorf("create storage file: %w", err)
			}
		} else {
			return nil, err
		}
	}

	return &FileExerciseStore{path: path}, nil
}

func NewExerciseStoreFromConfig(filePath, bucket, object string) (ExerciseStore, error) {
	if bucket != "" && object != "" {
		return NewGoogleCloudExerciseStore(bucket, object)
	}
	return NewFileExerciseStore(filePath)
}

func (s *FileExerciseStore) readAll() ([]models.Exercise, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return []models.Exercise{}, nil
	}
	var exercises []models.Exercise
	if err := json.Unmarshal(data, &exercises); err != nil {
		return nil, err
	}
	return exercises, nil
}

func (s *FileExerciseStore) writeAll(exercises []models.Exercise) error {
	data, err := json.MarshalIndent(exercises, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, append(data, '\n'), 0o644)
}

func (s *FileExerciseStore) List(_ context.Context, filters map[string]string) ([]models.Exercise, error) {
	exercises, err := s.readAll()
	if err != nil {
		return nil, err
	}

	focus := strings.TrimSpace(filters["focus"])
	typeFilter := strings.TrimSpace(filters["type"])
	muscle := strings.TrimSpace(filters["muscle"])

	filtered := make([]models.Exercise, 0, len(exercises))
	for _, exercise := range exercises {
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
	return filtered, nil
}

func (s *FileExerciseStore) GetByID(_ context.Context, id string) (models.Exercise, error) {
	var zero models.Exercise
	exercises, err := s.readAll()
	if err != nil {
		return zero, err
	}
	for _, exercise := range exercises {
		if exercise.ID == id {
			return exercise, nil
		}
	}
	return zero, ErrNotFound
}

func (s *FileExerciseStore) Create(_ context.Context, exercise models.Exercise) (string, error) {
	if strings.TrimSpace(exercise.Exercise) == "" {
		return "", errors.New("exercise name is required")
	}

	exercises, err := s.readAll()
	if err != nil {
		return "", err
	}
	if exercise.ID == "" {
		exercise.ID = fmt.Sprintf("%03d", len(exercises)+1)
	}
	for _, existing := range exercises {
		if existing.ID == exercise.ID {
			return "", fmt.Errorf("exercise %s already exists", exercise.ID)
		}
	}
	exercises = append(exercises, exercise)
	if err := s.writeAll(exercises); err != nil {
		return "", err
	}
	return exercise.ID, nil
}

func (s *FileExerciseStore) Update(_ context.Context, id string, exercise models.Exercise) error {
	exercises, err := s.readAll()
	if err != nil {
		return err
	}
	for i, existing := range exercises {
		if existing.ID == id {
			exercise.ID = id
			exercises[i] = exercise
			return s.writeAll(exercises)
		}
	}
	return ErrNotFound
}

func (s *FileExerciseStore) Delete(_ context.Context, id string) error {
	exercises, err := s.readAll()
	if err != nil {
		return err
	}
	for i, exercise := range exercises {
		if exercise.ID == id {
			exercises = append(exercises[:i], exercises[i+1:]...)
			return s.writeAll(exercises)
		}
	}
	return ErrNotFound
}

func (s *FileExerciseStore) Health(_ context.Context) error {
	if s == nil || s.path == "" {
		return errors.New("exercise storage path is not configured")
	}
	if _, err := os.Stat(s.path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("storage file does not exist: %s", s.path)
		}
		return err
	}
	return nil
}

func (s *FileExerciseStore) BackendName() string {
	return "file"
}
