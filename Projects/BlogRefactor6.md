# #12. EBS PersistenceVolume 설정
# project blog kubernetes aws

---

## 개요

기존에는 hostPath 볼륨을 이용해 블로그의 이미지파일을 저장했다. 백엔드 파드가 종료되고 재생성되더라도 hostPath 볼륨으로 데이터가 사라지지 않을 수 있었지만, 노드 자체가 삭제되거나 백엔드 파드가 다른 노드에서 생성된다면 여전히 데이터 소실 가능성이 있는 상태였다.

그래서 리팩토링 과정에서 AWS의 Elastic Block Store를 PV로 설정해서 쿠버네티스 상에서 별도 오브젝트로 관리되며 소실 위험성이 없도록 변경했다.

---

## EBS, PV, PVC, Volume

각 용어의 개념을 간단하게 짚고 넘어가자.

### EBS

EBS는 Elastic Block Store의 준말로, AWS에서 제공하는 스토리지이다. EC2 인스턴스에 붙였다 떼었다 할 수 있고, 용량을 늘릴 수도 있다. 프리티어는 달마다 30GiB의 스토리지가 무료로 제공된다.

### Volume

기본적으로 파드 생성 시 마운트될 수 있는 볼륨이다. 볼륨은 emptydir, hostPath 두 종류로 나뉘는데, **hostPath**는 현재 파드가 위치한 노드의 호스트 스토리지를 볼륨으로 사용하는 것, **emptyDir**은 다중 컨테이너 파드 내에서 컨테이너들끼리 디렉토리를 공유하기 위한 것이다.

### PV

PV는 Persistence Volume의 준말로, 영구적으로 보존해야하는 데이터들을 위한 볼륨이다. Volume의 emptyDir은 파드가 삭제되면 함께 사라지고, hostPath는 노드가 삭제되면 함께 사라진다. PV는 Volume처럼 노드나 파드에 종속되어있지않고, 하나의 쿠버네티스 오브젝트로서 생성되기 때문에 파드, 노드 등의 타 오브젝트의 삭제에 영향을 받지 않는다.

### PVC

PVC는 Persistence Volume Claim의 준말로, PV와 파드를 연결해주는 역할을 한다. 파드가 PV를 사용하고싶으면 해당 PV와 연결되어있는 PVC를 Claim해서 PV를 볼륨으로 사용할 수 있게된다.

---

## EBS Plugin deprecated

원래 EBS는 아래와 같이 쿠버네티스 manifest에서 쉽게 볼륨으로 지정할 수 있었다.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-ebs
spec:
  containers:
  - image: registry.k8s.io/test-webserver
    name: test-container
    volumeMounts:
    - mountPath: /test-ebs
      name: test-volume
  volumes:
  - name: test-volume
    awsElasticBlockStore:
      volumeID: "<volume-id>"
      fsType: ext4
```

그러나 이제 쿠버네티스는 v1.27부터 EBS 생성 플러그인(**awsElasticBlockStore:** 필드)을 더 이상 지원하지않는다. EBS 뿐만 아니라 Azure, Google Cloud 등의 전체적인 PV 플러그인들이 지원 중단됐거나 deprecated 상태이다. 그럼 이제 EBS는 더 이상 쿠버네티스에서 스토리지로 사용하지 못하는 걸까? 그렇지 않다.

### CSI

다양한 플랫폼의 스토리지들을 플러그인으로 사용하는 대신, CSI를 통해 PV를 설정할 수 있다. CSI는 Container Storage Interface의 준말이다. 컨테이너에 스토리지를 노출할 수 있게 해주는 API이다. 이 CSI를 통해 기존에 플러그인으로 제공했던 스토리지 블록들뿐만 아니라 앞으로 더 다양한 플랫폼의 스토리지들이 사용될 수 있을 것이다.

플러그인의 관리 비용 축소 + 각 플랫폼에서 제공하는 CSI를 통해 더 많은 종류의 스토리지 사용 가능. 개인적인 생각으로는 이 장점들로 인해 기존 플러그인들을 점점 지원 중단하고 CSI로 바뀌가고있는 것 같다.

[K8S Official Document](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)

---

## IAM 권한 설정

CSI로 사용자가 원하는 스토리지를 사용하고 관리하기 위해서는 CIS driver에 권한이 필요하다. 

나의 경우 kubeadm을 통한 베어메탈 클러스터를 구성했기 때문에 두 가지 선택지가 있다. (EKS로 구성하면 한 가지 선택지가 더 늘어난다.)

### 1. instance profile

IAM이 사용자(사람)을 구분하기 위한 것이라면, 인스턴스 프로파일은 인스턴스를 구분하기 위한 것이다.

IAM 서비스로 들어가서 역할 탭에 들어간다. 역할 생성을 눌러, EC2가 AmazonEBSCSIDriverPolicy권한을 가질 수 있는 역할을 생성한다.

그리고 AWS 콘솔의 인스턴스 서비스에 들어가서, driver가 배치될 인스턴스의 세부정보 - IAM 역할 이 - 로 비어있는 것을 확인한다. 작업 - 보안 - IAM 역할 수정에서 방금 생성한 역할을 지정해준다.

Driver pod는 이렇게 IAM Role을 통해 EBS CIS Driver에 대한 권한을 갖게된 이 인스턴스/노드에 배포되어야만 정상적으로 드라이버가 권한을 통해 EBS를 컨트롤할 수 있다. 그래서 Driver가 그 pod에만 배치될 수 있게 nodeSelector 설정을 해주어야한다.

### 2. K8S Secret

시크릿으로 직접 IAM의 AccessKey와 SecretKey를 전달해주어서 해당 pod가 직접 권한을 사용할 수 있도록 해줄 수도 있다.

우선 IAM를 생성하고 권한을 특정해서 Driver가 해당 IAM의 권한을 사용할 수 있게 AccessKey와 SecretKey를 CSI에게 알려주어야한다. 

    arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy (AmazonEBSCSIDriverPolicy)

IAM을 생성할 때 이 권한 정책을 설정해주면 CSI에 필요한 권한들을 IAM에 설정해줄 수 있다. 이후에 해당 IAM 사용자의 보안 자격 증명 섹션에서 액세스 키를 생성하고, AccessKey와 SecretKey를 이용해 secret을 아래와 같이 만들어준다.

```
kubectl create secret generic aws-secret \
    --namespace kube-system \
    --from-literal "key_id=${AWS_ACCESS_KEY_ID}" \
    --from-literal "access_key=${AWS_SECRET_ACCESS_KEY}"
```

---

기본적으로 EBS CSI Driver는 배치되는 노드에 CriticalAddonsOnly taint를 tolerate하고, tolerationSeconds가 300으로 지정되어있다. Driver가 배치되는 이 인스턴스에 다른 Pod들이 배치되지 않도록 taint하며, 만약 배치되면 300초의 여유시간을 두고 해당 Pod가 Gracefully 삭제될 수 있게 하는 것이다.

이게 싫으면(Driver와 다른 파드들이 같이 배치될 수 있게하려면) 헬름 차트에서 Value.node.tolerateAllTaints를 false로 변경하라고 한다.

---

만약 처음 클러스터가 구성되고 노드들과 파드들이 실행될 때, Driver가 필요한 (EBS가 필요한) Pod가 있는데 아직 Driver Pod는 생성 중인 상태라면, race condition이 생기게된다. 그래서 처음 클러스터가 구성되면 모든 노드들에 

    ebs.csi.aws.com/agent-not-ready:NoExecute

이 taint가 우선 설정되도록 클러스터를 구성해야한다. Driver Pod는 컨테이너가 생성되고 다 준비가 되면 이 taint를 자동으로 삭제하는 기능을 가지고있기 때문에, Driver Pod가 준비되면 taint가 노드들에서 사라지고, 그럼 다른 Pod들, 특히 Driver를 필요로 했던 Pod들이 race condirtion없이 동기적으로 노드에 배치될 수 있게된다.

---

## Deploy Driver

드라이버를 우리의 쿠버네티스 클러스터에 배포해야한다. 그래야 이 Driver Pod가 배치된 노드에 있는 IAM 권한을 이용해서 AWS EBS를 설정하고 관리할 수 있다. 

드라이버는 깃허브에 올라와있는 manifest를 통해 Driver를 배포할 수도 있고, 헬름차트를 이용해서 배포할 수도 있다.

    kubectl apply -k "github.com/kubernetes-sigs/aws-ebs-csi-driver/deploy/kubernetes/overlays/stable/?ref=release-1.22"

    helm repo add aws-ebs-csi-driver https://kubernetes-sigs.github.io/aws-ebs-csi-driver
    helm repo update
    helm upgrade --install aws-ebs-csi-driver \
    --namespace kube-system \
    aws-ebs-csi-driver/aws-ebs-csi-driver

이렇게 쿠버네티스에 배포된 Driver Pod는 kube-system namespace에서 확인할 수 있다.

---

## EBS 정적/동적 프로비저닝

Driver가 다 준비되었으니 이제 EBS를 프로비저닝하면 된다.

볼륨의 생성 및 적용 방식은 두 가지가 있다. 바로 정적/동적 프로비저닝이다.

정적 프로비저닝은 사전에 먼저 생성되어있는 볼륨을 가져다가 Claim해서 쓰는 것이고, 동적 프로비저닝은 볼륨을 원할 때 볼륨을 직접 생성한 후에 Claim해서 쓰는 것이다.

정적은 미리 사전에 IaC로 볼륨을 생성해두면 현재 인프라의 state를 코드로 관리하기가 쉽다는 장점이 있겠고, 동적은 유동적으로 볼륨을 생성할 수 있다는 장점이 있겠다.

### 정적 프로비저닝

정적 프로비저닝은 **PersistenceVolume** kind의 오브젝트를 이용해서 볼륨을 생성한다. 앞서 클러스터에 배포한 Driver를 통해 AWS에 미리 생성해둔 EBS를 연결하고 볼륨으로 지정한다.

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: ebs-csi-pv
spec:
  capacity:
    storage: 10Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: ""
  csi:
    driver: ebs.csi.aws.com
    volumeHandle: <EBS_VOLUME_ID>
    volumeAttributes:
      fsType: ext4
```

accessModes는 여러 모드가 있지만 EBS CSI에서는 ReadWriteOnce만 지원한다.

storageClassName은 말그대로 스토리지 클래스를 사용할 건지, 사용할 거면 이름이 뭔지를 지정하는 부분인데, 이 스토리지 클래스는 동적 프로비저닝에 사용된다. 이 경우는 정적 프로비저닝이기 때문에 스토리지 클래스가 필요없으면 ""로 사용하지 않음을 명시해준다.

csi 속성에서 driver와 volumeHandle, volumAttributes를 설정해준다.

다음은 Claim이다. PV가 볼륨을 정의하는 부분이라면, PVC는 볼륨과의 interface라고 이해하면 될 것 같다. Pod에서는 이 PVC를 통해 PV를 볼륨으로 마운트할 수 있다.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ebs-csi-pvc
spec:
  accessModes:
    - ReadWriteOnce
  volumeName: ebs-csi-pv
  resources:
    requests:
      storage: 10Gi
```

다음은 이 PVC를 서비스에서 어떻게 마운트하는지 예시를 보자.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app-label
  template:
    metadata:
      labels:
        app: app-label
    spec:
      containers:
        - name: app-container
          image: nginx
          volumeMounts:
            - mountPath: /var/www/html
              name: app-volume
      volumes:
        - name: app-volume
          persistentVolumeClaim:
            claimName: ebs-csi-pvc
```

이 경우는 EBS 볼륨을 nginx 컨테이너 내부의 /var/www/html 경로로 설정해서, 해당 경로에서 nginx가 제공하는 빌드파일이 EBS 볼륨에서 관리되도록 하는 예시이다.

이 Pod가 종료 후 재시작되더라도 EBS 볼륨이 /var/www/html에 마운트되어있기 때문에 데이터 손실 없이 안정적으로 빌드파일을 제공할 수 있게된다.

### 정적 프로비저닝

정적 프로비저닝은 앞에 언급한 것처럼 **StorageClass** kind의 오브젝트를 이용해서 볼륨을 생성한다. StorageClassName을 지정한 Pod가 생성될 때,  앞서 클러스터에 배포한 Driver를 통해 StorageClass에 정의된대로 EBS를 프로비저닝하게 된다.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: aws-ebs-sc
provisioner: ebs.csi.aws.com
parameters:
  type: gp2
  iopsPerGB: "10"
  fsType: ext4
reclaimPolicy: Delete
allowVolumeExpansion: true
mountOptions:
  - debug
volumeBindingMode: WaitForFirstConsumer
```

PV manifest와는 다르게 driver 속성 대신 provisioner 속성을 사용한다. 말그대로 새로운 EBS를 생성하는 것이기 때문에 이렇게 속성명을 설정한 것 같다.

아래는 이 storageClass와 Claim을 잇는 부분이다. Pod가 이 Claim을 통해 볼륨을 호출하면, Claim은 StorageClass대로 EBS를 프로비저닝해서 Pod에게 볼륨으로 전달할 것이다.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ebs-pvc
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: aws-ebs-sc
  resources:
    requests:
      storage: 10Gi
```

아래는 이 Claim을 어떻게 Pod 생성시 적용하는지에 대한 예시이다. 정적 프로비저닝과 동일하다. 여기서 PVC의 장점을 볼 수 있다.

만약 PVC가 없이 Pod에서 직접 볼륨을 지정하고 마운트하는 방식이었다면 코드도 코드대로 길어지고 지저분해지겠지만, 동적 프로비저닝을 해서 사용하던 EBS 대신 원래 생성되어있던 EBS로 볼륨을 교체하고 싶을 때, Pod의 manifest도 수정이 불가피해진다.

이렇게 PVC가 있음으로써 Pod와 볼륨간의 결합이 약해지고, Pod가 동적이 아닌 정적 프로비저닝으로 전환하고 싶다면 그냥 Claim 내용만 수정해주면 알아서 바뀐 볼륨으로 적용이 될 것이다. 이런 약결합이 가능해진다.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app-label
  template:
    metadata:
      labels:
        app: app-label
    spec:
      containers:
        - name: app-container
          image: nginx
          volumeMounts:
            - mountPath: /var/www/html
              name: app-volume
      volumes:
        - name: app-volume
          persistentVolumeClaim:
            claimName: ebs-pvc
```

---

## 정리

CSI를 사용해야만 EBS를 사용할 수 있게된 것은 참 안타깝다. 그래도 CSI가 없었으면 EBS 볼륨 설정하는 법만 알고, 다른 외부 스토리지를 가져다 쓰는 법은 알지 못했을텐데, 강제적으로 쿠버네티스가 어려운 길을 걸어가게 해줘서 많이 배운 것 같다.

---

## 참고

[Github: aws-ebs-csi-driver](https://github.com/kubernetes-sigs/aws-ebs-csi-driver/blob/master/docs/install.md)
[K8S: instance profile](https://repost.aws/ko/knowledge-center/attach-replace-ec2-instance-profile)
[CriticalAddsOn](https://unofficial-kubernetes.readthedocs.io/en/latest/concepts/cluster-administration/guaranteed-scheduling-critical-addon-pods/)
[Taint/Toleration](https://access.redhat.com/documentation/ko-kr/openshift_container_platform/4.6/html/post-installation_configuration/post-install-taints-tolerations#nodes-scheduler-taints-tolerations-about-seconds_post-install-node-tasks)




ebs 플러그인 지원 중단으로 csi이용해야함

PV는 정적/동적 프로비저닝 두가지 가능

ebs연결은 드라이버 직접 설치 및 정책 설정 방식과 매니페스트 이용 방식
차이는 커스터마이징과 간편성 차이

동적은 storage class 이용, 동적이면 PV kind 매니페스트는 없어도 됨
정적은 pv 이용

