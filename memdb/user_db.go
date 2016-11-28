package memdb

import (
	"errors"
	"github.com/murphysean/heimdall"
)

func (db *MemDB) NewUser() heimdall.User {
	u := new(User)
	u.Id = genUUIDv4()
	return u
}

func (db *MemDB) VerifyUser(username, password string) (heimdall.User, error) {
	db.m.RLock()
	defer db.m.RUnlock()
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
	db.m.Lock()
	defer db.m.Unlock()
	db.userMap[user.GetId()] = user
	if u, ok := user.(*User); ok {
		if u.Username != "" && u.Password != "" {
			db.loginMap[u.Username] = login{id: user.GetId(), password: u.Password}
		}
	}
	return user, nil
}

func (db *MemDB) GetUser(userId string) (heimdall.User, error) {
	db.m.RLock()
	defer db.m.RUnlock()
	user, ok := db.userMap[userId]
	if !ok {
		return user, errors.New("Not Found")
	}
	return user, nil
}

func (db *MemDB) UpdateUser(user heimdall.User) (heimdall.User, error) {
	db.m.Lock()
	defer db.m.Unlock()
	db.userMap[user.GetId()] = user
	return user, nil
}

func (db *MemDB) DeleteUser(userId string) error {
	db.m.Lock()
	defer db.m.Unlock()
	delete(db.userMap, userId)
	return nil
}
