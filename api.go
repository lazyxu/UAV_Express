package main

import (
	"errors"
	"math/rand"
	"time"
)

// These four structures is copy of DB-data
type User struct {
	DB_User
}

type Payment struct {
	DB_Payment
}

type Item struct {
	DB_Item
}

type UAV struct {
	DB_UAV
	lock_user_id chan int `json:"-"`
}

func (uav *UAV) LockForUser(userId int) bool {
	select {
	case uav.lock_user_id <- userId:
		return true
	default:
		return false
	}
}

func (uav *UAV) UnLock() bool {
	select {
	case <-uav.lock_user_id:
		return true
	default:
		return false
	}
}

func getUserById(id int) *User {
	return user_id_index.getUserById(id)
}

func (user *User) getRecentPayments(num int) []*Payment {
	return payment_user_id_time_index.getUserLastPayments(user.User_id, 0, num)
}

func getItemById(id int) *Item {
	return item_index.getItemById(id)
}

func getItemList(start, limit int) []*Item {
	return item_index.getItemList(start, limit)
}

func getUAVById(id int) *UAV {
	return uav_index.getUAVById(id)
}

func getUAVList(start, limit int) []*UAV {
	return uav_index.getUAVList(start, limit)
}

func (user *User) getAvailableUAV() *UAV {
	return uav_index.getAvailableUAV(user.User_id)
}

func (user *User) createPayment(pairs []ItemPair) bool {
	price := 0.0
	for _, p := range pairs {
		item := getItemById(p.Item_id)
		price += item.Item_price
	}

	uav := uav_index.getAvailableUAV(user.User_id)
	if uav == nil {
		return false
	}
	defer uav.UnLock()

	payment := Payment{
		DB_Payment{
			Payment_id:      rand.Int(), //TODO: Sync problem?
			Payment_time:    int(time.Now().UnixNano() / 1000000),
			Payment_price:   price,
			Payment_user_id: user.User_id,
			Payment_items:   pairs,
			Payment_number:  "S2Meteor", //TODO:
			Payment_uav_id:  uav.UAV_id,
		},
	}
	payment_user_id_time_index.insertPayment(payment)
	return true
}

func (user *User) Sync() error {
	user_id_index.lock.Lock()
	defer user_id_index.lock.Unlock()
	_userRecord, ok := user_id_index.tree.Get(user.User_id)
	if !ok {
		return errors.New("This user has been deleted!")
	}
	userRecord := _userRecord.(UserRecord)
	*userRecord.DB_User = user.DB_User
	return nil
}

func (item *Item) Sync() error {
	item_index.lock.Lock()
	defer item_index.lock.Unlock()
	_itemRecord, ok := item_index.tree.Get(item.Item_id)
	if !ok {
		return errors.New("This item has been deleted!")
	}
	itemRecord := _itemRecord.(ItemRecord)
	*itemRecord.DB_Item = item.DB_Item
	return nil
}

func (uav *UAV) Sync() error {
	uav_index.lock.Lock()
	defer uav_index.lock.Unlock()
	_uavRecord, ok := uav_index.tree.Get(uav.UAV_id)
	if !ok {
		return errors.New("This uav has been deleted!")
	}
	uavRecord := _uavRecord.(UAVRecord)
	*uavRecord.DB_UAV = uav.DB_UAV
	return nil
}
