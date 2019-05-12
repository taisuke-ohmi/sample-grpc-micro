package main

import (
	"fmt"
	pb "sample-grpc-micro/proto/user"
	"sample-grpc-micro/shared/inmemory"
)

type Store interface {
	CreateUser(*pb.User) (*pb.User, error)
	FindUser(uint64) (*pb.User, error)
	FindUserByEmail(string) (*pb.User, error)
}

type StoreOnMemory struct {
	users *inmemory.IndexMap
}

func NewStoreOnMemory() *StoreOnMemory {
	return &StoreOnMemory{inmemory.NewIndexMap()}
}

func (s *StoreOnMemory) CreateUser(u *pb.User) (*pb.User, error) {
	if _, err := s.FindUserByEmail(u.Email); err != nil {
		return nil, fmt.Errorf("already exists user %s", u.Email)
	}
	newUser := *u
	idx := s.users.Index()
	s.users.Set(idx, &newUser)
	return &newUser, nil
}

func (s *StoreOnMemory) FindUser(id uint64) (*pb.User, error) {
	v, ok := s.users.Get(id)
	if !ok {
		return nil, fmt.Errorf("not found user %d", id)
	}
	return v.(*pb.User), nil
}

func (s *StoreOnMemory) FindUserByEmail(email string) (*pb.User, error) {
	var user *pb.User
	s.users.Range(func(idx uint64, value interface{}) bool {
		u := value.(*pb.User)
		if u.Email == email {
			user = u
			return false
		}
		return true
	})
	if user == nil {
		return nil, fmt.Errorf("not found user")
	}
	return user, nil
}
