
---

## RBAC

마지막으로 RBAC에 대해 알아보자.

RBAC은 Roll Based Access Control의 약자이다. 말 그대로 Roll을 기반으로 접근제어를 할 수 있게하는 기능이다. 이 기능은 kubeadm으로 구성한 클러스터에는 기본적으로 활성화가 되어있다. 활성화 되어있지 않다면 쿠버네티스에서는 아래 커맨드로 K8S API server를 실행하라고 한다.

```
kube-apiserver --authorization-mode=Example,RBAC --other-options --more-options
```

RBAC은 Role, ClusterRole, RoleBinding, ClusterRoleBinding 네 가지 쿠버네티스 오브젝트를 통해 구현할 수 있다. 크게 Role과 Binding으로 나눌 수 있다.

Role은 접근에 대한 허가 규칙을 정의하게된다. 이 허가에는 접근에 대한 Deny는 없고 허가만 가능하다. 

Role과 ClusterRole 중 Role은 특정 Namespace에 대한 허가를 하는 규칙이어서, Role manifest를 작성할 때 Namespace를 꼭 명시해줘야한다. 반대로 ClusterRole은 namespace를 명시하지 않고 클러스터 전체에 적용되게 된다. Istio의 cert-manager가 issuers를 Issuer와 ClusterIssuer로 구분하는 것과 유사하다.

ClusterRole은 리소스에 대한 권한을 정의하고, 모든 또는 일부 Namespace에서 접근할 수 있는 규칙을 생성할 수 있다.




apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: default
  name: pod-reader
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "watch", "list"]


default Namespace의 Pod들에 대한 읽기권한을 부여하는 Role이다.


apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: secret-reader
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "watch", "list"]


전체 클러스터의 시크릿에 대한 읽기권한을 부여하는 ClusterRole이다. Role과 유사하나 namespace 필드가 빠져있는 것을 확인할 수 있다.

Role과 ClusterRole의 name은 unique 해야한다.



