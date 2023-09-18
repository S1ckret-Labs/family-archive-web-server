package helpers

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParamUint64(c *gin.Context, name string) (uint64, error) {
	param, present := c.Params.Get(name)
	if !present {
		return 0, fmt.Errorf("path parameter '%s' is not found", name)
	}
	i, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("path parameter '%s' must be an unsigned integer", name)
	}
	return i, nil
}

func QueryOptionalUint64(c *gin.Context, name string) (uint64, bool, error) {
	param, present := c.GetQuery(name)
	if present {
		i, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			return 0, present, fmt.Errorf("query parameter '%s' must be an unsigned integer", name)
		}
		return i, present, nil

	}
	return 0, present, nil
}
