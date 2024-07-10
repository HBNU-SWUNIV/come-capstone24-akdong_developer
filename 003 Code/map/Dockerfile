FROM golang:1.22.2

WORKDIR /server_map

COPY ./map_http_go /server_map

# 모듈 종속성을 정리하고 누락된 모듈을 추가
RUN go mod tidy

# 모듈을 다운로드하고 검증
RUN go mod download && go mod verify

# 필요한 패키지 설치
RUN go get github.com/mattn/go-isatty@v0.0.19

# main.go 파일을 직접 실행
CMD ["go", "run", "/server_map/main.go"]
