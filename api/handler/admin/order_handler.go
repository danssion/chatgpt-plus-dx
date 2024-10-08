package admin

import (
	"geekai/core"
	"geekai/core/types"
	"geekai/handler"
	"geekai/store/model"
	"geekai/store/vo"
	"geekai/utils"
	"geekai/utils/resp"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	handler.BaseHandler
}

func NewOrderHandler(app *core.AppServer, db *gorm.DB) *OrderHandler {
	return &OrderHandler{BaseHandler: handler.BaseHandler{App: app, DB: db}}
}

func (h *OrderHandler) List(c *gin.Context) {
	var data struct {
		OrderNo  string   `json:"order_no"`
		Status   int      `json:"status"`
		PayTime  []string `json:"pay_time"`
		Page     int      `json:"page"`
		PageSize int      `json:"page_size"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	session := h.DB.Session(&gorm.Session{})
	if data.OrderNo != "" {
		session = session.Where("order_no", data.OrderNo)
	}
	if len(data.PayTime) == 2 {
		start := utils.Str2stamp(data.PayTime[0] + " 00:00:00")
		end := utils.Str2stamp(data.PayTime[1] + " 00:00:00")
		session = session.Where("pay_time >= ? AND pay_time <= ?", start, end)
	}
	if data.Status >= 0 {
		session = session.Where("status", data.Status)
	}
	var total int64
	session.Model(&model.Order{}).Count(&total)
	var items []model.Order
	var list = make([]vo.Order, 0)
	offset := (data.Page - 1) * data.PageSize
	res := session.Order("id DESC").Offset(offset).Limit(data.PageSize).Find(&items)
	if res.Error == nil {
		for _, item := range items {
			var order vo.Order
			err := utils.CopyObject(item, &order)
			if err == nil {
				order.Id = item.Id
				order.CreatedAt = item.CreatedAt.Unix()
				order.UpdatedAt = item.UpdatedAt.Unix()
				list = append(list, order)
			} else {
				logger.Error(err)
			}
		}
	}
	resp.SUCCESS(c, vo.NewPage(total, data.Page, data.PageSize, list))
}

func (h *OrderHandler) Remove(c *gin.Context) {
	id := h.GetInt(c, "id", 0)

	if id > 0 {
		var item model.Order
		res := h.DB.First(&item, id)
		if res.Error != nil {
			resp.ERROR(c, "记录不存在！")
			return
		}

		if item.Status == types.OrderPaidSuccess {
			resp.ERROR(c, "已支付订单不允许删除！")
			return
		}

		res = h.DB.Unscoped().Where("id = ?", id).Delete(&model.Order{})
		if res.Error != nil {
			resp.ERROR(c, "更新数据库失败！")
			return
		}
	}
	resp.SUCCESS(c)
}
