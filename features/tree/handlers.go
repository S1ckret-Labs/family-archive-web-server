package tree

import (
	"database/sql"
	"fmt"
	"github.com/S1ckret-Labs/family-archive-web-server/helpers"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Feature struct {
	Db *sql.DB
}

func (f Feature) GetTree(c *gin.Context) {
	// Validation
	userId, err := helpers.ParamUint64(c, "id")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	rootObjectId, present, err := helpers.QueryOptionalUint64(c, "root_object_id")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if !present {
		// Fetch default root object id for a user
		query, err := f.Db.Query("select root_object_id from Users where user_id = ?;", userId)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		for query.Next() {
			err := query.Scan(&rootObjectId)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
		}
	}

	depth, present, err := helpers.QueryOptionalUint64(c, "depth")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if !present {
		// This is default
		depth = 1
	}

	// Getting data
	query, err := f.Db.Query(`select o.object_id, p.parent_object_id, o.object_key, o.size_bytes, 
		case
			when f.taken_at_sec is not null then 'obj'
			when d.objects_inside is not null then 'dir'
			when a.locked_until_sec is not null then 'zip'
			else '???'
		end as object_type,
		f.taken_at_sec, d.objects_inside, a.locked_until_sec
		from Objects as o
		left join (
			-- Select direct parents for objects
			select descendant as object_id, ancestor as parent_object_id from Paths 
			where path_length = 1
			) as p on p.object_id = o.object_id
		left join Files as f on f.object_id = o.object_id
		left join Directories as d on d.object_id = o.object_id
		left join Archives as a on a.object_id = o.object_id
		where o.object_id in (
			-- Select object_ids for a particular root_object_id with a particular depth.
			select descendant as object_id from Paths 
			where ancestor = ? and path_length != 0 and path_length <= ?
		);`, rootObjectId, depth)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	type ObjectRow struct {
		ObjectId       uint64
		ParentObjectId uint64
		ObjectKey      string
		SizeBytes      uint64
		ObjectType     string
		TakenAtSec     sql.NullInt64
		ObjectsInside  sql.NullInt64
		LockedUntilSec sql.NullInt64
	}

	var objects []map[string]any
	for query.Next() {
		var r ObjectRow
		err := query.Scan(&r.ObjectId, &r.ParentObjectId, &r.ObjectKey, &r.SizeBytes, &r.ObjectType, &r.TakenAtSec,
			&r.ObjectsInside, &r.LockedUntilSec)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if r.ObjectType == "obj" {
			objects = append(objects, map[string]any{
				"Id":             r.ObjectId,
				"ParentObjectId": r.ParentObjectId,
				"Key":            r.ObjectKey,
				"SizeBytes":      r.SizeBytes,
				"Type":           r.ObjectType,
				"TakenAtSec":     r.TakenAtSec.Int64,
			})
		} else if r.ObjectType == "dir" {
			objects = append(objects, map[string]any{
				"Id":             r.ObjectId,
				"ParentObjectId": r.ParentObjectId,
				"Key":            r.ObjectKey,
				"SizeBytes":      r.SizeBytes,
				"Type":           r.ObjectType,
				"ObjectsInside":  r.ObjectsInside.Int64,
			})
		} else if r.ObjectType == "zip" {
			objects = append(objects, map[string]any{
				"Id":             r.ObjectId,
				"ParentObjectId": r.ParentObjectId,
				"Key":            r.ObjectKey,
				"SizeBytes":      r.SizeBytes,
				"Type":           r.ObjectType,
				"LockedUntilSec": r.LockedUntilSec.Int64,
			})
		} else {
			c.AbortWithError(http.StatusInternalServerError,
				fmt.Errorf("encountered unknown object type '%s' for ObjectId '%d'", r.ObjectType, r.ObjectId))
			return
		}
	}

	// Returning data
	c.JSON(http.StatusOK, map[string]any{
		"RootObjectId": rootObjectId,
		"Objects":      objects,
	})
}