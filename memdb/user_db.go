package memdb

import (
	"code.google.com/p/go-uuid/uuid"
	"errors"
	"github.com/murphysean/heimdall"
)

func (db *MemDB) NewUser() heimdall.User {
	u := make(User)
	u["id"] = uuid.New()
	return u
}

func (db *MemDB) VerifyUser(username, password string) (heimdall.User, error) {
	if l, ok := db.loginMap[username]; ok {
		uid := l.id
		u := username
		p := l.password
		if u == username && p == password {
			return db.GetUser(uid)
		}
	}
	return nil, heimdall.ErrInvalidCredentials
}

func (db *MemDB) CreateUser(user heimdall.User) (heimdall.User, error) {
	db.userMap[user.GetId()] = user
	if u, ok := user.(User); ok {
		if p, pok := u["password"]; pok {
			if password, passwordok := p.(string); passwordok {
				if un, unok := u["username"]; unok {
					if username, usernameok := un.(string); usernameok {
						db.loginMap[username] = login{id: user.GetId(), password: password}
					}
				}
			}
		}
	}
	return user, nil
}

func (db *MemDB) GetUser(userId string) (heimdall.User, error) {
	user, ok := db.userMap[userId]
	if !ok {
		return user, errors.New("Not Found")
	}
	return user, nil
}

func (db *MemDB) UpdateUser(user heimdall.User) (heimdall.User, error) {
	db.userMap[user.GetId()] = user
	return user, nil
}

func (db *MemDB) DeleteUser(userId string) error {
	delete(db.userMap, userId)
	return nil
}
