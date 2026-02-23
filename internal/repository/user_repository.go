package repository

import (
	"database/sql"
	"time"

	"fiscalization-api/internal/models"

	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	// User operations
	Create(user *models.User) error
	GetByID(id int64) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByTaxpayerID(taxpayerID int64) ([]models.User, error)
	Update(user *models.User) error
	Delete(id int64) error
	
	// Security code operations
	SaveSecurityCode(userID int64, code string, expiresAt time.Time) error
	GetSecurityCode(userID int64) (string, time.Time, error)
	DeleteSecurityCode(userID int64) error
	
	// Password operations
	UpdatePassword(userID int64, passwordHash string) error
	
	// List operations
	List(taxpayerID int64, offset, limit int) ([]models.User, int, error)
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (
			taxpayer_id, username, password_hash, person_name, person_surname,
			email, phone_no, user_role, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(
		query,
		user.TaxpayerID,
		user.Username,
		user.PasswordHash,
		user.PersonName,
		user.PersonSurname,
		user.Email,
		user.PhoneNo,
		user.UserRole,
		user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) GetByID(id int64) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE id = $1`

	err := r.db.Get(&user, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE username = $1`

	err := r.db.Get(&user, query, username)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByTaxpayerID(taxpayerID int64) ([]models.User, error) {
	var users []models.User
	query := `SELECT * FROM users WHERE taxpayer_id = $1 ORDER BY username`

	err := r.db.Select(&users, query, taxpayerID)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) Update(user *models.User) error {
	query := `
		UPDATE users SET
			person_name = $1,
			person_surname = $2,
			email = $3,
			phone_no = $4,
			user_role = $5,
			status = $6,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $7`

	_, err := r.db.Exec(
		query,
		user.PersonName,
		user.PersonSurname,
		user.Email,
		user.PhoneNo,
		user.UserRole,
		user.Status,
		user.ID,
	)

	return err
}

func (r *userRepository) Delete(id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *userRepository) SaveSecurityCode(userID int64, code string, expiresAt time.Time) error {
	query := `
		INSERT INTO security_codes (user_id, code, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			code = EXCLUDED.code,
			expires_at = EXCLUDED.expires_at,
			created_at = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query, userID, code, expiresAt)
	return err
}

func (r *userRepository) GetSecurityCode(userID int64) (string, time.Time, error) {
	var code string
	var expiresAt time.Time

	query := `
		SELECT code, expires_at
		FROM security_codes
		WHERE user_id = $1 AND expires_at > CURRENT_TIMESTAMP`

	err := r.db.QueryRow(query, userID).Scan(&code, &expiresAt)
	if err == sql.ErrNoRows {
		return "", time.Time{}, nil
	}
	if err != nil {
		return "", time.Time{}, err
	}

	return code, expiresAt, nil
}

func (r *userRepository) DeleteSecurityCode(userID int64) error {
	query := `DELETE FROM security_codes WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *userRepository) UpdatePassword(userID int64, passwordHash string) error {
	query := `
		UPDATE users SET
			password_hash = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`

	_, err := r.db.Exec(query, passwordHash, userID)
	return err
}

func (r *userRepository) List(taxpayerID int64, offset, limit int) ([]models.User, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM users WHERE taxpayer_id = $1`
	err := r.db.Get(&total, countQuery, taxpayerID)
	if err != nil {
		return nil, 0, err
	}

	// Get users
	var users []models.User
	query := `
		SELECT * FROM users
		WHERE taxpayer_id = $1
		ORDER BY username
		LIMIT $2 OFFSET $3`

	err = r.db.Select(&users, query, taxpayerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
