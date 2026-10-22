package hash

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

type BCryptService struct {
	cost      uint
	maxLength uint
	charset   string
	dummy     []byte
}

func NewBCryptGenerator(cost uint) *BCryptService {
	gen := BCryptService{
		cost:      cost,
		maxLength: 0,
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
	hash, err := bcrypt.GenerateFromPassword(pwd, 12)
	clear(pwd)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (h *BCryptService) PasswordMatches(pwd Password, hash []byte) bool {
	if pwd == nil {
		pwd = h.dummy
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

func (h *BCryptService) IsValidPassword(passwd Password) bool {
	if len(passwd) == 0 || len(passwd) > int(h.maxLength) {
		return false
	}
	return true
}
