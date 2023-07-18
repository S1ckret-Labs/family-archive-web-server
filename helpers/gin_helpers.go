package helpers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
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
