# UDP Game Server with Go

※ Go Version : 1.21.5

## Project Structure

```
GameServer
|- controller
|   |- commandListener.go
|   |- controller.go
|   |- parser.go
|- go.mod
|- main.go
|- README.md
```

## Excute Sequence

```
main.main
-> controller.GetRequest
-> controller.HandleRequest
-> parser.MessageParser
-> controller.ValidateMessage
-> controller.BrodcastMessage
```

## Data in CommandListener
```go
type ReceiveMessage struct {
	CommandName  string
	SendUserId   string
	SendTime     string
	OtherMessage []string
}
```
example input :  
PlayerJoin$UserId;DataInputTime;UserNickname;MapId

```go
exampleStructure := ReceiveMessage{
    CommandName : "PlayerJoin",
    SendUserId : "UserId",
    SendTime : "DataInputTime",
    OtherMessage : ["UserNickname" "MapId"]
}
```

## Data in Controller
|Map Name|Key|Value|
|---|---|---|
|AddrUser|udp address of user (*net.UDPAddr)|user id (string)|
|UserAddr|user id (string)|udp address of user (*net.UDPAddr)|
|UserMapid|user id (string)|map id that user joined (string)|
|MapidUserList|map id (string)|user id's (string array)|
|ListenerMap|command name (string)|listener (function)

AddrUser <-> UserAddr  
양방향 접근 다수 예상

## Data Communication Example


input : 
PlayerJoin$UserId;DataInputTime;UserNickname;MapId

1. Get Command Name using MessageParser  

2. Validate Message using ValidateMessage  

3. Send Data to All Users where in same map   using BroadcastMessage  


