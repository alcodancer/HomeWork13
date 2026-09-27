// internal/service/user_service.go
package service

import (
    "context"
    "strings"

    "grpc-users/internal/domain"
    "grpc-users/internal/repository"
)

type UserService struct {
    repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, name, email string, age int) (*domain.User, error) {
    if err := validate(name, email, age); err != nil {
        return nil, err
    }
    if _, err := s.repo.FindByEmail(ctx, email); err == nil {
        return nil, domain.ErrEmailExists
    } else if err != domain.ErrUserNotFound {
        return nil, err
    }

    u := &domain.User{Name: name, Email: email, Age: age}
    id, err := s.repo.Create(ctx, u)
    if err != nil {
        return nil, err
    }
    u.ID = id
    return u, nil
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
    if id <= 0 {
        return nil, domain.ErrUserNotFound
    }
    return s.repo.GetByID(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context) ([]*domain.User, error) {
    return s.repo.List(ctx)
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, name, email string, age int) (*domain.User, error) {
    if err := validate(name, email, age); err != nil {
        return nil, err
    }
    existing, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    // если email меняется — проверяем уникальность
    if existing.Email != email {
        if other, err := s.repo.FindByEmail(ctx, email); err == nil && other.ID != id {
            return nil, domain.ErrEmailExists
        } else if err != nil && err != domain.ErrUserNotFound {
            return nil, err
        }
    }
    existing.Name = name
    existing.Email = email
    existing.Age = age
    if err := s.repo.Update(ctx, existing); err != nil {
        return nil, err
    }
    return existing, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
    return s.repo.Delete(ctx, id)
}

func validate(name, email string, age int) error {
    if strings.TrimSpace(name) == "" {
        return domain.ErrInvalidName
    }
    if !strings.Contains(email, "@") || strings.TrimSpace(email) == "" {
        return domain.ErrInvalidEmail
    }
    if age < 0 || age > 150 {
        return domain.ErrInvalidAge
    }
    return nil
}