# KIND에 Istio Ingress Gateway 배포하기
# study kubernetes Istio

---

## 개요

지난 글에서는 여러 로컬 쿠버네티스 개발환경 도구들과 그 중 Kind를 선택한 이유, 기본적인 Kind 사용법 및 배포 테스트에 대해 소개했다.

팀 프로젝트에서 서비스메시를 활용할 예정이기에 로컬에서부터 Istio를 클러스터에 배포하고, 개발하며 테스팅이 함께 되면 좋겠다고 판단했다.

이번 글에서는 Istio를 로컬 Kind 클러스터에 배포하고, Istio Ingress Gateway와 간단한 백엔드, 프론트엔드 서비스를 배포해서 테스트를 진행해보려고 한다.

---

## Istio 다운로드

Kind같은 경량화 로컬 쿠버네티스 클러스터 구현 도구가 과연 Istio를 지원할까?

놀랍게도 Kind 클러스터에 서비스메시인 istio도 배포가 가능하다. 심지어 Istio 공식 문서에서 Kind 관련 내용을 다룬다.

```
curl -L https://istio.io/downloadIstio | sh -
```

우선 Istio를 다운로드한다. 위 커맨드는 Linux 또는 MacOS에서 실행 가능한 커맨드이고, 알아서 가장 최신 버전의 Istio를 설치해준다.

Linux/ MacOS가 아니라면 [Istio release](https://github.com/istio/istio/releases/tag/1.18.2) 해당 링크에서 OS 및 아키텍처 별로 Istio 다운로드가 가능하다.

이후에 커맨드를 실행한 위치에 **istio-1.18.2**라는 이름의 디렉토리가 생성되는데, 이 디렉토리로 이동해준 뒤

```
export PATH=$PWD/bin:$PATH
```

위 커맨드로 환경 변수를 설정해준다.

---

## Istio 설치

Istio를 다운로드 한다고 클러스터에 설치되는 것이 아니다. 

```
istioctl install -y
```

istioctl 이라는 CLI를 사용해서 클러스터에 istio를 배포해야한다. istio에는 demo, minimum 등 여러 프로파일 설정이 있는데, 아무 옵션도 지정하지 않으면 default 프로파일로 설정된다.

default profile은 기본적으로 ingress-gateway와 istiod가 설치된다.

---

## 추가 설정

Istio를 설치하기만 한다고 Istio의 모든 기능들을 사용할 수 있는 것은 아니다.

### 사이드카 프록시 주입 설정

Istio의 주요 기능들은 클러스터에 배포되는 모든 Pod들의 내부에 주입되는 사이드카 프록시를 통해 이루어진다. 이 설정을 해주어야한다.

```
kubectl label namespace default istio-injection=enabled
```

사이드카 프록시 주입은 클러스터의 namespace를 기준으로 주입된다. 위와 같이 설정해두면 default ns에 파드가 배포될 때마다 Istio에 의해 사이드카 프록시가 주입되게 된다.

### 추가 대시보드 UI 배포

Istio는 다양한 외부 대시보드 UI 툴들과의 연동을 지원한다. 사이드카 프록시를 통해 수집되는 모든 로그, 메트릭, 모니터링, 트레이싱 데이터들을 이 도구들로 쉽게 확인할 수 있다.

```
kiali: kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.18/samples/addons/kiali.yaml
grafana:  kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.18/samples/addons/grafana.yaml
prometheus: kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.18/samples/addons/prometheus.yaml
jaeger:  kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.18/samples/addons/jaeger.yaml
```

내가 적용한 네 가지의 툴들을 적용하는 커맨드이다. 이 도구들은 모두 istio-system NS에 배포되고, 

    kubectl get pods -n istio-system

를 실행하면 해당 툴들이 잘 배포되었는지 확인할 수 있다.

각 도구들은

```
istioctl dashboard TOOL_NAME
```

을 통해 DASHBOARD UI에 접근할 수 있다.
 
![img](http://www.choigonyok.com/api/assets/65-1.png)

<br>

![img](http://www.choigonyok.com/api/assets/65-2.png)

<br>

![img](http://www.choigonyok.com/api/assets/65-3.png)

---

## Ingress-Gateway 설정

### 포트 변경

Ingress Gateway Service는 무작위 노드포트가 생성된다. 하지만 Ingress-Gateway를 통해 Kind 클러스터에 접근하려면 클러스터 생성시 설정한 extraPort와 노드포트가 일치해야한다. 그래서 무작위로 생성되어있는 Ingress-Gateway service의 노드포트를 변경해줘야한다.

```
kubectl edit service istio-ingressgateway -n istio-system
```

위 커맨드로 이름이 istio-ingressgateway인 서비스의 매니페스트를 수정할 수 있다.

테스트를 위한 작업이기에 다른 것들은 두고,

```
  - name: http2
    nodePort: 30080
    port: 80
    protocol: TCP
    targetPort: 3000
```

nodePort를 클러스터 생성시 설정한 extraPort로 변경시킨다. port는 nodePort 경로로 들어온 트래픽이 Gateway 오브젝트로 전달될 port를 의미하고, targetPort는 @

이렇게 변경해주고 :wq 로 저장 후 Vim 에디터를 빠져나와주면 된다. 

```
kubectl get svc -n istio-system
```

을 실행했을 때, 포트가 잘 바뀌어져있는 것을 확인할 수 있다.

### Gateway 오브젝트 생성

ingress-gateway 서비스가 트래픽을 전달할 gateway를 생성한다.

```
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: gateway
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
```

spec.selector의 istio: ingressgateway label을 가지고 있는 ingress-gateway 서비스는 이 Gateway로 트래픽을 전달한다.

특히 그 중 80번 포트로 오는 HTTP 요청을 처리하고, hosts에는 적절한 도메인을 설정해주는 것이 좋지만, 개발환경 특성상 따로 도메인이 있지 않기 때문에 *로 모든 호스트의 접근을 가능하게 함을 명시해준다.

### VirtualService 오브젝트 생성

VirtualService는 Ingress와 유사하다. 어떤 규칙으로 라우팅을 할 것인지를 지정한다.

VirtualService에서는 spec.gateways 속성을 통해서 어떤 gateway가 이 규칙들을 적용할지를 지정할 수 있다.

```
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: ingress-vs
  namespace: default
spec:
  hosts:
  - "*"
  gateways:
  - gateway
  http:
  - match:
    - uri:
        prefix: /
    route:
    - destination:
        port:
          number: 3000
        host: frontend-service
```

이렇게 VirtualService를 생성한다면, 모든 경로(/)에 대해 frontend-service의 3000번 포트로 트래픽을 전달하게 될 것이다.



Kind는 이전 "**KIND로 쿠버네티스 개발환경 구축하기**" 글에서 언급한 것처럼, kudeadm으로 구성된 쿠버네티스이다. 이 말은 곧, 로드밸런서 타입 서비스를 지원하지 않는다는 것이다. 이에 대해서는 "**Kubeadm으로 Ubuntu에 쿠버네티스 클러스터 스크래치 구성하기**" 글에 정리해두었다.

    kubectl get svc --all-namespaces

위 커맨드로 실행중인 ingress-gateway 서비스를 확인해보니 LoadBalancer 타입이었다. 이전에 블로그 서비스를 kubeadm으로 구성한 쿠버네티스로 이전할 때, LoadBalancer 타입 서비스가 지원되지 않아서 HAProxy 로드밸런서 / 리버스프록시 구현 도구를 사용했다. 그러나 이번에는 Istio까지 도입하게 되면서 Istio의 사이드카 프록시의 기능을 정상적으로 사용하기 위해, HAProxy 등의 다른 도구가 아닌 Istio-Ingress-Gateway를 꼭 도입해야만 했다.



## Istio


kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.7/config/manifests/metallb-native.yaml
istio-ingress-gateway를 사용하기 위한 네트워크 구현 툴 metallb 

kubectl wait --namespace metallb-system \
                --for=condition=ready pod \
                --selector=app=metallb \
                --timeout=90s
metallb pod 생성시까지 기다리기

docker network inspect -f '{{.IPAM.Config}}' kind
metallb가 관리할 네트워크 범위를 알려줘야함. 이 커맨드는 kind 클러스터의 네트워크 범위
아웃풋이 [{172.24.0.0/16  172.24.0.1 map[]} {fc00:f853:ccd:e793::/64  fc00:f853:ccd:e793::1 map[]}] 이런식으로 나오면

```
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: example
  namespace: metallb-system
spec:
  addresses:
  - 172.24.255.200-172.24.255.250
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: empty
  namespace: metallb-system
```
yml 파일 하나 만들고 

kubectl apply -f YAML_FILE_NAME
적용

---

## 정리

이렇게 설정하고

```
http://localhost:EXTRA_PORT
```

로 접근하면 Istio의 Ingress Gateway를 통해 정상적으로 서비스에 접근이 되는 것을 확인할 수 있다.

---

## 참고

[Istio Official Documentation : Kind](https://istio.io/latest/docs/setup/platform-setup/kind/)

[LoadBalancer Services using Kubernetes in Docker (kind)](https://medium.com/groupon-eng/loadbalancer-services-using-kubernetes-in-docker-kind-694b4207575d)

[Kind Official Documentation : LoadBalancer](https://kind.sigs.k8s.io/docs/user/loadbalancer/)