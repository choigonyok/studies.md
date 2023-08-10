# SSL/TLS TERMINATION IN HAPROXY

## 개요

HAPROXY에서 SSL/TLS종료 구현하기

HAProxy는 리버스프록시, 로드밸런서 오픈소스 툴이다. 쿠버네티스엔 IngressController라는 오브젝트가 있는데 왜 HAProxy를 쓰게 됐을까?

### Kubeadm을 채택한 계기

나의 경우에는 쿠버네티스 클러스터를 Kubeadm으로 구성했다. Kubeadm은 베어메탈 클러스터 프로비저닝 툴이다. 베어메탈이란 EKS 등의 클러스터 프로비저닝 자동화 툴의 도움을 받지 않고 직접 구성하는 방식을 말한다. 

직접 클러스터를 구성할 수 있는만큼 필요에 맞는 커스터마이징이 가능한 대신, 복잡성과 불편성을 감수해야한다. EKS 등의 서비스는 특정 클라우드 플랫폼에 종속되어있지만 Kubeadm으로 클러스터를 구성하면 플랫폼의 제약을 받지 않을 수 있어서 한 번 익혀두면 응용할 곳이 많고, 클러스터 구성 뒷단에서 무슨 일들이 일어나는 건지도 알고싶어서 클러스터 프로비저닝 툴로 Kubeadm을 선택하게 되었다.

Kubeadm이 커스터마이징을 강조하는만큼, 기능들 중 일부가 누락되어있다. 네트워크 인터페이스나 로드밸런서 등이 그러하다. 이미 많은 네트워크 인터페이스나 로드밸런서 오픈소스들이 많이 공개되어있기 때문에, 개발자는 어플리케이션에 가장 적합한 오픈소스를 가져와서 Kubeadm과 함께 사용할 수 있는 것이다.

### IngressController는 로드밸런서이다?

처음 쿠버네티스를 접했을 때, IngressController는 리버스프록시 겸 로드밸런서인 줄 알았다. 사실 쿠버네티스의 IngressController는 리버스프록시는 지원하지만 로드밸런싱은 지원하지 않는다. 

IngressController를 리버스프록시 겸 로드밸런서로 쓰려면 IngressController Service를 LoadBalancer 타입으로 설정해야한다. 사실 IngressController보다는 로드밸런서 타입 서비스가 실질적인 로드밸런싱을 한다. 만약 IngressController의 서비스 타입을 노드포트로 설정한다면 이 IngressController는 리버스프록시의 기능만 수행할 것이다.

### MetalLB vs HAProxy

앞서 말했던 Kubeadm이 커스터마이징을 이유로 로드밸런서 타입 서비스를 지원하지 않기 때문에, 클러스터 내부 로드밸런서를 구현하려면 오픈소스 툴을 가져다 사용해야한다. 내가 선택한 툴이 바로 HAProxy이다. 

널리 사용되는 로드밸런서/리버스프록시 구현 오픈소스 중 MetalLB, HAProxy가 후보에 들었는데, MetalLB는 클라우드와 온프레미스 환경에서 모두 사용되고, HAProxy는 클라우드에 좀 더 특화되어있다고 하여 HAProxy를 선택했다. 

추가적으로 몇몇 유명한 쿠버네티스 엔지니어들이 작성한 2023 쿠버네티스 표준 아키텍처에서도 API서버 로드밸런서로 HAProxy를 뽑았다. 이러한 이유로 HAProxy를 사용하기로 했다. 

### SSL/TLS Termination

HAProxy를 리버스프록시/로드밸런서로 쓰는 쉬운 방법은 HAProxy의 Helm 차트를 가져다가 클러스터에 배포하는 것이다. 이러면 HAProxy Pod는 리버스프록시 역할을 하며 정의된 Ingress 오브젝트의 rule을 따라서 서비스로 라우팅을 하고 로드밸런싱을 수행하게된다.

그러나 운영환경에서 이대로만 사용하기는 무리가 있다. 기본 HAProxy 차트에는 SSL/TLS 인증서가 담겨있지 않기 때문이다. 그럼 클라이언트와 HTTP 통신을 해야하고, 이 과정에서 중간자공격 등의 보안 위협이 발생할 수 있다. 때문에 직접 SSL/TLS 인증서를 받고, HAProxy Pod가 인증서를 보고 SSL/TLS 종료를 수행할 수 있게 설정해줘야한다. 이 글에서는 이러한 내용을 다룬다.

## SSL/TLS 인증 및 적용

크게 흐름을 먼저 보면, 인증서와 키가 담긴 시크릿 오브젝트를 생성하고, HAProxy가 사용하는 ConfigMap에 시크릿을 넣어주면 HAProxy가 SSL/TLS 종료를 수행할 수 있게된다.

진행하려면 Base64로 인코딩된 SSL/TLS 인증서와 키가 있어야한다.

### 인증서 키 발급



### TLS 시크릿 생성

TLS 인증서와 개인키를 위한 시크릿을 생성한다. tls.crt와 tls.key의 value로는 실제 Base64로 인코딩된 인증서와 키의 전문이 들어가야한다. 예시 코드는 아래와 같다.

    apiVersion: v1
    kind: Secret
    metadata:
        name: example-cert
    type: kubernetes.io/tls
    data:
        tls.crt: BASE64_ENCODED_CERT_CONTENT
        tls.key: BASE64_ENCODED_KEY_CONTENT

또는 HAProxy의 공식문서대로 Yaml파일을 만들지 않고 kubectl CLI로도 시크릿을 생성할 수 있다.

    kubectl create secret tls example-cert   --cert="example.crt"   --key="example.key"

### HAPROXY CONFIGMAP 수정

이제 HAProxy에도 ConfigMap 설정에 시크릿을 넣어줘야하는데, Helm으로 설치된 HAProxy의 컨피그맵 이름을 확인하려면 아래 명령어를 입력하면 된다.

    kubectl get configmaps --namespace haproxy-controller

그 다음 얻은 이름을 가지고, 컨피그맵을 수정한다.

    kubectl edit configmap --namespace haproxy-controller CONFIGMAPNAME

data 속성에 

    data:
        ssl-certificate: "default/example-cert"

이 내용을 추가해준다. default는 네임스페이스고 example-cert는 시크릿 이름이다. 상황에 맞게 설정하면 된다. edit 후 나오면 HAProxy가 알아서 configMap애 대한 정보를 업데이트한다.

이렇게 되면 모든 백엔드서비스(HAProxy가 라우팅하는 서비스)들에 대해 SSL/TLS 인증이 이루어진다. 

여러개의 HAProxy를 사용하고, HAProxy 별로 다른 Ingress를 사용하며, 각 Ingress마다 다른 인증서를 사용하고싶다면 아래 참고자료의 링크에서 확인할 수 있다.

---

## 보안적 위험성

이런 식으로 시크릿와 컨피그맵을 이용해 TLS 인증서와 개인키 말고도 여러 키들을 관리하는 건 단점이 있다. 키와 관련한 정보들은 공유저장소에 올리지 못한다는 것이다. 올리게되면 보안에 큰 위협이 될 수 있다. 올리고 싶지 않았더라도 실수로 올라가게 될 수도 있다. 이런 이유로 외부 버킷을 이용해 키들을 관리하거나 외부 키 관리 오픈 소스를 사용한다고 하는데, 키 관리 툴 중 인기가 많은 Vault에 대해서도 나중에 공부해봐야겠다.

---

## 참고자료

[HAProxy 공식문서](https://www.haproxy.com/documentation/kubernetes/latest/usage/terminate-ssl/)