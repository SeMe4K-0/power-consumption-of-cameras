package repository

import (
	"awesomeProject/internal/app/ds"
)

func (r *Repository) GetUserByID(id uint) (ds.User, error) {
	var u ds.User
	err := r.db.First(&u, id).Error
	return u, err
}

func (r *Repository) GetUserByUsername(username string) (ds.User, error) {
	var u ds.User
	err := r.db.Where("username = ?", username).First(&u).Error
	return u, err
}

func (r *Repository) UpdateUser(id uint, user ds.User) error {
	updates := map[string]interface{}{"username": user.Username}
	updates["is_leading_engineer"] = user.IsLeadingEngineer
	if user.Email != "" {
		updates["email"] = user.Email
	}
	return r.db.Model(&ds.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) CreateUser(user ds.User) (ds.User, error) {
	if user.Username == "" {
		user.Username = "user"
	}
	err := r.db.Create(&user).Error
	return user, err
}
