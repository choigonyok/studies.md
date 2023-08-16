## INGRESS CONTROLLER IN KUBERNETES

---

## 개요

nginx 인그레스 컨트롤러 설치방법과 함께 L4, L7로드밸런서, 리버스프록시의 역할과 차이에 대해 살펴보겠다.

---

인터넷에서 쿠버네티스 표준 아키텍처라는 이미지를 본 적이 있다.
API 서버 로드밸런서로는 HAPROXY가 올라와있었다. 쿠버네티스 네트워크에 대한 개념이 깊게 잡혀있지 않았던터라 "API 서버 로드밸런서가 뭐지? 지금은 클러스터 공부중이니까 넘어가~" 하고 넘겼다.
클러스터 프로비저닝 툴으로는 테라폼과 kubespray, kubeadm이 올라와있었다.
테라폼은 클러스터를 위한 인프라 프로비저닝, kubespray는 구성관리를 위한 자동화 툴, kubeadm은 쿠버네티스 설치를 위한 툴
다 알고있는 것들이었는데, 여기 쓰이는 kubeadm이 여기서 발목을 잡을 줄 몰랐다.

내가 많은 프로비저닝 툴 중 kubeadm을 클러스터 프로비저닝 툴로 선택한 이유는 다음과 같다.

1. 쿠버네티스 공식 툴이어서
2. 특정 클라우드 플랫폼에 종속되지 않아서
3. 온프레미스 환경에서도 구축가능해서
4. 편의성보다는 클러스터 구축에 대해 완벽한 이해를 하고싶어서

kubeadm같이 from scratch로 AtoZ 다 건들면서 구축하는 이런 방식을 baremetal 이라고도 한다.
물론 나중에는 EKS등의 high level 툴들도 다 써보겠지만 지금은 쿠버네티스를 처음만나 알아가는 단계니까 천천히 시작하고 있었다.

그래서 kubeadm이 인그레스 컨트롤러, HAPROXY와 무슨 상관이냐?

kubeadm은 kubernetes에서 LoadBalancer type의 서비스를 지원하지 않는다.
Service는 크게 ClusterIP, NodePort, LoadBalancer 로 나뉘는데, 이 LoadBalancer타입을 지원해야 External-IP도 생성하고 그 IP를 NLB와 연결해서 클라이언트의 접근이 인그레스 컨트롤러를 거쳐 인그레스 룰에 따라 알맞는 서비스와 파드로 라우팅될 수 있는데, 이걸 지원하지 않는다. 이 내용은 스택오버플로우를 뒤지다 발견했고, 발견하고의 충격은 이루 말할 수가 없었다...

[스택오버플로우](https://stackoverflow.com/questions/44110876/kubernetes-service-external-ip-pending)

이 글을 보고 어째야하나... 하다가 저 이미지를 봤던 기억이 스쳐갔다.

그리고 찾아보니 HAPROXY가 kubeadm 기반으로 구성된 bare metal 클러스터에서 ingress controller대신 사용할 수 있는 로드밸런서 겸 리버스프록시라는 것을 알게 되었다.

HAProxy는 강력한 로드밸런서, 리버스프록시 구현 오픈 소스 툴이다.
이름을 뜯어보면 HA와 Proxy를 합친 말이다. HA는 High Availability, 고가용성이다.
고가용성은 하나가 죽어도 서비스가 중단되지 않고 다른 대기하고 있던 컴포넌트가 죽은 컴포넌트 대신 역할을 수행할 수 있게 하는 것이다. 이런 걸 고가용성이 높다고 한다.
고가용성이 높은 리버스프록시 및 로드밸런서를 구현할 수 있게 해주는 툴인 것이다.

HAProxy를 이용한 클러스터 L7 로드밸런싱의 큰 개념은 다음과 같다.

HAProxy를 클러스터 내부 pod로 실행 -> HAProxy Service를 NodePort로 실행 -> Ingress 생성 -> 외부 L4 로드밸런서와 NodeIP:NodePort로 연결 -> 외부에서 요청시 HAProxy는 Ingress rule에 따라 API server를 확인해 알맞는 서비스로 라우팅, 이때 모든 다른 서비스들은 ClusterIP 타입으로 설정 

이러면 HAProxy의 포트만 NodePort로 열고 다른 백엔드 서비스들의 포트는 ClusterIP로 안전하게 유지될 수 있다. 관련한 좋은 글이 있어 밑 참고 탭에 남기겠다.

여담으로, medium이란 사이트는 참 유용하고 좋은 정보들이 잘 담겨있는 것 같다.

우선 helm 패키지 매니저를 설치한다. haproxy는 차트로 배포될 것이다.

    curl https://baltocdn.com/helm/signing.asc | sudo apt-key add -
    sudo apt-get install apt-transport-https --yes
    echo "deb https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
    sudo apt-get update
    sudo apt-get install helm   

HAProxy를 설치한다.

    helm repo add haproxytech https://haproxytech.github.io/helm-charts

차트를 새로고침한다.

    helm repo update

헬름으로 인그레스 컨트롤러를 등록한다. 

    helm install myingress-controller haproxytech/kubernetes-ingress

라우팅할 인그레스 룰을 정하기 위해 인그레스 리소스를 생성해준다.

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-controller
spec:
  ingressClassName: haproxy
  rules:
      - host: www.example.com
      http:
      paths:
        - path: /
          pathType: Prefix
          backend:
          service:
              name: backend-service
              port:
                  number: 8080

이제 haproxy가 배포된 pod가 배치된 노드의 publicIP:<NodePort>로 브라우저에서 접근이 가능하고 동시에 로드밸런싱도 인그레스 룰에 따라 진행할 수 있게 된다.
SSL 설정 등의 고급 기능을 포함한 haproxy를 구현하려면 차트를 커스터마이즈해서 배포해야하는데, helm에 대한 이해도가 아직 낮기때문에, Helm에 대해 먼저 공부하고 이어서 진행하도록 하겠다.

이제 로드밸런서에 연결해보겠다.

대상그룹을 설정해야하는데 대상이 HAPROXY가 위치한 노드의 IP와 NodePort니까 타겟구성을 IP로 지정하고, 로드밸런서와 일치하게 tcp 80으로 설정한다.
다음 단계인 대상으로는 HAPROXY가 위치한 노드의 **PrivateIP**를 입력해주고, 대상 포트로는 NodePort를 입력해준다.
이제 로드밸런서를 NLB로 생성하고 tcp 80번 포트로 지정한다. 그럼 로드밸런서가 프로비저닝되고 DNS네임이 생성된다. 프로비저닝되는데 시간이 좀 걸리고, 완료된 이후에는 해당 DNS네임으로 80번포트(=포트입력 없이)에서 http 요청을 성공적으로 보낼 수 있게된다. 이 과정을 요약해보자면

클라이언트에서 NLB DNSname으로 요청 -> NLB와 연결된 HAPROXY의 NodePort를 통해 트래픽이 들어온다 -> HAPROXY POD에서 Ingress 오브젝트와 K8S API서버를 참조해서 트래픽을 Ingress rules에 따라 라우팅한다 -> 트래픽이 정상적으로 백엔드 서비스에 도착한다 -> 백엔드 서비스가 트래픽을 Pod로 보낸다 -> Pod의 컨테이너에서 트래픽을 처리한다.

이렇게 Ingress Controller 오브젝트 없이 HAPROXY를 통해 LoadBalancer 타입의 서비스를 지원하지 않는 kubeadm으로 baremetal 구성된 클러스터에서 클라이언트의 80포트 요청을 처리하는 방법을 알아보았다.
참 한 줄로 정리하기도 길고 벅찬 내용이었다. 하지만 잘못알고있던 것들이 고쳐지고 새로운 지식을 흡수할 수 있어서 재밌었다!

지금 현재는 AWS에서 제공하는 의미없는 문자열의 나열인 DNS Name을 통해 접속하는데, 이걸 내가 이전에 구매하고 등록해둔 도메인과 연결시키는 방법에 대해 알아보겠다. 이 과정을 거치면 도메인인 www.choigonyok.com으로 클러스터 백엔드 서비스에 접근이 가능해질 것이다!

















---

## nginx ingress controller install

인그레스 컨트롤러를 설치하는 방법은 크게 두 가지가 있다.

1. Yaml파일 작성 후 kubectl apply -f로 실행
2. Helm 차트를 이용한 설치

헬름에 대한 부분은 추후에 다뤄보고, 1번 방식으로 컨트롤러를 설치해보겠다.

    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml

이 코드를 실행하면 nginx-ingress-controller를 클러스터에 설치할 수 있다. 이 yaml파일에 대한 내용은 따로 글을 작성하겠다. 대략적으로 service user 설정, role 배포, role 바인딩, 서비스 및 디플로이먼트 생성 등이 포함되어있다.

    kubectl get pods --namespace=ingress-nginx

이 커맨드로 컨트롤러를 위한 pod들을 확인할 수 있다. nginx-ingress-controller의 namespace는 ingress-nginx이다.

이렇게 nginx-ingress-controller가 클러스터에 배포되면 

    kubectl get svc ingress-nginx-controller --namespace=ingress-nginx

이 커맨드를 입력했을 때 external-ip가 생겨야한다. 만약 <pending>상태라면 클러스터가 로드밸런서를 프로비저닝할 수 없는 상태인 것이다.

external-ip가 있다면, DNS 레코드가 해당 ip를 가리키도록 설정한다.

그 다음 알맞은 인그레스 리소스를 생성하면 된다.


AWS에서 nginx 인그레스 컨트롤러가 외부로 노출되게 하려면 NLB가 필요하다. 여기서 AWS의 ELB, ALB, NLB에 대해 간략히 보고 넘어가자.

### AWS ELB

ELB는 Elastic LoadBalancer의 줄임말이다. ELB라는 서비스가 AWS에 있는 것은 아니고, AWS에서 제공하는 로드밸런서 서비스를 포괄적으로 ELB라고 부른다.

ELB에는 크게 두 가지가 있다.

ALB (Application LoadBalancer) : 어플리케이션 계층(OSI L7)에서 작동하는 로드밸런서이다. tcp/udp 기반으로 ip주소와 port를 통해 라우팅한다.
NLB (Network LoadBalancer) : 전송계층(OSI L4)에서 작동하는 로드밸런서이다. 고정 ip를 제공하고, http/https 헤더, 본문 내용, URL경로, 쿠키 정보 등을 통해 라우팅한다.

L4 로드밸런서는 전송계층에서 작동한다. TCP/UDP와 포트 기반으로 데이터 패킷을 라우팅한다. 요청의 내용을 보고 분석해서 라우팅하는게 아니라 단순히 프로토콜, ip주소와 포트만을 보기 때문에 암호화를 할 필요가 없어서 더 속도가 빠르다. 반대로 이게 단점이 되기도 하는데, 요청의 자세한 내용을 모르기떄문에 라우팅은 라운드로빈 등의 심플한 알고리즘을 사용해서 로드밸런싱 한다. (그래서 분산시스템을 위해 클러스터 앞단에 분배하는 듯, 클러스터가 여러개면 어디 주든 같은 클러스터라서 상관 없으니까)

이에 반해 L7 로드밸런서는 어플리케이션에서 작동한다. HTTP를 기반으로 패킷을 라우팅한다. 메시지를 읽고 HTTP헤더, 요청본문 내용, 요청 URL경로, 쿠키 내용을 보고 필요한 서비스에 라우팅하기 때문에, 시간이 상대적으로 더 걸리고, 암호화도 필요할 수 있다.(요청 내용을 읽기 때문에) 일반적으로 브라우저에서 SSL/TLS 암호화를 통해 HTTPS 요청을 전달하면 L7에서 요청을 해독하고, SSL/TLS 네트워크 트래픽이 종료시킨다. L7는 해독된 요청을 기반으로 라우팅 대상을 정해서 새로운 연결을 맺는다. 그리고 L7 로드밸런서에서 새롭게 해당 서버에 요청을 보낸다. 이 L7 로드밸런서가 없으면 클라이언트와 모든 서버 간 연결마다 SSL/TLS 인증을 해야하는데, L7 로드밸런서가 있으면 브라우저는 L7 이랑만 SSL/TLS 연결을 맺고 그 뒤는 로드밸런서가 해독해서 요청을 보내주기때문에, SSL/TLS 인증 과정이 훨씬 수월해진다.






## TLS termination in AWS NLB

디폴트로는, TLS종료는 인그레스 컨트롤러에서 이루어진다. 근데 AWS의 NLB에서 TLS종료가 이루어지도록 할 수도 있다.


---

## 추가 공부

L7 로드밸런서(이하 로드밸런서)와 리버스 프록시의 차이?

리버스프록시의 주목적은 **보안과 중개**이다. 리버스프록시는 인터넷과 서버 사이에 위치해서, 클라이언트가 서버를 알지 못하게해서 보안을 높이고, 모든 서버로의 요청은 리버스프록시를 거치게 되기 때문에 기존 각 서버마다 SSL인증을 해야하던 것에서 리버스프록시만 SSL인증을 하면 되도록 바뀌게 되고, SSL인증을 간편하게 설정할 수 있다는 장점을 준다. 리버스프록시는 받은 암호화 요청을 복호화하고 서버로 전달한다(SSL 종료). 서버와 클라이언트 사이에 위치하기 때문에 캐싱, 로그 수집, 압축등이 가능하다. 리버스프록시는 요청헤더와 본문을 읽고 URL경로, 도메엔, 쿼리 매개변수, 헤더 등을 기반으로 백엔드 서버에 중개가 가능하다.

로드밸런서의 주요 역할은 **분산**이다. 
중개와 분산의 차이는 이렇다.
예를들어 리버스프록시에서 요청 URL경로를 통해 해당 트래픽이 백엔드의 재고관리 서버로 가야하는 것을 지정한다. 그런데 백엔드에는 재고관리 서버의 수평적 확장으로 여러개의 재고관리 서버가 있을 수 있다. 그럼 로드밸런서는 각 서버의 헬스 체크, 서버 메트릭 모니터링을 통한 로드밸런싱, 세션 지속성 관리, 가중치 부하 분산 등의 기능을 제공한다.

흐름은 리버스프록시 -> 로드밸런서 -> 서버 -> 로드밸런서 -> 리버스프록시 이렇게 된다.

보통 리버스프록시와 로드밸런서의 기능이 함께 제공되는 경우가 많다. 리버스프록시와 로드밸런서는 같은 어플리케이션 계층이기도 하고 기능도 유사해서 용어가 자주 혼용된다.

---

## 참고자료

https://kubernetes.github.io/ingress-nginx/deploy (nginx 인그레스 컨트롤러 설치 공식 문서)
https://www.a10networks.com/glossary/how-do-layer-4-and-layer-7-load-balancing-differ/ (L4 L7 로드밸런서)
https://medium.com/@sujitthombare01/haproxy-smart-way-for-load-balancing-in-kubernetes-c2337f61d90b (HAProxy)