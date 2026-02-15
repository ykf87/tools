package users

import (
	"fmt"
	"net/http"
	"tools/runtimes/db"
	"tools/runtimes/db/admins"
	"tools/runtimes/db/medias"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func List(c *gin.Context) {
	admin, err := admins.GetAdminUser(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "请登录", nil)
		return
	}

	dt := new(db.ListFinder)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	list, total := medias.GetMediaUsers(admin.Id, dt)
	rsp := gin.H{
		"list":  list,
		"total": total,
	}

	response.Success(c, rsp, "")
}

func GetTags(c *gin.Context) {
	response.Success(c, medias.GetTags(), "")
}

func GetPlatforms(c *gin.Context) {
	response.Success(c, medias.GetUserPlatforms(), "")
}

func Editer(c *gin.Context) {
	admin, err := admins.GetAdminUser(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "请登录", nil)
		return
	}

	mu := new(medias.MediaUser)
	if err := c.ShouldBind(mu); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if mu.Id < 1 {
		response.Error(c, http.StatusBadRequest, "错误的请求:id", nil)
		return
	}

	// MediaUser不能让改其他数据,因此限制了可编辑的字段
	mmu := medias.GetMediaUserByID(mu.Id)
	if mmu == nil || mmu.Id < 1 {
		response.Error(c, http.StatusBadRequest, "错误的请求:nil", nil)
		return
	}
	// 其他人不允许编辑
	if mmu.AdminID != admin.Id {
		response.Error(c, http.StatusBadRequest, "错误的请求:uid", nil)
		return
	}

	if err := medias.GetDb().Write(func(tx *gorm.DB) error {
		fmt.Println("写入------")
		if err := mmu.EmptyClient(tx); err != nil {
			return err
		}
		if err := mmu.EmptyProxy(tx); err != nil {
			return err
		}
		if err := mmu.EmptyTag(tx); err != nil {
			return err
		}

		for tp, vls := range mu.Clients {
			var mutc []*medias.MediaUserToClient
			for _, v := range vls {
				mutc = append(mutc, &medias.MediaUserToClient{
					MUID:       mu.Id,
					ClientType: tp,
					ClientID:   v,
				})
			}
			if len(mutc) > 0 {
				if err := tx.Create(mutc).Error; err != nil {
					return err
				}
			}
		}

		if len(mu.Proxys) > 0 {
			var mutp []*medias.MediaUserProxy
			for _, v := range mu.Proxys {
				mutp = append(mutp, &medias.MediaUserProxy{
					MUID:    mu.Id,
					ProxyID: v,
				})
			}
			if len(mutp) > 0 {
				if err := tx.Create(mutp).Error; err != nil {
					return err
				}
			}
		}

		if len(mu.Tags) > 0 {
			tgs := medias.AddMUTagsBySlice(mu.Tags)
			var mutt []*medias.MediaUserToTag
			for _, v := range tgs {
				mutt = append(mutt, &medias.MediaUserToTag{
					UserID: mu.Id,
					TagID:  v.ID,
				})
			}
			if len(mutt) > 0 {
				if err := tx.Create(mutt).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	fmt.Println("其他信息处理完成... 保存")

	mmu.Autoinfo = mu.Autoinfo
	mmu.AutoDownload = mu.AutoDownload
	mmu.AutoTimer = mu.AutoTimer
	mmu.DownFreq = mu.DownFreq
	if err := mmu.Save(mmu, medias.GetDb().DB()); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	go mmu.AutoStart()

	// if err := mmu.EmptyClient(tx); err != nil {
	// 	tx.Rollback()
	// 	response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 	return
	// }
	// if err := mmu.EmptyProxy(tx); err != nil {
	// 	tx.Rollback()
	// 	response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 	return
	// }
	// if err := mmu.EmptyTag(tx); err != nil {
	// 	tx.Rollback()
	// 	response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 	return
	// }

	// for tp, vls := range mu.Clients {
	// 	var mutc []*medias.MediaUserToClient
	// 	for _, v := range vls {
	// 		mutc = append(mutc, &medias.MediaUserToClient{
	// 			MUID:       mu.Id,
	// 			ClientType: tp,
	// 			ClientID:   v,
	// 		})
	// 	}
	// 	if len(mutc) > 0 {
	// 		if err := tx.Create(mutc).Error; err != nil {
	// 			tx.Rollback()
	// 			response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 			return
	// 		}
	// 	}
	// }

	// if len(mu.Proxys) > 0 {
	// 	var mutp []*medias.MediaUserProxy
	// 	for _, v := range mu.Proxys {
	// 		mutp = append(mutp, &medias.MediaUserProxy{
	// 			MUID:    mu.Id,
	// 			ProxyID: v,
	// 		})
	// 	}
	// 	if len(mutp) > 0 {
	// 		if err := tx.Create(mutp).Error; err != nil {
	// 			tx.Rollback()
	// 			response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 			return
	// 		}
	// 	}
	// }

	// if len(mu.Tags) > 0 {
	// 	tgs := medias.AddMUTagsBySlice(mu.Tags)
	// 	var mutt []*medias.MediaUserToTag
	// 	for _, v := range tgs {
	// 		mutt = append(mutt, &medias.MediaUserToTag{
	// 			UserID: mu.Id,
	// 			TagID:  v.ID,
	// 		})
	// 	}
	// 	if len(mutt) > 0 {
	// 		if err := tx.Create(mutt).Error; err != nil {
	// 			tx.Rollback()
	// 			response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 			return
	// 		}
	// 	}
	// }

	// mmu.Autoinfo = mu.Autoinfo
	// mmu.AutoDownload = mu.AutoDownload
	// mmu.AutoTimer = mu.AutoTimer
	// mmu.DownFreq = mu.DownFreq
	// if err := mmu.Save(tx); err != nil {
	// 	tx.Rollback()
	// 	response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 	return
	// }
	// tx.Commit()
	mu.Commpare()

	response.Success(c, mu, "")
}

func Delete(c *gin.Context) {

}
