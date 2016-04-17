package sqldb

import (
	"encoding/json"
	giu "github.com/murphysean/gointerfaceutils"
	"github.com/murphysean/heimdall"
)

func (db *SqlDB) NewUser() heimdall.User {
	u := new(User)
	u.Id = genUUIDv4()
	return u
}

// Not part of the heimdall interface, but this will allow sso entities to create more
// complex user objects
func (db *SqlDB) InsertOrUpdateRawUser(userId, name string, user map[string]interface{}) (heimdall.User, error) {
	//Goal if the user exists, and there is json, merge it together and then resave it
	tx, err := db.Db.Begin()
	if err != nil {
		return nil, err
	}
	var ojson []byte
	ouser := make(map[string]interface{})
	var doc interface{}
	err = tx.QueryRow("SELECT json FROM users WHERE userId = ?", userId).Scan(&ojson)
	if err == nil {
		err = json.Unmarshal(ojson, &ouser)
		if err != nil {
			return nil, err
		}
		doc, err = giu.MergePatch(ouser, user)
		if err != nil {
			return nil, err
		}
	} else {
		doc = user
	}
	njson, err := json.Marshal(&doc)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec("INSERT OR REPLACE INTO users (id, name, json) VALUES (?,?,?)", userId, name, string(njson))
	if err != nil {
		return nil, err
	}
	tx.Commit()
	return db.GetUser(userId)
}

func (db *SqlDB) VerifyUser(username, password string) (heimdall.User, error) {
	var uid string
	var pw string
	err := db.Db.QueryRow("SELECT userid, password FROM users WHERE username = ?", username).Scan(&uid, &pw)
	if err != nil {
		return nil, heimdall.ErrInvalidCredentials
	}
	return db.GetUser(uid)
}

func (db *SqlDB) CreateUser(user heimdall.User) (heimdall.User, error) {
	tx, err := db.Db.Begin()
	if err != nil {
		return user, err
	}
	_, err = tx.Exec("INSERT OR REPLACE INTO users (id,name) VALUES (?,?)", user.GetId(), user.GetName())
	if err != nil {
		return user, err
	}
	if u, ok := user.(*User); ok {
		for k, v := range u.Clients {
			_, err := tx.Exec("DELETE FROM concents WHERE userid = ? AND clientid = ?", user.GetId(), k)
			if err != nil {
				return user, err
			}
			for _, c := range v.Concents {
				_, err := tx.Exec("INSERT OR REPLACE INTO concents (userid,clientid, concent) VALUES (?,?,?)", user.GetId(), k, c)
				if err != nil {
					return user, err
				}
			}
		}
	}
	tx.Commit()
	return user, nil
}

func (db *SqlDB) GetUser(userId string) (heimdall.User, error) {
	u := new(User)
	u.Id = userId
	err := db.Db.QueryRow("SELECT name FROM users WHERE id = ?", userId).Scan(&u.Name)
	if err != nil {
		return u, err
	}

	rows, err := db.Db.Query("SELECT clientid, concent FROM concents WHERE userid = ?", userId)
	for rows.Next() {
		var clientId string
		var concent string
		if err = rows.Scan(&clientId, &concent); err != nil {
			continue
		}
		client := u.Clients[clientId]
		client.Concents = append(client.Concents, concent)
		u.Clients[clientId] = client
	}

	rows, err = db.Db.Query("SELECT clientid, id FROM tokens WHERE userid = ? AND type = 'refresh' AND datetime(expired) > datetime('now')", userId)
	for rows.Next() {
		var clientId string
		var tokenId string
		if err = rows.Scan(&clientId, &tokenId); err != nil {
			continue
		}
		client := u.Clients[clientId]
		client.RefreshTokens = append(client.RefreshTokens, tokenId)
		u.Clients[clientId] = client
	}
	return u, nil
}

func (db *SqlDB) UpdateUser(user heimdall.User) (heimdall.User, error) {
	return db.CreateUser(user)
}

func (db *SqlDB) DeleteUser(userId string) error {
	_, err := db.Db.Exec("DELETE FROM users WHERE id = ?", userId)
	return err
}
