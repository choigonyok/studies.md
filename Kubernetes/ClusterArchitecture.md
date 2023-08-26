# 쿠버네티스 아키텍처 & 컴포넌트
# study kubernetes

---

## 개요

쿠버네티스는 컨테이너 오케스트레이션 플랫폼이다.

컨테이너 기반으로 배포되는 어플리케이션은 일반적으로 컨테이너 하나만으로는 배포가 불가능하다. 

컨테이너 오케스트레이션 플랫폼은 수많은 컨테이너들이 서로 통신하는 과정을 관리하고, 자원을 효율적으로 사용하고, 컨테이너의 상태를 모니터링하며, 컨테이너를 생성/변경/삭제하는 컨테이너 관련된 모든 기능들을 도와주는 플랫폼이다.

쿠버네티스의 아키텍처와 컴포넌트에 대해서 정리해본다.

---

## 쿠버네티스 아키텍처

![img](http://www.choigonyok.com/api/assets/4-1.png)

쿠버네티스에는 **클러스터**가 있다. 클러스터는 추상적인 개념이다.

하나의 네트워크로 묶여있는 마스터 노드, 워커노드들의 집합을 의미한다.

보통 어플리케이션은 하나의 클러스터로 구성되는데, 규모에 따라서 클러스터 여러개를 사용하고 L4 로드밸런서를 이용해서 트래픽을 여러개의 클러스터로 분산하는 경우도 있긴 한 것 같다.

controll plane, API server, Scheduler, etcd, kubelet, kube-proxy, nodes, control-manager, cloud-control-manager로 구성되어있는 걸 볼 수 있다.

### 마스터 노드

해당 다이어그램에 마스터노드는 따로 언급되지 않았지만, 마스터노드가 존재한다. 마스터노드는 클러스터를 관리하는 control plane 기능을 수행하는 노드이다. 

이 노드에는 일반적으로 어플리케이션 pod가 배치되지 않고, kube-system이라는 namespace의 쿠버네티스 기본 구성 pod들이 배치되어 각각 역할을 수행한다.

이 kube-system namespace에 배치되는 pod들이 각각 API server, Scheduler, etcd, kube-proxy, control-manager이다.

이중 kube-proxy만 마스터노드가 아닌 워커노드에 배치되고, 나머지는 모두 마스터노드에 배치된다.

마스터노드는 일반적으로 하나인데, 고가용성을 보장하기 위해 여러개의 마스터노드를 만들어두고 마스터노드에 장애가 생기면 다음 candidate가 마스터노드의 역할을 이어서 수행하는 식으로 시스템의 안정성을 보장할 수도 있다.

### 워커 노드

워커노드는 말 그대로 일하는 노드이다.

이 노드에는 사이드카 프록시나, BE,FE,DB pod 등 어플리케이션 배포에 필요한 다양한 pod들이 배치될 수 있다.

일반적으로는 워커노드 하나에 하나의 pod을 배치하는 게 일반적인 것 같다.

워커 노드 안에는 kube-proxy와 kubelet 컴포넌트가 들어가있다.

---

## 마스터노드 내부 컴포넌트

### Control Plane

control plane은 컴포넌트라고 보긴 어렵다. 클러스터의 기능을 관리하게 하는 컴포넌트들의 **기능 집합**을 추상화하는 용어이다.

control plane에 포함되는 컴포넌트들로는 etcd, proxy, Controller, APIserver, Scheduler 컴포넌트가 있다.

> Control Plane 기능을 수행하는 마스터노드 안에는 etcd, Controller, API server, Scheduler가 있다.

### etcd

클러스터를 위한 백엔드 데이터베이스이다. 여기에 클러스터 오브젝트 manifest, current state, desired state를 포함한 클러스터 전체의 설정, 구성 정보들이 저장되어있다.

etcd는 처음 클러스터를 설치하면 마스터노드에 포함되어있다. 이걸 마스터노드에서 별개의 노드로 분리해서 여러개의 etcd를 운영하는 게 일반적이다. 

etcd가 죽으면 아무 기능도 수행할 수가 없다. API 서버가 어디에 있는지, 워커노드들의 상태가 어떤지부터 모든 manifest 데이터도 다 사라지게된다.

쿠버네티스 클러스터에서 가장 중요한 컴포넌트라고 할 수 있다. 그래서 고가용성을 보장하기 위해서 etcd는 일반적으로 3개, 규모가 크면 5개로 구성한다.

하나의 etcd는 leader로 쓰기를 전담하고, 다른 etcd와 데이터 일관성을 유지하기 위해 다른 팔로워 etcd에 전달한다. leader가 아닌 follower etcd들은 읽기 전담이다. 

follower etcd중 하나는 follower이면서 동시에 candidate인데, leader etcd가 죽으면 candidate etcd가 새로운 leader로 선출되어 역할을 수행한다.

다른 컴포넌트들에서 etcd에 저장되어있는 데이터를 읽을 때 follower etcd를 이용한다.

### API server

클러스터 작업을 위한 인터페이스를 제공하는 말그대로 API이다. 

예를 들어, kubectl CLI로 클러스터를 관리할 때, 뒷단에서는 API서버로의 요청과 응답을 통해 모든 작업이 이루어진다. 

kubectl get pods로 pod들의 정보를 확인할 때도, kubectl create -f backend.yaml 으로 오브젝트를 생성할 때도 모두 API서버를 통해 작업이 이루어진다.

그래서 kubectl이 cert가 없거나, API서버가 다운되면 kubectl CLI로 아무 작업도 수행하거나 취소하거나 확인할 수 없다.

CLI 뿐만 아니라 모든 컴포넌트들간의 통신은 API서버를 통해 이루어진다. (etcd 제외)

### Scheduler

etcd manifest의 어피니티나 노드셀렉터 등을 체크하고, 파드의 request나 limit등의 리소스 요구사항과 노드의 리소스 상태도 확인하며 pod를 어디에 배치할지를 결정하는 컴포넌트이다. 

일치하는 노드가 여러 개이면 라운드로빈 알고리즘을 사용해서 배치시킨다.

### Controller

Controller는 etcd를 통해 현재 클러스터의 상태를 계속 업데이트하면서, 클러스터의 current state와 desired state를 비교한다.

current state와 desired state가 다를 때, API server를 통해 파드 재생성 요청을 하는 트리거가 된다.

---

이 모든 컴포넌트들은 앞서 말했듯이 kube-system이라는 namespace로 분리되어서 각 컴포넌트가 pod형태로 마스터노드에 배치되어있다.

이 pod들로 인해서 쿠버네티스는 마스터노드에 기본적으로 2GB 이상의 CPU와 RAM을 사용하도록 권장한다.

또 control plane 컴포넌트들이 하나라도 죽으면 전체 클러스터의 기능에 장애가 생기기 떄문에 일반적인 어플리케이션 pod는 마스터노드에 배치하지 않는 것을 원칙으로 한다.

> 참고로 쿠버네티스에서는 같은 클러스터여도 namespace가 다르면 논리적으로 다른 클러스터라고 인식한다. 네트워크에서 서브넷 나누는 느낌? 그래서 control plane 컴포넌트들이 kube-system namespace로 분리되어있는 것!

---

## 워커노드 내부 컴포넌트

### kube-proxy

kube-proxy는 클러스터 내의 네트워크 프록시 서비스를 제공한다.

쿠버네티스의 서비스 오브젝트를 내부적으로 구현하는 컴포넌트이다. 

1. 클러스터 내부 서비스를 외부에서 접근 가능하게 해주기도 하고, 
2. 서비스 오브젝트로 트래픽이 들어오면 해당 서비스와 연결된 pod들에게 로드밸런싱을 하기도 하고,
3. 서비스간 서비스 name으로 통신할 수 있는 서비스 디스커버리를 제공하기위해 API서버에게 노드에 있는 서비스들의 이름과 ip, port를 등록하기도 한다.

### kubelet

kubelet은 마스터노드의 API server와 통신하며 워커노드와 관련된 모든 작업들을 수행한다.

리소스를 생성하거나 상태를 모니터링하거나 하는 실질적인 모든 작업은 kubelet을 통해 이루어진다.

Controller가 desired state와 current state가 다른 것을 확인하고, Scheduler가 리소스를 어디에 배치해야할지를 판단해서 etcd에 저장되어있는 template을 API server가 kubelet에게 전달해주면, kubelet이 받은 데이터를 바탕으로 파드를 생성한다.

---

## 정리

![img](http://www.choigonyok.com/api/assets/48-2.png)

동작시나리오로 예시를 들어 정리하자면 아래와 같다.

### 가정 : 파드가 오류로 재생성되어야하는 상황

1. kubelet이 파드상태를 모니터링하다가 오류를 확인하고 API server로 상태를 보고한다.
2. API서버는 etcd에 파드 하나가 에러가 난 current state를 저장한다.
3. Controller는 주기적으로 etcd에서 current state를 확인하다가 desired state와 다름을 인지하고, API server에 파드 재생성을 요청한다.
4. API server는 Scheduler에게 어디에 배치하면 좋을지 데이터를 요청한다.
5. Scheduler는 etcd의 manifest를 참조해서 알고리즘, 어피니티, 노드셀렉터 등을 고려해 배치할 노드를 선정하고 API server에 응답한다.
6. API server는 Scheduler가 응답한 내용을 토대로 해당 노드에 있는 kubelet에게 etcd의 manifest 함께 파드 재생성을 요청한다.
7. 파드가 배치될 노드에 있는 kubelet은 받은 요청대로 파드를 재생성하고, 상태를 주기적으로 API서버에 보고한다.

> etcd는 API서버를 통해서 접근하지 않는다. Controller는 etcd를 모니터링하며 current state와 desired state를 비교하고, 둘이 일치되기 위해서 변경되어야할 사항을 API서버로 전달하는 것이다.