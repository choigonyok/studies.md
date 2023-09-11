# RabbitMQ 서버 쿠버네티스 클러스터에 구축하기
# kubernetes msa ops

---

## 개요

---

## rabbitMQ cluster-oprator 배포

kubernetes krew를 통해서 rabbitmq-management plugin을 설치해야 rabbitmq dashboard ui를 사용할 수 있다.

쿠버네티스 클러스터에 rabbitMQ를 배포하기 위해 아래 커맨드를 실행한다.

```
kubectl apply -f https://github.com/rabbitmq/cluster-operator/releases/latest/download/cluster-operator.yml
```

이 manifest 안에는 rabbitMQ를 배포할 수 있는 CRD, rabbitmq-system Namespace, rabbitMQ를 위한 RBAC, ServiceAccount 등의 설정들이 포함되어있다. 아래와 같은 오브젝트들이 생성된 걸 확인할 수 있다.

![img](http://www.choigonyok.com/api/assets/74-1.png)

이것만 한다고 rabbitMQ가 배포되는 것은 아니고, rabbitMQ를 위한 환경을 구성한 것이기 때문에 추가적으로 rabbitMQ를 배포해준다.

---

## RabbitmqCluster 배포

```yaml
apiVersion: rabbitmq.com/v1beta1
kind: RabbitmqCluster
metadata:
  name: production-rabbitmqcluster
spec:
  replicas: 3
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 1
      memory: 2Gi
  rabbitmq:
          additionalConfig: |
                  log.console.level = info
                  channel_max = 1000
                  default_user= guest 
                  default_pass = guest
                  default_user_tags.administrator = true
  service:
    type: LoadBalancer
```

request/limit 설정이나, spec.rabbitmq.additionalConfig는 프로젝트의 목적에 맞게 설정할 수 있다.

### spec.rabbitmq.additionalConfig.log.console.level

rabbitMQ의 로깅 레벨을 설정한다. info는 기본이고, debug로 설정하면 더 자세한 로그를 볼 수 있게된다.

### spec.rabbitmq.additionalConfig.default_user

User를 설정한다. 이 User는 이후에 rabbitMQ dashboard UI에서 사용될 아이디이다.

### spec.rabbitmq.additionalConfig.default_pass

Password를 설정한다. 이 Password는 이후에 rabbitMQ dashboard UI에서 사용될 비밀번호이다.

### spec.rabbitmq.additionalConfig.default_user_tags.administrator

위 default_user에서 선언한 유저에게 admin 권한을 줄 것을 명시한다.

### spec.rabbitmq.additionalConfig.channel_max

메시지를 전달할 수 있는 최대 채널의 수를 설정한다. 프로젝트의 구성, 서비스의 규모와 트래픽에 따라 다르게 설정해야할 것이다.

우선 예시로 1000을 설정해두었다.

<br/>

kubectl apply -f - 를 통해서 설정대로 RabbitmqCluster를 쿠버네티스 클러스터에 배포해준다. 그럼 아래와 같이 rabbitMQ Pod와 SVC가 잘 실행중인 걸 확인할 수 있다.

![img](http://www.choigonyok.com/api/assets/74-2.png)

![img](http://www.choigonyok.com/api/assets/74-3.png)

### EXTERNAL-IP PENDING

rabbitMQ를 로드밸런서 타입으로 생성했는데 EXTERNAL-IP가 pending 상태이다. 이건 내가 KIND 클러스터 위에 rabbitMQ를 배포했기 때문이다. KIND는 kubeadm을 기반으로 로컬 쿠버네티스 클러스터를 구성한다. kubeadm 등의 베어메탈 도구로 구성된 클러스터는 로드밸런서 타입 서비스를 지원하지 않는다.

자세한 내용과 metalLB를 통한 EXTERNAL-IP 구현 과정은 이 블로그의 게시글인

1.  kubeadm으로 클러스터 스크래치 구성
2.  KIND로 개발환경 구성
3.  KindIstio.md

에 작성되어있다.

어쨌든 EXTERNAL-IP를 할당하기 위해 metalLB를 사용하면 아래와 같이 metalLB Pod들이 생성되고,

![img](http://www.choigonyok.com/api/assets/74-4.png)

아래와 같이 pending 상태이던 rabbitmq Service의 EXTERNAL-IP도 자동으로 IP가 할당된 것을 확인할 수 있다.

![img](http://www.choigonyok.com/api/assets/74-5.png)

KIND로 클러스터를 생성할 때 --config 옵션으로 30080포트만을 열어주었기 때문에 rabbitmq Service의 포트도 변경해주어야한다. 

### Port Re-assign

```
kubectl edit svc RABBITMQ_SERVICE_NAME
```

로 rabbitMQ Service YAML 파일에 접근해서 http 프로토콜의 포트를 30080로 변경 후 저장하면 아래와 같이 포트가 잘 지정된 것을 확인할 수 있다.

![img](http://www.choigonyok.com/api/assets/74-6.png)

이 로드밸런서 타입 서비스는 Dashboard Web UI에 접근하기 위한 서비스이다.

---

## Web Dashboard UI 확인


이제 localhost:30080 또는 실제 운영환경에서 이 글을 따라 rabbitMQ를 배포했다면 rabbitMQ Loadbalancer type Service의 EXTERNAL-IP:30080으로 접근하면 아래 사진처럼 웹 UI가 잘 실행되는 걸 확인할 수 있다.

![img](http://www.choigonyok.com/api/assets/74-7.png)

RabbitmqCluster YAML 파일에서 설정한 user/pass인 guest:guest를 입력하면 정상적으로 로그인이 가능하다.

![img](http://www.choigonyok.com/api/assets/74-8.png)


실제 rabbitMQ를 이용해 통신을 할 때는 30080포트를 이용하지 않는다. 30080포트는 웹 UI를 사용하기 위한 포트이고, 서비스간 통신은 AMQP 프로토콜 기반으로 메시지를 주고받는다.

![img](http://www.choigonyok.com/api/assets/74-9.png)

따라서 5672포트를 이용해서 통신해야한다. rabbitMQ도 클러스터 내부에서 하나의 네트워크를 공유하기 때문에 NodePort를 사용하지 않고도 서비스 디스커버리나 ClusterIP를 이용해 rabbitMQ와의 통신을 구현하면 되겠다.

---

## 참고

[Deploying RabbitMQ on Kubernetes using RabbitMQ Cluster Operator](https://medium.com/nerd-for-tech/deploying-rabbitmq-on-kubernetes-using-rabbitmq-cluster-operator-ef99f7a4e417)

[How To Deploy RabbitMQ With The Kubernetes Operators In 1 Hour](https://getbetterdevops.io/how-to-deploy-rabbitmq-with-the-cluster-kubernetes-operators/)

[RabbitMQ Cluster Kubernetes Operator Quickstart](https://www.rabbitmq.com/kubernetes/operator/quickstart-operator.html)