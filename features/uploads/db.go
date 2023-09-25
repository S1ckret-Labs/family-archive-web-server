package uploads

import (
	"database/sql"
	"gopkg.in/guregu/null.v4"
	"log"
	
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// @Model UploadFile
// @Description Represents an uploaded file's information
// @ID upload-file
// @Property ObjectId uint64 true "unique identifier of the uploaded object"
// @Property ObjectKey string true "name of the uploaded object"
// @Property StatusName string true "status name of the upload"
// @Property SizeBytes uint64 true "size of the uploaded object in bytes"
// @Property TakenAtSec int false "unix timestamp when the object was taken"
type UploadFile struct {
	ObjectId   uint64
	ObjectKey  string
	StatusName string
	SizeBytes  uint64
	TakenAtSec null.Int
}

func FindUploadRequests(db *sql.DB, userId uint64) ([]UploadFile, error) {
	const selectUploadFilesForUser = `select o.request_id, o.object_key, o.size_bytes, o.taken_at_sec ,s.status_name from UploadRequests as o 
								  join UploadRequestStatuses as s using (status_id) 
   								  where user_id = ?;`

	rows, err := db.Query(selectUploadFilesForUser, userId)
	if err != nil {
		return nil, err
	}

	result := make([]UploadFile, 0)
	for rows.Next() {
		var file UploadFile
		err := rows.Scan(&file.ObjectId, &file.ObjectKey, &file.SizeBytes, &file.TakenAtSec, &file.StatusName)
		if err != nil {
			return nil, err
		}
		result = append(result, file)
	}
	return result, nil
}

func InsertUploadRequests(db *sql.DB, userId uint64, uploadRequests []CreateUploadRequest) ([]uint64, error) {
	const pendingUploadStatusId = 1
	const sqlStr = `INSERT INTO UploadRequests (user_id, status_id, object_key, size_bytes, taken_at_sec) VALUES (?, ?, ?, ?, ?)`
	var ids = make([]uint64, 0)

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	for _, request := range uploadRequests {
		exec, err := tx.Exec(sqlStr, userId, pendingUploadStatusId, request.ObjectKey, request.SizeBytes, request.TakenAtSec)
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
		ids = append(ids, uint64(id))
	}
	if err := tx.Commit(); err != nil {
		log.Panicln(err)
	}
	return ids, nil
}

func DeleteUploadRequestsWithS3Objects(db *sql.DB, s3Client *s3.S3, bucketName string, userId uint64, requestIDs []uint64) error {
  tx, err := db.Begin()
  if err != nil {
    return err
  }
  defer tx.Rollback()

  var objectKeys []string
  rows, err := tx.Query("SELECT object_key FROM UploadRequests WHERE user_id = $1 AND id = ANY($2)", userId, requestIDs)
  if err != nil {
    return err
  }
  defer rows.Close()
  for rows.Next() {
    var objectKey string
    if err := rows.Scan(&objectKey); err != nil {
      return err
    }
    objectKeys = append(objectKeys, objectKey)
  }

  _, err = tx.Exec("DELETE FROM UploadRequests WHERE user_id = $1 AND id = ANY($2)", userId, requestIDs)
  if err != nil {
    return err
  }

  err = tx.Commit()
  if err != nil {
    return err
  }

  for _, objectKey := range objectKeys {
    _, err = s3Client.DeleteObject(&s3.DeleteObjectInput{
      Bucket: aws.String(bucketName),
      Key:    aws.String(objectKey),
    })
    if err != nil {
      return err
    }
  }
	
	return nil
}
