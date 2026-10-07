package domain

import (
	"uuid"

	"github.com/distributed-programming-2026/lib/event"
	"github.com/pkg/errors"

	userevents "user/api/events"
)

type EchoService struct {
	repo       EchoRepository
	dispatcher event.Dispatcher
}

type UserService interface {
	CreateUser(login, firstName, lastName, email string) (uuid.UUID, error)
	UpdateUser(id uuid.UUID, login, firstName, lastName, email string) error
	RemoveUser(id uuid.UUID) error
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

type userService struct {
	repo       UserRepository
	dispatcher event.Dispatcher
}

func (u *userService) CreateUser(login, firstName, lastName, email string) (uuid.UUID, error) {
	id, err := u.repo.NextID()
	if err != nil {
		return uuid.Nil(), err
	}

	user := User{
		ID:        id,
		Login:     login,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}
	err = u.repo.Store(user)
	if err != nil {
		return uuid.Nil(), err
	}

	err = u.dispatcher.Dispatch(&userevents.UserCreated{
		ID:    user.ID.String(),
		Login: login,
	})

	return user.ID, nil
}

func (u *userService) UpdateUser(id uuid.UUID, login, firstName, lastName, email string) error {
	user, err := u.repo.Find(id)
	if err != nil {
		return err
	}

	if user.Login != login {
		user.Login = login
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
		ID:    id.String(),
		Login: login,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (u *userService) RemoveUser(id uuid.UUID) error {
	_, err := u.repo.Find(id)
	if err != nil {
		return err
	}

	err = u.repo.Delete(id)
	if err != nil {
		return err
	}

	err = u.dispatcher.Dispatch(&userevents.UserRemoved{ID: id.String()})
	if err != nil {
		return err
	}

	return nil
}

func NewEchoService(repo EchoRepository, dispatcher event.Dispatcher) *EchoService {
	return &EchoService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

func (s *EchoService) Echo(body string) (uuid.UUID, error) {
	id, err := s.repo.NextID()
	if err != nil {
		return uuid.Nil(), err
	}

	e := Echo{
		ID:   id,
		Body: body,
	}

	err = s.repo.Store(e)
	if err != nil {
		return uuid.Nil(), err
	}

	err = s.dispatcher.Dispatch(&userevents.Echo{
		ID:   e.ID.String(),
		Body: e.Body,
	})
	if err != nil {
		return uuid.Nil(), err
	}
	return id, nil
}
