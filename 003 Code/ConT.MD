## 이미지 수동 저장
- docker save [옵션] <파일명> [이미지명]
- 예시 : docker save -o nginx.tar nginx:latest

## rm: cannot remove 'nginx': Device or resource busy

mount | grep testcontainer // 어떤 파일 시스템이 마운트되어 있는지 확인
umount -l /CarteTest/container/testcontainer/sys // 강제 해제

- mount | grep /CarteTest/image/nginx
- umount /CarteTest/image/nginx

## tarball 이미지 생성
- mkdir -p /tmp/nginx-image
- cd /tmp/nginx-image
- docker run --name nginx-temp -d nginx
- docker export nginx-temp | gzip > nginx.tar.gz
- docker stop nginx-temp
- docker rm nginx-temp


## 이미지 -> 컨테이너
1. 이미지 pull
2. 컨테이너 파일 시스템 구성(레이어 병합, 루트 파일 시스템 생성)
3. 읽기-쓰기 계층 추가(컨테이너가 실행되는 동안 변경사항 저장)
4. 네임 스페이스, Cgroup추가(파일 시스템, 네트워크 격리) - 다른 컨테이너와 격리된 공간에서 사용
5. 프로세스 실행(컨테이너 내부 PID네임스페이스에서 실행), 명령어로 지정된 것을 따라 애플리케이션 실행
6. 리소스 할당 및 관리(cgroup을 통해 컨테이너에 할당된 cpu, 메모리 관리)
7. 컨테이너 실행 상태 유지


## 이미지 사용
### 1. 도커 이미지를 사용하는 경우
- 실제 도커 이미지를 컨테이너로 변환하기 위해서 도커 소스 코드 분석 필요
- 도커 명령어, 도커 API 사용하여 구현하는 것이 효율적

### 2. 자체 컨테이너 생성하는 경우
- 기본 컨테이너 구현 가능
- 특정 애플리케이션이나 프로세스를 격리된 환경에서 실행 가능함
- => 자체 컨테이너 생성으로 진행

## busybox 생성
wget https://busybox.net/downloads/busybox-1.36.0.tar.bz2
tar -xvjf busybox-1.36.0.tar.bz2
cd busybox-1.36.0

make defconfig
make -j$(nproc)

mkdir -p /CarteTest/image/mybusybox/bin
cp busybox /CarteTest/image/mybusybox/bin/

mkdir -p /CarteTest/image/mybusybox/{bin,sbin,etc,proc,sys,usr/bin,usr/sbin,dev,tmp,lib,lib64}

echo "root:x:0:0:root:/root:/bin/sh" > /CarteTest/image/mybusybox/etc/passwd
echo "root:x:0:" > /CarteTest/image/mybusybox/etc/group
echo "nameserver 8.8.8.8" > /CarteTest/image/mybusybox/etc/resolv.conf

ldd /CarteTest/image/mybusybox/bin/busybox

cp /lib/x86_64-linux-gnu/libm.so.6 /CarteTest/image/mybusybox/lib/
cp /lib/x86_64-linux-gnu/libresolv.so.2 /CarteTest/image/mybusybox/lib/
cp /lib/x86_64-linux-gnu/libc.so.6 /CarteTest/image/mybusybox/lib/
cp /lib64/ld-linux-x86-64.so.2 /CarteTest/image/mybusybox/lib64/

chmod +x /CarteTest/image/mybusybox/bin/busybox

ls -l /CarteTest/image/mybusybox/lib/
ls -l /CarteTest/image/mybusybox/lib64/

cd /CarteTest/image
tar -cvf busybox.tar -C mybusybox .
