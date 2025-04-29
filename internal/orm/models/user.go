package models

import (
	"nosql-db/internal/db"
	"nosql-db/internal/orm"
)

// User représente un utilisateur dans la base de données
type User struct {
	Name     string                 `json:"name"`
	Email    string                 `json:"email"`
	Age      int                    `json:"age"`
	IsActive bool                   `json:"is_active"`
	Metadata map[string]interface{} `json:"metadata"`
}

// UserCollection représente une collection d'utilisateurs
type UserCollection struct {
	*orm.Collection[User]
}

// NewUserCollection crée une nouvelle collection d'utilisateurs
func NewUserCollection(database *db.Database) (*UserCollection, error) {
	// Créer la collection
	collection, err := orm.NewCollection[User](database, "users")
	if err != nil {
		return nil, err
	}

	return &UserCollection{
		Collection: collection,
	}, nil
}

// FindByEmail recherche un utilisateur par son email
func (uc *UserCollection) FindByEmail(email string) (User, error) {
	users, err := uc.FindByField("Email", email)
	if err != nil || len(users) == 0 {
		return User{}, err
	}
	return users[0], nil
}

// FindActive récupère tous les utilisateurs actifs
func (uc *UserCollection) FindActive() ([]User, error) {
	return uc.FindByField("IsActive", true)
}

// FindByAge récupère tous les utilisateurs d'un âge spécifique
func (uc *UserCollection) FindByAge(age int) ([]User, error) {
	return uc.FindByField("Age", age)
}
