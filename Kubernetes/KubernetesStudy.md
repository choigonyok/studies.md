## ABOUT K8S A.K.A KUBERNETES

---

## 개요

쿠버네티스는 컨테이너 오케스트레이션 플랫폼이다.

컨테이너 기반으로 배포되는 어플리케이션은 일반적으로 컨테이너 하나만으로는 배포가 불가능하다. 

컨테이너 오케스트레이션 플랫폼은 수많은 컨테이너들이 서로 통신하는 과정을 관리하고, 자원을 효율적으로 사용하고, 컨테이너의 상태를 모니터링하며, 컨테이너를 생성/변경/삭제하는 컨테이너 관련된 모든 기능들을 도와주는 플랫폼이다.

쿠버네티스의 기능들과 공부하면서 배운 내용들을 정리해본다.

---

## 쿠버네티스 아키텍처

![img](/assets/KubernetesStudy.svg)

쿠버네티스에는 **클러스터**가 있다. 클러스터는 추상적인 개념이다.

하나의 네트워크로 묶여있는 마스터 노드, 워커노드들의 집합을 의미한다.

보통 어플리케이션은 하나의 클러스터로 구성되는데, 규모에 따라서 클러스터 여러개를 사용하고 L4 로드밸런서를 이용해서 트래픽을 여러개의 클러스터로 분산하는 경우도 있긴 한 것 같다.

controll plane, api server, scheduler, etcd, kubelet, kube-proxy, nodes, control-manager, cloud-control-manager로 구성되어있는 걸 볼 수 있다.

### 마스터 노드

해당 다이어그램에 마스터노드는 따로 언급되지 않았지만, 마스터노드가 존재한다. 마스터노드는 클러스터를 관리하는 control plane 기능을 수행하는 노드이다. 

이 노드에는 일반적으로 어플리케이션 pod가 배치되지 않고, kube-system이라는 namespace의 쿠버네티스 기본 구성 pod들이 배치되어 각각 역할을 수행한다.

이 kube-system namespace에 배치되는 pod들이 각각 api server, scheduler, etcd, kube-proxy, control-manager이다.

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

control plane에 포함되는 컴포넌트들로는 etcd, proxy, controller, apiserver, scheduler 컴포넌트가 있다.

> Control Plane 기능을 수행하는 마스터노드 안에는 etcd, proxy, controller, api server, scheduler가 있다.

### etcd

클러스터를 위한 백엔드 데이터베이스이다. 여기에 클러스터 오브젝트들에 대한 정보나, 클러스터에 포함된 노드 정보나, 이런 클러스터 관련 정보들이 저장되어있다.

오브젝트들의 template 데이터가 여기에 저장되어있고, replicaSet에 의해 pod를 배포해야할 때 etcd애 저장되어있는 template을 참조해 pod를 생성한다.

etcd는 처음 클러스터를 설치하면 마스터노드에 포함되어있다. 이걸 마스터노드에서 별개의 노드로 분리해서 여러개의 etcd를 운영하는 게 일반적이다. 

쿠버네티스의 중요한 역할 중 하나가 컨테이너에 오류가 발생하면 스스로 재시작하거나, 트래픽이 늘어날 때 auto-scailing하는 것인데, etcd가 죽으면 그 역할을 수행하지 못한다.

그래서 고가용성을 보장하기 위해서 etcd는 일반적으로 3개, 규모가 크면 5개로 구성한다.

하나의 etcd는 leader로 쓰기를 전담하고, 다른 etcd와 데이터 일관성을 유지하기 위해 다른 팔로워 etcd에 전달한다. leader가 아닌 follower etcd들은 읽기 전담이다. 

follower etcd중 하나는 follower이면서 동시에 candidate인데, leader etcd가 죽으면 candidate etcd가 새로운 leader로 선출되어 역할을 수행한다.

다른 컴포넌트들에서 etcd에 저장되어있는 데이터를 읽을 때 follower etcd를 이용한다.

### apiserver

클러스터 작업을 위한 인터페이스를 제공하는 말그대로 API이다. 

예를 들어, kubectl CLI로 클러스터를 관리할 때, 뒷단에서는 API서버로의 요청과 응답을 통해 모든 작업이 이루어진다. 

kubectl get pods로 pod들의 정보를 확인할 때도, kubectl create -f backend.yaml 으로 오브젝트를 생성할 때도 모두 API서버를 통해 작업이 이루어진다.

그래서 kubectl이 cert가 없거나, API서버가 다운되면 kubectl CLI로 아무 작업도 수행하거나 취소하거나 확인할 수 없다.

CLI 뿐만 아니라 모든 컴포넌트들간의 통신은 API서버를 통해 이루어진다.

### scheduler

pod를 어디에 배치할지를 결정하는 컴포넌트이다. 

어피니티나 노드셀렉터 등을 체크하고, 파드의 request나 limit등의 리소스 요구사항과 노드의 리소스 상태도 확인하며 pod를 배치한다.

일치하는 노드가 여러개이면 라운드로빈 알고리즘을 사용해서 배치한다고 한다.

### controller

이 모든 컴포넌트들은 앞서 말했듯이 kube-system이라는 namespace로 분리되어서 각 컴포넌트가 pod형태로 마스터노드에 배치되어있다.

이 pod들로 인해서 쿠버네티스는 마스터노드에 기본적으로 2GB 이상의 CPU와 RAM을 사용하도록 권장한다.

또 control plane 컴포넌트들이 하나라도 죽으면 전체 클러스터의 기능에 장애가 생기기 떄문에 일반적인 어플리케이션 pod는 마스터노드에 배치하지 않는 것을 원칙으로 한다.

> 참고로 쿠버네티스에서는 같은 클러스터여도 namespace가 다르면 논리적으로 다른 클러스터라고 인식한다. 네트워크에서 서브넷 나누는 느낌? 그래서 control plane 컴포넌트들이 kube-system namespace로 분리되어있는 것!

---

## 워커노드 내부 컴포넌트

### kube-proxy

kube-proxy는 클러스터 내의 네트워크 프록시 서비스를 제공한다.

클러스터의 큰 장점 중 하나가 서비스 디스커버리를 제공한다는 건데, 

---

## 파드

pod는 컨테이너가 배포되는 곳이다.


우선 pod가 있다. pod는 쿠버네티스 상에서 가장 작은 단위로 본다. 오토스케일링이나 레플리카셋을 이용한 재생성, 삭제 등 모두 파드 단위로 이루어진다.

이 pod안에는 컨테이너가 들어있다. 컨테이너는 하나 이상 들어있을 수 있다.
k8s에서는 yaml파일을 통해 컨테이너를 pod에 배포한다
deployment -> pod -> service의 과정

이미지를 containerilze해서 pod에 배포하려면 deployment로 만들어야하고
이 과정은 yaml파일을 통해 이루어진다.
yaml파일에 이미지 등 config -> yaml파일로 pod 생성

docker-compose는 개발환경을 위해 컨테이너를 묶는 기술
docker-compose와 k8s는 호환이 안되고 그래서 docker-compose의 볼륨이나 환경변수나 기타 설정들을 deployment.yml파일로 재정의 해야함

클러스터를 배포할 때 
vpc를 지정해서 인스턴스들이 같은 네트워크 대역을 쓰게 하는데
이건 기본적으로 private vpc로 선언되기 때문에, 외부에서 vpc인스턴스에 ssh등으로 접근하려면
internet-gateway 가 필요함
그리고 서브넷도 기본적으로 private 서브넷이 생성되는데, 사용자가 접근해야하는 서비스들은(주문, 목록조회 등)
퍼블릭 서브넷을 생성해서 그 안에 인스턴스 ip를 할당해야함
그러려면 map_public_ip_on_launch = true 설절을 서브넷에 해주면 퍼블릭 서브넷이 됨
라우팅테이블도 설정해야함
라우팅테이블은 외부에서 접근할 때 vpc가 여러 서브넷으로 나눠져있으니까 어떤 서브넷에 가야하는지
라우팅테이블을 갖고있는 것

pod에 컨테이너를 배포하는 과정은
dockerfile로 정의된 image를 빌드해서 컨테이너 레지스트리(도커허브)에 푸시
쿠버네티스에서 deploy yml파일에 도커허브에서 pull해올 이미지를 지정
yml파일 기반으로 deployment을 create하고, 이 deployment를 기반으로 pod를 실행

클러스터를 클라우드에 배포하는 과정은
우선 인프라를 구축하고
각 인프라 안에서 kubeadm과 kubelet kubectl로 마스터노드 설정, 워커노드 설정, 노드 join 등의 과정을 거침
이 때 스크립트 파일을 이용해서 자동화할 수 있고,

일반적으로는 클라우드에서 제공하는 k8s 클러스터 구축 툴(tsk?)을 이용해서 간편하게 구축

나는 과정을 공부해야하니까 nginx conf파일에서 수정했던 것처럼 가장 올드한 방식으로 도전

내일은 클러스터 스크립트 구성 좀 더 건드려보고,
chatservice프로젝트 클러스터에 배포 하는 거 건드려보고,
마이크로서비스 책 읽고

scale command로 template과 replica 를 통해 pod단위로 확장/축소가 가능
롤링업데이트를 지원해서 서비스가 중단되지 않으면서 확장 축소가 가능함
롤백을 통해 이전 버전으로 되돌릴 수 있음ㄴ
deployment kind를 통해 배포하면 service가 자동 생성
서비스는 해당 pod에 접근할 수 있게 endpoint를 열어주는데,
deployment로 만들어진 서비스는 클러스터 내부에서 해당 pod에 접근할 수 있도록 endpoint를 열어주는 역할이고,
expose command를 사용해 NodePort를 열어주면 외부(브라우저 등)에서도 접근 가능한 endpoint가 생성됨
서비스는 3종류가 있는데, NodePort, LoadBalancer, ClusterIP
노드포트는 외부접근 서비스
클러스터아이피는 내부접근 서비스
로드밸런서는 말그대로 로드밸런서 서비스

    로드 밸런서는 외부에서 들어오는 트래픽을 클러스터 내부의 서비스로 분산시켜주는 역할을 수행합니다. 이렇게 로드 밸런서를 설정하면 클러스터 내의 모든 노드에 노드 포트를 개별적으로 생성할 필요가 없어지며, 단일 IP 주소와 포트를 사용하여 서비스에 접근할 수 있습니다.
    로드밸런서가 없으면 nodeport가 있어야 외부접근 가능한데
    로드밸런서 있으면 로드밸런서가 외부접근을 받아서
    클러스터 내부 pod에 라우팅하는게 가능
    -> 근데 이 내용은 chatgpt가 오락가락해서 로드밸런서 있으면 nodeport 없어도 되는게 사실인지 확인해봐야할 것 같음

kubectl pause : 실행중인 pod을 중지함
kubectl rollout pause : deployment를 pause, 현 상태에서 pod가 하나 삭제되어도 replica에 맞게 새로운 pod를 생성하거나 하는 걸 멈추게 됨. 업데이트 등이 필요할 때 사용됨. 실행중이던 pod가 중단되진 않음
kubectl resume
kubectl rollout resume
kubectl rollout history : 버전 기록 확인
kubectl rollout undo : 이전 버전 돌아가기

label은 리소스(pod 등)를 식별
label로 로드밸런싱하거나, label별로 모니터링하거나 특정 노드에서 특정 파드만 실행되도록 할 수 있음
node에 label을 지정하고 pod에 nodeSelector을 지정하면 해당 파드는 일치하는 노드에서만 작동
label은 명령어도로 지정 가능
kubectl label nodes <name> <label key>=<label value>
kubectl label pods <name> <label key>=<label value>
이걸 통해 msa에서 백엔드와 고유한 폴리글랏 저장소가 서로를 인식할 수 있음
이러려면 go에서는 아예 db관련 코드가 싹 바뀜. import부터 k8s관련 패키지 수두룩

k8s api 서버 = 마스터노드 = 컨트롤플레인 이렇게 같은 의미로 사용하는 경우가 있는데
셋은 다른 것!!!
    k8s api 서버는 클러스터의 리소스에 접근해서 crud가능하게 해주는 서버, 마스터 노드 안에 위치함
    -> mysql pod에 접근해서 데이터 조회 등을 할 때 이용

    마스터노드는 etcd, 컨트롤러 매니저 등의 클러스터를 관리하는 컴포넌트와 k8s api서버를 포함한 여러 컴포넌트들이 모인 인스턴스

    컨트롤플레인은 클러스터 관리, pod 배포, 스케줄링등의 역할을 하는 컴포넌트의 그룹, 얘도 마스터 노드 안에 있음
정리하면, 마스터 노드 안에 여러 컴포넌트들이 있고, 컨트롤플레인은 클러스터 상대 관리하는 컴포넌트들, k8s api서버는 클러스터의 리소스에 접근해서 crud하는 컴포넌트를 의미함

어플리케이션이 오류가 생겨도 파드는 문제없이 실행될 수 있음
그래서 어플리케이션이 잘 작동하는지 헬스체크가 필요함
1. url이용 : pod 생성시 livenessProbe, readinessProbe 지정
   1. 컨테이너에 주기적으로 command 실행
헬스체크를 해야 문제가 생긴 파드를 재시작하고, 트래픽을 문제없는 다른 노드로 라우팅할 수가 있기때문에, 운영환경에서는 헬스체크를 꼭 포함시켜야함
livenessProbe : 컨테이너가 실행중인지를 확인, 실패하면 pod 재시작
readinessProbe : 컨테이너가 요청을 처리할 수 있는 상태인지를 확인, pod는 재시작하지 않고 검증이 실패하면 이 pod로 라우팅되지 않게하고, 다시 readiness 확인 후 확인되면 트래픽을 다시 받음
Probe는 컨테이너 단위로 진행되는데, Pod는 내부의 모든 컨테이너가 정상이어야 pod로 트래픽 라우팅을 할 수 있음
예를 들어, 한 파드에 컨테이너 A,B,C가 있는데, B가 livenessProbe에 실패하면 컨테이너B만 재시작하고 A,C는 그대로 있음
그래도 A,B,C는 한 파드에 있기 때문에 A,C는 정상상태이지만 트래픽을 라우팅받지는 못함


DNS 서비스
- 같은 클러스터 내의 다른 pod에 접근할 수 있게 해주는 k8s 기본 시스템
같은 pod의 다른 컨테이너끼리는 localhost:port로 접근 가능하고,
다른 pod끼리 접근할 때 사용
같은 클러스터라는 것은, namespace가 같다는 말임.
물리적으로 같은 클러스터 안에 있어도 namespace가 다르면 논리적으로 클러스터가 분리된 것이라 DNS로도 서칭이 불가능


볼륨과 볼륨마운트

apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
spec:
  containers:
    - name: nginx-container
      image: nginx
      volumeMounts:
        - name: config-volume -> 이 볼륨을 쓰겠다고 명시
          mountPath: /etc/nginx/conf.d -> 그 볼륨을 컨테이너의 이 위치로 마운트
  volumes:
    - name: config-volume -> 이 이름으로 볼륨을 설정, 로컬 볼륨
      configMap: -> 볼륨 내용으로 configMap을 이용해서 볼륨을 만듦
        name: nginx-map -> configMap의 이름, configMap은 만들기에 따라 여러개 있을 수 있음
        items:
          - key: proxy.conf -> nginx-map이라는 configMap의 proxy.conf키를 입력
            path: aaa/bbb.conf -> 그 키의 값을 볼륨의 이 경로에 저장
정리하면, volume의 디렉토리는 /aaa/bbb.conf 인거고
원래 컨테이너의 디렉토리는 /etc/nginx/conf.d 인건데
컨테이너의 /etc/nginx/conf.d path에 volume을 마운트했기에
컨테이너는 /etc/nginx/conf.d/aaa/bbb.conf 의 path를 갖게됨
근데 aaa/bbb.conf는 volume으로 참조하는 값이기에 이 path에서 변경된 사항은 컨테이너에 바로바로 적용가능

ingress controller
외부 로드밸런서 대신 라우팅해주는 쿠버네티스 기본 기능
url의 host에 따라 정해진 service로 라우팅할 수 있음
default를 설정해두면 일치하는 host가 없을 경우 default 서비스로 라우팅하게 됨
ingress의 포트를 80, 443 모두 설정해두면 http, https 모두 라우팅 가능

persistenceVolume

podpreset을 통해서 설정정보를 따로 관리할 수 있음
이걸 이용하면 해당 label에 일치하는 pod들은 모두 해당 설정정보를 갖게 됨(볼륨, 시크릿, 환경변수 등)
이건 모든 namespace에 동일하게 적용되서 장점이 있지만, 일반적으로 그런 특수한 상황이 아니면 deployment에 설정정보를 지정하는 방식을 더 많이 사용함

내부스토리지 등의 pv를 사용할 때
pv를 일단 선언해야하고
해당 pv를 claim하는 pvc를 선언해야하고
그 claim을 사용할 pod에서 persistentVolumeClaim의 claimname을 지정해줘야
pv를 사용할 수 있음

service는 pod로 라우팅해주는 창구느낌
pod를 재시작한다고 service까지 재시작할 필요는 없다.
**서비스 타입을 nodeport로 설정 안해도 파드끼리 클러스터ip로 통신이 가능함**
DNS서비스처럼 DB host ip를 매번 get svc해서 찾고싶지 않으면 go 코드에서 k8s 관련 함수를 이용해서 클러스터 ip를 얻을 수 있음! ->이건 나중에 k8s 패키지를 더 공부해봐야할 듯함 

statefullset
파드의 이름을 pod-랜덤스트링이 아닌 인덱스 기준으로 설정해서, 파드가 스케줄링되거나 생성/삭제 되어도 변함없이 DNS서비스를 유지할 수 있게 해주는 기능

daemonset
노드당 하나씩 pod가 존재하게 하는 것
노드가 하나 추가되면 그 노드에 pod가 생성되고,
삭제되면 그 pod도 삭제됨. -> 다른 노드로 스케줄링 되지 않음
노드별 모니터링, 로그수집 등에 사용

request/limit
리소스 모니터링을 통해 일정 cpu 사용량이 넘거나 부족하면 자동으로 노드를 확장/축소하는 기능
1000m이 1코어, 200m은 0.2코어 사용을 의미
4코어 노드에서 리소스를 1000m으로 설정하는 건 전체 사용량의 1/4 = 25%를 사용한다는 뜻
request와 limit이 있음
request를 200m으로 설정했다는 것 = 이 파드는 최소 0.2코어는 있어야만 한다는 것
limit을 200m으로 설정했다는 것 = 이 파드에는 0.2코어까지만 리소스를 주겠다는 것
limit이랑 request 수치를 동일하게 하는게 좋다고 함 (https://www.slideshare.net/try_except_/optimizing-kubernetes-resource-requestslimits-for-costefficiency-and-latency-highload)
ㄴ limit이 너무 작으면 cpu 쓰로틀링이 생기긴하지만 request가 너무 커서 메모리가 뻑나는 것보단 낫기 때문
request/limit 체크는 default로 30초마다하고, 변경 가능

autoscaling
HorizontalPodAutoscaler kind로 수평확장 오토스케일러 설정 가능
기준점은 이런게 있음
targetCPUUtilizationPercentage: 이미 언급한 것처럼, CPU 사용률의 목표치를 설정하여 해당 목표치를 유지하도록 오토스케일링합니다.
targetMemoryUtilizationPercentage: 메모리 사용률의 목표치를 설정하여 해당 목표치를 유지하도록 오토스케일링합니다.
targetCustomMetricValue: 사용자 정의 메트릭 값을 기준으로 오토스케일링할 수 있습니다. 이를 사용하여 특정 지표나 애플리케이션의 특정 동작에 따라 스케일링할 수 있습니다.
targetRequestPerSecond: 초당 요청 수를 기준으로 오토스케일링할 수 있습니다. 주로 HTTP 요청이 많이 발생하는 웹 서비스에서 사용됩니다.
    **pod들의 평균 값을 기준으로 높아지면 평균을 낮추기 위해 확장시키고, 낮아지면 높이기위해 파드 수를 축소시키는 것!!!**

affinity/antiaffinity
node affinity : nodeselector과 유사
pod affinity/anti affinity : 파드 간 관계를 정의해서 스케줄링에 참고하게 할 수 있음
1. 파드간 규칙을 정하는 것
2. 그 규칙을 얼마나 hard하게 지킬지 preference를 정하는것
이건 pod가 생성되고 처음 스케줄링시에만 적용되는거라서
만약 규칙이 있는데 어쩔 수 없이, 또 preference가 약해서 예외처리 된 게 있으면
나중에 클러스터 상황이 규칙적용이 가능해진 상태로 바뀌어도 자동으로 리스케줄링 되지 않음
직접 pod를 삭제하고 재생성해야함
node affinity의 기준으로 label, region, hostname, instance type 등이 있고, pod는 이 기준들을 점수로 매겨서 총합이 가장 높은(가장 weight가 높으면서도 여러개가 일치하는) 노드에 우선적으로 배포됨
노드에 어떤 label이 지정되어있는지 알고싶으면 kubectl get node <node name>을 통해서 알 수 있음

interpod affinity/anti-affinity
파드 간 어피티니/안티어피니티
label에 맞는 **pod**를 찾고 여기서는 **토폴리지키**라는 게 있는데
해당 파드의 토폴로지 값과 일치하는 노드에 pod가 배포됨
예를들어, node1은 ap-northeast-2a, node2,3은 ap-northeast-2b에 있고,
label이 app=backend라는 파드가 노드1에 배포되어있는데,
새로운 파드에 podAffinity를 label은 app=backend, 토폴로지키는 region is not ap-northeast-2a로 지정해서 pod를 생성하면
새로운 파드가 app=backend라는 레이블을 보고, 일치하는 pod와 토폴로지 키를 본 뒤
노드 2 또는 3에 새로운 pod를 배치하게 됨
podAffinity는 토폴로지키에 일치하는 곳으로, podAntiAffinity는 토폴로지키와 일치하지 않는 곳으로 배치되게 됨

toleration
affinity와는 다르게, pod에 toleration이 **지정되어있으면** 일치하는 노드에만 배포될 수 있는것
예를들어, toleration이 설정되어있지 않은 pod1은 어느 노드에나 배치가 가능함.
근데 toleration이 설정되어있으면 해당 toleration 조건에 맞는 노드에만 배치가 가능함
node affinity는 pod의 affinity rule과 node의 affinity rule이 일치하는 경우에만 배치가능
toleration은 노드엔 taint pod에는 toleration을 설정해서 비교해야하고, node affinity는 pod에서만 지정하면 되는것임
NoSchedule : 스케줄링 시 일치하지않는 node에는 pod를 배포시키지 않음
PreferNoSchedule : 우선적으로는 일치하는 node에는 pod를 배포시키지 않음
NoExecute : toleration이 일치하지 않는 pod가 노드에 배치되면 설정한 time만큼은 노드에서 pod를 실행시키고, 그 이후에는 삭제시킴
noschedule과는 아예 올리지도 못하게 하는거랑, 올려두고 언제 삭제할지를 정하는 것의 차이!
  뭔가 잘못알고있는 것 같음. 다시 공부하기

etcd의 분리
etcd는 리소스 정보 등을 담고있는 데이터베이스
이게 중요하기 때문에, default는 마스터노드에 포함되어있지만
1. 마스터노드는 작업량이 커 cpu가 많이 필요한데 반해, etcd는 그렇지 않고, 
2. etcd는 중요해서 안정성을 높여야하기 때문에
etcd를 마스터노드에서 분리해서 각 etcd를 개별 노드에 배치하는 게 일반적.
보통 3개의 etcd를 쓰고, 많으면 5개?
etcd는 leader, follower, candidate로 나뉘는데
leader는 읽기/쓰기/follower에게 데이터 전달을 담당.
follower는 leader가 준 데이터 동기화/읽기 담당
candidate는 leader가 죽으면 leader역할을 수행
보통 etcd가 3개니까, 1번 : leader 2번 : follower+candidate 3번: follower 이렇게 구성
쓰는 건 leader만 쓸 수 있는데 읽을 땐 아무 etcd나 가서 읽어도됨. (leader, follower 다 읽기 기능을 지원하기 때문)

마스터노드 고가용성(HA : High Availability)
etcd의 안정성을 위해 3개 운영하는 고가용성처럼 마스터노드도 여러개 운영 가능
같은 ap-northeast-2여도
리전마다(ex: ap-northeast-2a, ap-northeast-2b, ap-northeast-2c) 마스터노드를 지정해두면
한 지역에서 장애가 생겨도 (자연재해나, 오류 등) 다른 리전의 마스터가 바톤터치해서 클러스터를 유지시킬 수 있음