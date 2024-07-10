package controller

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/logrusorgru/aurora"
	"go.mongodb.org/mongo-driver/mongo"
)

// net address <-> user 용이하게 하기 위해 양방향으로 만듬

// *240512 추가*
// Go에서 Map의 Key값으로 구조체를 넣는 경우
// 안에 있는 값이 같더라도 다르게 인식할 수 있음
// 따라서, net address(String) <-> User로 변경함
// + 기본적으로 net.UDPAddr은 key로 사용할 수 있는 타입이 아님.
// 구조체 타입은 필드가 모두 comparable 한 타입이어야만 사용 가능함
// 그게 아니면 포인터를 써야됨

var AddrUser map[string]string = make(map[string]string)
var UserAddr map[string]string = make(map[string]string)
var UserMapid map[string]string = make(map[string]string)
var MapidUserList map[string][]string = make(map[string][]string)
var MapidLockedList map[string][]string = make(map[string][]string)
var LockObjUser map[string]string = make(map[string]string)

var ListenerMap = map[string]listeners{
	"PlayerJoin":    PlayerJoin,
	"PlayerLeave":   PlayerLeave,
	"PlayerMove":    PlayerMove,
	"AssetCreate":   AssetCreate,
	"AssetMove":     AssetMove,
	"AssetDelete":   AssetDelete,
	"AssetSelect":   AssetSelect,
	"AssetDeselect": AssetDeselect,
	"MapReady":      MapReady,
	"PlayerJump":    PlayerJump,
}

var DBClient *mongo.Database

// var timezone, _ = time.LoadLocation("Asia/Seoul")

// 클라이언트 요청 처리 함수
func GetRequest(conn *net.UDPConn) {
	// 클라이언트로부터 메시지 수신
	buf := make([]byte, 1024)
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println(
			aurora.Sprintf(
				aurora.Red("Error : reading from UDP:"), err))
		return
	}

	// 수신한 메시지 파싱
	msg := string(buf[:n])

	go HandleRequest(conn, addr, msg)

}

func HandleRequest(conn *net.UDPConn, addr *net.UDPAddr, msg string) {
	// userID, roomID := parseMessage(msg)

	recvMsg, recvResult := MessageParser(msg)
	if !recvResult {
		fmt.Println(
			aurora.Sprintf(
				aurora.Yellow("Cannot Parsing Message : %s"), msg))
		return
	}
	/*
		if recvMsg.CommandName == "MapReady" {
			fmt.Println("Received message from", addr, ":", string(msg))
		}
	*/

	if recvMsg.CommandName != "" {
		logData := LogMessage{
			Timestamp:       time.Now().UTC(),
			MapId:           UserMapid[recvMsg.SendUserId],
			Message:         recvMsg,
			OriginalMessage: msg,
		}

		if recvMsg.CommandName == "PlayerJoin" {
			logData.MapId = recvMsg.OtherMessage[1]
		}

		_, err := DBClient.Collection("log").InsertOne(context.TODO(), logData)
		if err != nil {
			fmt.Printf("insert result : %s\n", err)
		}
	}

	/*
		fmt.Println(
			aurora.Sprintf(
				aurora.Gray(12, "Receive Command : %s, Send User : %s"), recvMsg.CommandName, recvMsg.SendUserId))
	*/
	// TODO : ValidateMessage 구현
	// ValidateMessage(recvMsg)
	strAddr := addr.String()

	if ListenerMap[recvMsg.CommandName] != nil {

		listenerResult, listenerMsg := ListenerMap[recvMsg.CommandName](conn, recvMsg, strAddr)

		if (recvMsg.CommandName == "AssetCreate") || (recvMsg.CommandName == "AssetDelete") || (recvMsg.CommandName == "AssetSelect") || (recvMsg.CommandName == "AssetDeselect") || (recvMsg.CommandName == "MapReady") {
			fmt.Println(listenerMsg)
		}

		if listenerResult {
			if recvMsg.CommandName == "MapReady" {
				return
			}
			// fmt.Printf("User [%s] : Map [%s]\n\n", recvMsg.SendUserId, UserMapid[recvMsg.SendUserId])
			broadcast(conn, recvMsg.SendUserId, msg, includeSendUserCheck(recvMsg.CommandName))
		} else {
			errorReturn(conn, recvMsg.SendUserId, msg)
		}

		if recvMsg.CommandName == "PlayerLeave" {
			// fmt.Printf("Removed User [%s]\n\n\n", recvMsg.SendUserId)
			delete(UserMapid, recvMsg.SendUserId)
		}

	} else {
		fmt.Println(
			aurora.Sprintf(aurora.Yellow("Unavailable Command Input : %s"), recvMsg.CommandName))
	}

	// fmt.Println("-----------------------------------------------")
}

func includeSendUserCheck(commandName string) bool {
	switch commandName {
	case "PlayerJoin", "PlayerLeave", "AssetCreate", "AssetDelete":
		return true

	case "PlayerMove", "AssetMove", "MapReady", "PlayerJump":
		return false

	}
	return true
}

// 브로드캐스팅 함수
// 인자 목록
// 1. UDP Connection
// 2. 보낸 사람 ID
// 3. 원본 메시지
// 4. 본인 포함 여부
func broadcast(conn *net.UDPConn, sendUserId string, originalMessage string, includeSendUser bool) {

	/*
		fmt.Println(
			aurora.Sprintf(
				aurora.Green("BroadCasting Start From : %s (include Send User : %t)"), sendUserId, includeSendUser))
	*/
	// 보낸 사람이 존재하는 맵 ID
	sendUsersMap := UserMapid[sendUserId]
	// fmt.Printf("Send User Map ID : %s\n", sendUsersMap)

	// 보낸 사람이 존재하는 맵에 있는 유저들
	mapUsers := MapidUserList[sendUsersMap]
	// fmt.Printf("Map ID : %s \nUsers : %v\nUsers Address : [", sendUsersMap, mapUsers)
	/*
		for _, user := range mapUsers {
			fmt.Printf("%s ", UserAddr[user])
		}
		fmt.Printf("]\n")
	*/

	for _, user := range mapUsers {
		if (user == sendUserId) && !includeSendUser {
			continue
		} else {
			udpAddr, err := net.ResolveUDPAddr("udp", UserAddr[user])
			if err != nil {
				fmt.Println(aurora.Sprintf(aurora.Red("Error : Resolve UDP Address Error Occured.\nError Message : %s"), err))
			} else {
				// fmt.Println(originalMessage + ";s")
				go conn.WriteToUDP([]byte(originalMessage+";s"), udpAddr)
			}
		}
	}
}

func errorReturn(conn *net.UDPConn, sendUserId string, originalMessage string) {
	udpAddr, err := net.ResolveUDPAddr("udp", UserAddr[sendUserId])
	if err != nil {
		fmt.Println(aurora.Sprintf(aurora.Red("Error : Resolve UDP Address Error Occured.\nError Message : %s"), err))
	} else {
		go conn.WriteToUDP([]byte(originalMessage+";f"), udpAddr)
	}
	fmt.Println(aurora.Sprintf(aurora.Yellow("BroadCast Skipped : %s\n"), originalMessage))
}

/*
AssetCreate, AssetMove, AssetDelete 는 뒤에 s(success), f(fail)을 붙여서
본인에게 돌려줘야 할 것 같음
*/

/*
TODO : PlayerJoin하면 기존의 플레이어 리스트 보내줘야함
*/
