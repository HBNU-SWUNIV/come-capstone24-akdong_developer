FROM golang:1.22.3

WORKDIR /server_login

COPY ./login_http_go /server_login

# 모듈 종속성을 정리하고 누락된 모듈을 추가
RUN go mod tidy

# 모듈을 다운로드하고 검증
RUN go mod download && go mod verify

# main.go 파일을 직접 실행
CMD ["go", "run", "/server_login/main.go"]