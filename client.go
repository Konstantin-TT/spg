package client

import (
	"context"
	"fmt"
	"github.com/Konstantin-TT/configo"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

type Storage struct {
	Pool *pgxpool.Pool
}

func NewClient(ctx context.Context, cfg *configo.Database) (*pgxpool.Pool, error) {
	dsn := Dsn(cfg)

	var pool *pgxpool.Pool
	err := try(func() error {
		var err error
		ctx, cancel := context.WithTimeout(ctx, cfg.AttemptDelay)
		defer cancel()

		pool, err = pgxpool.New(ctx, dsn)
		if err != nil {
			return fmt.Errorf("ошибка при создании пула подключений к базе данных: %w", err)
		}

		err = pool.Ping(ctx)
		if err != nil {
			return fmt.Errorf("ошибка при проверке подключения к базе данных: %w", err)
		}

		return nil
	}, cfg.MaxAttempts, cfg.AttemptDelay)

	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных после %v попыток: %w", cfg.MaxAttempts, err)
	}

	log.Println("Успешное подключение к базе данных!")
	return pool, nil
}

func Dsn(cfg *configo.Database) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
}

func try(fn func() error, attempts int, delay time.Duration) (err error) {
	for attempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			attempts--
			continue
		}
		return nil
	}
	return
}
