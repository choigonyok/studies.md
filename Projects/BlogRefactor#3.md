# #8 쿠버네티스에 배포하기 : 운영환경 Dockerfile 작성
# project blog golang

---

## 개요

블로그 서비스의 기존 배포 방식에 문제가 많아서 어플리케이션을 쿠버네티스로 이전 하기로 했다.

---

## 기존 배포 방식

블로그 서비스는 프론트엔드, 백엔드, 데이터베이스, 리버스프록시인 Nginx로 이루어져있다.

AWS 프리티어로 무료 사용 가능한 m2.micro 인스턴스에 배포했다.

웹서버이기도 한 Nginx의 정적 파일 제공 기능을 통해서 프론트엔드의 빌드파일을 제공하고, 백엔드는 빌드파일을 실행하는 방식으로 배포했다.

운영서버에 원격으로 접속해서 백엔드를 실행시키고 접속을 종료하면, 동시에 실행되던 백엔드도 함께 종료되는 문제가 있어서, nohup으로 백엔드가 백그라운드에서 실행되도록 했다.

### 리소스 부족

운영 중 프론트엔드를 수정해야하는 일이 생겼다.

처음에는 원격서버에 직접 접속해서 수정하고 빌드까지 하려고했다.

1. 원격서버 접속
2. Vim 에디터로 코드 수정
3. 원격서버에서 코드 빌드

이 방식은 실패했다. t2.micro 인스턴스의 스펙은 0.5GiB CPU / 1GB RAM이다. 서버가 리소스 한계에 부딪혀 빌드가 정상적으로 이루어지지 않았다.

다음으로는 로컬에서 수정 후 빌드하는 방식으로 진행해봤다.

1. 로컬에서 코드 수정 후 빌드
2. 깃허브에 푸시
3. 원격서버에 접속
4. 원격서버에서 깃허브 레포지토리 풀

이 방식 역시 실패했다. 깃허브에서는 최대 100MB까지만 푸시를 지원하는데, 프론트엔드에 설치된 의존성 파일의 크기가 커서 푸시가 정상적으로 이루어지지 않았다.

그래서 빌드파일만을 전송하기 위한 레포지토리를 따로 만들기로 했다.

1. 로컬에서 코드 수정 후 빌드
2. 빌드파일만 푸시
3. 원격서버에 접속
4. 원격서버에서 깃허브 레포지토리 풀

이렇게 하니 정상적으로 수정된 코드가 배포될 수 있었다. 문제는 한 번 수정하고 결과를 보기 위해 수많은 과정들이 필요하다는 것이다.

만약 이렇게 수고해서 다시 배포한 코드가 오류가 있다면? 다시 코드를 수정하고 재배포해야한다.

크게는 **운영서버 리소스 부족 문제**와, **운영환경 상 재배포 과정의 문제**를 해결해야했다.

그래서 쿠버네티스를 배포에 적용하기로 했다.

---

## 쿠버네티스 배포의 장점

쿠버네티스로 배포하게 되면 여러 인스턴스를 이용해서 배포할 수 있다.

쿠버네티스 클러스터 내부에서 서비스 디스커버리를 제공하기 때문에, 각 서비스들이 서로 다른 인스턴스에 배포되어있어도 서비스 이름을 통해 쉽게 통신이 가능하다.

또 쿠버네티스로 배포하게 되면 자동확장이 가능하다.

쿠버네티스는 오토스케일링을 지원하기 때문에, 갑자기 늘어난 트래픽에 반응해서 임시로 인스턴스 수를 늘렸다가 줄일 수 있다.

또 쿠버네티스는 컨테이너 이미지를 통해 생성된 컨테이너를 배포하기 때문에, 버전 관리가 쉽다.

만약 어플리케이션을 업데이트했다가 문제가 생기면, 간단하게 이전 버전 이미지를 가져다가 롤백할 수 있다.

이런 장점들과 더불어서, 쿠버네티스를 학습한 내용들을 적용하기 위해서 쿠버네티스로 어플리케이션을 이전하기로 했다.

---

## 운영환경 도커파일 작성

운영환경용 컨테이너 이미지를 빌드하기 위해 도커파일을 작성했다.

개발환경 도커파일 작성은 커플 채팅 서비스를 개발할 때, 블로그 글을 게시해두었다. 

(링크)

```docker
FROM --platform=amd64 golang:latest

WORKDIR /app

COPY ./ ./

RUN go mod download

RUN go build ./src/main.go

EXPOSE 8080

CMD [ "./main" ]
```

--platform=amd64은 배포될 환경에 알맞게 이미지를 빌드할 수 있도록 해준다. 나의 경우엔 리눅스 우분투에 배포할 예정이었기 때문에 amd64로 설정해주었다.

    RUN go build ./src/main.go

개발환경 도커파일과는 다르게 **go run**이 아닌 **go build**로 바이너리 파일을 생성한다.

도커파일이 있는 백엔드 루트 디렉토리의 src 디렉토리 안에 main.go가 있기 때문에 경로를 이런 식으로 설정해준다.

    RUN cd ./src && go build main.go

이런 식으로 main.go가 있는 디렉토리로 이동한 뒤에 빌드하는 것도 가능하다.

대신 둘의 차이점은 있다.

위 방법은 바이너리 파일인 main이 루트 디렉토리에 생기고,
아래방법은 바이너리 파일인 main이 src 디렉토리 안에 생긴다.

빌드 명령을 어디서 시작했는지에 따라 바이너리 파일의 경로가 결정된다.

    EXPOSE 8080

EXPOSE는 컨테이너의 포트를 노출시킨다. 이 포트를 통해서 파드가 컨테이너로 요청을 전달해줄 수 있다.

    CMD [ "./main" ]

이 명령은 바이너리 파일을 실행하는 명령어이다. 개발환경 도커파일 작성 게시글에서 언급했던 것처럼, CMD는 컨테이너가 생성된 이후 실행될 명령어이다.

만약 위의 바이너리파일을 빌드하는 과정에서 

    RUN cd src && go build main.go

로 바이너리 파일을 생성했다면, main 파일의 위치는 도커파일과 같은 레벨에 있지 않기 때문에 

    CMD [ "./src/main" ]

이렇게 작성해줘야 올바르게 바이너리 파일이 실행될 수 있다.

참고로 바이너리 파일의 이름이 main인 이유는, 바이너리 파일 빌드 시 특정 이름을 지정해주지 않으면 소스파일의 이름대로 바이너리 파일 이름이 정해지기 때문이다. 만약 바이너리 파일의 이름을 변경하고 싶다면 빌드할 때,

    -o "NAME"

옵션을 넣어주면 된다. 바꿔주면 당연한 말이지만 CMD 명령을 통해 바이너리를 시작시키는 부분에도 main 대신 설정한 바이너리 파일 이름을 명시해줘야한다.

### 리액트 도커파일

FROM --platform=linux/amd64 node:16-alpine AS build

WORKDIR /app

COPY . .

RUN npm install

RUN npm run build

FROM --platform=linux/amd64 nginx

EXPOSE 3000

COPY ./nginx/default.conf /etc/nginx/conf.d/default.conf

COPY --from=build /app/build /usr/share/nginx/html

# image build시 docker build -t TAG -f build/Dockerfile . 로 빌드해야 부모요소도 컨테이너로 복사가 가능하다.
