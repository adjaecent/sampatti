package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"

	"github.com/adjaecent/sampatti/internal/repository"
	"github.com/adjaecent/sampatti/models"
)

type AuthService struct {
	userRepo      *repository.UserRepository
	planRepo      *repository.PlanRepository
	encryptionKey []byte
}

func NewAuthService(userRepo *repository.UserRepository, planRepo *repository.PlanRepository) *AuthService {
	// In production, this should come from environment variables or secure key management
	encryptionKey := []byte("sampatti-encryption-key-32bytes!") // 32 bytes for AES-256
	return &AuthService{
		userRepo:      userRepo,
		planRepo:      planRepo,
		encryptionKey: encryptionKey,
	}
}

func (s *AuthService) GetOrCreateUserPlan(email, name string) (*models.User, *models.Plan, error) {
	// First, get or create the user
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// User doesn't exist, create them
		user, err = s.userRepo.Create(email, name, nil)
		if err != nil {
			return nil, nil, err
		}
	}

	// Check if user owns a plan
	plan, err := s.planRepo.GetByOwnerUserID(user.ID)
	if err == nil {
		return user, plan, nil
	}

	// Check if user is a member of a plan
	plan, err = s.planRepo.GetByMemberUserID(user.ID)
	if err == nil {
		return user, plan, nil
	}

	// Create new plan for user
	planName := name + "'s Plan"
	plan, err = s.planRepo.Create(planName, user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, plan, nil
}

func (s *AuthService) EncryptPassword(password string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return hex.EncodeToString(ciphertext), nil
}

func (s *AuthService) DecryptPassword(encryptedPassword string) (string, error) {
	data, err := hex.DecodeString(encryptedPassword)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// Keep this for backward compatibility or if needed elsewhere
func (s *AuthService) HashPassword(password string) (string, error) {
	return s.EncryptPassword(password)
}

func (s *AuthService) ValidatePlatform(platform string) error {
	if platform != "kuvera" && platform != "stockal" {
		return errors.New("invalid platform")
	}
	return nil
}