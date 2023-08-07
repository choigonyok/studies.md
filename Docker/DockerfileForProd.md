## 풀스택 모놀리식 서비스를 위한 운영환경 도커파일 작성

---

## 개요

커플 채팅 서비스를 쿠버네티스 클러스터에 배포하기 위해서 도커파일을 작성해야했다.
도커파일을 처음 작성해보면서 배운 것들을 정리해보고 그 과정을 소개하겠다.

우선, 도커와 쿠버네티스를 이용한 어플리케이션 배포의 대략적인 큰 그림은 이렇게 된다.
코드 작성, 코드를 도커 이미지로 빌드, 도커 이미지를 컨테이너 이미지 레지스트리에 푸시, 쿠버네티스에서 이미지를 풀해서 pod 생성, pod와 서비스 연결
이 과정 중에서 도커파일은 코드를 도커 이미지로 빌드하는 데 쓰인다.
도커는 go, node, nginx, mysql 등 다양한 이미지를 제공한다.
이 기본 이미지를을 가지고 나만의 이미지를 빌드하면 그 이미지는 어느 환경에서든 동일하게 안정적인 운영이 가능해진다

---

## Go 백엔드 도커파일

    FROM --platform=linux/amd64 golang:latest

> FROM : 도커 이미지를 도커 허브에서 pull 해오겠다는 것이다.
> --platform=linux/amd64 : 이 이미지가 amd64 리눅스 OS 위에서 운영될 것이라고 명시한다.
> latest : 콜론 뒤의 내용은 버전을 의미한다. latest는 예약어로 쓰이는데 현재 레지스트리에 푸시되어있는 이미지중 가장 최신 버전으로 알아서 풀 하겠다는 의미이다. 


기본적으로 아무 옵션없이 이미지를 당겨오면 현재 로컬 os환경에 맞는 이미지가 당겨와진다.
근데 운영환경과 개발환경이 다르다면 운영환경에 맞는 이미지를 당겨오는 것이 안정적일 것이다.

golang은 크로스 플랫폼을 지원한다. 이따 코드가 또 나오겠지만
GOOS=linux GOARCH=amd64
이런식으로 환경변수를 설정해주는 것만으로 알아서 환경에 맞게 코드를 실행할 수 있다.
따라서 FROM --platform=linux/amd64 라고 명시하지 않아도 리눅스 환경에서 코드가 잘 작동한다.
그런데도 이 코드를 넣은 이유가 있다.
쿠버네티스 클러스터에 백엔드를 배포하고 내부 디렉토리 확인이나 다른 서비스간 통신을 확인하기 위해 컨테이너 내부로 bash를 통해 진입할 일이 생기는데,
이미지가 리눅스를 위한 이미지가 아니기 때문에 bash접근이 막힌다.

    WORKDIR /app

컨테이너 내부에서 작업할 디렉토리를 설정한다.
이 커맨드 이후로는 이 디렉토리에서 작업을 계속 수행하게된다.
기본 컨테이너 이미지에 있는 디렉토리면 그 안에서 하겠지만, 그렇지 않다면 저 경로를 만드는 것 같다.

    COPY . .

go.mod는 개발자가 편집 가능한, 패키지 이름과 버전 명시 파일이다.
go mod tidy 커맨드를 통해 현재 go.mod에 쓰여져있는 패지키와 버전들이 신뢰성을 체크하고 안전하면 go.sum이 생성된다.
go.sum과 go.mod의 정보는 항상 일치해야하고 이를 통해 go에서 쓰여지는 패키지들에 대한 안정성을 확보한다.
만약 일치하지 않으면 컴파일 시 오류가 발생한다.
현재 로컬에 있는 go.mod와 go.sum을 컨테이너에 복사한다.
이때 로컬디렉토리의 기준은 dockerfile이 위치한 곳을 기준으로 한다.
때에 따라서 COPY ../go.mod ./ 이렇게 변경될 수도 있는 것이다.
workdir이 /app이니까 이 커맨드를 실행하면 /app/go.mod와 /app/go.sum이 존재하게 될 것이다.
실제 패키지들이 담겨있는게 아니라 패키지에 대한 정보가 담겨있는 것이기 때문에 로컬의 파일을 그대로 복사해준다.

    RUN go mod download
go mod download는 go.mod와 go.sum의 내용을 바탕으로 컨테이너에서 의존성을 설치하는 명령어이다.
로컬에서는 패키지별로 go get ~ 커맨드를 통해 개별적으로 설치하는 경우가 대부분인데, 컨테이너에서는 모든 의존성을 한 번에 설치해준다.

    RUN cd src && GOOS=linux GOARCH=amd64 go build main.go
main 패키지가 있는 곳으로 디렉토리를 이동하면서 동시에 빌드파일을 생성한다.
&&를 쓰지 않고 두 명령어를 따로따로 수행하게 되면 오류가 생긴다.
RUN cd src를 하고서 다시 

    EXPOSE 8080
컨테이너의 port를 8080으로 열어준다.
이 포트는 클러스터에 컨테이너가 배포되면 서비스의 targetport로 작동한다.
백엔드에서 listen하고 있는 port와 동일하게 설정해야 컨테이너의 8080포트로 요청이 왔을 때 백엔드가 응답할 수 있다.

    CMD ["./src/main"]
./main 커맨드를 실행해서 빌드된 파일이 실행되도록 한다.


---

## React 프론트엔드 도커파일

    FROM --platform=linux/amd64 node:16-alpine AS build
# 이미지를 node:10으로 당겨왔다가 1시간동안 뻘짓함
# 도커 이미지 빌드 중 npm run build에서 막히길래 한참 찾았음
alpine 이미지는 조금 더 경량화된 컨테이너 이미지이다.
AS build는 이후에 nginx image를 당겨온 후에 nginx 이미지 안의 특정 경로에 빌드파일을 넣어줘야해서 체크포인트같은 느낌

    WORKDIR /app

    COPY . .

    RUN npm install

    RUN npm run build
# axios 등 다른 패키지를 사용하고 도커 이미지를 만드려면 package.json에 해당 패키지를 작성해줘야 함

    FROM --platform=linux/amd64 nginx

    EXPOSE 3000

    COPY ./nginx/default.conf /etc/nginx/conf.d/default.conf

    COPY --from=build /app/build /usr/share/nginx/html 
빌드 파일을 여기 넣어줘야 nginx가 빌드파일을 제공할 수 있음


---

## Mysql 데이터베이스 도커파일

    FROM mysql:latest

    ADD ./sqls/my.cnf ./etc/mysql/conf.d/my.cnf
# 한글 인식할 수 있게 설정 변경
# my.conf 아니고 my.cnf임!

---

## Nginx 리버스프록시 도커파일

    FROM nginx:latest

# FROM nginx

    COPY ./default.conf /etc/nginx/conf.d/default.conf

---

## 도커파일 빌드

    docker build -t <docker hub id>/<docker hub registry name> . 
. 이 중요함. 도커파일이 위치한 전체 파일을 이미지로 빌드하겠다는 뜻임

---

## 컨테이너 이미지 레지스트리 푸시

    docker push <docker hub id>/<docker hub registry name>