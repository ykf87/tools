package suggs

import (
	suggestions "tools/runtimes/db/Suggestions"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func SuggCate(c *gin.Context) {
	response.Success(c, suggestions.GetSuggCates(), "success")
}
