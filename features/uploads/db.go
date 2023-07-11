package uploads

import (
	"database/sql"
	"log"
)

type UploadFile struct {
	ObjectId   int64
	ObjectKey  string
	StatusName string
}

func FindUploadRequests(db *sql.DB, userId int64) ([]UploadFile, error) {
	const selectUploadFilesForUser = `select o.request_id, o.object_key, s.status_name from UploadRequests as o 
								  join UploadRequestStatuses as s using (status_id) 
   								  where user_id = ?;`

	rows, err := db.Query(selectUploadFilesForUser, userId)
	if err != nil {
		return nil, err
	}

	result := make([]UploadFile, 0)
	for rows.Next() {
		var file UploadFile
		err := rows.Scan(&file.ObjectId, &file.ObjectKey, &file.StatusName)
		if err != nil {
			return nil, err
		}
		result = append(result, file)
	}
	return result, nil
}

func InsertUploadRequests(db *sql.DB, userId int64, fileNames []string) ([]int64, error) {
	const pendingUploadStatusId = 1
	const sqlStr = `INSERT INTO UploadRequests (user_id, status_id, object_key) VALUES (?, ?, ?)`
	var ids = make([]int64, 0)

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	for _, fileName := range fileNames {
		exec, err := tx.Exec(sqlStr, userId, pendingUploadStatusId, fileName)
		if err != nil {
			_ = tx.Rollback()
			log.Panicln(err)
			return nil, err
		}
		id, err := exec.LastInsertId()
		if err != nil {
			_ = tx.Rollback()
			log.Panicln(err)
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := tx.Commit(); err != nil {
		log.Panicln(err)
	}
	return ids, nil
}
