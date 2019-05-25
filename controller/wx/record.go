package wx

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"
	"yanfei_backend/common"
	"yanfei_backend/controller"
	"yanfei_backend/model"

	"github.com/gin-gonic/gin"
)

// 代表不同的记录类型
const (
	HourRecord = 0
	ItemRecord = 1
)

// AddHourRecord 添加工时记录
// @Summary 添加工时记录
// @Description 添加工时记录
// @Tags 工作记录相关
// @Param token header string true "token"
// @Param 工时记录数据 body model.HourRecordRequest true "工时记录数据"
// @Produce json
// @Success 200 {object} controller.Message
// @Router /wx/record/add_hour_record [post]
func AddHourRecord(c *gin.Context) {
	claims, exist := c.Get("claims")
	// 获取数据失败
	if common.FuncHandler(c, exist, true, common.SystemError) {
		return
	}
	userID := claims.(*model.CustomClaims).UserID

	var hourRecordRequest model.HourRecordRequest
	if common.FuncHandler(c, c.BindJSON(&hourRecordRequest), nil, common.ParameterError) {
		return
	}

	if _, ok := UserExist(c, hourRecordRequest.WorkerID).(model.WxUser); !ok {
		return
	}
	if _, ok := UserExist(c, userID).(model.WxUser); !ok {
		return
	}
	if _, ok := GroupExistByID(c, hourRecordRequest.GroupID).(model.Group); !ok {
		return
	}

	var hourRecord model.HourRecord
	hourRecord.WorkHours = hourRecordRequest.WorkHours
	hourRecord.ExtraWorkHours = hourRecordRequest.ExtraWorkHours

	db := common.GetMySQL()

	var existRecord model.Record
	err := db.Where("group_id = ? AND worker_id = ? AND record_date = ?", hourRecordRequest.GroupID, hourRecordRequest.WorkerID, hourRecordRequest.RecordDate).First(&existRecord).Error
	if common.FuncHandler(c, err != nil, true, common.RecordHasExist) {
		return
	}

	tx := db.Begin()

	err = tx.Create(&hourRecord).Error
	// 数据库错误
	if common.FuncHandler(c, err, nil, common.DatabaseError) {
		// 发生错误时回滚事务
		tx.Rollback()
		return
	}

	var record model.Record
	record.CommonRecord = hourRecordRequest.CommonRecord
	record.AdderID = userID
	record.RecordType = HourRecord
	record.RecordID = hourRecord.ID
	record.AddTime = time.Now().Unix()

	err = tx.Create(&record).Error
	// 数据库错误
	if common.FuncHandler(c, err, nil, common.DatabaseError) {
		// 发生错误时回滚事务
		tx.Rollback()
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, controller.Message{
		Msg: "添加成功",
	})
}

// AddItemRecord 添加分项记录
// @Summary 添加分项记录
// @Description 添加分项记录
// @Tags 工作记录相关
// @Param token header string true "token"
// @Param 分项记录数据 body model.ItemRecordRequest true "分项记录数据"
// @Produce json
// @Success 200 {object} controller.Message
// @Router /wx/record/add_item_record [post]
func AddItemRecord(c *gin.Context) {
	claims, exist := c.Get("claims")
	// 获取数据失败
	if common.FuncHandler(c, exist, true, common.SystemError) {
		return
	}
	userID := claims.(*model.CustomClaims).UserID

	var itemRecordRequest model.ItemRecordRequest
	if common.FuncHandler(c, c.BindJSON(&itemRecordRequest), nil, common.ParameterError) {
		return
	}

	if _, ok := UserExist(c, itemRecordRequest.WorkerID).(model.WxUser); !ok {
		return
	}
	if _, ok := UserExist(c, userID).(model.WxUser); !ok {
		return
	}
	if _, ok := GroupExistByID(c, itemRecordRequest.GroupID).(model.Group); !ok {
		return
	}

	var itemRecord model.ItemRecord
	itemRecord.Subitem = itemRecordRequest.Subitem
	itemRecord.Quantity = itemRecordRequest.Quantity
	itemRecord.Unit = itemRecordRequest.Unit

	db := common.GetMySQL()

	var existRecord model.Record
	err := db.Where("group_id = ? AND worker_id = ? AND record_date = ?", itemRecordRequest.GroupID, itemRecordRequest.WorkerID, itemRecordRequest.RecordDate).First(&existRecord).Error
	if common.FuncHandler(c, err != nil, true, common.RecordHasExist) {
		return
	}

	tx := db.Begin()

	err = tx.Create(&itemRecord).Error
	// 数据库错误
	if common.FuncHandler(c, err, nil, common.DatabaseError) {
		// 发生错误时回滚事务
		tx.Rollback()
		return
	}

	var record model.Record
	record.CommonRecord = itemRecordRequest.CommonRecord
	record.AdderID = userID
	record.RecordType = ItemRecord
	record.RecordID = itemRecord.ID
	record.AddTime = time.Now().Unix()

	err = tx.Create(&record).Error
	// 数据库错误
	if common.FuncHandler(c, err, nil, common.DatabaseError) {
		// 发生错误时回滚事务
		tx.Rollback()
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, controller.Message{
		Msg: "添加成功",
	})
}

// CheckRecorded 检查某日是否记录
// @Summary 检查某日是否记录
// @Description 检查某日是否记录
// @Tags 工作记录相关
// @Param token header string true "token"
// @Param group_id query int true "班组id"
// @Param worker_id query int true "工人id"
// @Param date query string true "日期"
// @Produce json
// @Success 200 {object} controller.Message
// @Router /wx/record/check_recorded [get]
func CheckRecorded(c *gin.Context) {
	_, exist := c.Get("claims")
	// 获取数据失败
	if common.FuncHandler(c, exist, true, common.SystemError) {
		return
	}

	var groupID int64
	var workerID int64
	var err error

	groupID, err = strconv.ParseInt(c.Query("group_id"), 10, 64)
	if common.FuncHandler(c, err, nil, common.ParameterError) {
		return
	}
	workerID, err = strconv.ParseInt(c.Query("worker_id"), 10, 64)
	if common.FuncHandler(c, err, nil, common.ParameterError) {
		return
	}
	date := c.Query("date")

	if _, ok := UserExist(c, workerID).(model.WxUser); !ok {
		return
	}
	if _, ok := GroupExistByID(c, groupID).(model.Group); !ok {
		return
	}

	var record model.Record
	db := common.GetMySQL()

	err = db.Where("group_id = ? AND worker_id = ? AND record_date = ?", groupID, workerID, date).First(&record).Error

	if err == nil {
		switch record.RecordType {
		case HourRecord:
			var hourRecordRequest model.HourRecordRequest
			hourRecordRequest.CommonRecord = record.CommonRecord

			var hourRecord model.HourRecord
			err = db.First(&hourRecord, record.RecordID).Error
			if common.FuncHandler(c, err, nil, common.DatabaseError) {
				return
			}

			hourRecordRequest.WorkHours = hourRecord.WorkHours
			hourRecordRequest.ExtraWorkHours = hourRecord.ExtraWorkHours

			var AdderUser model.WxUser
			var ok bool
			if AdderUser, ok = UserExist(c, record.AdderID).(model.WxUser); !ok {
				return
			}

			var retHourInfo model.RetHourInfo
			retHourInfo.AdderInfo = AdderUser.WxUserInfo
			retHourInfo.AddTime = record.AddTime
			retHourInfo.HourRecordRequest = hourRecordRequest
			c.JSON(http.StatusOK, controller.Message{
				Data: retHourInfo,
			})
			break
		case ItemRecord:
			var itemRecordRequest model.ItemRecordRequest
			itemRecordRequest.CommonRecord = record.CommonRecord

			var itemRecord model.ItemRecord
			err = db.First(&itemRecord, record.RecordID).Error
			if common.FuncHandler(c, err, nil, common.DatabaseError) {
				return
			}

			itemRecordRequest.Subitem = itemRecord.Subitem
			itemRecordRequest.Quantity = itemRecord.Quantity

			var AdderUser model.WxUser
			var ok bool
			if AdderUser, ok = UserExist(c, record.AdderID).(model.WxUser); !ok {
				return
			}

			var retItemInfo model.RetItemInfo
			retItemInfo.AdderInfo = AdderUser.WxUserInfo
			retItemInfo.AddTime = record.AddTime
			retItemInfo.ItemRecordRequest = itemRecordRequest
			c.JSON(http.StatusOK, controller.Message{
				Data: retItemInfo,
			})
			break
		}

	} else {
		c.JSON(http.StatusOK, controller.Message{
			Msg: "无记录",
		})
	}
}

// GetMonthRecords 查看某月的工作记录
// @Summary 查看某月的工作记录
// @Description 查看某月的工作记录
// @Tags 工作记录相关
// @Param token header string true "token"
// @Param month query string true "月份，形如2019-04"
// @Produce json
// @Success 200 {object} controller.Message
// @Router /wx/record/get_month_records [get]
func GetMonthRecords(c *gin.Context) {
	claims, exist := c.Get("claims")
	// 获取数据失败
	if common.FuncHandler(c, exist, true, common.SystemError) {
		return
	}
	userID := claims.(*model.CustomClaims).UserID

	Month := c.Query("month")
	match, _ := regexp.MatchString("\\d{4}-\\d{2}", Month)
	fmt.Println(Month, match)
	if common.FuncHandler(c, len(Month) == 7 && match, true, common.ParameterError) {
		return
	}

	Month = Month + "%"
	db := common.GetMySQL()

	var records []model.Record
	err := db.Where("worker_id = ? AND record_date LIKE ?", userID, Month).Find(&records).Error

	if err == nil {
		var returnRecords []interface{}
		for _, record := range records {
			switch record.RecordType {
			case HourRecord:
				var hourRecordRequest model.HourRecordRequest
				hourRecordRequest.CommonRecord = record.CommonRecord

				var hourRecord model.HourRecord
				err = db.First(&hourRecord, record.RecordID).Error
				if common.FuncHandler(c, err, nil, common.DatabaseError) {
					return
				}

				hourRecordRequest.WorkHours = hourRecord.WorkHours
				hourRecordRequest.ExtraWorkHours = hourRecord.ExtraWorkHours

				returnRecords = append(returnRecords, hourRecordRequest)
				break
			case ItemRecord:
				var itemRecordRequest model.ItemRecordRequest
				itemRecordRequest.CommonRecord = record.CommonRecord

				var itemRecord model.ItemRecord
				err = db.First(&itemRecord, record.RecordID).Error
				if common.FuncHandler(c, err, nil, common.DatabaseError) {
					return
				}

				itemRecordRequest.Subitem = itemRecord.Subitem
				itemRecordRequest.Quantity = itemRecord.Quantity

				returnRecords = append(returnRecords, itemRecordRequest)
				break

			}
		}
		c.JSON(http.StatusOK, controller.Message{
			Data: returnRecords,
		})
	} else {
		c.JSON(http.StatusOK, controller.Message{
			Msg: "无记录",
		})
	}
}
