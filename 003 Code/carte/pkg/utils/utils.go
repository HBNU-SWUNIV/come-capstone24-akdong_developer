package utils

import (
    "os"
)

// createDirIfNotExists 함수는 지정된 경로에 디렉토리가 없으면 생성합니다.
func CreateDirIfNotExists(dir string) {
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        os.MkdirAll(dir, 0755)
    }
}
