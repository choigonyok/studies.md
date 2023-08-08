## PROD DOCKERFILE FOR FULLSTACK MONOLITHIC WEB APPLICATION

---

## 개요

커플 채팅 서비스를 컨테이너로 배포하게 되면서 도커파일을 작성해야했다. 도커파일은 개발환경에서는 도커컴포즈에 쓰이고, 운영환경에서는 쿠버네티스 클러스터 오브젝트에 쓰였다.

도커파일을 처음 작성해보면서 배운 것들을 정리해보고, 개발환경와 운영환경으로 나누어서 도커파일 작성 과정을 기술하려고한다.

우선, 도커와 쿠버네티스를 이용한 어플리케이션 배포의 대략적인 큰 그림은 이렇게 된다.

    운영환경 : 코드 작성 -> 컨테이너 이미지 빌드 -> 컨테이너 이미지 레지스트리 푸시 -> 쿠버네티스 오브젝트 생성시 컨테이너 이미지 레지스트리 풀 -> 서비스와 오브젝트 연결
    개발환경 : 코드 작성 -> 도커 컴포즈 생성시 컨테이너 이미지 빌드 -> 컨테이너 실행

이 과정 중에서 도커파일은 컨테이너 이미지를 빌드하는 데 쓰인다.

도커는 go, node, nginx, mysql 등 다양한 이미지와 버전들을 제공한다.
이 이미지들을 가지고 나만의 이미지를 빌드하면 그 이미지는 어느 환경에서든 동일하게 안정적인 운영이 가능해진다.

도커파일을 작성하기 위해서는 프로젝트의 루트 디렉토리에 "Dockerfile"이라는 파일을 생성해주고, 그 안에 도커파일의 내용을 작성해주면 된다. 이제 내가 진행한 프로젝트의 도커파일을 살펴보겠다.

---

## GO 백엔드 운영환경 도커파일

    FROM --platform=linux/amd64 golang:latest

> FROM : 컨테이너 이미지를 도커허브에서 pull 해오는 커맨드

> 도커허브 : 도커에서 제공하는 컨테이너 이미지 레지스트리

--platform=linux/amd64 : 이 이미지가 amd64 리눅스 OS 위에서 운영될 것이라고 명시한다. 기본적으로 아무 옵션없이 이미지를 pull하게 되면 도커데몬이 실행되고있는 호스트에 맞는 이미지가 pull된다.

나의 경우는 MacOS를 사용중이기 때문에 이에 맞는 golang이미지가 불러와질 것이다. 그러나 내가 원하는 건 MacOS에서 잘 작동하는 이미지가 아닌 운영환경인 우분투(linux/amd64)에서 잘 작동하기를 원하기 때문에, 이렇게 명시해준다.

콜론 뒤의 값은 버전을 의미한다. 특정버전을 지정하고 싶다면 버전명을 확인 후 명시해주면 된다. latest는 가장 최신 버전의 이미지를 pull하는 것이다.

golang은 크로스 플랫폼을 지원한다. golang은 low-level의 컴파일이 이루어지기 때문에 환경에 크게 제약받지 않고 여러 환경에서도 잘 동작한다. golang의 장점 중 하나이다.
만약 특정 환경에서만 쓰이는 코드를 작성했다면 해당 환경을 명시해주기 위해서 

> GOOS=linux GOARCH=amd64
        
이런식으로 환경변수만 설정해주면 알아서 환경에 맞게 코드를 컴파일할 수 있다.

이 말은 곧 --platform=linux/amd64 라고 옵션을 주지않아도 MacOS에서 작성한 golang 코드가 리눅스 환경에서 코드가 잘 작동한다는 것이다. 그럼 굳이 왜 도커파일에서 golang 이미지를 pull할 때 옵션을 지정하는 것일까?
실제로 해당 옵션 없이 클러스터에 배포를 해도 코드는 정상적으로 잘 동작한다. 그러나 클러스터 상에서 백엔드 컨테이너 내부 bash로 들어가려고하면 이미지와 환경이 충돌해서 컨테이너 내부로 접근할 수 없다는 오류가 발생한다.
코드는 잘 작동하더라도 컨테이너로 정상적 접근이 가능하게 하기 위해서 platform 옵션을 지정해주도록 하자.

    WORKDIR /app

컨테이너 내부에서 작업할 디렉토리를 설정한다. 이 커맨드 이후로는 이 디렉토리에서 작업을 계속 수행하게된다. 기본 컨테이너 이미지에 있는 디렉토리면 그 안에서 하겠지만, 그렇지 않다면 저 경로를 만드는 것 같다.

    COPY . .

.은 현재 디렉토리를 가리킨다. 누구의 현재 디렉토리일까? 지금 작성하고 있는 도커파일의 디렉토리이다. 이 코드의 의미는 현재 도커파일이 위치한 디렉토리에 있는 모든 것들을(왼쪽 .), 위에서 지정한 컨테이너 내부의 work directory(오른쪽 .)에 복사하겠다는 것을 의미한다.

그럼 소스파일과, go.mod, go.sum, readme.md, .env 등 모든 파일들이 복사된다.

    RUN go mod download

이 명령어에 대해 알아보기 전에 go.mod와 go.sum에 대해 알아보자.

go.mod는 개발자가 직접 편집이 가능한, 소스파일에서 사용하는 패키지들의 이름과 버전을 명시해둔 파일이다. 개발자는 자신이 사용할 패키지들을 이곳에 정의할 수 있고, 삭제 및 변경도 가능하다. 또는 외부 패키지 다운로드 명령어인 **go get ~**를 이용하면 자동으로 go.mod에 패키지가 추가된다.

go.sum은 go.mod가 정의되어있는 상태에서 **go mod tidy** 커맨드를 통해 현재 go.mod에 쓰여져있는 패지키와 버전들이 신뢰성을 체크한 후 신뢰할만한 패키지들이면 go.mod가 위치한 디렉토리에 go.sum이 생성된다.

go.sum과 go.mod의 정보는 항상 일치해야하고, 일치하지 않으면 컴파일 시 에러가 발생한다. 예를 들어 개발자가 go mod tidy로 go.sum을 생성한 후 go.mod를 수정했다. 그리고 다시 go mod tidy를 하지 않은채로 컴파일을 시도한다면 오류가 발생한다. 사용자가 go.mod에 추가적으로 정의한 패키지에 대한 신뢰성이 go.sum을 통해 확인되지 않기 떄문이다. 이런 방식으로 go.mod와 go.sum을 통해 go에서 쓰여지는 패키지들에 대한 안정성을 확보한다.

돌아와서, COPY . . 명령어를 수행하면 로컬에 있는 go.sum과 go.mod도 컨테이너 안으로 복사가 되는데, 이 go.mod는 실제 의존성이 설치된 파일이 아니라, 필요한 의존성들의 이름과 버전을 나열해둔 파일이기 때문에, 이 go.mod를 바탕으로 실제 의존성 설치를 해주어야한다. 이 명령어가 바로 **go mod download**이다. go.mod는 설치리스트, go.sum은 인증서라고 이해하면 될 것 같다.

    RUN go build ./src/main.go

main.go는 go workspace에서 src 디렉토리 안에 위치하는 경우가 많기 때문에 이렇게 빌드를 해준다. 만약 main 패키지가 위치한 소스파일이 다른 경로에 있다면 그 경로로 빌드르 실행해준다.

    EXPOSE 8080

컨테이너의 port를 8080으로 열어준다. 이 포트는 컨테이너가 쿠버네티스 클러스터에 배포되면, 서비스와 연결될 targetport로 작동할 것이다. 백엔드에서 listen하고 있는 port와 동일하게 설정해야 컨테이너의 8080포트로 요청이 왔을 때 백엔드가 요청을 처리할 수 있다.

    CMD ["./main"]
./main 커맨드를 실행해서 빌드된 파일이 실행되도록 한다. main.go파일은 src디렉토리 안에 있지만, 빌드르 실행한 디렉토리게 루트이기 때문에 이렇게 main 빌드파일은 src 디렉토리 안이 아닌 도커파일과 같은 루트 디렉토리에 위치하게 된다.

전체적인 백엔드 도커파일 코드를 보면 이러하다.

    FROM --platform=amd64 golang:latest

    WORKDIR /app

    COPY . .

    RUN go mod download

    RUN go build ./src/main.go

    EXPOSE 8080

    CMD ["./main" ]

---

## React 프론트엔드 운영환경 도커파일

리액트도 go처럼 운영환경에서는 빌드파일을 이용한다. 왜 빌드파일은 운영환경에서는 많이 이용할까? 빌드파일은 압축된 형태로 존재해서 브라우저에 더 좋은 성능으로 정적파일들을 제공할 수 있고, 빌드하는 과정에서 개발환경에서 쓰였던 불필요한 코드들을 제거하기 때문에 운영환경에 더 적합한 경우가 많다.

리액트는 빌드파일을 nginx의 정적파일 서빙 기능을 이용해서 배포할 예정이기 때문에 리액트와 nginx를 위한 두 개의 이미지를 필요로 한다. 리액트 도커파일을 살펴보자.

    FROM --platform=linux/amd64 node:16-alpine AS build

alpine이미지는 일반 이미지보다 조금 더 경량화된 컨테이너 이미지이다. 자금의 한계로 운영환경에서 최대한 적은 리소스를 사용하며 배포해야하기도 했고, 내가 개발한 프로젝트가 큰 규모의 프로젝트도 아니기 때문에 alpine이미지를 사용했다.

AS build는 빌드파일 빌드를 위한 리액트의 설정들을 마친 후에 nginx 이미지 안의 특정 경로에 빌드파일을 넣어줘야하기 때문에 사용된다. 변수를 설정하는 느낌으로 보면 될 것 같다.

    WORKDIR /app

go와 마찬가지로 work directory를 설정해준다.

    COPY . .
    RUN npm install

go는 go.mod와 go.sum으로 의존성을 관리했는데, 리액트는 package.json이라는 파일로 의존성을 관리한다. 이 파일은 COPY . . 커맨드를 통해 컨테이너 안으로 복사되었기 때문에, **npm install** 커맨드를 실행해서 package.json에 있는 의존성들을 설치해준다. 여기서 주의할 것이 있다. golang은 go get ~ 커맨드로 패키지를 설치하면 자동으로 go.mod가 업데이트되는데, 리액트는 그렇지 않다. 따라서 흔히 웹 어플리케이션 개발에 쓰이는 axios같은 외부 패키지를 사용하려면 실제 코드상 사용을 위한 **npm install axios**등의 커맨드를 실행했더라도, 추가로 package.json에 직접 이름과 버전을 명시해주어야한다.

    RUN npm run build

이 명령어를 통해서 빌드파일을 생성한다. 빌드파일은 따로 옵션을 지정하지 않으면 build라는 이름의 폴더 안에 필요한 모든 종속성들이 들어있는 빌드파일이 생성되게 된다.

    FROM --platform=linux/amd64 nginx

nginx 이미지도 pull한다. 여기서 쓰이는 nginx는 리버스프록시를 위한 nginx가 아닌 정적파일 서빙을 위한 nginx이다.

    EXPOSE 3000

컨테이너의 3000번 포트를 열어준다. 이 포트를 통해 외부와 서비스 오브젝트를 통해 통신할 수 있고, 클러스터 내부에서도 이 포트를 통해 클러스터ip나 DNS서비스네임으로 통신할 수 있다. 

    COPY ./nginx/default.conf /etc/nginx/conf.d/default.conf

로컬에는 project/frontend/nginx/default.conf 파일이 있다. 이 파일은 아래와 같이 구성되어있다. 

    server {
            listen 3000;

            location / {
                    root /usr/share/nginx/html;
                    # 정적 파일을 브라우저에 제공하기 위해. 리액트에서 빌드한 파일이 어디있는지 지정하는 거
                    index index.html index.htm;
                    # 처음 시작을 뭘로 할 건지 설정
                    try_files $uri $uri/ /index.html;
                    # SPA만 만들 수 있는 리액트에서 라우팅을 가능하게 해줌
            }
    }

몇 번 포트에서 listen할 것인지, nginx가 서빙해야할 빌드파일은 어디에 위치하는지, 처음 서빙할 파일의 이름은 무엇인지 등을 설정해준다. 이 파일을 nginx 이미지의 /etc/nginx/conf.d/default.conf 이 경로에 넣어줘야 nginx가 이 config파일을 인식할 수 있다.

    COPY --from=build /app/build /usr/share/nginx/html 

/usr/share/nginx/html 이 경로는 위의 default.conf 파일에서 root로 정의한, 빌드파일이 올라가야하는 경로이다. nginx의 이미지 안에 이 경로에 빌드파일을 넣어줘야 nginx가 빌드파일을 정상적으로 브라우저에 제공할 수 있다.

nginx에 빌드파일이 올라갈 경로를 정해주고, 그 경로에 빌드파일을 올려서 nginx가 빌드파일을 서빙하게 되는 과정인 것이다.

전체적인 도커파일을 보면 아래와 같다.

    FROM --platform=linux/amd64 node:16-alpine AS build

    RUN mkdir app

    WORKDIR /app

    COPY . .

    RUN npm install

    RUN npm run build

    FROM --platform=linux/amd64 nginx

    EXPOSE 3000

    COPY ./nginx/default.conf /etc/nginx/conf.d/default.conf

    COPY --from=build /app/build /usr/share/nginx/html 

---

## Mysql 데이터베이스 운영환경 도커파일

    FROM --platform=linus/amd64 mysql:latest


    COPY ./sqls/my.cnf ./etc/mysql/conf.d/my.cnf

한글은 UTF8 인코딩방식을 사용한다. Mysql 버전에 따라 기본 문자셋이 다르게 설정되어있을 수 있기 때문에 레코드에 한글도 저장할 수 있도록 해당 config를 mysql이 인식할 수 있는 지정된 위치에 추가해준다.

---

원래 같으면 추가적으로 리버스프록시를 위한 nginx 이미지를 생성하고 설정해서 빌드하겠지만, 나는 쿠버네티스 클러스터의 ingress nginx controller를 이용해 리버스프록시 기능을 수행할 예정이어서 nginx를 위한 도커파일은 따로 작성하지 않았다.


## 도커파일 이미지 빌드

이렇게 작성된 도커파일들은 이미지로 빌드해서 컨테이너 이미지 레지스트리에 푸시해야한다.

    docker build -t <dockerhub ID>/<dockerhub registry name> .

-t는 태그를 지정하는 옵션이다. 아이디/레지스트리 이름으로 설정하면 후에 이미지를 레지스트리에 푸시할 때 해당 레지스트리 이름으로 푸시되게 된다.

.은 현재 도커파일이 위치한 디렉토리와 하위 디렉토리 전체를 이미지로 빌드하겠다는 뜻이다.

---

## 컨테이너 이미지 푸시

이렇게 빌드된 이미지는

    docker images

커맨드나, 혹은 도커 데스크탑이 실행중이라면 image 탭에서 확인할 수 있다.

이 이미지를 도커허브 레지스트리에 푸시하려면, 빌드시 입력한 ID와 일치하는 도커허브 계정이 생성된 상태여야한다. 그리고 터미널에서 도커에 로그인이 된 상태여야하는데, 이 명령은 dockerCLI를 이용해서 할 수 있다.

DockerCLI는 도커를 install할 때 일반적으로 함께 설치된다.

    docker login

커맨드로 도커에 로그인하고,

    docker push <created image name>

을 입력하면 이미지가 푸시된다. 물론 이 created image name은 도커ID/레지스트리명 형식의 이름이어야한다. 그래야 도커허브에 푸시할 수 있다.

---

이렇게 하면 쿠버네티스 클러스터에서 오브젝트 생성 시 도커허브 레지스트리에 푸시되어있는 이미지를 풀해다가 사용할 수 있다.

다음 글에선 개발환경용 Dockerfile.dev 작성에 대해 알아보겠다.