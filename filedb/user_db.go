package filedb

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/csv"
	"encoding/json"
	"errors"
	"github.com/murphysean/heimdall"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	USERS_DIRECTORY = "users"
)

func (db *FileDB) NewUser() heimdall.User {
	u := make(User)
	u["id"] = uuid.New()
	return u
}

func (db *FileDB) VerifyUser(username, password string) (heimdall.User, error) {
	f, err := os.Open(filepath.Join(db.Directory, "login.csv"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = 3

	for {
		record, err := r.Read()
		if err != nil {
			break
		}
		uid := record[0]
		u := record[1]
		p := record[2]
		if u == username && p == password {
			return db.GetUser(uid)
		}
	}
	return nil, heimdall.ErrInvalidCredentials
}

func (db *FileDB) CreateUser(user heimdall.User) (heimdall.User, error) {
	db.cache.Put(user.GetId(), user)
	b, err := json.Marshal(&user)
	if err != nil {
		return user, err
	}
	err = ioutil.WriteFile(filepath.Join(db.Directory, USERS_DIRECTORY, user.GetId()+".json"), b, os.ModePerm)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (db *FileDB) GetUser(userId string) (heimdall.User, error) {
	u, err := db.cache.GetWithValueLoader(userId, db.getUser)
	if err != nil {
		return nil, err
	}
	if user, ok := u.(heimdall.User); ok {
		return user, nil
	}
	return nil, errors.New("Unknown Value")
}

func (db *FileDB) getUser(userId string) (interface{}, error) {
	b, err := ioutil.ReadFile(filepath.Join(db.Directory, USERS_DIRECTORY, userId+".json"))
	if err != nil {
		return nil, err
	}
	var user User
	err = json.Unmarshal(b, &user)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (db *FileDB) UpdateUser(user heimdall.User) (heimdall.User, error) {
	return db.CreateUser(user)
}

func (db *FileDB) DeleteUser(userId string) error {
	db.cache.Invalidate(userId)
	return os.Remove(filepath.Join(db.Directory, USERS_DIRECTORY, userId+".json"))
}
