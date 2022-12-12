package repository

import (
	"Uploader/internal/models"
	logging "Uploader/pkg"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type Repository struct {
	Connection *gorm.DB
	log        logging.Logger
}

func NewRepository(conn *gorm.DB) *Repository {
	return &Repository{Connection: conn}
}

type Temp struct {
	Name     string
	ID       string
	Password string
}

func (r *Repository) Registration(name string, login string, password []byte) error {
	log := logging.GetLogger()

	sqlQwery := `insert into cloud_user(name, login, password)
				values (?,?,?);`

	tx := r.Connection.Exec(sqlQwery, name, login, password)

	if tx.Error != nil {
		log.Println(tx.Error, "help in tx")
		return tx.Error
	}

	return nil
}

func (r *Repository) CreateUserFolder(login string) error {
	log := logging.GetLogger()
	var user Temp
	sqlQuery := `select id from cloud_user where login =?;`
	if err := r.Connection.Raw(sqlQuery, login).Scan(&user).Error; err != nil {
		return err
	}

	sqlQwery := `insert into cloud_folders(name, user_id)
				values (?,?);`

	tx := r.Connection.Exec(sqlQwery, login, user.ID)

	if tx.Error != nil {
		log.Println(tx.Error, "help in tx 'CreateUserFolder'")
		return tx.Error
	}

	return nil

}

func (r *Repository) Login(login string) (*Temp, error) {
	log := logging.GetLogger()
	var user Temp
	sqlQuery := `select name, password, id from cloud_user where login = ?;`

	if err := r.Connection.Raw(sqlQuery, login).Scan(&user).Error; err != nil {
		log.Println(err)
		return nil, err
	}
	if user.Name == "" {
		log.Println("adfasjkkjafdsdkfjs;fdk;jl")
		return nil, errors.New("Takogo logina net uvazhaemiy")
	}

	//log.Println(user.Name, user.ID, user.Password, "fayt")
	if user.Password == "" {
		return nil, errors.New("Takogo palolya net uvazhaemiy")
	}

	return &user, nil
}

func (r Repository) SetToken(token string, userId string) error {
	log := logging.GetLogger()
	//log.Println(token)
	sqlQwery := `insert into cloud_tokens (token, user_id)
				values (?, ?);`

	tx := r.Connection.Exec(sqlQwery, token, userId)
	if tx.Error != nil {
		log.Println("tx error", tx.Error)
		return tx.Error
	}
	return nil
}

type TokenStruct struct {
	ID     string
	UserId string
	Expire time.Time
}

func (r *Repository) ValidateToken(token string) (string, string, error) {
	log := logging.GetLogger()
	var tokenS TokenStruct
	log.Println("324")
	sqlQuery := `select *from cloud_tokens where token =?;`

	if err := r.Connection.Raw(sqlQuery, token).Scan(&tokenS).Error; err != nil {
		return "", "", nil
	}

	TimeChecker := time.Now().After(tokenS.Expire)
	log.Println(tokenS.ID, tokenS.Expire)
	if TimeChecker == true {
		return "", "", errors.New(" TimeChecker is true")
	}
	return tokenS.ID, tokenS.UserId, nil
}

func (r *Repository) FolderCreationForUser(fileName, userId, folderId string) error {
	log := logging.GetLogger()
	log.Println(userId, folderId)
	sqlQwery := `insert into cloud_folders(name, user_id, folder_id)
				values (?, ?, ?);`

	tx := r.Connection.Exec(sqlQwery, fileName, userId, folderId)

	if tx.Error != nil {
		log.Println("tx error", tx.Error)
		return tx.Error
	}
	return nil
}

func (r *Repository) GetFoldersFromParent(count, page int, folder *models.Folder) ([]*models.Folder, error) {
	log := logging.GetLogger()
	var folders []*models.Folder
	offset := (page - 1) * count
	sqlQwery := `select * from cloud_folders where user_id= ?
                              and folder_id= ? order by id limit ? offset ?;`
	tx := r.Connection.Raw(sqlQwery, folder.UserID, folder.FolderID, count, offset).Scan(&folders)
	if tx.Error != nil {
		log.Println("tx error", tx.Error)
		return nil, tx.Error
	}
	//log.Println("test", folders.Name, folders.FolderID, "one")

	return folders, nil
}

//coalesce(cd.folder_id, '')

type folderStruct struct {
	Id       string
	FolderId string
}

func (r *Repository) GetParentFolders(folder *models.Folder) (string, []*models.Folder, error) {
	log := logging.GetLogger()
	var folders []*models.Folder
	var id folderStruct
	sql := `select id, coalesce((select coalesce(folder_id, null))::text, ' ') from cloud_folders
		where user_id= ?;`
	tx := r.Connection.Raw(sql, folder.UserID).Scan(&id)
	if tx.Error != nil {
		log.Println("tx error", tx.Error)
		return "", nil, tx.Error
	}
	log.Println(id.FolderId, id.Id, "hehe")

	sqlQwery := `select *from cloud_folders where folder_id= ?;`
	tx2 := r.Connection.Raw(sqlQwery, id.Id).Scan(&folders)
	if tx2.Error != nil {
		log.Println("tx error", tx2.Error)
		return "", nil, tx2.Error

	}
	//log.Println("test", folders.Name, folders.FolderID, "one")
	log.Println(id.Id, folders)
	return id.Id, folders, nil
}

func (r *Repository) GetFiles(count, page int, file *models.File) ([]*models.File, error) {
	log := logging.GetLogger()
	var files []*models.File
	offset := (page - 1) * count
	sqlQwery := `select *from cloud_files where folder_id=? order by id limit ? offset ?;`
	tx := r.Connection.Raw(sqlQwery, file.FolderID, count, offset).Scan(&files)
	if tx.Error != nil {
		log.Println("tx error", tx.Error)
		return nil, tx.Error
	}
	log.Println(files)
	return files, nil
}

func (r *Repository) UploadFile(name, userId, folderId string) error {
	log := logging.GetLogger()
	//log.Println(name, url, userId, folderId)
	sqlQwery := `insert into cloud_files(name, user_id, folder_id)
				values (?, ?, ?); `

	tx := r.Connection.Exec(sqlQwery, name, userId, folderId)
	if tx.Error != nil {
		log.Println("error in uploadFile", tx.Error)
		return tx.Error
	}
	return nil
}

func (r *Repository) DownloadFiles(id string) (*models.File, error) {
	log := logging.GetLogger()
	var files *models.File
	sqlQwery := `select *from cloud_files where id= ?;`
	tx := r.Connection.Raw(sqlQwery, id).Scan(&files)
	if tx.Error != nil {
		log.Println("tx error", tx.Error)
		return nil, tx.Error
	}
	return files, nil
}

func (r *Repository) ValidationForDownload(files *models.File) (string, error) {
	log := logging.GetLogger()

	sqlQwery := `select user_id from cloud_files where id= ?;`
	tx := r.Connection.Raw(sqlQwery, files.ID).Scan(&files)
	if tx.Error != nil {
		log.Println("tx error", tx.Error)
		return "", tx.Error
	}
	//log.Println("test", folders.Name, folders.FolderID, "one")

	return files.UserID, nil
}

func (r *Repository) ChangeFileName(files *models.File) error {
	log := logging.GetLogger()

	sqlQwery := `update cloud_files set name= ? where id = ?;`
	tx := r.Connection.Raw(sqlQwery, files.Name, files.ID).Scan(&files)
	if tx.Error != nil {
		log.Println("tx error", tx.Error)
		return tx.Error
	}

	return nil
}

func (r *Repository) ChangeFolderName(folder *models.Folder) error {
	log := logging.GetLogger()

	sqlQwery := `update cloud_folders set name= ? where id = ?;`
	tx := r.Connection.Raw(sqlQwery, folder.Name, folder.ID).Scan(&folder)
	if tx.Error != nil {
		log.Println("tx error", tx.Error)
		return tx.Error
	}

	return nil
}

func (r *Repository) DeleteFile(files *models.File) error {
	log := logging.GetLogger()
// todo add select file
	sqlQwery := `DELETE FROM cloud_files where id=?;`
	tx := r.Connection.Raw(sqlQwery, files.ID).Scan(&files)
	if tx.Error != nil {
		log.Println("tx error", tx.Error)
		return tx.Error
	}

	return nil
}
