package filedb

type FileDB struct {
	Directory string
}

func NewFileDB(dir string) *FileDB {
	db := new(FileDB)
	db.Directory = dir
	return db
}
