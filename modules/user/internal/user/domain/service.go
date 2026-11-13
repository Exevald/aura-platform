package domain

import (
	"uuid"

	"github.com/distributed-programming-2026/lib/event"
	"github.com/pkg/errors"

	userevents "user/api/events"
)

type UserService interface {
	CreateUser(displayName, firstName, lastName, email string) (uuid.UUID, error)
	CreateUserWithID(id uuid.UUID, displayName, firstName, lastName, email string) error
	GetUser(id uuid.UUID) (User, error)
	FindUserByEmail(email string) (User, error)
	UpdateUser(id uuid.UUID, displayName, firstName, lastName, email string) error
	RemoveUser(id uuid.UUID) error
}

type userService struct {
	repo       UserRepository
	dispatcher event.Dispatcher
}

func NewUserService(
	repo UserRepository,
	dispatcher event.Dispatcher,
) UserService {
	return &userService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

func (u *userService) CreateUser(displayName, firstName, lastName, email string) (uuid.UUID, error) {
	id, err := u.repo.NextID()
	if err != nil {
		return uuid.Nil(), err
	}
	if err = u.CreateUserWithID(id, displayName, firstName, lastName, email); err != nil {
		return uuid.Nil(), err
	}
	return id, nil
}

func (u *userService) CreateUserWithID(id uuid.UUID, displayName, firstName, lastName, email string) error {
	if id == uuid.Nil() || displayName == "" || firstName == "" || lastName == "" || email == "" {
		return ErrInvalidUser
	}
	if _, err := u.repo.FindByEmail(email); err == nil {
		return ErrEmailTaken
	} else if !errors.Is(err, ErrUserNotFound) {
		return err
	}

	user := User{
		ID:          id,
		DisplayName: displayName,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
	}
	err := u.repo.Store(user)
	if err != nil {
		return err
	}

	err = u.dispatcher.Dispatch(&userevents.UserCreated{
		Id:          user.ID.String(),
		DisplayName: displayName,
	})
	if err != nil {
		return err
	}
	return nil
}

func (u *userService) FindUserByEmail(email string) (User, error) {
	if email == "" {
		return User{}, ErrInvalidUser
	}
	return u.repo.FindByEmail(email)
}

func (u *userService) GetUser(id uuid.UUID) (User, error) {
	if id == uuid.Nil() {
		return User{}, ErrInvalidUser
	}

	return u.repo.Find(id)
}

func (u *userService) UpdateUser(id uuid.UUID, displayName, firstName, lastName, email string) error {
	if id == uuid.Nil() || displayName == "" || firstName == "" || lastName == "" || email == "" {
		return ErrInvalidUser
	}
	user, err := u.repo.Find(id)
	if err != nil {
		return err
	}
	if existing, findErr := u.repo.FindByEmail(email); findErr == nil && existing.ID != id {
		return ErrEmailTaken
	} else if findErr != nil && !errors.Is(findErr, ErrUserNotFound) {
		return findErr
	}

	if user.DisplayName != displayName {
		user.DisplayName = displayName
	}
	if user.FirstName != firstName {
		user.FirstName = firstName
	}
	if user.LastName != lastName {
		user.LastName = lastName
	}
	if user.Email != email {
		user.Email = email
	}

	err = u.repo.Store(user)
	if err != nil {
		return err
	}

	err = u.dispatcher.Dispatch(&userevents.UserUpdated{
		Id:          id.String(),
		DisplayName: displayName,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (u *userService) RemoveUser(id uuid.UUID) error {
	if id == uuid.Nil() {
		return ErrInvalidUser
	}
	_, err := u.repo.Find(id)
	if err != nil {
		return err
	}

	err = u.repo.Delete(id)
	if err != nil {
		return err
	}

	err = u.dispatcher.Dispatch(&userevents.UserRemoved{Id: id.String()})
	if err != nil {
		return err
	}

	return nil
}
