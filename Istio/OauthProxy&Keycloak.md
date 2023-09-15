각 오브젝트가 하는 일

Keycloak 서비스 : 구글
Oauth2-Proxy : 
EnvoyFilter : 특정 헤더 유무에 따른 트래픽 인가/비인가
RequestAuthentication : 서비스 앞단에서 토큰의 유효성 검증
AuthorizationPolicy : 어떤 경로/호스트로 향하는 트래픽을 intercept해서 검증할지 지정

1. kind 클러스터 extraportmappings 설정
2. kind 클러스터 생성
3. istio 설치
4. 사이드카 프록시 주입 설정
5. keycloak 헬름 차트 배포 후 테스트
6. ingress gateway, gateway, vs, echo test 배포 후 테스트
7. oauth2-proxy config 파일 설정
8. oauth2-proxy 헬름 차트 배포 (with REDIS) 후 테스트
9. requestAuthentication 설정
10. EnvoyFilter 설정


# istioctl meshConfig

meshConfig는 istio에게 외부인가서비스를 선언하기 위한 부분
AuthorizationPolicy로 intercept된 트래픽의 헤더 중, includeHeaderInCheck에 포함된 헤더가 외부인가서비스인 oauth2-proxy로 전달된다

```conf
meshConfig:
  extensionProviders:
    - name: "oauth2-proxy"
      envoyExtAuthzHttp:
        service: "oauth2-proxy.default.svc.cluster.local"
        port: "80"
        includeHeadersInCheck: ["authorization", "cookie"]
        headersToUpstreamOnAllow: ["authorization", "path", "x-auth-request-user", "x-auth-request-email", "x-auth-request-access-token"]
        headersToDownstreamOnDeny: ["content-type", "set-cookie"]
```

# AuthorizationPolicy

AuthorizationPolicy는 어느 workload로 향하는 트래픽을 intercept할지를 정의하는 부분
아래 예시에는 /app 경로로 향하는 트래픽을 intercept해서 인가 확인
/app 경로 이외에는 모두 접근을 허용
selector는 대상 POD의 label을 지정해야함

```yaml 
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: echo-policy
spec:
  selector:
    matchLabels:
      app: echo
  action: CUSTOM
  provider:
    name: "oauth2-proxy"
  rules:
  - to:
    - operation:
        paths: ["/app"]
```

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
 name: allow-all
 namespace: default
spec:
 rules:
 - {}
```

# Oauth2-proxy config 파일

```conf
config:
  clientID: "oauth2-proxy-client"
  clientSecret: "urkT6SppXC5dpoKegJRH8HfpekxcdUnc"
  cookieSecret: "P7IZzGBjksqcmrfFJ8BYdw=="
  configFile: |-
    provider = "keycloak"
    redirect_url = "http://localhost/realms/master/broker/google/endpoint"
    oidc_issuer_url="http://localhost/realms/master"
    cookie_httponly = true
    cookie_refresh = "1h"
    code_challenge_method = "S256"
    cookie_secure = true
    email_domains = "*"
    pass_access_token = true
    pass_authorization_header = true
    session_store_type = "cookie"
    set_authorization_header = true
    silence_ping_logging = true
    skip_provider_button = true
    skip_auth_strip_headers = false
    skip_jwt_bearer_tokens = true
    standard_logging = true
    upstreams = [ "file:///dev/null" ]
    http_address = "0.0.0.0:80"
    set_xauthrequest = true
    profile_url = "http://localhost/realms/master/protocol/openid-connect/userinfo"
    validate_url = "http://localhost/realms/master/protocol/openid-connect/certs"
```

-approval-prompt string: OAuth approval_prompt (default "force")
  -authenticated-emails-file string: authenticate against emails via file (one per line)
  -azure-tenant string: go to a tenant-specific or common (tenant-independent) endpoint. (default "common")
  -basic-auth-password string: the password to set when passing the HTTP Basic Auth header
  # -client-id string: the OAuth Client ID: ie: "123456.apps.googleusercontent.com"
  # -client-secret string: the OAuth Client Secret
  -config string: path to config file
  -cookie-domain string: an optional cookie domain to force cookies to (ie: .yourcompany.com)
  -cookie-expire duration: expire timeframe for cookie (default 168h0m0s)
  -cookie-httponly: set HttpOnly cookie flag (default true)
  -cookie-name string: the name of the cookie that the oauth_proxy creates (default "_oauth2_proxy")
  -cookie-refresh duration: refresh the cookie after this duration; 0 to disable
  # -cookie-secret string: the seed string for secure cookies (optionally base64 encoded)
  -cookie-secure: set secure (HTTPS) cookie flag (default true)
  -custom-templates-dir string: path to custom html templates
  -display-htpasswd-form: display username / password login form if an htpasswd file is provided (default true)
  # -email-domain value: authenticate emails with the specified domain (may be given multiple times). Use * to authenticate any email
  -footer string: custom footer string. Use "-" to disable default footer.
  -github-org string: restrict logins to members of this organisation
  -github-team string: restrict logins to members of any of these teams (slug), separated by a comma
  -google-admin-email string: the google admin to impersonate for api calls
  -google-group value: restrict logins to members of this google group (may be given multiple times).
  -google-service-account-json string: the path to the service account json credentials
  -htpasswd-file string: additionally authenticate against a htpasswd file. Entries must be created with "htpasswd -s" for SHA encryption
  # -http-address string: [http://]<addr>:<port> or unix://<path> to listen on for HTTP clients (default "127.0.0.1:4180")
  -https-address string: <addr>:<port> to listen on for HTTPS clients (default ":443")
  # -login-url string: Authentication endpoint
 # -pass-access-token: pass OAuth access_token to upstream via X-Forwarded-Access-Token header
  -pass-basic-auth: pass HTTP Basic Auth, X-Forwarded-User and X-Forwarded-Email information to upstream (default true)
  -pass-host-header: pass the request Host Header to upstream (default true)
  -pass-user-headers: pass X-Forwarded-User and X-Forwarded-Email information to upstream (default true)
  # -profile-url string: Profile access endpoint
  # -provider string: OAuth provider (default "google")
  -proxy-prefix string: the url root path that this proxy should be nested under (e.g. /<oauth2>/sign_in) (default "/oauth2")
  # -redeem-url string: Token redemption endpoint
  # -redirect-url string: the OAuth Redirect URL. ie: "https://internalapp.yourcompany.com/oauth2/callback"
  # -request-logging: Log requests to stdout (default true)
  -request-logging-format: Template for request log lines (see "Logging Format" paragraph below)
  -resource string: The resource that is protected (Azure AD only)
  -scope string: OAuth scope specification
  # -set-xauthrequest: set X-Auth-Request-User and X-Auth-Request-Email response headers (useful in Nginx auth_request mode)
  -signature-key string: GAP-Signature request signature key (algorithm:secretkey)
  # -skip-auth-preflight: will skip authentication for OPTIONS requests
  -skip-auth-regex value: bypass authentication for requests path's that match (may be given multiple times)
  -skip-provider-button: will skip sign-in-page to directly reach the next step: oauth/start
  -ssl-insecure-skip-verify: skip validation of certificates presented when using HTTPS
  -tls-cert string: path to certificate file
  -tls-key string: path to private key file
  # -upstream value: the http url(s) of the upstream endpoint or file:// paths for static files. Routing is based on the path
  # -validate-url string: Access token validation endpoint
  -version: print version string


apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: authn-filter
  #namespace: istio-system
  namespace: foo
spec:
  workloadSelector:
    labels:
      app: httpbin
  configPatches:
  - applyTo: HTTP_FILTER
    match:
      context: SIDECAR_INBOUND
      #context: GATEWAY
      listener:
        portNumber: 80
        filterChain:
          filter:
            name: "envoy.http_connection_manager"
            subFilter:
              name: "envoy.router"
    patch:
      operation: INSERT_BEFORE
      value:
        #name: envoy.filters.http.ext_authz
        name: envoy.ext_authz
        typed_config:
          "@type": type.googleapis.com/envoy.config.filter.http.ext_authz.v2.ExtAuthz
          http_service:
            server_uri:
              uri: http://oauthproxy-service.default.svc.cluster.local:4180
              cluster: outbound|4180||oauthproxy-service.default.svc.cluster.local
              timeout: 3s

            authorizationRequest:
              allowedHeaders:
                patterns:
                 - exact: "cookie"
                 - exact: "x-forwarded-access-token"
                 - exact: "x-forwarded-user"
                 - exact: "x-forwarded-email"
                 - exact: "authorization"
                 - exact: "x-forwarded-proto"
                 - exact: "proxy-authorization"
                 - exact: "user-agent"
                 - exact: "x-forwarded-host"
                 - exact: "from"
                 - exact: "x-forwarded-for"
                 - exact: "accept"
                 - prefix: "x-forwarded"
                 - prefix: "x-auth-request"
                 - prefix: _oauth2_proxy
                 - exact: cookie
            authorizationResponse:
               allowed_upstream_headers:
                 patterns:
                 - exact: authorization
                 - exact: Cookie
                 - exact: cookie
                 - prefix: x-forwarded
                 - prefix: x-auth-request
                 - prefix: _oauth2_proxy
               allowedClientHeaders:
                 patterns:
                 - exact: "location"
                 - exact: "proxy-authenticate"
                 - exact: "set-cookie"
                 - exact: "authorization"
                 - exact: "www-authenticate"
                 - prefix: "x-forwarded"
                 - prefix: "x-auth-request"
               allowedUpstreamHeaders:
                 patterns:
                 - exact: "location"
                 - exact: "proxy-authenticate"
                 - exact: "set-cookie"
                 - exact: "authorization"
                 - exact: "www-authenticate"
                 - prefix: "x-forwarded"
                 - prefix: "x-auth-request"      




  workloadSelector:
    labels:
      app: reviews
  configPatches:
    # The first patch adds the lua filter to the listener/http connection manager
  - applyTo: HTTP_FILTER
    match:
      context: SIDECAR_INBOUND
      listener:
        portNumber: 8080
        filterChain:
          filter:
            name: "envoy.filters.network.http_connection_manager"
            subFilter:
              name: "envoy.filters.http.router"

    patch:
      operation: INSERT_BEFORE
      value:
        name: envoy.filters.http.ext_authz
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz
          http_service:
            server_uri:
              uri: http://oauthproxy-service.default.svc.cluster.local:4180
              cluster: outbound|80||oauth2-proxy.default.svc.cluster.local
              timeout: 3s

apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: myns-ext-authz
  namespace: myns
spec:
  configPatches:
  - applyTo: HTTP_FILTER
    match:
      context: SIDECAR_INBOUND
    patch:
      operation: ADD
      filterClass: AUTHZ # This filter will run *after* the Istio authz filter.
      value:
        name: envoy.filters.http.ext_authz
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz
          grpc_service:
            envoy_grpc:
              cluster_name: acme-ext-authz
            initial_metadata:
            - key: foo
              value: myauth.acme # required by local ext auth server.


apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  labels:
    app.kubernetes.io/name: myapp
  name: myapp
  namespace: istio-system
spec:
  configPatches:
  - applyTo: HTTP_FILTER
    match:
      context: GATEWAY
      listener:
        filterChain:
          filter:
            name: envoy.http_connection_manager
            subFilter:
              name: istio.metadata_exchange
          sni: echo.default.svc.cluster.local
    patch:
      operation: INSERT_AFTER
      value:
        name: envoy.filters.http.ext_authz
        typed_config:
          '@type': type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz
          http_service:
            authorizationRequest:
              allowedHeaders:
                patterns:
                - exact: accept
                - exact: authorization
                - exact: cookie
                - exact: from
                - exact: proxy-authorization
                - exact: user-agent
                - exact: x-forwarded-access-token
                - exact: x-forwarded-email
                - exact: x-forwarded-for
                - exact: x-forwarded-host
                - exact: x-forwarded-proto
                - exact: x-forwarded-user
                - prefix: x-auth-request
                - prefix: x-forwarded
            authorizationResponse:
              allowedClientHeaders:
                patterns:
                - exact: authorization
                - exact: location
                - exact: proxy-authenticate
                - exact: set-cookie
                - exact: www-authenticate
                - prefix: x-auth-request
                - prefix: x-forwarded
              allowedUpstreamHeaders:
                patterns:
                - exact: authorization
                - exact: location
                - exact: proxy-authenticate
                - exact: set-cookie
                - exact: www-authenticate
                - prefix: x-auth-request
                - prefix: x-forwarded
            server_uri:
              cluster: outbound|80||oauth2-proxy.default.svc.cluster.local
              timeout: 1.5s
              uri: http://oauth2-proxy.default.svc.cluster.local