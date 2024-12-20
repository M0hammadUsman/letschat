package domain

import (
	"context"
	"regexp"
	"time"
)

var (
	RgxEmail      = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	AnonymousUser = &User{}
)

type User struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Password   []byte     `json:"-"`
	Activated  bool       `json:"-"`
	LastOnline *time.Time `json:"lastOnline,omitempty" db:"last_online"`
	CreatedAt  time.Time  `json:"createdAt"  db:"created_at"`
	Version    int        `json:"-"`
	// Websocket related
	Messages  MsgChan `json:"-"`
	CloseSlow func()  `json:"-"`
}

type UserService interface {
	RegisterUser(ctx context.Context, u *UserRegister) (string, error)
	ExistsUser(ctx context.Context, email string) (bool, error)
	GetByUniqueField(ctx context.Context, fieldValue string) (*User, error)
	UpdateUser(ctx context.Context, u *UserUpdate) error
	UpdateUserOnlineStatus(ctx context.Context, usr *User, online bool) error
	GetForToken(ctx context.Context, scope string, plainToken string) (*User, error)
	ActivateUser(ctx context.Context, user *User) error
	AuthenticateUser(ctx context.Context, u *UserAuth) (string, error)
	GetByQuery(ctx context.Context, queryParam string, filter Filter) ([]*User, *Metadata, error)
	SetOnlineUsersLastSeen(ctx context.Context, t time.Time) error
}

type UserRepository interface {
	RegisterUser(ctx context.Context, u *User) (string, error)
	ExistsUser(ctx context.Context, email string) (bool, error)
	GetByUniqueField(ctx context.Context, fieldName, fieldValue string) (*User, error)
	UpdateUser(ctx context.Context, u *User) error
	GetForToken(ctx context.Context, scope string, hash []byte) (*User, error)
	ActivateUser(ctx context.Context, user *User) error
	GetByQuery(ctx context.Context, paramName string, paramValue string, filter Filter) ([]*User, *Metadata, error)
	SetOnlineUsersLastSeen(ctx context.Context, t time.Time) error
}

// DTOs

type UserRegister struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserAuth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserUpdate struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	NewPassword     *string `json:"newPassword"`
	CurrentPassword *string `json:"currentPassword"`
}

func (u *User) IsAnonymousUser() bool {
	return u == AnonymousUser
}

func ValidateName(name string, ev *ErrValidation) {
	ev.Evaluate(name != "", "name", "must be provided")
	ev.Evaluate(len(name) >= 3, "name", "must be 3 bytes long")
	ev.Evaluate(len(name) <= 30, "name", "must be no more than 30 bytes long")
}

func ValidateEmail(email string, ev *ErrValidation) {
	ev.Evaluate(email != "", "email", "must be provided")
	if len(email) > 254 || !RgxEmail.MatchString(email) {
		ev.AddError("email", "must be valid")
	}
}

func ValidPlainPassword(pass string, ev *ErrValidation) {
	ev.Evaluate(pass != "", "password", "must be provided")
	ev.Evaluate(pass == "" || len(pass) >= 8, "password", "must be at least 8 bytes long")
	ev.Evaluate(len(pass) <= 72, "password", "must no be more than 72 bytes long")
}

func ValidPlainPasswordWithKey(pass string, ev *ErrValidation, errKey string) {
	ev.Evaluate(pass != "", errKey, "must be provided")
	ev.Evaluate(pass == "" || len(pass) >= 8, errKey, "must be at least 8 bytes long")
	ev.Evaluate(len(pass) <= 72, errKey, "must no be more than 72 bytes long")
}
