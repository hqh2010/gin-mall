package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	logging "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"mall/pkg/e"
	util "mall/pkg/utils"
	"mall/repository/db/dao"
	"mall/repository/db/model"
	"mall/serializer"
	"mall/types"
)

var PaymentSrvIns *PaymentSrv
var PaymentSrvOnce sync.Once

type PaymentSrv struct {
}

func GetPaymentSrv() *PaymentSrv {
	PaymentSrvOnce.Do(func() {
		PaymentSrvIns = &PaymentSrv{}
	})
	return PaymentSrvIns
}

func (s *PaymentSrv) PayDown(ctx context.Context, uId uint, req *types.PaymentServiceReq) serializer.Response {
	code := e.SUCCESS

	err := dao.NewOrderDao(ctx).Transaction(func(tx *gorm.DB) error {
		util.Encrypt.SetKey(req.Key)

		payment, err := dao.NewOrderDaoByDB(tx).GetOrderById(req.OrderId)
		if err != nil {
			logging.Info(err)
			return err
		}
		money := payment.Money
		num := payment.Num
		money = money * float64(num)

		userDao := dao.NewUserDaoByDB(tx)
		user, err := userDao.GetUserById(uId)
		if err != nil {
			logging.Info(err)
			code = e.ErrorDatabase
			return err
		}

		// 对钱进行解密。减去订单。再进行加密。
		moneyStr := util.Encrypt.AesDecoding(user.Money)
		moneyFloat, _ := strconv.ParseFloat(moneyStr, 64)
		if moneyFloat-money < 0.0 { // 金额不足进行回滚
			logging.Info(err)
			code = e.ErrorDatabase
			return errors.New("金币不足")
		}

		finMoney := fmt.Sprintf("%f", moneyFloat-money)
		user.Money = util.Encrypt.AesEncoding(finMoney)

		err = userDao.UpdateUserById(uId, user)
		if err != nil { // 更新用户金额失败，回滚
			logging.Info(err)
			code = e.ErrorDatabase
			return err
		}
		boss := new(model.User)
		boss, err = userDao.GetUserById(uint(req.BossID))
		moneyStr = util.Encrypt.AesDecoding(boss.Money)
		moneyFloat, _ = strconv.ParseFloat(moneyStr, 64)
		finMoney = fmt.Sprintf("%f", moneyFloat+money)
		boss.Money = util.Encrypt.AesEncoding(finMoney)

		err = userDao.UpdateUserById(uint(req.BossID), boss)
		if err != nil { // 更新boss金额失败，回滚
			logging.Info(err)
			code = e.ErrorDatabase
			return err
		}

		product := new(model.Product)
		productDao := dao.NewProductDaoByDB(tx)
		product, err = productDao.GetProductById(uint(req.ProductID))
		if err != nil {
			return err
		}
		product.Num -= num
		err = productDao.UpdateProduct(uint(req.ProductID), product)
		if err != nil { // 更新商品数量减少失败，回滚
			logging.Info(err)
			code = e.ErrorDatabase
			return err
		}

		// 更新订单状态
		payment.Type = 2
		err = dao.NewOrderDaoByDB(tx).UpdateOrderById(req.OrderId, payment)
		if err != nil { // 更新订单失败，回滚
			logging.Info(err)
			code = e.ErrorDatabase
			return err
		}

		productUser := model.Product{
			Name:          product.Name,
			CategoryID:    product.CategoryID,
			Title:         product.Title,
			Info:          product.Info,
			ImgPath:       product.ImgPath,
			Price:         product.Price,
			DiscountPrice: product.DiscountPrice,
			Num:           num,
			OnSale:        false,
			BossID:        uId,
			BossName:      user.UserName,
			BossAvatar:    user.Avatar,
		}

		err = productDao.CreateProduct(&productUser)
		if err != nil { // 买完商品后创建成了自己的商品失败。订单失败，回滚
			logging.Info(err)
			code = e.ErrorDatabase
			return err
		}

		return nil

	})

	if err != nil {
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}

	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}
