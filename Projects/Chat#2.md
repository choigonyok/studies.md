# #2. 풀스택 개발환경 도커파일 설정
# project couple-chat docker

---

## 개요

Docker를 이번 프로젝트부터 도입했다. 도커는 뛰어난 격리성과 간편성을 제공하는 가상화 오픈소스이다.

이 글은 커플 채팅 서비스의 개발환경을 구축하며 작성한 golang Dockerfile.dev 파일을 설명한 글이다.

---

## 도커

---

## Dockerfile

컨테이너는 컨테이너 이미지를 기반으로 생성된다. 

컨테이너 이미지 안에는 독립적인 실행을 위해 필요한 모든 의존성이 다 설치되어있다. 예를 들어, golang 이미지에는 golang과 기본 패키지 등의 golang을 사용하기 위한 모든 사전 설정들이 다 되어있다.

개발자는 그저 필요한 이미지를 가져다가 커스터마이징을 약간 더해 이미지를 빌드하고 컨테이너를 실행하면 된다.

이 Dockerfile 안에서 어떤 이미지를 불러올지, 어떤 커스마이징을 할지, 어떻게 빌드할지에 대한 설정을 할 수 있다. 그럼 도커 데몬과 컨테이너 런타임이 설정파일을 보고 이미지를 생성해 컨테이너가 실행될 수 있다.

### 개발환경과 운영환경의 차이

똑같은 어플리케이션의 이미지를 빌드한다고 하더라도, 개발환경과 운영환경에서의 도커파일은 차이가 있다.

1. OS 차이

개발환경과 운영환경은 OS가 다를 수 있다. 개발은 맥북으로 하는데 운영서버는 우분투라면 OS가 다른 것이다.

각 OS마다 이미지를 최적화하는 방식이 달라서 OS에 맞게 이미지를 빌드해줘야한다. 이 과정에서 도커파일이 달라지게 된다.

2. 실행 방식의 차이

당장 리액트만 보더라도 개발환경에서는

    npm run start

커맨드를 이용하고, 운영환경에서는

    npm run build

커맨드를 이용해 빌드파일을 생성한다.

왜 이런 차이가 있을까?

답은 운영환경과 개발환경의 특성 차이에 있다.

**개발환경은 빠른 재실행이 가능해야한다.** 만약 스트링 출력을 A에서 B로 수정해야하는데 매번 빌드파일을 새로 생성하느라 몇 분, 몇 십 분씩 컴파일 및 빌드 시간이 소요된다면 상당히 불편할 것이다.

**반대로 운영환경은 빠른 재실행보다는 빠른 성능이 중요하다.** 운영환경에서는 상대적으로 개발환경에 비해 수정이 일어나는 횟수가 적다. 또 빌드파일이 한 번 정적으로 생성되면, 클라이언트에서 요청이 왔을 때 뚝딱 빌드파일 내용물을 제공해주면 되기 때문에 상당히 성능이 빠르다. nginx나 apache등의 웹서버가 이런 역할을 해준다.

이런 이유로 개발환경에서는 빌드파일을 생성하지 않고, 운영환경에서는 빌드파일을 생성하는 그런 차이가 있다. 그래서 이 부분에서도 도커파일이 달라진다.

물론 운영환경과 개발환경의 실행방식이 동일하고, 운영/개발 서버의 OS가 동일하다면 굳이 도커파일을 분리할 이유는 없다.

---

## Golang Dockerfile.dev

```dockerfile
FROM golang:alpine3.18

WORKDIR /app

COPY ./go.mod ./
COPY ./go.sum ./

RUN go mod download

RUN go install -mod=mod github.com/githubnemo/CompileDaemon

COPY . .

ENTRYPOINT CompileDaemon -polling -log-prefix=false -build="go build -o main ./cmd" -command="./main" -directory="./"
```

먼저 백엔드의 도커파일을 먼저 살펴보자.

    FROM golang:alpine3.18

golang 이미지를 도커허브(컨테이너 이미지 레지스트리)에서 pull 해온다.

> 컨테이너 이미지 레지스트리 : 컨테이너 이미지 공유저장소, 도커에서 제공하는 공식 이미지들도 있고, 개인이 빌드한 이미지를 올릴 수도 있다. 

alpine3.18 태그로 특정 버전의 golang 이미지를 당겨온다. alpine은 가벼운 이미지버전인데, 꼭 필요한 의존성만 포함하고, 크기를 줄여둔 이미지이다.

    WORKDIR /app

컨테이너에서 작업 기준점을 잡는다. 설정해두면 이후의 작업이 이 WORKDIR를 기준으로 경로설정이 이루어지게 된다.

    COPY go.mod ./
    COPY go.sum ./

이제 로컬의 go.mod와 go.sum을 컨테이너로 복사한다. ./로 상대경로가 지정되어있는데, WORKDIR이 /app 이므로 go.mod와 go.sum은 컨테이너 내부에 /app/go.sum, /app/go.mod 이 경로에 저장되어있다.

### go.mod & go.sum

go.mod와 go.sum이 뭔데 복사하는 걸까?

go.mod는 개발자가 직접 편집이 가능한, 소스파일에서 사용하는 패키지들의 이름과 버전을 명시해둔 파일이다.

개발자는 자신이 사용할 패키지들을 이곳에 정의할 수 있고, 삭제 및 변경도 가능하다. 또는 외부 패키지 다운로드 명령어인 **go get ~**를 이용하면 자동으로 go.mod에 패키지가 추가된다.

go.sum은 go.mod가 정의되어있는 상태에서 **go mod tidy** 커맨드를 통해 현재 go.mod에 쓰여져있는 패지키와 버전들이 신뢰성을 체크한 후 신뢰할만한 패키지들이면 go.mod가 위치한 디렉토리에 go.sum이 생성된다.

go.sum과 go.mod의 정보는 항상 일치해야하고, 일치하지 않으면 컴파일 시 에러가 발생한다.

예를 들어 개발자가 go mod tidy로 go.sum을 생성한 후 go.mod를 수정했다. 그리고 다시 go mod tidy를 하지 않은채로 컴파일을 시도한다면 오류가 발생한다. 사용자가 go.mod에 추가적으로 정의한 패키지에 대한 신뢰성이 go.sum을 통해 확인되지 않기 떄문이다.

이런 방식으로 go.mod와 go.sum을 통해 go에서 쓰여지는 패키지들에 대한 안정성을 확보한다.

---

    RUN go mod download

위에서 설명한 go.mod는 실제 의존성이 설치된 파일이 아니라, 필요한 의존성들의 이름과 버전을 나열해둔 파일이다. 때문에 이 go.mod를 바탕으로 실제 의존성 설치를 해주어야한다. 이 명령어가 바로 **go mod download**이다. go.mod는 설치리스트, go.sum은 인증서라고 이해하면 될 것 같다.

    RUN go install -mod=mod github.com/githubnemo/CompileDaemon

중요한 부분이다. 이 ComepileDaemon은 golang에서 코드 수정사항을 실시간으로 재빌드해주는 오픈소스이다.

### go run main.go VS npm run start

리액트는 실행명령인 npm run start와 빌드파일을 생성하는 npm run build가 가능하다.

node.js의 기반인 자바스크립트는 인터프리터 언어이다. 인터프리터 언어는 변경된 코드만 다시 실행시키는 핫리로딩 또는 라이브리로딩 기능을 지원한다.

그래서 npm run start로 리액트를 실행하면 이후 설명할 도커컴포즈의 볼륨을 적용했을 때 로컬에서 변경된 리액트의 코드가 그대로 컨테이너로 복사되어서 로컬에서 수정한 사항이 컨테이너에 실시간으로 적용될 수 있다.

자비스크립트는 인터프리터 언어이지만 npm run build 명령어를 통해 빌드파일을 만들 때는 컴파일 과정을 거치기 때문에, 한 번 빌드파일을 만들어두면 인터프리터 언어이더라도 수정사항이 적용되지 못한다. 이미 엎질러진 물을 다시 담을 수 없는 것과 같다.

그래서 npm run start는 개발환경에, npm run build는 운영환경에서 일반적으로 사용된다.

go도 개발환경을 위한 go run main.go와 운영환경을 위한 go build main.go 명령어가 존재한다.

npm run start와 go run main.go 둘 다 개발환경을 위한 명령어이지만 차이점이 있다.

go는 인터프리터 언어가 아닌 컴파일 언어이기 때문에, npm run start는 컴파일 없이 실행되고, go run main.go는 컴파일을 한다는 것이다.

컴파일을 한다는 것을 빌드파일을 만든다는 것과 유사하다. 먼저 코드를 실행해 오류가 없는지를 검증하고 바이너리파일을 만들어두기 때문에, 이후에 코드를 수정하더라도 수정한 내용이 적용되지 못한다. 이미 다 포장해버린 선물상자와도 같다.

### go run main.go VS go build main.go

그럼 어차피 둘 다 컴파일을 하는데 왜 run은 개발환경에서 쓰이고, build는 운영환경에서 쓰인다는 걸까? 무슨 차이일까?

차이는 실제 사용가능한 바이너리 파일을 만드느냐 안만드느냐에 있다. 이 파일을 만들게 되면 빌드 시간이 더 오래 걸리게 된다.

개발환경에서는 실제 빌드파일이 필요한 것은 아니기 때문에, 둘 다 컴파일을 하긴 하지만 좀 더 가볍고 빠른 go run main.go를 사용하게 되는 것이다.

---

이런 이유로 인해 개발환경에서 go 코드의 수정을 바로바로 실시간 적용하기 위해서는 그걸 가능하게 해주는 툴이 필요하고, 그 툴로 CompileDaemon이라는 오픈소스 도구를 사용한 것이다.

CompileDaemon은 인터프리터 언어처럼 핫리로딩 기능을 제공해서, 수정된 코드 부분만을 다시 빌드할 수 있게해주는 기능을 제공한다.

    COPY . .

go.mod와 go.sum 이외의 나머지도 컨테이너로 복사해준다. 어차피 COPY . . 하면 go.sum go.mod도 다 COPY가 될텐데 왜 굳이 go.mod와 go.sum을 먼저 COPY했던 걸까?

### 의존성 설치 막기

로컬 프로젝트 디렉토리 안에는 소스파일 말고도 많은 파일들이 존재한다.

디렉토리 구조를 변경한다고 가정하자. 이 때 이미지를 재빌드해줄 필요성이 있다. 소스 코드가 변경된 게 아니기 때문에 CompileDaemon도 이런 변경사항은 감지하지 못한다.

이미지를 재빌드할 때 도커엔진은 수정된 부분부터 재빌드를 시작한다. 만약 go.mod와 go.sum이 다른 파일들과 같이 COPY . . 커맨드를 통해 복사되었다면, 변경된 부분이 go.mod와 go.sum이 아닐 때에도 go.mod와 go.sum까지 재빌드가 되게 된다. 

의존성을 재설치하는 과정은 상대적으로 시간이 훨씬 많이 소요된다.

그래서 go.mod와 go.sum만 먼저 복사해서 의존성을 설치해두고 나머지는 따로 복사해두어서, 이미지를 재빌드해야할 때 go.mod와 go.sum은 재빌드할 필요가 없게하기 위해 따로 분리해서 복사해주는 것이다.

예시처럼 디렉토리 구조를 변경한다면 도커엔진은 이미지를 도커파일 처음부터 다시 빌드하지 않고, COPY . . 부터 빌드할 것이다.

이를 통해서 이미지 빌드 시간을 줄이고 효율적으로 개발을 할 수 있게된다.

---

    ENTRYPOINT CompileDaemon -polling -log-prefix=false -build="go build -o main ./cmd" -command="./main" -directory="./"

CompileDaemon을 컨테이너 내부에서 실행시키는 부분이다.

./cmd 디렉토리 안에 있는 go 소스파일을 빌드해서 main이라는 이름의 바이너리파일을 생성하고, 그 바이너리 파일을 "./main"을 통해 실행한다. directory는 main 빌드파일의 위치이다.

---

다음은 리액트 도커파일을 살펴보자.

## React Dockerfile.dev

```dockerfile
FROM node:16-alpine

WORKDIR /app

COPY ./package.json ./

RUN npm install

COPY . .

CMD ["npm", "run", "start"]
```

프론트엔드의 리액트 도커파일을 Go 도커파일에 비해 상대적으로 간단하다.

    FROM node:16-alpine

리액트를 위해 node alpine 이미지를 pull한다.

    WORKDIR /app

작업 기준점을 /app 으로 설정한다.

    COPY ./package.json ./
    RUN npm install

package.json 역시 golang의 go.mod, go.sum처럼 의존성 파일이다. 마찬가지로 다른 파일들보다 먼저 복사해서 의존성을 설치해준다.

    COPY . .

다른 파일들도 복사해주고,

    CMD ["npm", "run", "start"]

개발환경이므로 빠른 재실행을 위해서 npm run start 명령을 실행하도록 한다. CMD는 컨테이너 안에서 명령어를 실행하게해준다.

---

## Mysql Dockerfile.dev

다음은 데이터베이스를 위한 도커파일이다.

```dockerfile
FROM mysql:latest

COPY ./sqls/my.cnf ./etc/mysql/conf.d/my.cnf

COPY ./sqls /docker-entrypoint-initdb.d/
```

    FROM mysql:latest

mysql 이미지를 다운받는다. :latest 태그는 관용구처럼 쓰이는 건데, 여러 버전 중 가장 최신 버전을 사용하겠다는 의미이다. 실제로 latest라는 태그의 이미지가 있는 게 아니고, 이미지 중 가장 최근에 push된 이미지를 가져온다.

    COPY ./sqls/my.cnf ./etc/mysql/conf.d/my.cnf

이 코드는 데이터베이스 레코드로 한글을 사용할 수 있게 UTF-8 설정을 해주는 설정파일을 컨테이너 내부로 복사하는 것이다.

./sqls/my.cnf는 로컬에 있는 설정파일의 경로이고, ./etc/mysql/conf.d/my.cnf는 컨테이너 안에 복사할 경로이다.

로컬 경로는 경로가 동일하지 않아도 괜찮은데, 컨테이너 안에 복사할 경로는 꼭 저 경로와 저 파일명이어야한다. 그래서 mysql이 정해진 저 경로의 my.cnf파일을 읽고 설정을 적용할 수 있다.

my.cmf의 파일 내용은 아래와 같다.

```
[mysqld]
character-set-server=utf8

[mysql]
default-character-set=utf8

[client]
default-character-set=utf8
```

디렉토리 sqls/에는 my.cnf 말고도 initialize.sql파일이 있다. 이 파일은 mysql의 데이터베이스와 테이블, 컬럼등의 초기 설정을 담고있다. 예시는 아래와 같다.

```
DROP DATABASE IF EXISTS example;

CREATE DATABASE example;

USE example;

CREATE TABLE `usrs` (
        `id` INT NOT NULL PRIMARY KEY,
        `password` VARCHAR(255) NOT NULL DEFAULT "1234");
```

---

## 정리

이렇게 개발환경을 위한 Dockerfile 설정에 대해 알아보았다. 이 도커파일의 이름은 Dockerfile이 아니라 Dockerfile.dev인데, 이 이유는 다음 글인 docker-compose 설정 글에서 설명하겠다.