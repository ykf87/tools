package task

import (
	"tools/runtimes/db/tasks"

	"github.com/gin-gonic/gin"
)

func List(c *gin.Context) {
	tasks.GetTasks(1, 10, 1)
}
