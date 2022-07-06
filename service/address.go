package service

import (
	"context"
	logging "github.com/sirupsen/logrus"
	"mall/dao"
	"mall/model"
	"mall/pkg/e"
	"mall/serializer"
	"strconv"
)

type AddressService struct {
	Name    string `form:"name" json:"name"`
	Phone   string `form:"phone" json:"phone"`
	Address string `form:"address" json:"address"`
}

func (service *AddressService) Create(ctx context.Context, uId uint) serializer.Response {
	var address model.Address
	code := e.SUCCESS
	addressDao := dao.NewAddressDao(ctx)
	address = model.Address{
		UserID:  uId,
		Name:    service.Name,
		Phone:   service.Phone,
		Address: service.Address,
	}
	err := addressDao.CreateAddress(address)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	addressDao = dao.NewAddressDaoByDB(addressDao.DB)
	var addresses []model.Address
	addresses, err = addressDao.ListAddressByUid(uId)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Data:   serializer.BuildAddresses(addresses),
		Msg:    e.GetMsg(code),
	}
}

func (service *AddressService) Show(ctx context.Context, uId string) serializer.Response {
	code := e.SUCCESS
	addressDao := dao.NewAddressDao(ctx)
	userId, _ := strconv.Atoi(uId)
	addresses, err := addressDao.ListAddressByUid(uint(userId))
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Data:   serializer.BuildAddresses(addresses),
		Msg:    e.GetMsg(code),
	}
}

func (service *AddressService) Delete(ctx context.Context, aId string) serializer.Response {
	addressDao := dao.NewAddressDao(ctx)
	code := e.SUCCESS
	addressId, _ := strconv.Atoi(aId)
	err := addressDao.DeleteAddressById(uint(addressId))
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
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

func (service *AddressService) Update(ctx context.Context, uid uint, aid string) serializer.Response {
	code := e.SUCCESS

	addressDao := dao.NewAddressDao(ctx)
	address := model.Address{
		UserID:  uid,
		Name:    service.Name,
		Phone:   service.Phone,
		Address: service.Address,
	}
	addressId, _ := strconv.Atoi(aid)
	err := addressDao.UpdateAddressById(uint(addressId), address)
	addressDao = dao.NewAddressDaoByDB(addressDao.DB)
	var addresses []model.Address
	addresses, err = addressDao.ListAddressByUid(uid)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Data:   serializer.BuildAddresses(addresses),
		Msg:    e.GetMsg(code),
	}
}
