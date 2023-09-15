# Istio OAuth2-Proxy & Ingress Gateway로 구글 소셜로그인 구현하기
# istio sec

---

## 개요

팀 프로젝트를 구성하면서 OAuth, JWT부터 AutorizationPolicy, RequestAuthentication 등 여러 인증/인가 기능에 대해 알게되었다. Istio에서도 OAuth를 통한 인증/인가 및 소셜로그인 기능을 구현할 수 있다. 

이전에는 OAuth는 소셜로그인 기능을 통틀어서 부르는 단어인 줄 알았고, 모든 건 백엔드 소스코드로 구현하는 것인 줄 알았다. OAuth & JWT & OIDC에 대한 올바른 정리와 Istio에서 OAuth2-Proxy & Ingress Gateway를 사용해 소셜로그인을 구현하는 방법을 알아보려고한다.

---

## 소셜 로그인 과정

Istio, Oauth2-proxy, Keycloak가 실질적으로 어떻게 작동하는지에 대해 알아보겠다. 내가 정리한 작동 순서는 다음과 같다. 경우의 수가 발생할 수 있는 이벤트에 대해서는 분기로 처리했다.

```
1. 사용자(이하 브라우저)가 게이트웨이로 접근한다.
2. 게이트웨이는 X-forwarded-access-token 헤더를 확인한다.
3-1. x-forwared-access-token 헤더가 있으면 요청된 경로의 서비스로 포워딩한다.

3-2. 헤더가 없으면 oauth2-proxy로 요청을 포워딩한다.
4. 프록시는 브라우저의 쿠키에 SessionID가 있는지 확인한다.
5-1. SessionID가 있으면 레디스 key에 일치하는 SessionID가 있는지 확인한다.
6. 일치하는 SessionID가 있으면 해당 SessionID의 value인 AccessToken의 role를 확인한다.
7-1. 요청의 경로(향하는 서비스)와 role의 권한이 일치한다면,
8. 요청 헤더에 x-forwarded-access-token 헤더를 추가하고 값으로 accessToken을 넣어 200상태코드로 게이트웨이에 응답한다. (3-2로 이동)

5-2. SessionID가 없으면 요청을 Keycloak으로 리다이렉트시킨다.
6. Keycloak은 idP(ex: Google)에 요청을 포워딩한다.
7. Google은 구글 로그인 사이트로 리다이렉트시킨다.
8. 구글은 사용자의 ID/PW가 확인되면 accessCode를 경로에 담아서 게이트웨이로 리다이렉트한다.
9. 게이트웨이는 요청을 oauth2-proxy로 포워딩한다.
10. oauth2-proxy는 accessCode를 쿼리파라미터에 담아 Keycloak에 토큰을 요청한다.
11. KeyCloak은 토큰을 응답한다.
12. oauth2-proxy는 토큰을 자신의 세션에 저장하고 브라우저의 쿠키에 세션ID 추가해서 게이트웨이로 리다이렉트시킨다.
13. 게이트웨이는 요청을 oauth2-proxy로 포워딩한다. (5-1로 이동)

7-2. 요청의 경로(향하는 서비스)와 role의 권한이 일치하지 않는다면,
8. 게이트웨이에 상태코드로 권한 없음을 응답한다.
9. 게이트웨이는 브라우저의 접근을 거부한다.
```

---

## 구현

kind 클러스터 config파일과 함께 create

istioctl install -f config.yml로 meshConfig 설정하면서 생성
생성후 수정도 같은 명령어로 가능


### gateway 생성

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

### 게이트웨이가 라우팅할 규칙
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: vs
spec:
  hosts:
  - "*"
  gateways:
  - gateway
  http:
  - match:
    - uri:
        prefix: /oauth2
    route:
    - destination:
        host: oauth2-proxy.default.svc.cluster.local
        port:
          number: 80          
  - match:
    - uri:
        prefix: /app
    route:
    - destination:
        host: echo.default.svc.cluster.local
        port:
          number: 80        
  - match:
    - uri:
        prefix: /
    route:
    - destination:
        host: keycloak.default.svc.cluster.local
        port:
          number: 8080  

### oauth2-proxy의 포워딩을 위한 규칙
kind: AuthorizationPolicy
apiVersion: security.istio.io/v1beta1
metadata:
  name: ext-authz-oauth2-proxy
  namespace: istio-system
spec:
  selector:
    matchLabels:
      istio: ingressgateway
  action: CUSTOM
  provider:
    name: oauth2-proxy
  rules:
    - to:
        - operation:
            hosts: ["localhost"]
            notPaths: ["/realm/*"]
### oauth2-proxy가 배포될 때 적용할 config 파일
  config:
    clientID: "1003120694761-605icm9ht94b2advttpgobld809hu8bm.apps.googleusercontent.com"
    clientSecret: "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"
    cookieSecret: "TnpGZkFIbER6Rm5RR1NOazkwMnN5cFRuckt5czFYVUw="
    cookieName: "my-cookie"
    configFile: |-
      provider = "oidc"
      oidc_issuer_url="https://localhost/auth/realms/my-realm"
      profile_url="https://localhost/auth/realms/my-realm/protocol/openid-connect/userinfo"
      validate_url="https://localhost/auth/realms/my-realm/protocol/openid-connect/userinfo"
      scope="my-scope openid email profile"
      pass_host_header = true
      reverse_proxy = true
      auth_logging = true
      cookie_httponly = true
      cookie_refresh = "4m"
      cookie_secure = true
      email_domains = "*"
      pass_access_token = true
      pass_authorization_header = true
      request_logging = true
      session_store_type = "cookie"
      set_authorization_header = true
      set_xauthrequest = true
      silence_ping_logging = true
      skip_provider_button = true
      skip_auth_strip_headers = false
      skip_jwt_bearer_tokens = true
      ssl_insecure_skip_verify = true
      standard_logging = true
      upstreams = [ "static://200" ]
      whitelist_domains = ["localhost"]

### 적용해서 oauth2-proxy 배포
helm repo add oauth2-proxy https://oauth2-proxy.github.io/manifests
helm install oauth2-proxy oauth2-proxy/oauth2-proxy -f updated_values.yaml

### Keycloak K8S 배포

helm repo add bitnami https://charts.bitnami.com/bitnami
helm install bitnami/keycloak --generate-name --set auth.createAdminUser=true,auth.adminUser=ID,auth.adminPassword=PASSWORD
: PASSWORD는 K8S 시크릿에 알아서 저장됨

kubectl edit service keycloak
for nodeport

apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
   meshConfig:
     extensionProviders:
       - name: "oauth2-proxy"
         envoyExtAuthzHttp:
           service: "oauth2-proxy.default.svc.cluster.local"
           port: "80"
           includeHeadersInCheck: ["authorization", "cookie"]
           headersToUpstreamOnAllow: ["authorization", "path", "x-auth-request-user", "x-auth-request-email", "x-auth-request-access-token"]
           headersToDownstreamOnDeny: ["content-type", "set-cookie"]


Istio AuthorizationPolicy의 Custom Action을 통해 간단하게 외부 인가 서비스(구글 등)의 대리자가 될 수 있다.

이 custom action을 통해 ingress gateway와 oauth2 proxy를 사용해서 인가 로직을 구현할 수 있다.

1. 사용자 접근
2. AuthorizationPolicy의 Custom action 설정대로, 요청이 서비스로 안가고 인터셉트돼서 외부 인가 서비스로 전달
3. 외부 인가 서비스(구글) 등이 허가할지 안할지를 결정(구글 로그인이 되는지 안되는지)
4. 허가가 되면(로그인이 되면) AuthorizationPolicy의 custom말고 allow/deny(ip,port,method,header 등 네트워크 기반의 접근제어) action을 통해 알맞은 곳으로 트래픽이 라우팅
5. 허가가 안되면(로그인 안되면) 사용자의 요청 기각

> 생각보다는 간단하다


외부 Oauth2/OIDC 공급자에게 사용자의 요청을 전달하기 위해서 OAuth2-proxy를 사용할 수 있다.
외부 서비스의 응답으로 토큰을 받고, 토큰이 유효한지 확인하고, 사용자의 정보를 읽고, 원래 사용자가 가고자 했던 서비스로 트래픽을 포워딩한다.
OAuth2-proxy는 helm chart를 공식적으로 지원해서, values.yaml에서 외부 공급자(구글, 네이버, 카카오 등)를 정의하고, helm install cli로 클러스터에 설치할 수 있다.


# Oauth 사전 설정

```
config:
  # OAuth client ID
  clientID: "<CLIENT_ID>"
  # OAuth client secret
  clientSecret: "<CLIENT_SECRET>"
  # Create a new secret with the following command
  # openssl rand -base64 32 | head -c 32 | base64
  # Use an existing secret for OAuth2 credentials (see secret.yaml for required fields)
  # Example:
  # existingSecret: secret
  cookieSecret: "<COOKIE_SECRET>"
  # The name of the cookie that oauth2-proxy will create
  # If left empty, it will default to the release name
  cookieName: ""
  google: {}
    # adminEmail: xxxx
    # serviceAccountJson: xxxx
    # Alternatively, use an existing secret (see google-secret.yaml for required fields)
    # Example:
    # existingSecret: google-secret
  # Default configuration, to be overridden
  configFile: |-
    email_domains = [ "*" ]
    upstreams = [ "file:///dev/null" ]
  # Custom configuration file: oauth2_proxy.cfg
  # configFile: |-
  #   pass_basic_auth = false
  #   pass_access_token = true
  # Use an existing config map (see configmap.yaml for required fields)
  # Example:
  # existingConfig: config

image:
  repository: "quay.io/oauth2-proxy/oauth2-proxy"
  tag: "v7.1.3"
  pullPolicy: "IfNotPresent"

# Optionally specify an array of imagePullSecrets.
# Secrets must be manually created in the namespace.
# ref: https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod
# imagePullSecrets:
  # - name: myRegistryKeySecretName

extraArgs: 
  provider: oidc
  cookie-secure: false
  cookie-domain: "<COOKIE_DOMAIN>"
  cookie-samesite: lax
  cookie-refresh: 1h
  cookie-expire: 4h
  set-xauthrequest: true
  reverse-proxy: true
  pass-access-token: true # X-Auth-Request-Access-Token, must first enable --set-xauthrequest
  set-authorization-header: true # Authorization: Bearer <JWT>
  pass-authorization-header: true # pass OIDC IDToken to upstream via Authorization Bearer header
  pass-user-headers: true
  pass-host-header: true # pass the request Host Header to upstream
  pass-access-token: true
  scope: "openid email"
  upstream: static://200
  skip-provider-button: true
  whitelist-domain: <WHITELIST_DOMAIN>
  login-url: <LOGIN_URL>
  oidc-jwks-url: <JWKS_URL> # this is accessed by proxy in-mesh - http
  redeem-url: <REDEEM_URL> # This is accessed by proxy in-mesh - http
  skip-oidc-discovery: true
  redirect-url: <REDIRECT_URL>
  oidc-issuer-url: <ISSUER_URL>
  standard-logging: true
  auth-logging: true
  request-logging: true
extraEnv: []
```

위 코드 예제에서 < >로 묶인 부분들 수정하면 된다.

# 그 다음에 OAuth2-proxy helm chart 설치
helm repo add oauth2-proxy https://oauth2-proxy.github.io/manifests
helm install oauth2-proxy oauth2-proxy/oauth2-proxy -f updated_values.yaml

# istio는 이 OAuth2-proxy가 요청을 가로채서 외부서비스에 전달하고, 토큰 확인 및 검증하고, 다시 서비스로 트래픽을 보낼수 있게 하기위해서 istio의 profile을 수정해야한다. 특히 meshConfig 부분을 수정해야한다.


meshConfig:
  extensionProviders:
    - name: "oauth2-proxy"
      envoyExtAuthzHttp:
        service: "oauth2-proxy.default.svc.cluster.local"
        port: "80"
        includeHeadersInCheck: ["authorization", "cookie"]
        headersToUpstreamOnAllow: ["authorization", "path", "x-auth-request-user", "x-auth-request-email", "x-auth-request-access-token"]
        headersToDownstreamOnDekny: ["content-type", "set-cookie"]

# 수정하고 profile을 업데이트한다. 굳이 이거때문에 새로 설치할 필요는 없고 수정이 가능한 것 같다.

istioctl install -f updated-profile.yaml

# AuthorizationPolicy를 이용해서 어떤 요청을 가로채서 외부서비스에 인가를 받게할지 결정한다.


apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: example-auth-policy
  namespace: istio-system
spec:
  action: CUSTOM
  provider:
    name: "oauth2-proxy"
  rules:
  - to:
    - operation:       
        paths: ["/app"]
        notPaths: ["/oauth2/*"]
  selector:
    matchLabels:
      app: istio-ingressgateway

이 예시는 demo.example.com에서 오는 요청이나, 요청 경로가 /api로 시작하는 경우에 요청을 가로채서 외부서비스에 인가받게하는 룰을 설정하는 예시이다.




https://medium.com/@senthilrch/api-authentication-using-istio-ingress-gateway-oauth2-proxy-and-keycloak-a980c996c259
https://medium.com/@senthilrch/api-authentication-using-istio-ingress-gateway-oauth2-proxy-and-keycloak-part-2-of-2-dbb3fb9cd0d0
https://keycloakthemes.com/blog/how-to-setup-sign-in-with-google-using-keycloak
https://freestrokes.tistory.com/148
https://medium.com/@lucario/istio-external-oidc-authentication-with-oauth2-proxy-5de7cd00ef04
https://oingdaddy.tistory.com/252


Istio는 요청에 대한 인가를 JWT토큰으로 판단한다.

Keycloak는 OIDC 프로토콜을 기반으로 유저를 인가할 수 있게해주는 오픈소스 idP이다.
idP는 identiㅅy Provider의 준말이다.

OIDC는 OAuth2.0을 기반으로 하는데, OAuth는 Open Authentication의 약자로 외부 인가 서비스의 대리자역할로 인가받는 것들 도와주는 여러 메커니즘의 집합을 의미한다.
OAuth는 사용자가 Kekcloak과만 크리덴셜을 공유하면 되기 때문에 직접 외부서비스로부터의 인가를 스크래치로 구성할 때보다 안전하다.

개발자는 Keycloak에 계정을 만들고, 이 계정의 ID/PW로 인증을 한다. OAuth 서버와 유사한 역할을 한다.

OAuth2-Proxy

OAuth2-Proxy는 오픈소스 리버스프록시 솔루션이다. OAuth Client의 역할을 한다. Resource Owner(사용자)와 외부 서비스 간 중재를 하는 서버역할이다.
Oauth server역할인 Keycloak과 연동해서 유저의 인가를 확인한다. Keycloak에서 토큰을 받고 Oauth 백엔드에 저장한다. 그리고 토큰이 만료되면 알아서 refresh 토큰으로 토큰을 갱신한다.

Ingress Controller는 들어오는 모든 요청 중 경로가 /auth/* 이 아닌 모든 요청을 OAuth2-proxy로 전달한다. /auth/*를 제외하는 이유는 이 경로로 오는 요청은 oauth2-proxy가 keycloak으로 리다이렉트를 요청하는 경우이기 때문에, oauth2-proxy가 아닌 keycloak으로 요청이 포워딩 되어야한다. OAuth2-proxy는 인가가 필요하면 Keycloak으로 트래픽을 리다이렉팅한다. 인가가 정상적으로 이루어지면 oauth2-proxy는 access token, id token and refresh token을 idP(Keycloak)으로부터 받아서 자신의 백엔드(redis를 사용한다고함)에 저장하고, 헤더에 Authorization, X-Auth-Request-Access-Token 헤더를 추가해서 ingress gateway에 200 코드를 응답한다.
인가가 이미 이루어진 상태여도 ingress gateway에 200 상태코드를 응답한다. Ingress gateway는 oauth2-proxy의 응답 상태코드를 확인하고, 헤더에 담겨있는 JWT를 확인해서 토큰이 유효한지(변형되지 않았는지) 확인후 토큰에 적혀있는 서비스에게로 라우팅을 한다.

Access token은 oauth2-proxy가 저장하고있는 것이고, 이 토큰은 gateway에 전달되어서 복호화된 후 토큰 내용에 따라 해당 서비스로 라우팅되게된다.
ID token은 사용자(브라우저)가 가지고있는 토큰이고, 이 토큰은 oauth2-proxy에 전달되어서 oauth2-proxy가 복호화한 후 내용 기반으로 인가 여부를 확인하게 된다. 인가가 되어있는 상태면 ID token에 알맞은 Access token을 gateway에 전달해주게된다.

인가가 되어있지 않은(비로그인) 상태일 때, oauth2-proxy가 요청을  keycloak으로 포워딩하는 게 아니라, oauth는 브라우저를 /auth 경로로 리다이렉트를 시키고, 브라우저는 gateway에 /auth 경로로 접근하게 되며, ingress gateway는 /auth로 오는 요청을 keycloak으로 라우팅하는 방식이다.
순서가 
사용자 - gateway - oauth2-proxy - keycloak이 아니라
사용자 - gateway - oauth2-proxy, 사용자 - gateway - keycloak인 것이다.




Ingress gateway는 인가된 트래픽을 요청하는 서비스로 라우팅할 때, 요청에 ID token이 담기는 Authorization 헤더와 ACCESS token이 담기는 X-Auth-Request-Access-Token헤더를 추가해서 라우팅한다.
* 이 헤더는 게이트웨이가 붙이는 것? 아니면 요청에 붙어져있는것?








만약 Spring cloud gateway가 서버처럼 따로 구축하는 거라면
- Istio ingress gateway 뒷단에 바로 붙여서 사용할 수 있음. 이러면 Ingress gateway는 트래픽 제어는 하지 않고 그냥 Istio Proxy로 트래픽을 받기 위한 입구역할만 하고, 인증/인가+트래픽제어는 Spring Cloud gateway에서 하게 됨
- 단점은 추가적인 서버 구축 비용이 들어감. Istio만 하면 될 걸 Spring까지 추가하기 때문임

만약 Spring Cloud Gateway가 서버처럼 따로 구축하는 게 아니라면
- Istio ingress gateway와 Spring cloud gateway 둘 다 쓰면 됨. 가장 앞단에서 istio gateway가 받아서 전달하는 방식. 이게 맞으면 가장 베스트이긴 함.

어쨌든 Istio Ingress Gateway는 꼭 쓰긴 써야함. 서비스메시를 사용하기 위해서 Istio ingress gateway가 필요




Istio ingress gateway
- Spring cloud gateway + 고급 보안(mTLS, 서비스 간 인가설정)



# Keycloak K8S 배포

helm repo add bitnami https://charts.bitnami.com/bitnami
helm install keycloak bitnami/keycloak --set auth.adminUser=admin,auth.adminPassword=admin

helm install keycloak oci://registry-1.docker.io/bitnamicharts/keycloak --set auth.adminUser=admin,auth.adminPassword=admin

: PASSWORD는 K8S 시크릿에 알아서 저장됨

kubectl edit service keycloak
for nodeport




oauth2-proxy를 외부 공급자로 설정

/auth 경로 제외하고는 모두 프록시에게 검증받도록 authorizationpolicy 설정

사용자 브라우저 클러스터에 접근

게이트웨이는 외부provider인 oauth2-proxy에 cookie, authrization 헤더 전달

프록시는 헤더 확인

- cookie 없으면 keycloak에 요청

keycloak은 브라우저에 access code 응답

브라우저는 프록시에 요청

토큰은 프록시 세션에 저장하고 auth

- cookie는 있는데 authorization 없으면 프록시 세션 확인

확인되면 authrization 헤더에 세션에 저장되어있던 토큰 담아서 200 응답

- authorization 있는데 access code면 코드 가지고 키클락에 요청

키클락은 코드 보고 프록시에 토큰 발급해서 응답

프록시 세션이 토큰 저장하고 authorization 헤더에 토큰 담아서 리다이렉션

브라우저는 토큰 헤더 가지고 게이트웨이 접근

게이트웨이는 프록시에게 인가 포워딩

프록시는 authorization 헤더에 토큰 보고 200 응답 및 헤더 추가

- authorization 있는데 토큰이면 200 응답하고 헤더 추가



https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider#keycloak-oidc-auth-provider

https://wiki.onap.org/pages/viewpage.action?pageId=162105420


apiVersion: v1
kind: Service
metadata:
  labels:
    app: oauth-proxy
  name: oauth-proxy
spec:
  type: NodePort
  selector:
    app: oauth-proxy
  ports:
  - name: http-oauthproxy
    port: 4180
    nodePort: 31023
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: oauth-proxy
  name: oauth-proxy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: "oauth-proxy"
  template:
    metadata:
      labels:
        app: oauth-proxy
    spec:
      containers:
      - name: oauth-proxy
        image: "quay.io/oauth2-proxy/oauth2-proxy:v7.2.0"
        ports:
        - containerPort: 4180
        args:
          - --http-address=0.0.0.0:4180
          - --upstream=http://echo:80
          - --set-xauthrequest=true
          - --pass-host-header=true
          - --pass-access-token=true
        env:
          # OIDC Config
          - name: "OAUTH2_PROXY_PROVIDER"
            value: "google"
          - name: "OAUTH2_PROXY_OIDC_ISSUER_URL"
            value: "http://test-nlb-b4a2aa0a60c67255.elb.ap-northeast-2.amazonaws.com/realms/my-realm"
          - name: "OAUTH2_PROXY_CLIENT_ID"
            value: "oauth2-proxy-client"
          - name: "OAUTH2_PROXY_CLIENT_SECRET"
            value: "JGEQtkrdIc6kRSkrs89BydnfsEv3VoWO"
          # Cookie Config
          - name: "OAUTH2_PROXY_COOKIE_SECURE"
            value: "false"
          - name: "OAUTH2_PROXY_COOKIE_SECRET"
            value: "ZzBkN000Wm0pQkVkKUhzMk5YPntQRUw_ME1oMTZZTy0="
          - name: "OAUTH2_PROXY_COOKIE_DOMAINS"
            value: "*"
          # Proxy config
          - name: "OAUTH2_PROXY_EMAIL_DOMAINS"
            value: "*"
          - name: "OAUTH2_PROXY_WHITELIST_DOMAINS"
            value: "*"
          - name: "OAUTH2_PROXY_HTTP_ADDRESS"
            value: "0.0.0.0:4180"
          - name: "OAUTH2_PROXY_SET_XAUTHREQUEST"
            value: "true"
          - name: OAUTH2_PROXY_PASS_AUTHORIZATION_HEADER
            value: "true"
          - name: OAUTH2_PROXY_SSL_UPSTREAM_INSECURE_SKIP_VERIFY
            value: "true"
          - name: OAUTH2_PROXY_SKIP_PROVIDER_BUTTON
            value: "true"
          - name: OAUTH2_PROXY_SET_AUTHORIZATION_HEADER
            value: "true"





apiVersion: v1
kind: Service
metadata:
  name: echo
spec:
  type: ClusterIP
  selector:
    app: echo
  ports:
  - name: echo
    port: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo
spec:
  selector:
    matchLabels:
      app: echo
  template:
    metadata:
      labels:
        app: echo
    spec:
      containers:
      - name: port
        image: nginxdemos/hello
        ports:
        - containerPort: 80





apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  meshConfig:
    accessLogFile: /dev/stdout
    extensionProviders:
    - name: "oauth2-proxy"
      envoyExtAuthzHttp:
        service: "oauth2-proxy-1694621050.default.svc.cluster.local"
        port: "4180"
        includeHeadersInCheck: ["authorization", "cookie","x-forwarded-access-token","x-forwarded-user","x-forwarded-email","x-forwarded-proto","proxy-authorization","user-agent","x-forwarded-host","from","x-forwarded-for","accept","x-auth-request-redirect"]
        headersToUpstreamOnAllow: ["authorization", "path", "x-auth-request-user", "x-auth-request-email", "x-auth-request-access-token","x-forwarded-access-token"]
        headersToDownstreamOnDeny: ["content-type", "set-cookie"]









apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: gateway
  namespace : istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
    - port:
        number: 80
        name: http
        protocol: HTTP
      hosts:
        - '*'







metadata:
  name: gateway-vs
spec:
  hosts:
    - '*'
  gateways: 
    - istio-system/test-gateway
  http:
    - match:
      - uri:
          prefix: /oauth2
      route:
      - destination:
          host: oauth-proxy.default.svc.cluster.local
          port:
            number: 4180
    - match:
      - uri:
          prefix: /
      route:
      - destination:
          host: echo.default.svc.cluster.local
          port:
            number: 3000





apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: policy
  namespace: istio-system  
spec:
  action: CUSTOM
  provider:
    name: "oauth2-proxy"
  rules:
  - to:
    - operation:
        paths: ["/app"]
  selector:
    matchLabels:
      app: istio-ingressgateway
# 이건 app만 제외하고 검증하겠다는 뜻
# /app 만 검증하려면 아래같이 수정

apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: policy
  namespace: default
spec:
  action: CUSTOM
  provider:
    name: "oauth2-proxy"
  rules:
  - to:
    - operation:
        paths: ["/oauth"]

apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
 name: allow-all
 namespace: istio-system
spec:
 rules:
 - {}
 selector:
  matchLabels:
   app: istio-ingressgateway




apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: oauth2-ingress
  namespace: istio-system
spec:
  workloadSelector:
    labels:
      istio: ingressgateway
  configPatches:
  - applyTo: CLUSTER
    match:
      cluster:
        service: oauth-proxy
    patch:
      operation: ADD
      value:
        name: oauth
        dns_lookup_family: V4_ONLY
        type: LOGICAL_DNS
        connect_timeout: 10s
        lb_policy: ROUND_ROBIN
        transport_socket:
          name: envoy.transport_sockets.tls
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
            sni: oauth2.googleapis.com
        load_assignment:
          cluster_name: oauth
          endpoints:
          - lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: oauth2.googleapis.com
                    port_value: 443
  - applyTo: HTTP_FILTER
    match:
      context: GATEWAY
      listener:
        filterChain:
          filter:
            name: "envoy.http_connection_manager"
            subFilter:
              name: "envoy.filters.http.jwt_authn"
    patch:
      operation: INSERT_BEFORE
      value:
       name: envoy.filters.http.oauth2
       typed_config:
         "@type": type.googleapis.com/envoy.extensions.filters.http.oauth2.v3alpha.OAuth2
         config:
          token_endpoint:
            cluster: oauth
            uri: https://oauth2.googleapis.com/token
            timeout: 3s
          authorization_endpoint: https://accounts.google.com/o/oauth2/v2/auth
          redirect_uri: "https://%REQ(:authority)%/_oauth2_callback"
          redirect_path_matcher:
            path:
              exact: /_oauth2_callback
          signout_path:
            path:
              exact: /signout
          credentials:
            client_id: myclientid.apps.googleusercontent.com
            token_secret:
              name: token
              sds_config:
                path: "/etc/istio/config/token-secret.yaml"
            hmac_secret:
              name: hmac
              sds_config:
                path: "/etc/istio/config/hmac-secret.yaml"

          inline_code: |
            function envoy_on_response(response_handle)
              if response_handle:headers():get(":status") == "401" then
                response_handle:logInfo("Got status 401, redirect to login...")
                response_handle:headers():replace(":status", "302")
                response_handle:headers():add("location", "https://naver.com")
              end
            end


--set config.provider=keycloak 
--set config.cient-id=oauth2-proxy-client 
--set config.client-secret=urkT6SppXC5dpoKegJRH8HfpekxcdUnc
-set config.redirect-url=http://localhost/oauth2/callback --
set config.oidc-issuer-url=http://localhost/realms/master -
set -cookie-secret=
set config.email-domain="*" --set config.code-challenge-method=S256
 --set config.upstream=http://localhost/app --
 set config.login_url=http://localhost/realms/master/protocol/openid-connect/auth --
 set config.redeem_url=https://localhost/realms/master/protocol/openid-connect/token --
 set config.validate-url=http://localhost/realms/master/protocol/openid-connect/userinfo --
 set config.set-xauthrequest=true --
 # -http-address 
 set config.pass-access-token=true --
 set config.pass-authorization-header=true --
 set config.pass-basic-auth=true 
 --set config.pass-host-header=true --
 set config.pass-user-headers=true


config:
  clientID: "oauth2-proxy-client"
  clientSecret: "01K8s7J1xV8gU8XaotdRUXlF4NHK8329"
  cookieSecret: "UmRaMTlQajM1a2ordWFYRnlJb2tjWEd2MVpCK2grOFM="
  cookieName: "_proxy_auth"
  configFile: |-
    provider = "keycloak_oidc"
    oidc_issuer_url="https://localhost:8080/realms/master"
    profile_url="https://localhost:8080/realms/master/protocol/openid-connect/userinfo"
    validate_url="https://localhost:8080/realms/master/protocol/openid-connect/userinfo"
    scope="my-scope openid email profile"
    pass_host_header = true
    reverse_proxy = true
    auth_logging = true
    cookie_httponly = true
    cookie_refresh = "4m"
    cookie_secure = true
    email_domains = "*"
    pass_access_token = true
    pass_authorization_header = true
    request_logging = true
    session_store_type = "cookie"
    set_authorization_header = true
    set_xauthrequest = true
    silence_ping_logging = true
    skip_provider_button = true
    skip_auth_strip_headers = false
    skip_jwt_bearer_tokens = true
    ssl_insecure_skip_verify = true
    standard_logging = true
    upstreams = [ "localhost" ]
    whitelist_domains = [".localhost"]



  ## Trouble Shooting

  1. sidecar proxy가 없으면 gateway로 라우팅을 할 수 없다.
  2. 권한이 없는 (/api) 경로로 접근하면 로그인 페이지로 리다이렉팅이 되지 않는다.
  3. 로그인시 토큰이 생성되는 거 같긴하다. 요청에 담겨서 보내지는데, 403(Forbiden) 오류가 출력된다. 왜? 권한이 없는 토큰인가?
- role이 할당되지 않아서 그랬음. admin role을 추가해주니까 admin 권한으로 대시보드에 접근 가능.
- 중요한 건 keycloak말고 /api에 접근했을 때 로그인으로 리다이렉트 시키고, 로그인되면 다시 /api로 보내는게 필요
  1. Keycloak 서버를 클러스터 외부에 별도 구축하는 과정에서 HTTPS REQUIRED 오류 발생run






docker run -p 8080:8080 -d -e KEYCLOAK_ADMIN=admin -e KEYCLOAK_ADMIN_PASSWORD=admin -e KC_HTTPS_CERTIFICATE_FILE=/opt/keycloak/conf/server.crt.pem -e KC_HTTPS_CERTIFICATE_KEY_FILE=/opt/keycloak/conf/server.key.pem -e KC_HOSTNAME_STRICT_HTTPS=false quay.io/keycloak/keycloak:latest start-dev

proxy-address-forwarding="true"