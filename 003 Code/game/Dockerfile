# Go 언어를 기반으로 한 빌드 환경 설정
FROM golang:1.21.5

# 작업 디렉토리 설정
WORKDIR /server_game

# 종속성 다운로드 및 정리
# 소스 코드 복사
COPY ./game_udp_go/go.mod ./game_udp_go/go.sum /server_game/
RUN go mod tidy
RUN go mod download && go mod verify

# 소스 코드 복사
COPY ./game_udp_go /server_game

# 필요한 패키지 설치
RUN go get github.com/mattn/go-isatty@v0.0.19

# 애플리케이션 실행
CMD ["go", "run", "main.go"]
