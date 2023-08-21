# 쿠버네티스 ingress 오브젝트
# study kubernetes

---

## 개요

쿠버네티스에서는 인그레스와 인그레스컨트롤러라는 오브젝트를 제공한다. 이 아이들은 리버스프록시 및 로드밸런서 역할을 해주는 오브젝트이다. 이번 커플 채팅 서비스를 쿠버네티스에 배포할 예정이어서, 인그레스 컨트롤러에 대해 공부하고 적용하기로 했다.

---

## 사전 용어 정리

인그레스와 인그레스 컨트롤러에 대한 내용은 전체적으로 쿠버네티스 공식 도큐먼트를 참고했다.

수많은 훌륭한 개발자들의 블로그와 질문/답변들이 인터넷 상에 많지만, 아무래도 쿠버네티스를 개발한 사람들이 구글이라는 대기업 안에서 작성한 공식 도큐먼트는 못이기지 않을까 싶은 생각이었다.

도큐먼트를 읽다고 처음 보는 용어들이 많이 나와서 용어들에 대해 먼저 알아보겠다.

---

## 인그레스란?

인그레스는 클러스터 가장 앞단에서 내부 서비스를 외부로 HTTP, HTTPS 경로를 노출시키는 오브젝트이다.
NodePort나 service의 loadbalancer type도 외부에서 접근할 수 있게 해주지만, 인그레스는 http와 https만을 관리한다.

인그레스가 정상적으로 포트를 노출시키기 위해서는 인그레스 컨트롤러가 필요하다. 인그레스는 k8s에 정의된 오브젝트이지만, 인그레스 컨트롤러는 원하는 다양한 이미지들 중 선택해서 구성할 수 있다.

가장 많이 쓰이는 인그레스 컨트롤러는 nginx-ingress-controller가 있다.

인그레스 오브젝트 설정 yaml 파일을 보자.

    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
    name: minimal-ingress
    annotations:
        nginx.ingress.kubernetes.io/rewrite-target: /
    spec:
    ingressClassName: nginx-example
    rules:
    - http:
        paths:
        - path: /testpath
            pathType: Prefix
            backend:
            service:
                name: test
                port:
                number: 80

### 어노테이션

annotation은 오브젝트에 메타데이터를 첨부할 수 있게해주는 속성이다. 키와 값으로 이루어진 map배열인데, 키와 값으로는 **문자열**만 사용할 수 있다. 위 설정파일에 나온 어노테이션의 예시는 아주 흥미로운데,

    annotations:
        nginx.ingress.kubernetes.io/rewrite-target: /

이 어노테이션을 적용하면, 모든 path가 /로 rewrite되어서 라우팅된다.

FE/BE 간 리버스프록시를 사용할 때, 백엔드 경로에 /api를 붙여서 백엔드와 프론트엔드를 구분해서 라우팅을 하는 경우가 많은데, 이 rewrite을 이용하면 프론트엔드에서는 HOST+"/api/~" 경로로 요청을 보내지만, 실제 백엔드에서는 "/api/~" 경로가 아닌 "/~" 경로로 요청을 받을 수 있게 된다.

/api는 백엔드를 구분하기 위한 목적밖에 없기 때문에, /api를 보고 어디로 라우팅을 해야할지 판단한 후에는 /로 rewrite을 해줄 수 있게되는 것이다. 이 부분이 아주 흥미로웠다.

### 인그레스 클래스 네임

ingressClassName은 인그레스 클래스의 이름을 선언하는 부분이다. 

한 클러스터 안에 여러개의 인그레스 컨트롤러가 실행될 수 있고, 각 인그레스 컨트롤러마다 다른 인그레스를 처리할 수 있다. 

예를 들면, 컨트롤러1은 프론트엔드로의 라우팅을 수행하고, 컨트롤러2는 마이크로서비스에서 서비스간의 통신을 수행할 수 있다. 

사용자가 브라우저를 통해 도메인에 접속하면 컨트롤러1이 라우팅을 하고, 접속해서 여러 마이크로서비스의 기능들을 호출할 때는 컨트롤러2가 라우팅을 수행하게 되는 것이다.

인그레스 컨트롤러를 생성할 때, 인그레스 name으로 인그레스를 식별하는 게 아니기 때문에, class를 통해 인그레스를 식별할 수 있게 된다.

인그레스 클래스 이름을 지정하기 전에 인그레스 클래스 오브젝트가 배포되어있어야한다. 설정파일 예시는 아래와 같다.

    apiVersion: networking.k8s.io/v1
    kind: IngressClass
    metadata:
    name: nginx-example
    spec:
    controller: example.com/ingress-controller
    parameters:
        apiGroup: k8s.example.com
        kind: IngressParameters
        name: nginx-example

metadata.name은 이후에 인그레스 리소스 설정에서 ingressClassName의 값으로 사용된다.

spec.controller는 어떤 인그레스 컨트롤러를 쓸 것인지 명시해주는 부분이다.

parameters는 인그레스 컨트롤러에 전달될 설정들을 정의하는 부분이다.

### rules

rules는 어떤 기준으로 라우팅을 할 것이냐에 관한 부분인데, 크게 두 가지로 나뉜다. 서브도메인으로 나눌 것인지, URL 경로로 나눌 것인지. 

FE와 BE를 예시로 서브도메인을 적용하면, www.example.com HOST에서는 프론트엔드 서비스로, api.example.com HOST에서는 백엔드 서비스로 라우팅을 할 수 있다.

설정파일은 아래와 같이 구성된다.

    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
    name: my-ingress
    spec:
    rules:
        - host: www.example.com
        http:
            paths:
            - path: /
                pathType: Prefix
                backend:
                service:
                    name: frontend-service
                    port:
                    number: 3000
        - host: api.example.com
        http:
            paths:
            - path: /
                pathType: Prefix
                backend:
                service:
                    name: backend-service
                    port:
                    number: 8080

서브도메인이 아닌 URL경로로 라우팅을 하는 건 위에서 예를 들었듯이, /api 경로로 올 때는 백엔드서비스로, 그 외에는 프론트엔드서비스로 라우팅하는 방법이 가장 일반적이다. 예시 설정파일은 아래와 같다.

    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
    name: my-ingress
    spec:
    rules:
        - host: www.example.com
        http:
            paths:
            - path: /
                pathType: Prefix
                backend:
                service:
                    name: frontend-service
                    port:
                    number: 3000
            - path: /api
                pathType: Prefix
                backend:
                service:
                    name: backend-service
                    port:
                    number: 8080

이제 rules의 속성들을 살펴보자.

#### host:

host는 말그대로 host이다. 도메인 대신 ip주소로도 가능하다. URL경로로 라우팅하는 경우 + EIP(Elastic IP)를 사용하는 경우라면 IP로 호스트를 설정해도 큰 문제는 없을 것 같다.

#### http:

http는 http 포트(80)로 오는 트래픽에 대한 rule을 설정하겠다는 것이다. 인그레스는 http, https만을 다룬다. 예시에는 http만 나와있지만 SSL/TLS 적용이 된다면 https 속성도 사용해서 규칙을 정의할 수 있다.

#### path:

path는 해당 경로와 일치하는 지에 따라서 라우팅을 하는 기준이다. 만약 서브도메인을 기준으로 나눈다면, 서브도메인의 모든 경로로 오는 트래픽을 라우팅해야하기 때문에 path를 /로 설정해야할 것이다. path는 여러개를 설정할 수 있고, 각 path는 ,로 구분한다.

#### pathType:

pathType은 Prefix, Exact, Mixed 가 있다. 

1. Prefix는 접두사를 기준으로 판단한다는 것이다. 

path가 /, pathType이 Prefix로 되어있다면, 모든 경로는 기본적으로 가장 앞에 / 가 붙어있기 때문에 모든 경로를 일치하는 경로로 보고 라우팅을 한다. 그래서 /가 모든 트래픽을 라우팅할 수 있는 것이다. 

/api로 path를 설정했을 때도 접두사를 보는 것이기 때문에 /api/usr/id 경로로 요청이 들어와도 가장 앞에 /api가 있으므로 일치하는 경로로 판단하고 라우팅을 할 수 있는 것이다. 그리고 겹치는 경로가 있으면 더 긴 경로로 라우팅한다.

/api 경로에는 /도 포함되어있지만 /로 라우팅 되지않고 /api로 라우팅되는 이유가 바로 거기에 있다.

2. Exact는 정확히 그 경로만 보는 것이다.

path가 /로 정의되어있어도, pathType이 Exact면 / 경로로 오는 요청을 제외하고는 모든 요청이 거부된다. 참고로 경로의 가장 마지막 /는 무시될 수 있다. 

예를 들어 www.google.com을 host로 지정하고, path는 /, pathType은 Exact로 설정했다고 가정하자. 

Exact에 따르면 www.google.com/ 로 들어온 경로만 라우팅해야할 것 같지만, 경로상 마지막 /은 무시될 수 있기 때문에, www.google.com/ 과, www.google.com은 모두 라우팅될 수 있다. 

물론 www.google.com/123은 라우팅 될 수 없다. pathType이 Exact이기 때문에! 이 마지막 / 무시는 Prefix에서도 동일하게 적용된다.

#### backend:

여기서 사용되는 백엔드 속성은 실제 백엔드 서비스가 아니라, 뒷단을 의미한다. 특정 경로로 접근이 왔을 때 rule에서 설정한 path와 일치하면 뒷단에서는 무엇을 할 것인지를 설정한다.

#### service:

서비스에 라우팅한다고 명시한다.

#### name:

라우팅할 서비스의 이름이다. DNS서버를 통해 서비스 이름만으로 해당 서비스에 접근할 수 있다.

#### port:

해당 서비스에 몇 번 포트로 라우팅할지를 정한다.

#### number:

포트의 번호를 의미한다.

---

## 인그레스 유형 다섯가지

인그레스 유형이 꼭 다섯가지로 제한되는 것은 아니겠지만, K8S Official Document에서 소개된 유형들을 정리해보겠다.

### 1. 단일 서비스로 지원되는 인그레스

만약 인그레스 리소스에 rule을 정의하지 않는다면, 적어도 defaultBackend는 정의를 해야한다.

defaultBackend는 해당하는 rule이 없을 때, 트래픽을 defaultBackend로 라우팅하게 된다. 근데 rule자체가 없다면, 이 인그레스는 요청받는 모든 트래픽을 defaultBackend service(단일 서비스)로 라우팅하게된다. 예시 코드는 아래와 같다.

    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
    name: test-ingress
    spec:
    defaultBackend:
        service:
        name: test
        port:
            number: 80
        
### 2. 간단한 팬아웃(fanout)

위에서 보았던 URL경로 기반의 라우팅을 설명한다.

### 3. 이름 기반의 가상 호스팅

위에서 보았던 서브도메인 기반의 라우팅을 설명한다.

### 4. TLS

인그레스 컨트롤러는 TLS종료를 수행할 수 있다. 리버스프록시의 역할을 하는 것이다. 방법은 시크릿에 인증서와 개인키를 정의하고, ingress 설정파일에서 시크릿을 불러오면 된다. 

TLS의 종료지점이 되기 때문에, 인그레스 컨트롤러 이후의 트래픽은 일반 TEXT형태로 복호화된다. 시크릿 설정 예시 코드는 아래와 같다.

    apiVersion: v1
    kind: Secret
    metadata:
        name: testsecret-tls
        namespace: default
    data:
        tls.crt: <base64 encoded cert>
        tls.key: <base64 encoded key>
    type: kubernetes.io/tls

이 TLS시크릿을 이용해 TLS종료 지점인 ingress 설정파일을 구성하는 코드 예시는 아래와 같다.

    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
    name: my-ingress
    spec:
    rules:
        - host: www.example.com
        http:
            paths:
            - path: /
                pathType: Prefix
                backend:
                service:
                    name: frontend-service
                    port:
                    number: 80
        - host: api.example.com
        http:
            paths:
            - path: /
                pathType: Prefix
                backend:
                service:
                    name: backend-service
                    port:
                    number: 80
    tls:
        - hosts:
            - www.example.com
            - api.example.com

### 5. 로드 밸런싱

인그레스 컨트롤러는 로드밸런싱도 수행한다. 서비스로 트래픽이 들어왔을 때, 서비스에 연결된 pod가 여러 개라면, 인그레스 컨트롤러는 로드밸런싱 정책 설정에 따라 부트스트랩하게 된다. 이 때 고급 로드밸런싱(지속적인 세션, 동적 가중치 등)은 수행하지 못하고 기본적인 로드밸런싱 정책을 적용한다고 한다. 

도큐먼트에서 기본적인 로드밸런싱 정책으로는 가중치를 예시르 들었다. 

가중치는 처음 서버를 구성할 때 지정한 가중치가 쭉 이어지는 방식이다. 노드A는 리소스가 크고 B는 작다고 했을 떄, A에 더 큰 가중치를 부여해서 A로 더 많은 트래픽이 전달될 수 있게한다. 인그레스 컨트롤러는 이런 기본적인 로드밸런싱 정책을 이용해서 라우팅을 한다.

도큐먼트에서 소개한 고급 로드밸런싱 중 지속적인 세션은 한 번 연결된 클라이언트와 서비스는 쭉 연결이 이어지게 하는 것이다.

만약 일반적으로 라운드로빈 로드밸런싱을 한다면 클라이언트는 매번 서비스에 접근할 때마다 다른 pod로 라우팅될텐데, 지속적인 세션은 한 번 연결된 pod에만 접근되게 함으로써 캐싱을 이용해 빠른 작업이 이루어질 수 있다. 

또 동적 가중치는 리소스의 상태, 응답시간에 따라 가중치를 다르게 부여하는 것이다. 응답시간이 긴 pod가 있으면 가중치를 낮게 주는 방식으로 각 노드와 pod의 동적인 상태에 따라 효율적으로 트래픽을 라우팅할 수 있게 된다. 이러한 기능을 인그레스 컨트롤러는 제공하지 않는다.

---

## 인그레스 편집

인그레스 rule을 수정해야하는 상황이 생길 수 있다. 이럴 때 인그레스를 지웠다가 다시 배포하면 그동안 서비스가 중단되기 때문에, 그 방법보다는

    kubectl edit ingress <ingress name>

이 커맨드를 사용해서 인그레스를 수정하도록 하자.

---

다음 글에서는 ingress controller에 대해 글을 써보도록 하겠다.

## 참고자료

https://kubernetes.io/ko/docs/concepts/services-networking/ingress/