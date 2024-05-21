## 개요
- 한밭대학교 공학설계입문 1분반 수업 진행
- 약 40명의 학생들이 실제 서비스 사용

## Login Err
1. too many connections 에러 발생
- login-mysql 서버와 연결과정에서 연결된 클라이언트 일정수치 이상인 경우 발생하는 에러
### 1-1. check_1
- mysql 연결 가능한 최대 클라이언트 갯수 확인
### 1-2. check_2
- mysql 지속시간 확인
