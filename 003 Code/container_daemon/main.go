package main

import(
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
)

// test : host name 출력
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hostname, err := os.Hostname()
	if err == nil {
		fmt.Fprint(w, "Welcome! "+hostname+"\n")
	} else {
		fmt.Fprint(w, "Welcome! Error\n")
	}
}
func main() {
	router := httprouter.New()
	router.GET("/", Index)

	log.Fatal(http.ListenAndServe(":8080", router))
}

// Carte_Daemon 실행(서버, 컨테이너 생성 구현), Carte_Client 실행(이미지 전달)
// 시스템 호출, 네임 스페이스,, fork 부모 자식 프로세스 필요
