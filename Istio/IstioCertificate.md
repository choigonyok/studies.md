# Istio Certificate
# study istio kubernetes

# Certificate 생성하려면 K8S의 CRD (Custom Resource Definition)을 정의해야하는데, CRD를 정의하려면 유효한 도메인이 있어야하고, 그러려면 개발환경에서는 테스트가 불가능

---

# Istio cert-manager

Istio cert-manager는 그라파나, 프로메테우스같이 외부 오픈소스이고, Istio는 이 cert-manager와의 쉬운 연동을 지원한다. 
인증서는 cert-manager를 통해 외부 인증서를 발급받을 수도 있고, 자체적으로 self-signed 인증서를 만들어서 사용할 수도 있지만, 브라우저 상 사용자 경험이 좋지 않다.

cert-manager는 letsencrypt 또는 vault와도 연동이 가능하다.

vault는 시크릿 관리 오픈소스 서비스로, 별도의 vault서버를 생성하고 인증서를 통해 vault서버에 접근해서 시크릿, 환경변수, 패스워드 등을 저장/관리하고 공유할 수 있다.

Grafana, Prometheus, Kiali 등의 대쉬보드UI 들이 사전에 정의된 파일을 kubectl apply -f - 를 통해서 쉽게 적용할 수 있는 것과 달리, cert-manager는 따로 Certificate Kind의 오브젝트를 생성해야한다.

Ceretificate 리소스는 인증서를 요청하기 위한 리소스이다.

인증서를 발급하기 위해서는 Certificate 이전에 issuer, ClusterIssuer 둘 중 하나의 리소스가 필요하다.

issuer는 지정한 namespace에서 사용되는 인증서이고, 클러스터 전반적으로 하나의 인증서를 사용하고 싶다면, ClusterIssuer를 사용하면 된다.

Certificate는 인증서 발급을 요청할 issuer가 필요함.

cert-manager는 istio-ingressgateway가 위치한 namespace에서 실행해야한다. 

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-com
  namespace: default
spec:
  # Secret names are always required.
  secretName: example-com-tls

  duration: 2160h # 인증서 유효기간, h,m,s만 사용가능 day의 d는 사용 불가함
  renewBefore: 360h # 인증서 만료 몇 시간 전부터 인증서 재발급 시작할 건지
  subject:
    organizations:
      - jetstack
  # The use of the common name field has been deprecated since 2000 and is
  # discouraged from being used.
  commonName: example.com
  isCA: false
  privateKey:
    algorithm: RSA
    encoding: PKCS1
    size: 2048
  usages:
    - server auth
    - client auth

  # dnsNames, uris, ipAddresses 중 적어도 하나는 있어야함
  dnsNames:
    - example.com
    - www.example.com
  uris:
    - spiffe://cluster.local/ns/sandbox/sa/example
  ipAddresses:
    - 192.168.0.5

  # 어느 issuer가 발급하게 할지 꼭 명시해야하는 부분
  issuerRef:
    name: ca-issuer
    kind: Issuer # 또는 ClusterIssuers. default는 issuer임
```

```yaml
# Check that CRDs and codegen are up to date
make verify-crds verify-codegen

# Update CRDs based on code
make update-crds

# Update generated code based on CRD defintions in code
make update-codegen
```

apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: crontabs.stable.example.com
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: stable.example.com
  # list of versions supported by this CustomResourceDefinition
  versions:
    - name: v1
      # Each version can be enabled/disabled by Served flag.
      served: true
      # One and only one version must be marked as the storage version.
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                cronSpec:
                  type: string
                image:
                  type: string
                replicas:
                  type: integer
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: crontabs
    # singular name to be used as an alias on the CLI and for display
    singular: crontab
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CronTab
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - ct

이렇게 할당된 인증서는 spec.secretName에 지정한 이름으로 Secret 오브젝트가 생성되어 쿠버네티스에 저장된다.

renewBefore는 너무 길게 duration과 비슷하게 설정하면 계속 인증서 재발급 상태 루프에 빠질 수 있으니 주의하고, 어떤 issuer는 notBefore를 설정하는 issuer도 있다.

cert-manager는 커스텀 키나 확장 키를 사용한 인증서 요청을 지원한다.



secretTemplate이 뭔지?


---

## 참고

[Istio Certificate Resources](https://cert-manager.io/docs/usage/certificate/)

