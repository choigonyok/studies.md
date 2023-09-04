# Skaffold로 KIND 클러스터 CI/CD 파이프라인 구축하기
# study kubernetes

---

## 개요

Kind로 로컬 쿠버네티스를 구축했다. 개발 편의성을 높이기위해, 개발자가 코드를 수정하면 자동으로 이미지가 재빌드되고, Kind 클러스터에 Load 되고, 매니페스트로 쿠버네티스에 배포되는 파이프라인을 구축하고 싶었다.

알아보던 중 클러스터 + 개발환경에 Skaffold라는 도구가 많이 사용된다는 것을 알게되었고, 이 글은 Kind 로컬 쿠버네티스 클러스터에 Skaffold로 CI/CD 파이프라인을 구축하고 적용하는 과정을 기록한 글이다.

---

## Skaffold

Skaffold(이하 스캐폴드)는 컨테이너 기반 또는 쿠버네티스 기반의 어플리케이션을 지속적 개발이 가능하도록 돕는 커맨드라인 도구이다. 소스 코드의 변경을 감지하고, 컨테이너 이미지 빌드, 테스트, 컨테이너 레지스트리 푸시, 쿠버네티스 클러스터로의 배포 파이프라인을 설정하고 관리할 수 있다. 경량화되어있어서 변경사항에 대한 빠른 배포가 가능하다.

스캐폴드의 워크플로우는 다음과 같다.

![img](http://www.choigonyok.com/api/assets/66-1.png)

---

## skaffold install

아래 링크에서 로컬 OS에 맞는 스캐폴드 바이너리 파일을 다운로드한다.

[Skaffold Install](https://skaffold.dev/docs/install/)

---

## skaffold init

프로젝트의 루트 디렉토리에서 아래 커맨드를 실행한다.

```
skaffold init --generate-manifests
```

스캐폴드는 파일 싱크 기반으로 소스 파일 변경사항을 추적한다. 위 커맨드는 skaffold.yaml 파일을 생성하고, 커맨드가 실행되는 디렉토리부터 스캐폴드가 소스파일을 감지한다. 때문에 해당 커맨드는 프로젝트의 Root Directory에서 실행되어야한다.

---

커맨드를 실행한 루트 디렉토리에 deployment.yaml, skaffold.yaml 두 파일이 생성되었을 것이다.

### deployment.yaml

deployment.yaml 파일에는 스캐폴드가 자동으로 빌드하고 쿠버네티스에 재배포할 때 쓰일 매니페스트들을 입력한다. 나의 경우에는 테스트를 위해 블로그 서비스의 프론트엔드만 작성했다.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: frontend
  name: frontend-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - image: frontend
        name: frontend
        ports:
        - containerPort: 3000
---
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
spec:
  selector:
    app: frontend
  type: ClusterIP
  ports:
    - port: 3000
```

그냥 일반적인 매니페스트와 같다. 다만, skaffold를 통해 tag를 생성해줄 것이기 때문에, spec.template.spec.containers.image에는 태그를 입력하지 않는다.

### skaffold.yaml

skaffold.yaml 파일은 스캐폴드 파이프라인을 설정하는 yaml 파일이다.

```yaml
apiVersion: skaffold/v4beta6
kind: Config
metadata:
  name: frontend
build:
  tagPolicy:
    dateTime:
      format: 01/02_15:04:05
  artifacts:
  - image: frontend
    context: .
    docker:
      dockerfile: Dockerfile.dev
manifests:
  rawYaml:
    - deployment.yaml
```

예시 메인 속성에는 build, manifests 뿐이지만, 추가적으로 deploy, test, portForward, varify, customActions, profiles 가 있다.

build.tagPolicy는 생성되는 이미지의 태그를 어떻게 설정할 것인지 정의할 수 있다.

build.tagPolicy.dateTime으로 태그를 이미지 빌드 시간으로 설정할 것임을 선언하고, 뒤따라오는 format으로 시간의 형식을 지정한다. 스캐폴드는 go로 개발된 툴이기 때문에 시간 형식 설정도 go의 문법을 따라간다.

build.artifacts는 빌드할 이미지 관련 내용을 설정한다.

image는 빌드할 이미지의 이름, context는 도커파일의 위치, docker.dockerfile은 도커파일의 이름을 설정한다.

사실 context는 도커파일의 위치라고 정확히 말하기는 어려운데, 이 context에 관한 내용은 커플채팅 서비스를 개발하며 작성한 "**#3. docker-compose.yaml 작성하기**" 글에서 자세히 확인할 수 있다.

manifests는 쿠버네티스 오브젝트를 생성할 때 쓰일 yaml파일을 지정하는 부분이다.

<br>

사실 skaffold.yaml 파일에는 정말 많고 다양한 속성들이 있다. 예를 들어, artifacts.docker.dockerfile.requires에서는 이미지의 종속성을 설정할 수 있고, artifacts.local에서는 동시에 빌드할 이미지의 수를 설정할 수도 있다.

이 글에서는 skaffold CI/CD 파이프라인 구축 테스트를 위한 필수적인 속성들만 작성했다. 추후 팀프로젝트를 하며 실 서비스 개발을 위한 skaffold.yaml 작성을 다시 한 번 다를 예정이다.

skaffold.yaml의 다양한 속성들은 공식문서에서 확인할 수 있는데, 글 마지막에 링크를 올려놓았다.

---

## skaffold dev

아래 커맨드를 통해 Dev Loop를 실행시킨다. skaffold.yaml에 정의된 설정대로 이미지 빌드, 테스트, 푸시, 배포를 실행한다. 이후로는 마치 루프를 도는 것처럼 소스파일이 변경되면 설정된 파이프라인을 재실행한다.

```
skaffold dev
```

이렇게 파이프라인을 구축하면 소스파일에 변경(소스코드 수정 등)이 될 때마다 자동으로 이미지 빌드, Kind 클러스터에 load, Kind 클러스터에 배포가 이루어진다.

스캐폴드는 CI/CD를 진행하다가 과정 중 오류가 발생하면 Dev Loop를 종료시켜버린다. 그럼 개발자가 실수로 컴파일이 안되는 코드를 입력하는 등의 사소한 실수에도 Dev Loop가 종료되고, 개발자는 그 때마다 skaffold dev 커맨드를 다시 입력해야한다.

이를 방지하기 위해서 스캐폴드에서는

```
--keep-running-on-failure
```

플래그를 제공한다.

처음 skaffold dev를 실행할 때, 

```
skaffold dev --keep-running-on-failure
```

플래그를 달아서 시작하면 오류가 발생하면 해당 변경사항에 대한 빌드->배포 과정은 종료되지만 CI/CD 파이프라인은 중단되지 않을 수 있게된다.

---

## 정리

쿠버네티스, 서비스메시, MSA를 도입할 팀프로젝트를 위한 개발환경 구성에 대한 공부가 얼추 끝나간다. 

로컬 멀티노드 쿠버네티스 클러스터를 Kind로 생성하고, 그 안에 서비스도 배포하고, Istiod와 Ingress-Gateway 배포하고, 여러 모니터링 도구들 배포하고, Skaffold로 CI/CD 파이프라인까지 구축하니 RAM 용량이 상당히 부족하다.

한 페이지짜리 프론트엔드의 CI/CD 파이프라인 테스트를 했을 뿐인데 CPU가 메모리 스왑까지하며 발열이 95도까지 올라갈 정도로 메모리가 부족했다.

실제 개발을 위해 환경을 구성하면 더 많은 설정이 들어가고 추가적인 툴들과 배포되는 서비스들의 크기도 상당히 클텐데 로컬머신이 감당할 수 있을지 모르겠다.

다음은 Helm 차트에 대해 깊게 알아볼 예정이다.

---

## 참고

[Skaffold Official Documentation](https://skaffold.dev/docs/)

[skaffold.yaml](https://skaffold.dev/docs/references/yaml/)