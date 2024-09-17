package controller

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/logrusorgru/aurora"
)

var consoleListenerMap = map[string]consoleListeners{
	"help": consoleHelpCommand,
	"map":  consoleMapCommand,
	"user": consoleUserCommand,
	"exit": consoleExitCommand,
}

type consoleListeners func([]string)

var conn *net.UDPConn

func ConsoleController(_conn *net.UDPConn) {
	conn = _conn
	for {
		reader := bufio.NewReader(os.Stdin)
		consoleInput, _ := reader.ReadString('\n')
		// CRLF 해결하려고 -2 함
		// 캐리지 리턴!!!!!!!!! 매우 중요함
		consoleInputSplit := consoleInput[:len(consoleInput)-2]

		inputSlice := strings.Fields(consoleInputSplit)

		if len(inputSlice) == 0 {
			continue
		}

		commandName := inputSlice[0]
		if _, exist := consoleListenerMap[commandName]; !exist {
			go consoleListenerMap["help"](inputSlice)
		} else {
			go consoleListenerMap[commandName](inputSlice)
		}

	}
}

// 도움말 출력
func consoleHelpCommand(consoleInput []string) {
	defer func() {
		fmt.Println("Usage :\n<command> [arguments]\nCommand : map, user, exit")
		fmt.Println("-----------------------------------------------")
	}()
}

// 맵 관련
func consoleMapCommand(consoleInput []string) {

	if len(consoleInput) != 2 || consoleInput[1] == "help" {
		fmt.Println("map all : Output all map ID where the user exists\nmap [MapId] : Print User List of Map ID")
	} else if consoleInput[1] == "all" {
		var allMapIds []string

		for key, _ := range MapidUserList {
			allMapIds = append(allMapIds, key)
		}

		fmt.Printf("Current Open Map ID List :\n%v\n", allMapIds)
	} else {
		if _, exist := MapidUserList[consoleInput[1]]; !exist {
			fmt.Printf("There's no Player in Map [%s]\n", consoleInput[1])
		} else {
			fmt.Printf("Map [%s] User List :\n", consoleInput[1])
			fmt.Printf("%v\n", MapidUserList[consoleInput[1]])
		}
	}
	fmt.Println("-----------------------------------------------")
}

// 접속한 유저관련
func consoleUserCommand(consoleInput []string) {
	fmt.Printf("input command : %v\n", consoleInput)
	if len(consoleInput) == 2 {
		userId := consoleInput[1]
		userAddress, userAddrExist := UserAddr[userId]
		userMap, userMapExist := UserMapid[userId]

		if userAddrExist && userMapExist {
			fmt.Printf("User [%s] Info : \nIP Address : [%s]\nMap ID : [%s]\n", userId, userAddress, userMap)
		} else {
			fmt.Printf("Cannot Found User [%s]\n", userId)
		}
	} else if len(consoleInput) == 3 && consoleInput[2] == "leave" {
		userId := consoleInput[1]
		userAddress, userAddrExist := UserAddr[userId]
		_, userMapExist := UserMapid[userId]

		if userAddrExist && userMapExist {
			// FIXME : 1234 시간 부분 하드코딩 빼야함
			msg := fmt.Sprintf("PlayerLeave$%s;1234", userId)

			udpAddr, err := net.ResolveUDPAddr("udp", userAddress)
			if err != nil {
				fmt.Println(aurora.Sprintf(aurora.Red("Error : Resolve UDP Address Error Occured.\nError Message : %s"), err))
			} else {
				HandleRequest(conn, udpAddr, msg)
			}

		} else {
			fmt.Printf("Cannot Found User [%s]\n", userId)
		}

	} else {
		fmt.Println("user [id] : Print user map and user address")
	}

	fmt.Println("-----------------------------------------------")
}

func consoleExitCommand(consoleInput []string) {
	fmt.Println(aurora.Green("============ Game Server Shut Down ============"))
	os.Exit(0)
}
