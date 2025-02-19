package models

import (
	"regexp"
	"time"

	"promoservice/proto/user"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex    = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]{8,}$`)
	validate      = validator.New()
)

type User struct {
	ID           uuid.UUID `json:"id" validate:"required"`
	Login        string    `json:"login" validate:"required,min=3,max=50"`
	PasswordHash string    `json:"-"`
	Email        string    `json:"email" validate:"required,email"`
	FirstName    string    `json:"first_name" validate:"max=50"`
	LastName     string    `json:"last_name" validate:"max=50"`
	Phone        string    `json:"phone" validate:"omitempty,phone"`
	BirthDate    time.Time `json:"birth_date" validate:"omitempty"`
	CreatedAt    time.Time `json:"created_at" validate:"required"`
	UpdatedAt    time.Time `json:"updated_at" validate:"required"`
}

func (u *User) Validate() error {
	return validate.Struct(u)
}

func validateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func validatePhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

func validatePassword(password string) bool {
	return passwordRegex.MatchString(password)
}

func init() {
	validate.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		return validatePhone(fl.Field().String())
	})
}

func (u *User) ToProto() *user.User {
	return &user.User{
		Id:        u.ID.String(),
		Login:     u.Login,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		BirthDate: timestamppb.New(u.BirthDate),
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

func UserFromProto(protoUser *user.User) (*User, error) {
	id, err := uuid.Parse(protoUser.Id)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        id,
		Login:     protoUser.Login,
		Email:     protoUser.Email,
		FirstName: protoUser.FirstName,
		LastName:  protoUser.LastName,
		Phone:     protoUser.Phone,
		BirthDate: protoUser.BirthDate.AsTime(),
		CreatedAt: protoUser.CreatedAt.AsTime(),
		UpdatedAt: protoUser.UpdatedAt.AsTime(),
	}, nil
}

type RegisterRequest struct {
	Login     string     `json:"login" validate:"required"`
	Password  string     `json:"password" validate:"required"`
	Email     string     `json:"email" validate:"required,email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Phone     string     `json:"phone"`
	BirthDate *time.Time `json:"birth_date"`
}

type LoginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
	Email     string     `json:"email" validate:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Phone     string     `json:"phone"`
	BirthDate *time.Time `json:"birth_date"`
}
