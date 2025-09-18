package user

import (
	"net/http"
	"tools/runtimes/db/admins"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"
	"tools/runtimes/parses"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func Info(c *gin.Context) {
	u, ok := c.Get("_user")
	if !ok {
		response.Error(c, http.StatusNotFound, i18n.T("Please Login first"), nil)
		return
	}
	adm := u.(*admins.Admin)
	r, _ := parses.Marshal(adm, c)
	response.Success(c, r, "success")
}

func Editer(c *gin.Context) {
	adm := new(admins.Admin)
	if err := c.ShouldBind(adm); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	u, ok := c.Get("_user")
	if !ok {
		response.Error(c, http.StatusNotFound, i18n.T("Please Login first"), nil)
		return
	}
	dbadm := u.(*admins.Admin)

	if dbadm.Main == 1 {
		if adm.Status == 0 {
			response.Error(c, http.StatusNotFound, i18n.T("Superusers cannot modify"), nil)
			return
		}
	}

	if adm.Password != "" {
		adm.Password, _ = funcs.GenPassword(adm.Password, 0)
	}

	if err := adm.Save(nil); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(c, adm, "")
}

type removeStruct struct {
	Id int64 `json:"id"`
}

func Remove(c *gin.Context) {
	u, ok := c.Get("_user")
	if !ok {
		response.Error(c, http.StatusBadRequest, i18n.T("Please Login first"), nil)
		c.Abort()
		return
	}

	rm := new(removeStruct)
	if err := c.ShouldBind(rm); err != nil {
		response.Error(c, http.StatusBadRequest, i18n.T("Error"), nil)
		return
	}
	us := u.(*admins.Admin)
	if us.Id == rm.Id {
		response.Error(c, http.StatusBadRequest, i18n.T("Can't delete yourself"), nil)
		return
	}

	if rm.Id < 1 {
		response.Error(c, http.StatusBadRequest, i18n.T("Error"), nil)
		return
	}

	if err := admins.DeleteAdminById(rm.Id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, gin.H{"status": "ok"}, "Success")
}
