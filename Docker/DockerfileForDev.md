    FROM golang:latest
개발환경이기 때문에 --platform=linux/amd64 옵션을 주지 않아도 된다.
latest 태그로 가장 최신 버전의 golang 이미지를 당겨온다.

    WORKDIR /app
컨테이너에서 작업 기준점을 잡는다.

    COPY go.mod ./
    COPY go.sum ./
Prod 도커파일과 다르게 go.mod와 go.sum을 먼저 컨테이너로 카피한다.
왜냐하면 운영환경과 개발환경의 특성 차이 때문.
운영환경은 수정이 적고, 성능이 중요하다.
개발환경은 성능보다는 빠른 수정이 가능한 것이 중요하다.
개발환경에서 docker-compose를 통해서 개발할 때, 볼륨을 이용한다.
원래는 필요한 모든 종속성을 컨테이너로 옮겨서 컨테이너에서 실행하는데,
그렇게 되면 코드를 수정할 때마다 새롭게 이미지를 빌드하고 도커컴포즈를 올려야한다.
이 시간이 상당히 길기 때문에, 코드 관련된 부분은 컨테이너가 로컬의 코드를 참조하도록 구현할 수 있다.
이게 바로 볼륨
볼륨을 설정하게 되면 코드를 수정하자마자 컨테이너는 그 파일을 참조하고있기 때문에 바로바로 수정된 내용이 적용될 수 있다.
근데 한 번에 묶에서 카피를 하게 되면 볼륨을 

    RUN go mod download
Prod와 마찬가지로 의존성을 설치한다.

    RUN go install -mod=mod github.com/githubnemo/CompileDaemon
# go get github~ 로는 설치가 안됨
# 특히 저 -mod=mod가 뭔지 찾아봐야할 듯

    COPY . .
# CMD ["go", "run", "main.go" ]
# build 뿐만 아니라 run도 코드 수정을 반영할 수 없음
# 그래서 compileDaemon 오픈 소스 활용
go.mod와 go.sum 이외의 나머지도 컨테이너로 복사해준다.

    ENTRYPOINT CompileDaemon -polling -log-prefix=false -build="go build -o main ./src" -command="./main" -directory="./"
# build로 build 명령어 작성하고
# command로 빌드한 빌드파일을 실행할 명령어 작성
컴파일데몬은 golang에서 실시간 수정을 가능하게 해주는 데몬이다.
golang은 빌드파일을 생성하고 그 빌드파일을 실행하는 방식으로 작동한다.
이게 문제가 뭐냐면, 실시간으로 볼륨을 통해 참조하려면 계속 실행되고있는 상태여야하는데
빌드파일은 빌드할 당시의 코드만 가지고있고, 파일이기 때문에 계속 수정이 어려움
그래서 이걸 이용해서 필요한 부분만 빠르게 재빌드할 수 있도록 도와주는 툴임
그래서 종속성 먼저 카피한게 저 부분이 코드파일과 같이 묶여있으면 코드부분만 재빌드하면 될 걸 묶여있어서 의존성까지 재설치하는 일이생기고
그럼 시간이 상당히 길어짐
코드를 바꿀 때 새로운 패키지를 사용하게 됐다거나 하면 의존성도 다시 설치해야하는 게 맞지만
보통의 코드 작성시에는 의존성에 변화가 생기는 일은 많지 않음
그래서 이 데몬을 이용


---

## 프론트엔드 개발환경 도커파일

    FROM node:16-alpine

    WORKDIR /app

    COPY ./package.json ./

    RUN npm install

    COPY . .
이거 역시 볼륨을 이용해서 실시간 수정이 가능하도록 하기 위해 의존성 먼저 설치

    CMD ["npm", "run", "start"]
node는 npm run build로 빌드파일도 만들 수 있지만
빌드파일을 만들지 않고도 런타임으로 실행할 수 있는 기능을 제공
golang의 컴파일데몬이 node에 기본 내장되어있는 거라고 생각하면 될 것 같음
빌드파일을 통해서 실행되는 게 아니기때문에 운영환경과 달리 정적파일을 제공해주는 nginx 이미지를 당겨오지 않음
# dev 파일 추가 작성
# 운영환경은 빌드파일로 빠르게 사용자에게 제공하는 목적이 있고
# 개발환경은 잦은 수정과 빌드의 필요성이 있어서
# 개발환경인 Dockerfile.dev에서는 start를 하고
# 운영환경인 Dockerfile에서는 build를 함
# 빌드를 하면 이미 정적인 빌드 파일이 생긴 것이기 때문에 volumes으로 코드 수정을 적용하는 게 불가능함

---

mysql과 nginx는 운영환경과 개발환경의 도커파일 차이가 없음
도커파일이 분리되는 건 코드가 작성되는 부분이 있을 때 분리되는 듯함

---

## 도커컴포즈 작성

version: '3'
services:
  frontend:
    build:
      dockerfile: Dockerfile.dev
      context: ./frontend
    volumes:
      - /app/node_modules
      - ./frontend/:/app
    # 다 설정 마치고 react의 코드를 수정했는데 volumes가 적용되지 않았음
    # 이건 운영환경용 도커파일이어서 빌드된 이미지를 컨테이너화 하는 거라서
    # 이미 빌드된 이미지에 아무리 코드 수정을 한다고 한들
    # 코드가 바뀔리가 없었음
    # 그래서 개발환경용 도커파일을 따로 만들어서 거기서는 빌드하지 않고 npm run start로 하는 거였음!!
    # 대박, 그래서 백엔드와 프론트엔드 모두 Dockerfile.dev를 추가해서 작성해줬음
    stdin_open: true
    environment:
      - WDS_SOCKET_PORT=${WS_PORT}

  nginx:
    restart: no
    build:
      dockerfile: Dockerfile
      context: ./nginx
    ports:
      - "80:80"
      
  mysql:
    build: ./mysql
    restart: no
    ports:
      - "3306:3306"
    volumes:
      - ./mysql/mysql_data:/var/lib/mysql
      - ./mysql/sqls:/docker-entrypoint-initdb.d/
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_PASSWORD}
      MYSQL_DATABASE: ${DB_NAME}
      # .env를 활용해 환경변수를 설정하고 
      # .env는 .gitignore로 공유저장소에 올라가지 않도록 설정
      TZ: Asia/Seoul

  backend:
    build:
      dockerfile: Dockerfile
      # 안쓰면 Dockerfile만 읽음
      context: ./backend
      # docker-compose.yml파일의 directory 기준으로 상대경로 작성
    volumes:
      - ./backend/:/app
      # 오른쪽 : 컨테이너의 해당 디렉토리가, 왼쪽 : 호스트(로컬)의 해당 디렉토리를 참조하도록 함
      # 백엔드에선 go.mod나 go.sum을 node_modules처럼 제외시킬 필요가 없음
      # -> 클라이언트에서는 package.json으로 종속성을 다시 설치하는데, 서버에서는 로컬의 go.mod, go.sum을 카피해와서 사용하기 때문
      # CompileDaemon을 설치하고 빌드/실행 했다고해서 volumes을 안적어주면 데몬이 적용 안됨
    ports:
      - "8080:8080"
      # react는 이미지로 컨테이너 생성시 포트지정을 안해줘도 nginx가 3000을 listen하고 있어서
      # 실행이 가능한데 go는 그렇지 않아서 포트 지정을 해줘야함

# build는 docker-compose up
# 두 번째부터는 docker-compose up --build