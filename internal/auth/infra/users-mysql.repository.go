package infra

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"neuro.app.jordi/database/dbmodels"
	"neuro.app.jordi/internal/auth/domain"
)

type UsersMYSQLRepository struct {
	Exec boil.ContextExecutor // <- puede ser *sql.DB o *sql.Tx

}

// NewUsersMYSQLRepository crea una nueva instancia de UsersMYSQLRepository
func NewUseMYSQLRepository(db *sql.DB) *UsersMYSQLRepository {
	return &UsersMYSQLRepository{Exec: db}
}

func userToDomain(user *dbmodels.User) domain.User {
	mail, _ := domain.NewMail(user.Email)
	name, _ := domain.NewUserName(user.Name)
	return domain.User{
		ID:                            user.ID,
		Email:                         mail,
		Name:                          name,
		IsActive:                      user.IsActive,
		HasAcceptedTermsAndConditions: user.HasAcceptedLatestTerms,
		CreatedAt:                     user.CreatedAt,
		UpdatedAt:                     user.UpdatedAt,
	}
}

func userToDBModel(user domain.User) *dbmodels.User {
	dbUser := &dbmodels.User{
		ID:                     user.ID,
		CognitoID:              user.ID,
		Email:                  user.Email.Mail,
		Phone:                  null.StringFrom(user.Phone),
		HasAcceptedLatestTerms: user.HasAcceptedTermsAndConditions,
		IsActive:               user.IsActive,
		Name:                   user.Name.Name,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}
	return dbUser
}

// GetUserById obtiene un usuario por su ID
func (repo *UsersMYSQLRepository) GetUserById(ctx context.Context, id string) (domain.User, error) {
	user, err := dbmodels.Users(
		dbmodels.UserWhere.ID.EQ(id),
	).One(ctx, repo.Exec)
	if err != nil {
		return domain.User{}, err
	}

	return userToDomain(user), nil

}

// GetUserByMail obtiene un usuario por su correo electrónico
func (repo *UsersMYSQLRepository) GetUserByMail(ctx context.Context, mail string) (domain.User, error) {
	user, err := dbmodels.Users(
		dbmodels.UserWhere.Email.EQ(mail),
	).One(ctx, repo.Exec)
	if err != nil {
		return domain.User{}, err
	}

	return userToDomain(user), nil

}

// Insert inserta un nuevo usuario
func (repo *UsersMYSQLRepository) Insert(ctx context.Context, user domain.User) error {
	dbUser := userToDBModel(user)
	return dbUser.Insert(ctx, repo.Exec, boil.Infer())
}

// Delete elimina un usuario por su ID
func (repo *UsersMYSQLRepository) Delete(ctx context.Context, id string) error {
	_, err := dbmodels.Users(
		dbmodels.UserWhere.ID.EQ(id),
	).DeleteAll(ctx, repo.Exec)
	return err
}

// Exists verifica si un usuario existe por su correo electrónico
func (repo *UsersMYSQLRepository) Exists(ctx context.Context, mail string) bool {
	ok, _ := dbmodels.Users(
		dbmodels.UserWhere.Email.EQ(mail),
	).Exists(ctx, repo.Exec)
	return ok
}

// Update actualiza un usuario existente
func (repo *UsersMYSQLRepository) Update(ctx context.Context, user domain.User) error {
	dbUser := userToDBModel(user)
	_, err := dbUser.Update(ctx, repo.Exec, boil.Infer())
	return err
}
