// internal/repository/memory.go
package repository

import (
    "context"
    "sync"

    "grpc-users/internal/domain"
)

type InMemoryUserRepo struct {
    mu     sync.RWMutex
    nextID int64
    users  map[int64]*domain.User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
    return &InMemoryUserRepo{users: make(map[int64]*domain.User)}
}

func (r *InMemoryUserRepo) Create(_ context.Context, u *domain.User) (int64, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.nextID++
    u.ID = r.nextID
    cp := *u
    r.users[cp.ID] = &cp
    return cp.ID, nil
}

func (r *InMemoryUserRepo) GetByID(_ context.Context, id int64) (*domain.User, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    u, ok := r.users[id]
    if !ok {
        return nil, domain.ErrUserNotFound
    }
    cp := *u
    return &cp, nil
}

func (r *InMemoryUserRepo) List(_ context.Context) ([]*domain.User, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    res := make([]*domain.User, 0, len(r.users))
    for _, u := range r.users {
        cp := *u
        res = append(res, &cp)
    }
    return res, nil
}

func (r *InMemoryUserRepo) Update(_ context.Context, u *domain.User) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.users[u.ID]; !ok {
        return domain.ErrUserNotFound
    }
    cp := *u
    r.users[cp.ID] = &cp
    return nil
}

func (r *InMemoryUserRepo) Delete(_ context.Context, id int64) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.users[id]; !ok {
        return domain.ErrUserNotFound
    }
    delete(r.users, id)
    return nil
}

func (r *InMemoryUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    for _, u := range r.users {
        if u.Email == email {
            cp := *u
            return &cp, nil
        }
    }
    return nil, domain.ErrUserNotFound
}