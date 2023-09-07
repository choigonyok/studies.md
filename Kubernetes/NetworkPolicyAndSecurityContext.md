# NetworkPolicy, SecurityContext in K8S
# study kubernetes sec

---

## 개요

쿠버네티스는 보안을 위해 NetworkPolicy, SecurityContext를 제공한다. 각 오브젝트에 대해 깊게 알아보려고 한다.

---

## NetworkPolicy

쿠버네티스는 3,4 계층의 트래픽제어를 가능하게하는 NetworkPolicy를 제공한다. IP주소와 Port기반으로 트래픽을 제어하는 4 계층 방화벽과 유사한 기능을 한다. 이 NetworkPolicy는 **Pod**에 적용되는데, 특정 Pod들이 어떤 다른 Pod의 접근을 허용할지, 어떤 Namespace의 접근을 허용할지, 어떤 IP 블록에서의 접근을 허용할지를 설정할 수 있다.

쿠버네티스의 NetworkPolicy는 네트워크 플러그인(CNI)에 의해 구현된다. 실질적으로는 CNI가 쿠버네티스의 NetworkPolicy대로 트래픽을 제어하는 것이다. 따라서 쿠버네티스 클러스터에서 사용하는 네트워크 플러그인의 NetworkPolicy 지원여부를 확인해야한다.

---

## NetworkPolicy Manifest

설명에 앞서서, NetworkPolicy의 기본값은 모두 허용이다. 네트워크 정책을 따로 설정하지 않으면 파드로 들어오고, 파드에서 나가는 모든 트래픽이 허용된다. 물론 Pod 앞단에서 다른 접근제어나 방화벽, 인증/인가 시스템이 있다면 먼저 걸러질 것이다. 적어도 Pod level까지 도달한 트래픽, 또는 Pod에서 막 나가려는 트래픽은 모두 허용된다는 뜻이다.

쿠버네티스 공식 문서의 NetworkPolicy 매니페스트를 살펴보자

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-network-policy
  namespace: default
spec:
  podSelector:
    matchLabels:
      role: db
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - ipBlock:
            cidr: 172.17.0.0/16
            except:
              - 172.17.1.0/24
        - namespaceSelector:
            matchLabels:
              project: myproject
        - podSelector:
            matchLabels:
              role: frontend
      ports:
        - protocol: TCP
          port: 6379
  egress:
    - to:
        - ipBlock:
            cidr: 10.0.0.0/24
      ports:
        - protocol: TCP
          port: 5978
```

4개의 만다토리 필드가 있다. spec의 내부를 살펴보겠다.

### podSelector

podSelector는 이 네트워크 정책을 적용할 Pod를 선택한다. 만약 podSelector에 따로 값을 정의하지 않으면, default로 NetworkPolicy가 배치되는 Namespace의 모든 Pod들이 이 네트워크 정책대로 트래픽을 제어받게 된다. 예시에서는 role: db 라는 label을 가진 Pod들에 대해 네트워크 규칙을 정의한다.

만약 Namespace의 Pod들 중 일부에는 A 네트워크 정책, 다른 일부에는 B 네트워크 정책을 구성하고 싶다면, NetworkPolicy를 여러 개 구성해도 상관없다. 네트워크 정책은 허용보다는 restric의 개념이기 때문에 네트워크 정책 간 충돌이 일어나지 않는다.

예를 들어, 1 ~ 10 중 첫 번째 네트워크 정책이 1 ~ 8만 허용하고, 두 번째 네트워크 정책은 6 ~ 10만 허용한다고 가정하자. 두 네트워크 정책이 설정한 podSelector 규칙에 둘 다 일치하는 Pod가 있을 수 있다. 그럼 그 Pod는 두 네트워크 정책 모두 적용해서 6 ~ 8만 허용하게 된다.

네트워크 정책을 잘못 구성해서 1 ~ 5만 허용하는 규칙과, 6 ~ 10만 허용하는 규칙이 하나의 Pod에 같이 적용되는 상황이라면, 해당 Pod는 아무 트래픽도 받거나 보내지 못할 것이다.

### policyType

policyType은 Ingress/Egress 두 가지 중 선택할 수 있다. 둘 다 선택할 수도 있다. Ingress는 podSelector에서 지정한 Pod들로 들어오는 트래픽에 대한 제어를 하겠다는 뜻이고, 반대로 Egress는 podSelector에서 지정한 Pod들에서 나가는 트래픽에 대한 제어를 하겠다는 뜻이다.

### ingress/egrass

policyType으로 Ingress/Egress가 선언되었다면, Pod로 들어오고 나가는 트래픽을 어떻게 제어할 것인지 규칙을 설정해주어야한다.

from과 to는 큰 의미 없이, ingress니까 from이고, egress니까 to이다. 왜 굳이 from/to 필드를 만들었는지 잘 이해는 가지 않는다. ingress/egress의 의미를 좀 더 명확하게 하기 위한 것 같다.

from/to 내부의 속성들이 중요하다.

1. podSelector
2. namespaceSelector
3. podSelector & namespaceSelector
4. ipBlock

**podSelector, namespaceSelector**는 이름 그대로 어떤 pod와 namespace가 spec.podSelector에 선언한 Pod에 접근하는 것을 허용할지를 정한다. 위 예시 코드의 일부를 가져와보면,

```yaml
- namespaceSelector:
   matchLabels:
     project: myproject
- podSelector:
   matchLabels:
     role: frontend
```

이 NetworkPolicy의 spec.podSelector에 선언된 Pod는 **project=myproject** label을 가진 모든 Namespace에 속한 모든 Pod가 접근할 수 있고, **role=frontend** label을 가진 모든 Pod가 접근할 수 있다.

project: myproject label을 가진 Namespace안에 Pod A, B가 있고, 레이블이 없는 Namespace안에 Pod C, D가 있는데, Pod B와 C는 role=frontend label을 가지고있다면, 이 경우엔 project=myproject Namespace에 속한 A, B와 role=frontend label을 가진 B, C, 총 3개의 Pod가 접근할 수 있게된다.

3번인 **podSelector & namespaceSelector**는 1, 2번을 함께 사용하는 것과는 다른 결과를 가져온다.

```yaml
- namespaceSelector:
   matchLabels:
     project: myproject
  podSelector:
   matchLabels:
     role: frontend
```

위 코드와 아래 코드의 차이가 보일지 모르겠다. podSelector 앞에 하이픈 하나가 없다.

아래 같은 경우가 3번 케이스이다. 네트워크 정책을 이렇게 설정하면, **project=myproject** label을 가진 namespace **"안의 Pod 중에서"** role=frontend label을 가진 Pod만 접근을 허용하게 된다.

이렇게 되면 아까 A, B, C 총 세 개의 Pod가 spec.podSelector의 Pod에 접근할 수 있었던 것과 달리, project=myproject label을 가진 Namespace 안의 Pod들 (A, B) **중에서** role=frontend label을 가진 B만 접근이 가능해지게 된다. 결국 하나의 파드만 접근이 가능한 것이다.

이렇게 하이픈 하나로 NetworkPolicy의 결과가 확 바뀌기 때문에, 주의해서 잘 설정해야할 것이다.

마지막으로 4번인 **ipBlock** CIDR 블록 기반으로 트래픽을 제어한다. 

```yaml
- ipBlock:
    cidr: 172.17.0.0/16
    except:
      - 172.17.1.0/24
```

cidr 블록의 네트워크 범위는 spec.podSelector에 접근할 수 있도록 허용하는 것이다. except field를 통해 cidr 블록 중에서도 접근을 막을 예외 서브넷을 설정할 수 있다.

ipBlock은 Ingress 뿐만 아니라 Egress에서eh 설정할 수 있다. Egress에서 ipBlock을 설정하면 spec.podSelector Pod는 정의된 cidr 네트워크 범위 안에 있는 IP로만 트래픽을 보낼 수 있게된다.

### default NetworkPolicy

앞서 말했듯이 NetworkPolicy를 정의하지 않으면 모든 Ingress/Egress를 허용한다.

만약 ingress 전체만 restrict하거나, egress 전체만 restrict하고싶다면 어떻게 할까? ingress 전체를 차단하고싶다면 아래와 같이

```yaml
policyTypes:
  - Ingress
ingress:
  - {}
```

egress 전체를 차단하고싶다면 아래와 같이 

```yaml
policyTypes:
  - Engress
engress:
  - {}
```

설정할 수 있다. 

Ingree 또는 Egress에 대해 정책이 적용되긴 하지만, Ingress/Egress의 rule을 비워줌으로써 허용하는 것이 아무것도 없다고 선언할 수 있다.

### port & endPort

Ingress, Egress에서는 포트 설정도 가능하다. Ingress에서 ports가 쓰이면 spec.podSelector Pod가 Ingress를 받는 port, Egress에서 ports가 쓰이면 spec.podSelector Pod에서 나가는 port를 지정할 수 있다.

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: multi-port-egress
  namespace: default
spec:
  podSelector:
    matchLabels:
      role: db
  policyTypes:
    - Egress
  egress:
    - to:
        - ipBlock:
            cidr: 10.0.0.0/24
      ports:
        - protocol: TCP
          port: 32000
          endPort: 32768
```

위 예시처럼 port + endPort로 port의 범위를 지정할 수도 있으나, endPort는 선택사항이다. 

### multiple namespace label

Ingress나 Egress 안에서 허용할 namespace의 label을 여러 개 설정할 수도 있다.

```yaml
- namespaceSelector:
    - matchExpressions:
        - key: namespace
          operator: In
          values: ["frontend", "backend"]
```

values의 frontend, backend는 Namespace의 이름이 아니다. label의 value이다.

key가 namespace이고, values가 frontend, backend이기 때문에, 이 네트워크 정책이 적용되면, **namespace=frontend**라는 label을 가지고있거나 **namespace=backend**라는 label을 가지고있는 Namespace의 모든 Pod들이 접근 허용될 것이다.

---

## SecurityContext

이제 SecurityContext에 대해 알아보자. SecurityContext는 Pod/Container에 대한 권한 및 접근제어를 관리한다. SecurityContext는 NetworkPolicy처럼 따로 kind로 정의되는 리소스 형태가 아니라, Pod를 구성할 때 설정할 수 있는 spec 만다토리 필드의 일부이다.

쿠버네티스 공식 문서의 예시 코드를 살펴보자.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: security-context-demo
spec:
  securityContext:
    runAsUser: 1000
    runAsGroup: 3000
    fsGroup: 2000
  volumes:
  - name: sec-ctx-vol
    emptyDir: {}
  containers:
  - name: sec-ctx-demo
    image: busybox:1.28
    command: [ "sh", "-c", "sleep 1h" ]
    volumeMounts:
    - name: sec-ctx-vol
      mountPath: /data/demo
    securityContext:
      allowPrivilegeEscalation: false
```

securityContext는 Pod 및 Container에 대한 권한/접근제어를 각각 설정할 수 있다. 위의 manifest 예시에서도 spec.securityContext와 spec.containers.securityContext가 별도로 존재하는 걸 볼 수 있다. 컨테이너의 securityContext 설정은 Pod의 설정보다 우선순위가 높다. 같은 필드를 다른 값으로 설정했다면 Pod securityContext의 값은 Container securityConetxt 값으로 덮어씌워진다.

Pod의 securityContext 내용부터 살펴보자.

---

## SecurityContext in Pod

### runAsUser / runAsGroup

runAsUser가 설정되지 않으면 default는 root 권한으로 컨테이너 작업이 실행된다. runAsUser는 해당 파드 안에 있는 컨테이너들이 UID 1000 권한으로 실행되도록 하고, runAsGroup은 GID 3000 권한으로 실행되도록한다. 컨테이너 실행 권한을 지정하는 것과 보안이 무슨 관계일까?

컨테이너는 가상화를 기반으로한다. 컨테이너가 실행될 때의 root는 호스트의 root와 다르지 않다. 컨테이너가 root 권한으로 실행되면 다른 컨테이너들의 namespace나 볼륨 접근등을 root 권한으로 제어할 수 있고, 시크릿에 접근하는 등의 보안 문제가 생길 수 있다. root는 호스트 커널에도 접근할 수 있다. 

평시에는 root 권한으로 컨테이너를 실행하는 것이 큰 문제가 되지 않을 것이다. 만약 공격자에게 Pod 중 하나가 공격당해서 공격자가 Pod에 접근해있는 상태라고 가정하자. 이 공격자는 root 권한의 컨테이너를 이용해서 같은 네트워크 상 연결된 개인 정보들을 탈취할 수 있다. 그래서 컨테이너 실행시 root권한으로 실행하지 않도록 적절한 설정을 해줘야한다. 그렇다고 권한을 다 뺏어버리면 컨테이너 프로세스를 실행할 권한마저 사라질 수 있다. 따라서 적절히 필요한 권한만을 허용해주는 것이 중요하다.

### runAsNonRoot: true

특정권한이 필요할 때는 runAsUser, runAsGroup으로 알맞은 UID, GID를 설정했다. 컨테이너 내부에서 특정 권한이 아예 필요하지 않은 경우에는 어떻게 해야할까? root 권한으로는 실행하면 안되니까 아무 UID, GID나 입력해야하는걸까?

```yaml
runAsNonRoot: true
```

securityContext에 위 내용을 넣어주면 된다. 특수한 상황을 제외하고는 기본적으로 모든 파드에 이 runAsNonRoot를 설정해주는 것이 보안적으로 좋을 것이다.

### fsGroup

fsGroup은 이 Pod에 마운트되는 볼륨의 권한에 대한 것이다. 볼륨이 Pod로 마운트 될 때, Pod는 볼륨에 대한 권한을 fsGroup에 명시된 groupsID로 변경한다. 이를 통해 Pod 안에 여러 개의 컨테이너가 있더라도 같은 fsGroup을 사용함으로써 볼륨을 공유할 수 있게된다. 

만약 fsGroup이 설정되지 않는다면, 볼륨은 처음 자신에게 접근한 user에게 권한을 부여하게 되고, 볼륨을 공유하고싶은 다른 컨테이너는 권한이 없어서 볼륨에 접근하지 못하게된다.

여러 컨테이너가 하나의 공유 아이디로 볼륨에 접근할 수 있게 하기 위한 것이 fsGroup이다.

---

## SecurityContext in Container

이번엔 컨테이너의 securityContext 설정을 살펴보자. 예시 코드는 아래와 같다. 더 많은 설명을 위해 쿠버네티스 공식문서 예시 코드에서 임의로 약간 수정했다.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: security-context-demo-2
spec:
  securityContext:
    runAsUser: 1000
  containers:
  - name: sec-ctx-demo-2
    image: gcr.io/google-samples/node-hello:1.0
    securityContext:
      runAsUser: 2000
      capabilities:
        drop:
            - all
        add: ["NET_ADMIN", "SYS_TIME"]
      seccompProfile:
        type: Localhost
        localhostProfile: my-profiles/profile-allow.json
      seLinuxOptions:
        level: "s0:c123,c456"
      readOnlyRootFilesystem: true
```

앞서 언급한 것처럼, Pod에서 runAsUser를 1000으로, Container에서 runAsUser를 2000으로 설정하게되면, 1000이 2000으로 덮어씌워져서 컨테이너는 UID 2000으로 실행된다.

### capabilities

컨테이너 root권한은 호스트 root처럼 커널에 대한 접근 권한까지 갖게된다고 이야기했다. 보안을 위해 root 권한으로 실행시키지 않는 것까진 좋다. 근데 일부 커널 접근같은 추가 권한이 필요할 때가 있으면 어떻게 해야할까?

```yaml
capabilities:
  drop:
      - all
  add: ["NET_ADMIN", "SYS_TIME"]
```

이런 식으로 capabilities를 지정해줄 수 있다. capabilities field는 특정 리눅스 커널의 기능에 접근할 수 있는 권한을 준다. 모든 커널의 권한을 갖는 root와 달리, 커널이 필요하다면 필요한 커널의 권한만을 부여하게된다. 예시코드처럼 capabilities 설정은 add와 drop으로 구성된다. drop은 가지고있는 권한을 삭제하는 것이고, add는 권한을 추가하는 것이다.

아래는 기본적으로 Pod가 가지고있는 리눅스 커널 기능이다. 아무 설정 없이 생성된 Pod 내부로 아래 커맨드를 통해 들어가면 확인할 수 있다.

```
kubectl exec --stdin --tty POD_NAME -- ash 
/ # capsh --print
```

그럼 아래와 같은 커널 기능들이 출력될 것이다.

> cap_chown,cap_dac_override,cap_fowner,cap_fsetid,cap_kill,cap_setgid,cap_setuid,cap_setpcap,cap_net_bind_service,cap_net_raw,cap_sys_chroot,cap_mknod,cap_audit_write,cap_setfcap+eip

아직 내가 익숙히 아는 기능은 chown, kill밖에 없다..

```yaml
drop: ["CHOWN"]
```

위와 같이 securityContext 설정을 하고 다시 한 번 Pod 내부로 접근해서 현재 기능을 확인해보면, cap_chown이 사라진 것을 확인할 수 있는 것이다. 반대로 예시처럼 특정 기능을 add할 수도 있다. 보안을 위해서는 - all 로 모든 커널 기능을 다 drop 한 뒤에, 꼭 필요한 커널 기능들만을 add하는 방식이 가장 좋다고 한다. 물론 모든 기능들을 drop하기 전에 리눅스와 커널 기능에 대한 이해가 충분히 있어야할 것이다.

capabilities는 컨테이너 런타임에게 전달되어서, 컨테이너 런타임이 컨테이너를 생성할 떄 적용된다.

### seccompProfile

리눅스에서는 secure computing mode라는 기능을 제공한다. 이 기능은 사용자의 커널 호출을 제한한다.

```yaml
seccompProfile:
  type: Localhost
  localhostProfile: my-profiles/profile-allow.json
```

localhostProfile은 로컬에 있는 seccomp profile 파일의 경로를 지정하게된다. 이 파일에 정의된대로 사용자의 커널 호출이 제한된다. 이 역시도 seccomp profile을 작성해서 Pod에 반영시키는 것이기 때문에, seccomp profile에 대한 깊은 이해가 있어야 적용할 수 있다.

### seLinuxOptions

SELinux는 어플리케이션, 파일시스템, 프로세스 등에 대한 정책이다.

```yaml
securityContext:
  seLinuxOptions:
    level: "s0:c123,c456"
```    

이 정책은 user:role:type:level 형식의 label로 구성되어있다. 쿠버네티스 공식문서 예시로는 level만 나와있다.

서로 다른 label간 접근이 일어날 때, 접근을 거절하는 방식으로 설정될 수도 있고, 접근은 가능하지만 대신 log를 남기는 방식으로 설정될 수도 있다. seLinuxOptions를 통해 컨테이너의 권한을 축소시킬 수 있다. 컨테이너는 기본적으로 root권한으로 실행되지 않는 이상, 컨테이너 내부로 권한이 제한되지만 seLinuxOptions를 통해 더 축소된 권한을 갖도록 해서 보안을 더 안전하게 할 수 있다.

예시 코드 level에서 s0은 보안 레벨, c123,c456은 다른 리소스에 대한 접근제어를 하기 위한 항목이라고 한다.

### readonlyRootFilesystem

볼륨 등의 stateful data에 접근할 때 특정 컨테이너는 볼륨에 write하지 않고 read만 한다면, readonlyRootFilesystem 설정을 통해 불필요한 컨테이너의 권한을 제어할 수 있다. default는 false로 설정되어있고, 예시코드는 아래와 같다.

```yaml
readOnlyRootFilesystem: true
```

---

## 정리

참 공부하기 쉽지 않은 보안 영역이었다. 특히 SecurityContext의 리눅스 커널 관련된 다양한 설정을 공부하면서, 데브옵스 엔지니어의 역할과 범위가 여기까지 닿는건가? 하는 의문도 조금 생겼다. 그래도 만약 실무에서 보안 전문가가 이런저런 보안 설정이 리눅스 서버에 필요하다고 하면, 그 내용대로 쿠버네티스에 적용은 할 줄 알아야하지 않을까? 라는 생각이 들어 중간에 멈추지 않고 끝까지 알아보았다.

그리고 Istio의 Documentation과는 다르게 Kubernetes의 Documentation은 아주 정리가 잘 되어있다. 그러나 쿠버네티스의 모든 기능들을 하나하나 상세히 설명하지는 않는 것 같아 아쉬웠다. 수업 내용은 몇 개 빼먹지만 강의력이 훌륭한 일타강사 느낌의 문서였다. 그래도 Istio의 Documentation에 비하면 내용을 파악하기 전혀 부족함이 없었다..

---

## 참고

[K8S Official Documentation](https://kubernetes.io/docs/concepts/services-networking/network-policies/)

[K8S SecurityContext Documentation](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

[10 Kubernetes Security Context settings you should understand](https://snyk.io/blog/10-kubernetes-security-context-settings-you-should-understand/)

[Why Processes In Docker Containers Shouldn't Run as Root](https://www.howtogeek.com/devops/why-processes-in-docker-containers-shouldnt-run-as-root/)

[Why non-root containers are important for security](https://docs.bitnami.com/tutorials/why-non-root-containers-are-important-for-security)