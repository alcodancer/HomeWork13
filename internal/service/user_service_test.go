// internal/service/user_service_test.go
package service_test // <- внешний пакет = чёрный ящик

import (
    "context"
    "errors"
    "testing"

    "grpc-users/internal/domain"
    "grpc-users/internal/repository"
    "grpc-users/internal/service"
)

func newSvc() *service.UserService {
    return service.NewUserService(repository.NewInMemoryUserRepo())
}

func TestUserService_CreateUser(t *testing.T) {
    type want struct {
        id    int64
        name  string
        email string
        age   int
        err   error
    }
    tests := []struct {
        name string
        // предзаполнение хранилища
        seed []struct{ name, email string; age int }
        in   struct{ name, email string; age int }
        want want
    }{
        {
            name: "success",
            in:   struct{ name, email string; age int }{"Alice", "a@mail.com", 30},
            want: want{id: 1, name: "Alice", email: "a@mail.com", age: 30},
        },
        {
            name: "empty name",
            in:   struct{ name, email string; age int }{"", "a@mail.com", 30},
            want: want{err: domain.ErrInvalidName},
        },
        {
            name: "whitespace name",
            in:   struct{ name, email string; age int }{"   ", "a@mail.com", 30},
            want: want{err: domain.ErrInvalidName},
        },
        {
            name: "invalid email (no @)",
            in:   struct{ name, email string; age int }{"Alice", "mail.com", 30},
            want: want{err: domain.ErrInvalidEmail},
        },
        {
            name: "negative age",
            in:   struct{ name, email string; age int }{"Alice", "a@mail.com", -1},
            want: want{err: domain.ErrInvalidAge},
        },
        {
            name: "too big age",
            in:   struct{ name, email string; age int }{"Alice", "a@mail.com", 200},
            want: want{err: domain.ErrInvalidAge},
        },
        {
            name: "duplicate email",
            seed: []struct{ name, email string; age int }{{"Bob", "a@mail.com", 25}},
            in:   struct{ name, email string; age int }{"Alice", "a@mail.com", 30},
            want: want{err: domain.ErrEmailExists},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := newSvc()
            ctx := context.Background()
            for _, s := range tt.seed {
                if _, err := svc.CreateUser(ctx, s.name, s.email, s.age); err != nil {
                    t.Fatalf("seed failed: %v", err)
                }
            }

            got, err := svc.CreateUser(ctx, tt.in.name, tt.in.email, tt.in.age)

            if tt.want.err != nil {
                if !errors.Is(err, tt.want.err) {
                    t.Fatalf("want err %v, got %v", tt.want.err, err)
                }
                return
            }
            if err != nil {
                t.Fatalf("unexpected err: %v", err)
            }
            if got.ID != tt.want.id || got.Name != tt.want.name ||
                got.Email != tt.want.email || got.Age != tt.want.age {
                t.Fatalf("want %+v, got %+v", tt.want, got)
            }
        })
    }
}

func TestUserService_GetUser(t *testing.T) {
    ctx := context.Background()
    svc := newSvc()
    created, _ := svc.CreateUser(ctx, "Alice", "a@mail.com", 30)

    tests := []struct {
        name string
        id   int64
        want error
    }{
        {"success", created.ID, nil},
        {"not found", 999, domain.ErrUserNotFound},
        {"zero id", 0, domain.ErrUserNotFound},
        {"negative id", -5, domain.ErrUserNotFound},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := svc.GetUser(ctx, tt.id)
            if tt.want != nil {
                if !errors.Is(err, tt.want) {
                    t.Fatalf("want %v, got %v", tt.want, err)
                }
                return
            }
            if err != nil || got.ID != tt.id {
                t.Fatalf("unexpected: user=%+v err=%v", got, err)
            }
        })
    }
}

func TestUserService_ListUsers(t *testing.T) {
    ctx := context.Background()

    t.Run("empty", func(t *testing.T) {
        svc := newSvc()
        list, err := svc.ListUsers(ctx)
        if err != nil || len(list) != 0 {
            t.Fatalf("want empty, got %v err=%v", list, err)
        }
    })

    t.Run("multiple", func(t *testing.T) {
        svc := newSvc()
        svc.CreateUser(ctx, "A", "a@mail.com", 20)
        svc.CreateUser(ctx, "B", "b@mail.com", 21)
        list, err := svc.ListUsers(ctx)
        if err != nil || len(list) != 2 {
            t.Fatalf("want 2, got %d err=%v", len(list), err)
        }
    })
}

func TestUserService_UpdateUser(t *testing.T) {
    ctx := context.Background()

    tests := []struct {
        name    string
        prepare func(svc *service.UserService) int64
        update  struct{ name, email string; age int }
        want    error
    }{
        {
            name: "success",
            prepare: func(svc *service.UserService) int64 {
                u, _ := svc.CreateUser(ctx, "Alice", "a@mail.com", 30)
                return u.ID
            },
            update: struct{ name, email string; age int }{"Alice2", "a2@mail.com", 31},
        },
        {
            name: "not found",
            prepare: func(_ *service.UserService) int64 { return 123 },
            update: struct{ name, email string; age int }{"X", "x@mail.com", 1},
            want:   domain.ErrUserNotFound,
        },
        {
            name: "invalid name",
            prepare: func(svc *service.UserService) int64 {
                u, _ := svc.CreateUser(ctx, "Alice", "a@mail.com", 30)
                return u.ID
            },
            update: struct{ name, email string; age int }{"", "a@mail.com", 30},
            want:   domain.ErrInvalidName,
        },
        {
            name: "email taken by other",
            prepare: func(svc *service.UserService) int64 {
                svc.CreateUser(ctx, "Bob", "b@mail.com", 20)
                u, _ := svc.CreateUser(ctx, "Alice", "a@mail.com", 30)
                return u.ID
            },
            update: struct{ name, email string; age int }{"Alice", "b@mail.com", 30},
            want:   domain.ErrEmailExists,
        },
        {
            name: "same email ok",
            prepare: func(svc *service.UserService) int64 {
                u, _ := svc.CreateUser(ctx, "Alice", "a@mail.com", 30)
                return u.ID
            },
            update: struct{ name, email string; age int }{"Alice", "a@mail.com", 31},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := newSvc()
            id := tt.prepare(svc)
            got, err := svc.UpdateUser(ctx, id, tt.update.name, tt.update.email, tt.update.age)
            if tt.want != nil {
                if !errors.Is(err, tt.want) {
                    t.Fatalf("want %v, got %v", tt.want, err)
                }
                return
            }
            if err != nil {
                t.Fatalf("unexpected: %v", err)
            }
            if got.Name != tt.update.name || got.Email != tt.update.email || got.Age != tt.update.age {
                t.Fatalf("update not applied: %+v", got)
            }
        })
    }
}

func TestUserService_DeleteUser(t *testing.T) {
    ctx := context.Background()
    tests := []struct {
        name    string
        prepare func(svc *service.UserService) int64
        want    error
    }{
        {
            name: "success",
            prepare: func(svc *service.UserService) int64 {
                u, _ := svc.CreateUser(ctx, "Alice", "a@mail.com", 30)
                return u.ID
            },
        },
        {
            name:    "not found",
            prepare: func(_ *service.UserService) int64 { return 42 },
            want:    domain.ErrUserNotFound,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := newSvc()
            id := tt.prepare(svc)
            err := svc.DeleteUser(ctx, id)
            if !errors.Is(err, tt.want) {
                t.Fatalf("want %v, got %v", tt.want, err)
            }
        })
    }
}