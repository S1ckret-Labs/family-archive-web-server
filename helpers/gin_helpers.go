package helpers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

func ParamInt64(c *gin.Context, name string) (int64, error) {
	param, present := c.Params.Get(name)
	if !present {
		return 0, fmt.Errorf("path parameter '%s' is not found", name)
	}
	i, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("path parameter '%s' must be an integer", name)
	}
	return i, nil
}
