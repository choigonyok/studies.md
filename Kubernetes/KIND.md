# Setup Dev Environment w/ KIND for Kubernetes
# study kubernetes

---

## 개요

MSA를 쿠버네티스 클러스터에 배포하는 새로운 팀 프로젝트를 기획하고있다.

블로그 서비스를 쿠버네티스 클러스터를 이전해본 경험은 있지만, 처음부터 쿠버네티스를 염두에 두고 개발을 시작한 적은 없었다. 운영환경에 쿠버네티스가 사용될 때 개발환경은 어떻게 구축하면 좋을까?

이 글은 다양한 쿠버네티스를 위한 개발환경들과, 그 중 내가 Kind를 선택한 이유 및 기본적인 Kind 사용법에 대해 작성한 글이다.

---

## 다양한 쿠버네티스 개발환경

개발환경 클라우드 서버 구축, 도커 컴포즈, minikube, vagrant, 마지막으로 kind에 대해 알아보겠다.

### 개발환경 클라우드 서버 구축

가장 효과적이고 쉬운 방법은 운영환경과 같은 개발환경 서버를 구축하는 것이다. 이 방식은 운영환경과 개발환경의 차이가 없다는 점에서 효과적이다.

대신 개발하는 기간동안 발생하는 서버비용이 고스란히 청구되기 때문에, 비용적인 측면에서 단점이 있다.

실제 실무에서는 이 방식을 가장 많이 채택하는 것 같다.

### 도커 컴포즈

도커는 모든 컨테이너들을 한 번에 실행할 수 있게 해주는 도커 컴포즈 기능을 제공한다. 자세한 내용은 커플 채팅 어플리케이션을 개발하면서 작성한
    
    #3. Config docker-compose.yaml

게시글에서 확인할 수 있다.

상대적으로 아주 간편한 설정과 네트워크 환경 구성이 가능하고, 서비스 디스커버리 기능도 제공하면서 볼륨을 이용한 빠른 빌드로 개발 편의성을 높일 수 있다.

다른 방식들과 비교해 가장 큰 단점은 실제 쿠버네티스 위에서 개발하지 못한다는 것이다. 도커 컴포즈로 개발된 서비스를 운영환경의 쿠버네티스로 이전할 때, 쿠버네티스 manifest부터 PV설정, 네트워크 설정, 파드 스케줄링 설정 등을 재구성해야한다.

도커 컴포즈를 사용해서 개발시간이 줄어든만큼 운영환경을 구축하는데 시간이 추가적으로 소요되기 때문에, 결국 전체적인 시간은 유사하다고 볼 수 있다.

### minikube

미니큐브는 유명한 로컬 **단일 노드** 클러스터 구현체이다. 많은 개발자들이 로컬에서 테스트 및 개발 목적으로 사용하고있다.

미니큐브는 VM을 기반으로 로컬 쿠버네티스 클러스터를 생성한다. 즉 로컬의 호스트OS와 별개로 미니큐브 VM의 OS를 운영환경의 OS와 동일하게 설정이 가능하다는 것이다.

이로인해 쿨라우드 개발환경 서버를 구축하는 것만큼은 아니더라도 나름 실제 운영환경과 유사한 환경을 구성할 수 있다.

VM 기반이기에 생기는 단점도 있는데, VM은 별도의 게스트OS가 존재하기 때문에, 로컬에서 리소스 사용량을 많이 차지하게된다.

또 VM을 가동하기 위해서 VirtualBox 등의 가상화 구현 툴을 사용하게 되면, 여기서 또 추가적인 리소스가 사용된다.

추가적으로 앞에서 강조한 것처럼 **단일 노드** 클러스터여서 일반적으로 멀티 노드 클러스터로 구축되는 운영환경과는 차이가 있을 수 밖에 없고, 이로인해 스케줄링, 오토스케일링 등의 테스트는 불가하다.

희박한 가능성이지만, 배포하고자하는 서비스가 단일 노드로 구성되어있고, 자신의 로컬 개발환경 리소스 용량이 넉넉하다면 minikube는 더할 나위 없이 좋은 선택지가 될 것이다.

### vagrant

vagrant는 그 유명한 Terraform을 개발한 HashCorp 사의 프로비저닝 툴이다. 이미지를 통해 쉽게 VM 설정이 가능하다.

Vagrant는 쿠버네티스 클러스터 프로비저닝 툴이 아니라 VM 프로비저닝 툴이다. VM은 간편하게 설정하더라도, 그 VM 안에서 쿠버네티스 클러스터는 직접 구성해야한다. 여기서 복잡성이 크게 증가하게 된다.

또 minikube와 마찬가지로 VM 기반이기 때문에 로컬에서 상당한 리소스 사용량을 갖게된다.

### Kind

kind는 다중 노드 클러스터를 지원한다. 또 minikube와 vagrant와는 다르게 컨테이너 기반으로 클러스터를 구성한다. 

컨테이너 기반이기에 상대적으로 가볍고 로컬 리소스를 덜 차지한다는 장점과 함께, 컨테이너는 로컬의 호스트OS를 그대로 사용하기 때문에 실제 운영환경과는 이 부분에서 차이가 생길 수 밖에 없다.

---

## Kind를 선택한 이유

이번 팀 프로젝트의 개발환경 구축 도구로 Kind를 선택한 가장 큰 이유는 아래 두 가지이다.

1. 멀티 노드 클러스터 지원
2. 컨테이너 기반의 가벼운 환경

### 멀티 노드 클러스터 지원

실제 운영환경과 유사하게 멀티 노드에 서비스를 배포해 테스트할 수 있다. 오토스케일링, 노드셀렉터나 어피니티 등의 스케줄링, 서비스 메시를 활용한 서비스 간 통신을 테스트할 수 있다.

### 컨테이너 기반의 가벼운 환경

이 이유는 첫 번째 이유보다 더 큰 이유인데, 내 로컬의 리소스 용량은 8코어 CPU, 8GiB RAM, 256GiB 스토리지로 구성되어있다.

이번 프로젝트에서 컨테이너, 쿠버네티스, 서비스 메시까지 다룰 예정인데, 다른 리소스는 여자저차 하더라도 8GiB의 RAM은 정말 턱없이 모잘라도 너무 모자란다.

<br>

이러한 이유로 Kind를 선택하게 되었다.

---

## Kind Definition

Kind 공식 도큐먼트에서는 Kind를 아래와 같이 소개한다.

    kind is a tool for running local Kubernetes clusters using Docker container “nodes”.

nodes에 강조를 해서 멀티 클러스터 구축이 가능한 것을 강조하고 있다. 귀엽다.

처음 kind는 로컬에서 간단한 클러스터 테스트를 위해 개발되었지만 로컬 개발이나 CI에도 사용될 수 있다고한다.

Kind는 golang으로 작성된 도구이다. Go 최고.

---

## Kind Installation

조금 특이하다고 느낀 부분은 Kind를 사용하기 위해서 go를 설치해야한다.

go는 v1.16 이상, 그리고 도커도 설치해야한다. 둘 다 설치한 이후에

    go install sigs.k8s.io/kind@v0.20.0

커맨드를 실행하면 Kind가 설치된다. 만약 다른 방식으로 설치하고 싶다면

[패키지매니저/바이너리파일 설치](https://kind.sigs.k8s.io/docs/user/quick-start#installing-from-release-binaries)

brew나 choco 같은 패키지매니저, 혹은 바이너리 파일을 통한 설치도 가능하다. **어떤 방식으로 설치하든 go와 docker가 필요한 것은 변함 없다.**

go는 클러스터 생성, 이미지 빌드를 하기 위해 필요하고, docker는 컨테이너, 노드 등의 격리성을 구현하기 위해 필요하다고 한다. 

그리고 Kind는 멀티 노드 클러스터를 구축하기 위해 kubeadm을 사용한다고 한다! 

kubeadm에 대해 궁금하다면 나의 블로그 게시글

    Config Kubernetes Cluster w/ Kubeadm in Ubuntu from Scratch

에서 kubeadm에 대해 알 수 있다.

Kind가 잘 설치되었는지는

    kind version

커맨드나

    docker ps

로 현재 실행중인 컨테이너를 확인해볼 수도 있다. 

---

## Cluster Create / Delete

심플하다. 클러스터를 생성할 떄는

    kind create cluster

삭제할 때는

    kind delete cluster

---

## Container Image Deploy

실제 쿠버네티스 환경과 유사하게, 이미지 빌드 -> 이미지 pull -> manifest apply의 과정을 거쳐 배포가 이루어진다.

로컬 클러스터에서는 굳이 이미지를 이미지 레지스트리에 배포할 이유가 없기 때문에 빌드 후 이미지 pull이 아닌 이미지를 load 해주고 manifest를 apply하는 방식으로 진행된다.

    docker build -t 이미지이름:태그 ./도커파일 경로
    kind load docker-image 이미지이름:태그
    kubectl apply -f manifest이름:태그

현재 클러스터에서 사용중인 전체 이미지를 확인하고 싶으면 아래 커맨드를 사용한다.

    docker exec -it 노드이름 crictl images

예를 들어, 노드 이름으로 kind-control-plane을 입력하면 아래와 같이 출력된다.

![img](http://www.choigonyok.com/api/assets/64-1.png)

아무 복잡한 설정 없이 로컬에서 쿠버네티스 control plane Pod들이 잘 실행되고 있는 걸 확인할 수 있다.

## Build Container Image

## multi-node Cluster

Kind의 장점 중 하나인 멀티노드 클러스터를 구현하기 위해서는 yaml 파일이 필요하다. yaml 파일에 정의한만큼 워커노드가 생성된다.

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
```

더 많은 워커노드를 생성하고 싶다면 - role: worker를 원하는만큼 추가하면 된다.

이 yaml 파일의 파일명이 config.yaml이라고 가정하면, 클러스터를 생성할 때, --config attribute를 config.yaml으로 지정해주면 해당 yaml파일이 적용된다.

    kind create cluster --config config.yaml

kind에서는 멀티 노드 클러스터를 넘어서 클러스터 자체도 여러개를 생성할 수 있다. 물론 이름은 다르게 설정해줘야하는데, 그 부분은 생략하고 넘어간다.

만약 이미 default name으로 생성된 클러스터가 있다면, 앞서 말한 것처럼 같은 이름으로는 클러스터를 추가 생성할 수 없기 때문에 기존 클러스터를 
    
    kind delete cluster

커맨드로 삭제해준 후 config를 적용해서 재생성하면 된다.

## 리소스 사용량

우선 마스터노드 1개로 테스트 클러스터를 생성했다. 마스터노드에만 여러 컨트롤 플레인 파드들이 배치되어있는 상태이다.

![img](http://www.choigonyok.com/api/assets/64-2.png)

420m CPU와 700MB RAM이 사용되는 것을 볼 수 있다.

이번엔 마스터노드 1개, 워커노드 4개로 테스트 클러스터를 생성했다. 마스터노드에 컨트롤 플레인 파드들만 배치되어있으며 어플리케이션은 아직 워커노드에 배포되지 않은 상태에서 로컬 리소스 사용량은 아래와 같았다.

![img](http://www.choigonyok.com/api/assets/64-3.png)

570m CPU, 1.16GB RAM이 사용된다.

이후에 배포될 Istio가 일반적으로 3~4GiB의 Memory를 사용하는 것을 감안하면 상당히 큰 용량을 차지하는 것을 알 수 있다.

## 외부 포트매핑

외부에서 이 로컬 클러스에 접근하기 위해서는 포트매핑이 필요하다. Istio의 Ingress Gateway로 연 포트에 

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
```

## Istio와 연동

kind는 놀랍게도 서비스메시인 istio도 설치가 가능하다. Istio 공식 문서에서 kind 관련 내용을 다룬다.

https://istio.io/latest/docs/setup/platform-setup/kind/
