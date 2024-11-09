# BusyBox 다운로드 및 압축 해제
mkdir /CarteDaemon
mkdir /CarteDaemon/test
cd /CarteDaemon/test

# BusyBox 다운로드
wget https://busybox.net/downloads/busybox-1.36.0.tar.bz2
tar -xvjf busybox-1.36.0.tar.bz2
cd /CarteDaemon/test/busybox-1.36.0
make distclean  # 이전 빌드 정리
make defconfig  # 기본 구성 설정
echo "CONFIG_IP=y" >> .config  # 'ip' 지원이 포함되도록 설정
make -j$(nproc)  # BusyBox 빌드

# 필요한 디렉토리 생성 및 파일 복사
mkdir -p /CarteDaemon/container/testContainer/bin
cp busybox /CarteDaemon/container/testContainer/bin/

# 기본 디렉토리 구조 생성
mkdir -p /CarteDaemon/container/testContainer/{bin,sbin,etc,proc,sys,usr/bin,usr/sbin,dev,tmp,lib,lib64}
mkdir /CarteDaemon/image
mkdir /CarteDaemon/cgroup

# /etc에 필요한 파일 추가
echo "root:x:0:0:root:/root:/bin/sh" > /CarteDaemon/container/testContainer/etc/passwd
echo "root:x:0:" > /CarteDaemon/container/testContainer/etc/group
echo "nameserver 8.8.8.8" > /CarteDaemon/container/testContainer/etc/resolv.conf
echo "127.0.0.1 localhost" > /CarteDaemon/container/testContainer/etc/hosts
echo "testContainer" > /CarteDaemon/container/testContainer/etc/hostname

# /dev 디렉토리에 장치 파일 생성
mknod -m 666 /CarteDaemon/container/testContainer/dev/null c 1 3
mknod -m 666 /CarteDaemon/container/testContainer/dev/zero c 1 5
mknod -m 666 /CarteDaemon/container/testContainer/dev/random c 1 8
mknod -m 666 /CarteDaemon/container/testContainer/dev/urandom c 1 9
mknod -m 666 /CarteDaemon/container/testContainer/dev/tty c 5 0

# 필요한 라이브러리 복사
cp /lib/x86_64-linux-gnu/libc.so.6 /CarteDaemon/container/testContainer/lib/
cp /lib/x86_64-linux-gnu/libm.so.6 /CarteDaemon/container/testContainer/lib/
cp /lib/x86_64-linux-gnu/libresolv.so.2 /CarteDaemon/container/testContainer/lib/

# lib64 디렉토리에 라이브러리 복사
cp /lib64/ld-linux-x86-64.so.2 /CarteDaemon/container/testContainer/lib64/

# 권한 설정
chmod +x /CarteDaemon/container/testContainer/bin/busybox

# /bin/sh를 busybox로 연결 (심볼릭 링크)
ln -s /CarteDaemon/container/testContainer/bin/busybox /CarteDaemon/container/testContainer/bin/sh

# /bin/sh, /bin/ps, /bin/ls, /bin/ip를 busybox로 복사
# cp /CarteDaemon/container/testContainer/bin/busybox /CarteDaemon/container/testContainer/bin/ps
cp /CarteDaemon/container/testContainer/bin/busybox /CarteDaemon/container/testContainer/bin/ls
cp /CarteDaemon/container/testContainer/bin/busybox /CarteDaemon/container/testContainer/bin/ip
cp /CarteDaemon/container/testContainer/bin/busybox /CarteDaemon/container/testContainer/bin/sh

# 이런식으로 써야 복사됨, 안그러면 오류 남
# ip 명령어와 모든 필요한 라이브러리 다시 복사
cp /usr/sbin/ip /CarteDaemon/container/testContainer/bin/
ldd /usr/sbin/ip | grep "=>" | awk '{print $3}' | xargs -I {} cp {} /CarteDaemon/container/testContainer/lib/

cp /usr/bin/ping /CarteDaemon/container/testContainer/bin/
ldd /usr/bin/ping | grep "=>" | awk '{print $3}' | xargs -I {} cp {} /CarteDaemon/container/testContainer/lib/

# 이미지 압축
cd /CarteDaemon/container
tar -cvf /CarteDaemon/image/testImage.tar -C . testContainer

# Cleanup
cd /CarteDaemon/test || exit
rm -rf busybox-1.36.0 busybox-1.36.0.tar.bz2

echo "BusyBox container creation complete."
