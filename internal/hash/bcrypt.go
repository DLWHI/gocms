package hash

import (
	"crypto/rand"
	"gocms/internal/logger"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

type BCryptService struct {
	cost      uint
	maxLength uint
	charset   string
	dummy     []byte
}

func NewBCryptService(cost uint) *BCryptService {
	if cost < 4 {
		logger.Warn("BCryptService", "Provided cost is to low, clamped it to 4", nil)
		cost = 4
	} else if cost > 31 {
		logger.Warn("BCryptService", "Provided cost is to high, clamped it to 31", nil)
		cost = 31
	}

	gen := BCryptService{
		cost:      cost,
		maxLength: 72,
		charset:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?",
	}

	pwd, err := gen.GeneratePassword(16)
	if err != nil {
		return &gen
	}
	dummyHash, err := gen.Hash(pwd)
	if err != nil {
		return &gen
	}
	gen.dummy = dummyHash
	return &gen
}

func (h *BCryptService) Hash(pwd Password) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, int(h.cost))
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (h *BCryptService) PasswordMatches(pwd Password, hash []byte) bool {
	if len(hash) == 0 {
		if len(h.dummy) == 0 {
			return false
		}
		hash = h.dummy
	}
	if err := bcrypt.CompareHashAndPassword(hash, pwd); err != nil {
		return false
	}
	return true
}

func (h *BCryptService) GeneratePassword(length int) (Password, error) {
	bytes := make([]byte, length)

	for i := range bytes {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(h.charset))))
		if err != nil {
			return nil, err
		}
		bytes[i] = h.charset[randomIndex.Int64()]
	}

	return bytes, nil
}

func (h *BCryptService) IsValidPassword(password Password) bool {
	if len(password) == 0 || len(password) > int(h.maxLength) {
		return false
	}
	return true
}
