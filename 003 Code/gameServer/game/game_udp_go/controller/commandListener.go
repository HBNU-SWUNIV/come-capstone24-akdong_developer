package controller

import (
	controllerhttp "GameServer/controller-http"
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/logrusorgru/aurora"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Listener는 ReceiveMessage와 연결 정보를 받고
// 서버에 유효한 결과를 도출했는지 bool로 반환함
// 에러가 있을 경우 string도 추가로 반환
type listeners func(*net.UDPConn, ReceiveMessage, string) (bool, string)

/*
func ValidateMessage(m ReceiveMessage) {

}
*/
func ItemLock(userId string, itemId string) {
	mapId := UserMapid[userId]

	_, lockedListResult := MapidLockedList[mapId]
	// if !lockedListResult => 해당 방에 Lock된 아이템이 없음 => 새로 만들어줘야함
	if !lockedListResult {
		MapidLockedList[mapId] = []string{itemId}
	} else { // 해당 방에 Lock 된 아이템이 있음 => 뒤에 append
		MapidLockedList[mapId] = append(MapidLockedList[mapId], itemId)
	}
	LockObjUser[itemId] = userId // 해당 아이템이 userId에 의해 Lock 된 것임을 확인
}

func ItemUnlock(userId string, itemId string) {
	mapId := UserMapid[userId]

	lockedList, _ := MapidLockedList[mapId]

	var itemIndex int

	for index, item := range lockedList {
		if item == itemId {
			itemIndex = index
		}
	}

	if itemIndex != -1 {
		lockedList = append(lockedList[:itemIndex], lockedList[itemIndex+1:]...)
		MapidLockedList[mapId] = lockedList
	}

	delete(LockObjUser, itemId)
	fmt.Printf("Item [%s] Unlocked", itemId)
}

func isLocked(userId string, itemId string) bool {
	mapId := UserMapid[userId]

	lockedList, lockedListResult := MapidLockedList[mapId]
	if !lockedListResult {
		return false
	}
	for _, item := range lockedList {
		if item == itemId {
			return true
		}
	}
	return false
}

func otherMessageLengthCheck(commandName string, messageLength int) bool {
	switch commandName {
	case "AssetCreate":
		return messageLength == 8
	case "AssetMove":
		return messageLength == 4
	case "PlayerJoin", "PlayerMove":
		return messageLength == 2
	case "AssetDelete", "AssetSelect", "AssetDeselect":
		return messageLength == 1
	case "PlayerLeave", "MapReady", "PlayerJump":
		return messageLength == 0
	}
	return false
}

func isUserExists(userId string, addr string) bool {
	_, usrExt := UserAddr[userId]
	_, addExt := AddrUser[addr]
	mapid, mapExt := UserMapid[userId]
	userListExt := false
	if mapExt {
		mapidUserList := MapidUserList[mapid]
		mapidUserIndex := sort.SearchStrings(mapidUserList, userId)

		if mapidUserIndex != len(mapidUserList) {
			userListExt = true
		}
	}
	return usrExt && addExt && mapExt && userListExt
}
func createCreatorList(mapid string) bool {
	datas := CreatorLists{
		MapId:    mapid,
		Creators: []string{},
	}
	_, err := DBClient.Collection("creators").InsertOne(context.TODO(), datas)
	return err == nil
}

func isCreator(userId string) bool {
	if isAdmin(userId) {
		return true
	}

	var datas CreatorLists

	mapid, mapResult := UserMapid[userId]
	if !mapResult {
		return false
	}

	filter := bson.M{"MapId": mapid}

	err := DBClient.Collection("creators").FindOne(context.TODO(), filter).Decode(&datas)
	// fmt.Printf("creatorResult : [%v]\n", creatorResult)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			createCreatorList(mapid)
		}
		return false
	}
	// fmt.Printf("Map [%s] Creators : [%v]", mapid, datas.Creators)
	for _, user := range datas.Creators {
		if userId == user {
			return true
		}
	}
	return false
}

func isAdmin(userId string) bool {
	if userId == "abc" || userId == "abcd" || userId == "abcde" {
		return true
	} else {
		return false
	}
}

func PlayerJoin(conn *net.UDPConn, m ReceiveMessage, addr string) (bool, string) {
	// PlayerJoin 메시지 형태 :
	// PlayerJoin$sendUserId;SendTime;SendUserNickname;MapId
	// otherMessage length: 2

	// otherMessage 길이 체크
	if otherMessageLengthCheck(m.CommandName, len(m.OtherMessage)) {
		mapid := m.OtherMessage[1]

		fmt.Println(
			aurora.Sprintf(
				aurora.Gray(12, "Player [%s] Join Map [%s] | (%s)"), m.SendUserId, mapid, addr))

		_, isUserExists := UserAddr[m.SendUserId]
		_, isAddrExists := AddrUser[addr]
		/*
			fmt.Println(
				aurora.Sprintf(
					aurora.Gray(12, "Users Address : %s"), usersAddress))

			fmt.Println(
				aurora.Sprintf(
					aurora.Gray(12, "Users Name : %s"), usersName))
		*/

		// id와 주소가 접속자 map에 있을 때
		if isUserExists && isAddrExists {
			// 같은 맵에 접속하려고 하는거면 true.
			if UserMapid[m.SendUserId] == mapid {
				fmt.Println("User Join same map")
				return true, aurora.Sprintf(aurora.Green("Success : User [%s] joined Map [%s]"), m.SendUserId, mapid)
			}
		}

		// id와 주소가 접속자 map에 없을 때
		if !isUserExists && !isAddrExists {
			UserAddr[m.SendUserId] = addr
			AddrUser[addr] = m.SendUserId
			UserMapid[m.SendUserId] = mapid

			mapUsers := MapidUserList[mapid]
			fmt.Printf("map Users : %v\n", mapUsers)
			if len(mapUsers) > 0 {
				// IP:Port 형태를 UDPAddr 로 변경해서 저장 시도
				udpAddr, err := net.ResolveUDPAddr("udp", addr)
				if err != nil {
					fmt.Println(aurora.Sprintf(aurora.Red("Error : Resolve UDP Address Error Occured.\nError Message : %s"), err))
				} else {
					for _, user := range mapUsers {
						conn.WriteToUDP([]byte("PlayerJoin$"+user+";12345678;"+user+";"+mapid+";s"), udpAddr)
					}
				}
			}

			mapUsers = append(MapidUserList[mapid], m.SendUserId)
			if len(mapUsers) > 1 {
				sort.Strings(mapUsers)
			}
			MapidUserList[mapid] = mapUsers

			return true, aurora.Sprintf(aurora.Green("Success : User [%s] joined Map [%s]"), m.SendUserId, mapid)
		} else if isUserExists { // 이미 유저ID가 있을 경우
			return false, aurora.Sprintf(aurora.Yellow("Error : User [%s] is already in this game."), m.SendUserId)
		} else if isAddrExists { // 이미 접속한 IP가 등록되어 있었을 경우
			return false, aurora.Sprintf(aurora.Yellow("Error : Address [%s] is already in this game."), addr)
		}
	} else {
		return false, aurora.Sprintf(aurora.Yellow("Error : PlayerJoin Message Length Error : need 2, received %d"), len(m.OtherMessage))
	}
	return false, aurora.Sprintf(aurora.Yellow("Error : Unknown (in PlayerJoin)"))
}

func MapReady(conn *net.UDPConn, m ReceiveMessage, addr string) (bool, string) {
	// MapReady 메시지 형태 :
	// MapReady$sendUserId;SendTime;

	// fmt.Printf("MapReady Start\n")
	if otherMessageLengthCheck(m.CommandName, len(m.OtherMessage)) {
		_, isUserAddrExists := UserAddr[m.SendUserId]

		if isUserAddrExists {
			boolChan := make(chan bool)
			go SendBeforeLog(UserMapid[m.SendUserId], m.SendUserId, boolChan)
			result := <-boolChan
			if result {
				return true, aurora.Sprintf(aurora.Green("Send After Log Complete"))
			} else {
				return false, aurora.Sprintf(aurora.Yellow("Error : Cannot Send After Log"))
			}
		} else {
			return false, aurora.Sprintf(aurora.Yellow("Error : Cannot Found User [%s]"), m.SendUserId)
		}

	} else {
		return false, aurora.Sprintf(aurora.Yellow("Error : MapReady Message Length Error : need 0, received %d"), len(m.OtherMessage))
	}
}

// TODO : 유정이 답장오면 만들기
func SendBeforeLog(mapid string, userId string, result chan bool) {
	mapSaveTimeString := controllerhttp.GetMapTime(mapid)
	mapSaveTime, err := time.Parse(time.RFC3339Nano, mapSaveTimeString.Message)
	if err != nil {
		fmt.Println(aurora.Sprintf(aurora.Red("Error in Parsing Saved Time : %s"), err))
		result <- false
	}

	logResult, err := FindDocumentsAfterTime(mapSaveTime, mapid)
	if err != nil {
		fmt.Println(aurora.Sprintf(aurora.Red("Error in Load Log Result : %s"), err))
		result <- false
	}

	udpAddr, err := net.ResolveUDPAddr("udp", UserAddr[userId])
	if err != nil {
		fmt.Println(aurora.Sprintf(aurora.Red("Error : Resolve UDP Address Error Occured.\nError Message : %s"), err))
		result <- false
	}

	for _, logOne := range logResult {
		// fmt.Printf("Log : %s\n", logOne.OriginalMessage)
		conn.WriteToUDP([]byte(logOne.OriginalMessage+";s"), udpAddr)
		time.Sleep(50 * time.Millisecond)
	}
	result <- true
}

func FindDocumentsAfterTime(parsedTimeFromMaptime time.Time, mapid string) ([]LogResponseData, error) {
	collection := DBClient.Collection("log")

	fmt.Printf("Find start\n\n")
	currentTime := time.Now().UTC()
	// fmt.Printf("Parsed Time : %s\nCurrent Time : %s\nTime Duration : %f\n", parsedTimeFromMaptime, currentTime, currentTime.Sub(parsedTimeFromMaptime).Seconds())

	// fmt.Printf("Parsed Time : %s\nCurrent Time : %s\n", strconv.FormatInt(parsedTimeFromMaptime.UnixMilli(), 10), strconv.FormatInt(currentTime.UnixMilli(), 10))

	filter := bson.M{
		"timestamp": bson.M{
			"$gt": parsedTimeFromMaptime,
			"$lt": currentTime,
		},
		"mapid":           mapid,
		"originalmessage": bson.M{"$regex": "^Asset"},
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []LogResponseData
	// fmt.Printf("Next Cursor Start\n")
	for cursor.Next(context.TODO()) {
		var elem LogResponseData
		err := cursor.Decode(&elem)
		if err != nil {
			fmt.Printf("Error : %s\n", err)
			return nil, err
		}
		// fmt.Printf("LogResponseData : %s\n", elem.OriginalMessage)
		results = append(results, elem)
	}
	if err := cursor.Err(); err != nil {
		fmt.Printf("Error : %s\n", err)
		return nil, err
	}

	return results, nil
}

func PlayerLeave(conn *net.UDPConn, m ReceiveMessage, addr string) (bool, string) {
	// PlayerLeave 형태 :
	// PlayerLeave$SendUserId;SendTime
	// otherMessage length : 0
	if otherMessageLengthCheck(m.CommandName, len(m.OtherMessage)) {
		_, isUserAddrExists := UserAddr[m.SendUserId]
		_, isUserMapidExists := UserMapid[m.SendUserId]

		if isUserAddrExists {
			delete(UserAddr, m.SendUserId)
			delete(AddrUser, addr)
		} else {
			return false, aurora.Sprintf(aurora.Yellow("Error : Cannot Found User [%s]"), m.SendUserId)
		}
		if isUserMapidExists {
			userMapid := UserMapid[m.SendUserId]
			mapidUserList := MapidUserList[userMapid]
			mapidUserIndex := -1

			if len(mapidUserList) == 1 {
				delete(MapidUserList, userMapid)
			} else {
				for index, users := range mapidUserList {
					if users == m.SendUserId {
						mapidUserIndex = index
					}
				}

				if mapidUserIndex != -1 {
					mapidUserList = append(mapidUserList[:mapidUserIndex], mapidUserList[mapidUserIndex+1:]...)
					MapidUserList[userMapid] = mapidUserList
				}
			}

			lockedList, lockedListResult := MapidLockedList[userMapid]
			if lockedListResult {
				for _, item := range lockedList {
					if LockObjUser[item] == m.SendUserId {
						ItemUnlock(m.SendUserId, item)

						unlockMessage := "AssetDeselect$" + m.SendUserId + ";" + strconv.FormatInt(time.Now().UnixMilli(), 10) + ";" + item

						fmt.Printf("Unlock Message : %s\n", unlockMessage)

						structUnlockMessage, err := MessageParser(unlockMessage)

						if !err {
							fmt.Printf("unlock message parsing error\n")
							return false, aurora.Sprintf(aurora.Yellow("Error : Cannot Parsing Unlock Message"))
						}

						logData := LogMessage{
							Timestamp:       time.Now().UTC(),
							MapId:           userMapid,
							Message:         structUnlockMessage,
							OriginalMessage: unlockMessage,
						}

						_, insertErr := DBClient.Collection("log").InsertOne(context.TODO(), logData)
						if insertErr != nil {
							fmt.Printf("insert result : %s\n", insertErr)
						}
					}
				}
			}

			// delete(UserMapid, m.SendUserId)
			fmt.Println(
				aurora.Sprintf(
					aurora.Gray(12, "Player [%s] left Map [%s] | (%s)"), m.SendUserId, userMapid, addr))

			return true, aurora.Sprintf(aurora.Green("Success : User [%s] left Map [%s]\n"), m.SendUserId, userMapid)
		} else {
			return false, aurora.Sprintf(aurora.Yellow("Error : Cannot Found Map ID of User [%s]"), m.SendUserId)
		}
	}

	return false, aurora.Sprintf(aurora.Yellow("Error : Unknown (in PlayerLeave)"))
}

func PlayerMove(conn *net.UDPConn, m ReceiveMessage, addr string) (bool, string) {
	// PlayerMove 형태 :
	// PlayerMove$SendUserId;SendTime;Position;Rotation
	// otherMessage length : 2

	// 보낸 유저가 있는 유저면 return true
	// 딱히 더 할 작업은 없음
	if isUserExists(m.SendUserId, addr) {
		return true, aurora.Sprintf(aurora.Green("Success : User [%s] move\n"), m.SendUserId)
	} else {
		return false, aurora.Sprintf(aurora.Yellow("Error : Cannot Found User [%s]"), m.SendUserId)
	}
}

func AssetCreate(conn *net.UDPConn, m ReceiveMessage, addr string) (bool, string) {
	// AssetCreate 형태 :
	// AssetCreate$SendUserId;SendTime;AssetId;ObjectId;Position;Rotation;Scale;Type;MeshCollider;Rigidbody
	// otherMessage length : 8

	// 보낸 유저가 있는 유저고, Creator List에 있으면 return true
	// 딱히 더 할 작업은 없음
	if otherMessageLengthCheck(m.CommandName, len(m.OtherMessage)) {
		if isUserExists(m.SendUserId, addr) {
			if isCreator(m.SendUserId) {
				return true, aurora.Sprintf(aurora.Green("Success : User [%s] Asset [%s] Create\n"), m.SendUserId, m.OtherMessage[1])
			}
			return false, aurora.Sprintf(aurora.Yellow("Error : User [%s] is not Creator"), m.SendUserId)
		} else {
			return false, aurora.Sprintf(aurora.Yellow("Error : Cannot Found User [%s]"), m.SendUserId)
		}
	} else {
		return false, aurora.Sprintf(aurora.Yellow("Error : AssetCreate Message Length Error : need 8, received %d"), len(m.OtherMessage))
	}
}

func AssetMove(conn *net.UDPConn, m ReceiveMessage, addr string) (bool, string) {
	// AssetMove 형태 :
	// AssetMove$SendUserId;SendTime;ObjectId;Position;Rotation;Scale
	// otherMessage length : 4

	// 보낸 유저가 있는 유저고, Creator List에 있으면 return true
	// 딱히 더 할 작업은 없음
	if otherMessageLengthCheck(m.CommandName, len(m.OtherMessage)) {
		if isUserExists(m.SendUserId, addr) {
			if isCreator(m.SendUserId) {
				itemId := m.OtherMessage[0]
				if isLocked(m.SendUserId, itemId) {
					if LockObjUser[itemId] == m.SendUserId {
						return true, aurora.Sprintf(aurora.Green("Success : User [%s] Asset [%s] Move\n"), m.SendUserId, m.OtherMessage[0])
					} else {
						return false, aurora.Sprintf(aurora.Yellow("Error : Item [%s] is not locked by [%s]"), itemId, m.SendUserId)
					}
				} else {
					return false, aurora.Sprintf(aurora.Yellow("Error : Unavailable Access to Asset Move"))
				}
			}
			return false, aurora.Sprintf(aurora.Yellow("Error : User [%s] is not Creator"), m.SendUserId)
		} else {
			return false, aurora.Sprintf(aurora.Yellow("Error : Cannot Found User [%s]"), m.SendUserId)
		}
	} else {
		return false, aurora.Sprintf(aurora.Yellow("Error : AssetMove Message Length Error : need 8, received %d"), len(m.OtherMessage))
	}
}

func AssetDelete(conn *net.UDPConn, m ReceiveMessage, addr string) (bool, string) {
	// PlayerMove 형태 :
	// PlayerMove$sendUserId;sendTime;ObjectId
	// otherMessage length : 1

	// 보낸 유저가 있는 유저고, Creator List에 있으면 return true
	// 딱히 더 할 작업은 없음
	if otherMessageLengthCheck(m.CommandName, len(m.OtherMessage)) {
		if isUserExists(m.SendUserId, addr) {
			if isCreator(m.SendUserId) {
				itemId := m.OtherMessage[0]
				if isLocked(m.SendUserId, itemId) {
					if LockObjUser[itemId] == m.SendUserId {
						ItemUnlock(m.SendUserId, itemId)
						return true, aurora.Sprintf(aurora.Green("Success : User [%s] Asset [%s] Delete\n"), m.SendUserId, m.OtherMessage[0])
					} else {
						return false, aurora.Sprintf(aurora.Yellow("Error : Item [%s] is not locked by [%s]"), itemId, m.SendUserId)
					}
				} else {
					return false, aurora.Sprintf(aurora.Yellow("Error : Unavailable Access to Asset Delete"))
				}
			}
			return false, aurora.Sprintf(aurora.Yellow("Error : User [%s] is not Creator"), m.SendUserId)
		} else {
			return false, aurora.Sprintf(aurora.Yellow("Error : Cannot Found User [%s]"), m.SendUserId)
		}
	} else {
		return false, aurora.Sprintf(aurora.Yellow("Error : AssetDelete Message Length Error : need 8, received %d"), len(m.OtherMessage))
	}
}

func AssetSelect(conn *net.UDPConn, m ReceiveMessage, addr string) (bool, string) {
	// AssetSelect 형태 :
	// AssetSelect$SendUserId;SendTime;ObjectId;
	// otherMessage length : 1

	// Lock 함. 이미 누가 Lock 해놨으면 false 날리기

	if otherMessageLengthCheck(m.CommandName, len(m.OtherMessage)) {
		if isUserExists(m.SendUserId, addr) {
			if isCreator(m.SendUserId) {
				itemId := m.OtherMessage[0]

				if !isLocked(m.SendUserId, itemId) {
					ItemLock(m.SendUserId, itemId)
					fmt.Printf("MapidLockedList : %v\n", MapidLockedList[UserMapid[m.SendUserId]])
					fmt.Printf("LockObjUser : %s\n", LockObjUser[itemId])
					return true, aurora.Sprintf(aurora.Green("Success : User [%s] Asset [%s] Select\n"), m.SendUserId, m.OtherMessage[0])
				}
				return false, aurora.Sprintf(aurora.Yellow("Error : Item [%s] is Locked"), itemId)
			}
			return false, aurora.Sprintf(aurora.Yellow("Error : User [%s] is not Creator"), m.SendUserId)
		} else {
			return false, aurora.Sprintf(aurora.Yellow("Error : Cannot Found User [%s]"), m.SendUserId)
		}
	} else {
		return false, aurora.Sprintf(aurora.Yellow("Error : AssetSelect Message Length Error : need 1, received %d"), len(m.OtherMessage))
	}
}

func AssetDeselect(conn *net.UDPConn, m ReceiveMessage, addr string) (bool, string) {
	// AssetDeselect 형태 :
	// AssetDeselect$SendUserId;SendTime;ObjectId;
	// otherMessage length : 1

	// Unlock 함.

	if otherMessageLengthCheck(m.CommandName, len(m.OtherMessage)) {
		if isUserExists(m.SendUserId, addr) {
			if isCreator(m.SendUserId) {
				itemId := m.OtherMessage[0]

				if isLocked(m.SendUserId, itemId) {
					lockUser := LockObjUser[itemId]
					if m.SendUserId == lockUser {
						ItemUnlock(m.SendUserId, itemId)
						return true, aurora.Sprintf(aurora.Green("Success : User [%s] Asset [%s] Deselect\n"), m.SendUserId, m.OtherMessage[0])
					} else {
						return false, aurora.Sprintf(aurora.Yellow("Error : Item [%s] is Locked"), itemId)
					}
				} else {
					return false, aurora.Sprintf(aurora.Yellow("Error : Unavailable Access to Asset Deselect"), itemId)
				}

			}
			return false, aurora.Sprintf(aurora.Yellow("Error : User [%s] is not Creator"), m.SendUserId)
		} else {
			return false, aurora.Sprintf(aurora.Yellow("Error : Cannot Found User [%s]"), m.SendUserId)
		}
	} else {
		return false, aurora.Sprintf(aurora.Yellow("Error : AssetSelect Message Length Error : need 1, received %d"), len(m.OtherMessage))
	}
}

func PlayerJump(conn *net.UDPConn, m ReceiveMessage, addr string) (bool, string) {
	// PlayerJump 형태 :
	// PlayerJump$SendUserId;SendTime
	// otherMessage length : 0

	// 보낸 유저가 있는 유저면 return true
	// 딱히 더 할 작업은 없음
	if isUserExists(m.SendUserId, addr) {
		return true, aurora.Sprintf(aurora.Green("Success : User [%s] Jump\n"), m.SendUserId)
	} else {
		return false, aurora.Sprintf(aurora.Yellow("Error : Cannot Found User [%s]"), m.SendUserId)
	}
}
