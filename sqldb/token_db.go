package sqldb

import (
	"github.com/murphysean/heimdall"
	"strings"
)

func (db *SqlDB) NewToken() heimdall.Token {
	t := new(Token)
	t.Id = genUUIDv4()
	return t
}

func (db *SqlDB) CreateToken(token heimdall.Token) (heimdall.Token, error) {
	_, err := db.Db.Exec("INSERT OR REPLACE INTO tokens (id,type,userid,clientid,expires,scope,accesstype,refreshtokenid) VALUES (?,?,?,?,?,?,?,?)", token.GetId(), token.GetType(), token.GetUserId(), token.GetClientId(), token.GetExpires(), strings.Join(token.GetScope(), ","), token.GetAccessType(), token.GetRefreshToken())
	if err != nil {
		return token, err
	}
	return token, nil
}
func (db *SqlDB) GetToken(tokenId string) (heimdall.Token, error) {
	t := new(Token)
	t.Id = tokenId
	var scope string
	err := db.Db.QueryRow("SELECT type,userid,clientid,expires,scope,accesstype,refreshtokenid FROM tokens WHERE id = ?", tokenId).Scan(&t.Type, &t.UserId, &t.ClientId, &t.Expires, &scope, &t.AccessType, &t.RefreshToken)
	t.Scope = strings.Split(scope, ",")
	if err != nil {
		return t, err
	}
	return t, nil
}

func (db *SqlDB) UpdateToken(token heimdall.Token) (heimdall.Token, error) {
	return db.CreateToken(token)
}

func (db *SqlDB) DeleteToken(tokenId string) error {
	_, err := db.Db.Exec("DELETE FROM tokens WHERE id = ?", tokenId)
	return err
}
