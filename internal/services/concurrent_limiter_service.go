package services

import (
	"sync"

	"github.com/google/uuid"
)

type UserConcurrentLimiter struct {
	mu    sync.Mutex
	limit int
	count map[uuid.UUID]int
}

func NewUserConcurrentLimiter(limit int) *UserConcurrentLimiter {
	return &UserConcurrentLimiter{
		limit: limit,
		count: make(map[uuid.UUID]int),
	}
}

func (l *UserConcurrentLimiter) Increment(userID uuid.UUID) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.count[userID]++
	return l.count[userID]
}

func (l *UserConcurrentLimiter) Decrement(userID uuid.UUID) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.count[userID] <= 1 {
		delete(l.count, userID)
		return
	}
	l.count[userID]--
}
