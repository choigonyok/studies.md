# #3. docker-compose.yaml 파일 설정
# project couple-chat docker

---

## 개요

이전 글에서 개발환경을 위한 도커파일 작성을 살펴봤다. 이제 도커 컴포즈를 활용해서 각 서비스의 도커파일을 읽고, 이미지를 빌드하고, 빌드한 이미지를 기반으로 컨테이너를 실행하는 과정을 살펴보겠다.

---

## 도커 컴포즈

docker compose는 로컬환경에서 컨테이너 기반의 어플리케이션을 쉽게 빌드할 수 있게 해준다. 여러 장점이 있지만 내가 느낀 가장 큰 장점 두 가지를 꼽아보겠다.

1. 여러 컨테이너를 하나의 커맨드로 동시에 실행이 가능하다.

하나의 컨테이너만 실행한다면 아래 커맨드로 실행할 수 있다.

    docker run -

만약 개발환경에서 여러 개의 컨테이너를 실행해야한다면? 풀스택이나 혹은 마이크로서비스를 개발하기 위해 100개가 되는 컨테이너를 동시에 실행해야한다면?

~~docker run 커맨드를 manually 100번 실행할 순 없다.~~ 할 수는 있는데 아주 비효율적이다. 도커 컴포즈는 yaml파일에 정의된 모든 컨테이너를 한 번에 실행할 수 있게 해주는 간편성을 제공한다.

2. 컨테이너 간 네트워크 설정이 간편하다.

컨테이너를 일반적으로 하나씩 실행한다면 각 컨테이너는 개별적인 네트워크를 가지게 된다. 프론트와 백엔드가 통신하려면 각자의 ip주소와 포트를 통해서 통신해야만한다.

도커 컴포즈를 사용하면 모든 컨테이너를 하나의 네트워크로 묶을 수 있고, localhost로 통신할 수 있어 개발과정이 간편해진다.

도커 컴포즈는 내부적으로 DNS서비스도 제공하기 때문에, 도커 컴포즈 파일에 정의한 서비스 이름을 통해 다른 서비스와 통신할 수 있다.

---

## Docker-Compose.yml

이제 도커컴포즈 파일을 작성해보자.

```yaml
version: '3'
services:
  frontend:
    build:
      dockerfile: Dockerfile.dev
      context: ./frontend
    volumes:
      - /app/node_modules
      - ./frontend/:/app
    stdin_open: true
    environment:
      - WDS_SOCKET_PORT=80
    ports:
      - "8080:8080"
      
  mysql:
    build: ./mysql
    ports:
      - "3306:3306"
    volumes:
      - ./mysql/mysql_data:/var/lib/mysql
      - ./mysql/sqls:/docker-entrypoint-initdb.d/
    environment:
      MYSQL_ROOT_PASSWORD: 1234
      MYSQL_DATABASE: example
      TZ: Asia/Seoul

  backend:
    build:
      dockerfile: Dockerfile.dev
      context: ./backend
    volumes:
      - ./backend/:/app
    ports:
      - "8080:8080"
```

### services

    services: 

하위 항목들을 서비스 이름을 지정한다. 이 이름은 DNS서비스에 등록되어서 서비스명을 통해 해당 컨테이너와 통신할 수 있게 된다.

### build

    build:
        dockerfile: Dockerfile.dev
        context: ./frontend

**dockerfile** 속성의 값에는 빌드해야하는 도커파일이 어떤 건지를 명시해야한다. 만약 dockerfile 속성을 사용하지 않는다면 도커는 default로 Dockerfile이라는 이름을 가진 파일을 찾는다.

운영환경과 개발환경의 도커파일을 분리할 때, 개발환경 도커파일은 Dockerfile.dev로 이름 짓는 것이 일반적이다.

만약 이 build 속성에서 dockerfile을 Dockerfile.dev로 명시해주지 않으면 앞서 말한 것처럼 도커는 Dockerfile을 찾아서 이미지를 빌드하게 될 것이다.

**context** 속성은 경로를 지정한다.

위의 리액트 Dockerfile.dev를 예시로 들면, COPY ./package.json ./ 이 코드가 실행될 때 ./package.json 부분에서 .이 context가 되는 것이다.

WORKDIR는 컨테이너 내부의 경로, context: 는 로컬의 경로라고 생각하면 된다.

또 다른 예를 들어보겠다.

    도커파일 경로 : /project/backend/Dockerfile
    도커컴포즈 경로 : /docker-compose.yml
    도커파일 내용 : COPY . .

이런 상황일 때, 도커컴포즈의 내용이 아래와 같다면

```
1번
build:
    dockerfile: backend/Dockerfile
    context: ./project
```

```
2번
build:
    dockerfile: Dockerfile
    context: ./project/backend
```

1번과 2번 중 어느 것이 정상적으로 실행될까?

정답은 **둘 다** 이다.

그렇다고 결과가 같은 것은 아니다.

COPY . . 명령을 실행하면 1번은 context가 ./project이므로 project 디렉토리 안의 전체를 컨테이너로 복사할 것이고, 2번은 context가 ./project/backend이므로 backend 디렉토리 안의 전체를 컨테이너로 복사할 것이다.

dockerfile, context 속성은 이런 관계가 있다.

### volumes

    volumes:
      - /app/node_modules
      - ./frontend/:/app

volumes는 볼륨을 생성하고 컨테이너로 마운트시킨다. 컨테이너 안에서 생성된 데이터는 본래 컨테이너가 사라지면 같이 사라진다. 같은 이미지로 컨테이너가 재시작되어도 이전 컨테이너가 가지고있던 정보는 남아있지 않게된다.

볼륨을 이용하면 컨테이너와 host간에 통신을 해서, 같은 데이터를 공유할 수 있다.

이러면 컨테이너가 사라져도 컨테이너와 공유한 호스트 볼륨에 데이터가 남아있게 되고, 새로운 컨테이너가 생성되면 다시 볼륨을 연결해서 이전 컨테이너가 가지고있던 데이터를 새로 생성된 컨테이너가 그대로 이어받을 수 있게 된다.

볼륨은 데이터 보존 뿐만 아니라 컨테이너에 스크립트 등의 데이터를 전달하는 등 다양한 목적으로 사용될 수 있다.

예시 코드에는 : 심볼이 있는 부분과 없는 부분 두 라인이 있다.

1. : 심볼이 없으면 해당 경로는 볼륨에서 제외시키겠다는 의미이다. .gitignore 그런 느낌이다.
2. : 심볼이 있으면 : 심볼 왼쪽의 로컬 경로를 볼륨으로 설정하고, : 심볼 오른쪽의 컨테이너 경로에 볼륨을 마운트시키겠다는 것이다.

* 둘 다 유효한(실존하는) 경로여야한다.

이렇게 볼륨을 설정하고나면, 로컬에서 ./frontend/text.txt 생성 시 컨테이너에 /app/text.txt가 생성되고, 컨테이너에서 /app/text.txt를 삭제하면 로컬에서도 ./frontend/text.txt가 삭제되는 식으로 연동되게 된다.

이 볼륨을 통해서 1편에서 소개한 실시간 코드 수정 적용이 가능한데, 위 예시에서 ./frontend 경로 안에는 리액트 소스파일들이 있고, 이 경로는 볼륨으로서 /app 경로에 마운트되어있기 때문에, 코드 수정시 컨테이너 내부 소스파일도 동일하게 변경되게 된다.

흐름을 요약하자면 다음과 같다.

    볼륨 설정 -> 볼륨 안에서 코드 수정 -> 볼륨이 마운트된 컨테이너 내부에도 파일 수정이 적용 -> 변경사항 감지 후 핫리로딩으로 수정사항 적용

대신 컨테이너의 /app 경로에 **./frontend/node_modules는 마운트되지 않는다.** 왜?

/app/node_modules 라인을 설정해둬서 node_modules는 동기화가 되지 않고 제외되기 때문이다.

근데 node_modules는 왜 제외시키는 걸까?

node_modules는 의존성 파일이다. package.json처럼 의존성 설치 리스트가 아니라, package.json을 통해 npm install으로 설치된 실제 의존성들이 설치되어있는 파일이다.

이 파일은 용량도 클 뿐더러, 운영환경에서는 도커파일에 정의한 RUN npm install 명령어로 해당 운영환경에서 필요한 의존성들만을 알아서 설치할 것이기 때문에 볼륨에서 제외시켜준다.

제외시키지 않으면 운영환경에는 필요없지만 개발환경에는 필요한 의존성들이 불필요하게 운영환경에까지 복사되어서 리소스를 낭비하게된다.

### environment

```
environment:
    MYSQL_ROOT_PASSWORD: 1234
    MYSQL_DATABASE: example
    TZ: Asia/Seoul
```

이 속성을 통해 컨테이너 내부에 환경변수를 주입할 수 있다.

mysql의 경우 기본적으로 적어도 MYSQL_ROOT_PASSWORD와 MYSQL_DATABASE 환경변수는 설정을 해줘야 실행이 가능하다. 어느 DB를 사용할 것인지, 또 root 사용자가 접근할 때 password는 뭘로할 건지를 설정하는 부분이다.

TZ는 TimeZone의 줄임말로 mysql의 시간 설정을 하는 부분이다.

### ports

    ports:
      - "8080:8080"

ports는 라우팅 포트와 리스닝 포트를 설정하는 부분이다. : 심볼 오른쪽은 컨테이너가 전달할 서비스의 포트, 심볼 왼쪽은 외부에서 컨테이너가 리스닝할 포트를 의미한다.

백엔드 같은 경우 RUN(":8080") 이런 식으로 리스너를 구현할텐데, 만약 다 같은데 RUN(":8000")로 백엔드 코드를 수정한다면 

    ports:
      - "8080:8000"

포트 설정도 이렇게 변경해주어야 할 것이다. 만약 프런트엔드에서 API요청을 4000번 포트로 변경하고싶다면, 백엔드의 ports는 

    ports:
      - "4000:8000"

이렇게 수정되어야 할 것이다.

---

## 추가 설명

추가적으로 frontend 서비스의

```
environment:
    - WDS_SOCKET_PORT=80
```

이 부분은 웹소켓 커넥션을 위해 웹소켓 port를 설정해주는 부분이다. 설정해주지 않으면 웹소켓 연결이 정상적으로 이루어지지 않는다.

```
volumes:
      - ./mysql/mysql_data:/var/lib/mysql
      - ./mysql/sqls:/docker-entrypoint-initdb.d/
```

컨테이너 내부 /var/lib/mysql는 mysql의 현재 DB 데이터가 저장되는 곳이다.

이 경로와 로컬의 ./mysql/mysql_data를 볼륨으로 동기화시키면, 컨테이너가 종료 후 재시작해도 로컬에 볼륨이 있기 때문에 기존 컨테이너가 가지고있던 레코드들을 그대로 유지할 수 있다. 백업 용도라고 볼 수 있겠다.

컨테이너 내부 /docker-entrypoint-initdb.d/는 mysql의 초기 테이블, 컬럼 설정 정보가 들어있는 곳이다. 

로컬에서 config 파일을 생성해서 호스트 볼륨인 ./mysql/sqls/ 안에 넣어주면, mysql이 /docker-entrypoint-initdb.d/경로의 sql 파일을 읽고 초기 설정을 할 수 있게된다. 이 경우엔 백업보다는 데이터 전달 목적이라고 볼 수 있겠다.

---

## 실행

docker-compose로 이미지를 빌드하고 컨테이너를 실행시키려면

    docker-compose up

커맨드를 사용하면 된다. 만약 이미지가 이미 docker-compose up으로 생성되어있는데 이미지를 수정해서 재빌드하고싶다면

    docker-compose up --build

커맨드를 통해 이미지를 재빌드하고 도커 컴포즈를 실행할 수 있다.

---

## 정리

docker-compose.yml 파일 내용을 정리하면서 모호하게 인식하고 있던 내용들도 글을 쓰며 깔끔하게 정리가 된 것 같아서 좋다.

다음 커플 채팅 서비스 포스트는 리액트에서 캘린터 구현하는 과정에 대해 다뤄보겠다.