## K8S CLUSTER CONFIGURATION W/ KUBEADM, CONTAINERD IN UBUNTU

---

## 개요

github에 알아서 쿠버네티스 클러스터가 구성되도록 잘 만들어진 스크립트 파일이 많이 있다.

그래도 직접 AtoZ 클러스터를 구성해보면서 쿠버네티스에서 주요하게 쓰이는 툴들(kubectl kubeadm kubelet iptables ...)의 역할과 클러스터 구성과정, Ubuntu의 command들에 대한 이해를 높히고 싶어서 kubeadm을 이용해 직접 클러스터 구성을 해보기로 했다.

약 2주의 시간 동안 거의 하루 10시간씩 K8S 클러스터만 붙잡고 공부했다.

---

해당 글은 아래와 같은 환경에서 작성된 글이다.

**OS - Linux Ubuntu 22.04 TLS (jammy)**

**Instance - AWS EC2 t3.small**

**Container Runtime - Containerd**

**Kubernetes Install - Kubeadm**

**Binary Files' Version : Containerd(v1.7.3), Runc(v1.1.8)**



---

## 초기 구성

> cat : 파일 내용을 출력

> COMMAND | tee FILEPATH : COMMAND 실행 후 그 값을 FILEPATH에 입력하고, 입력한 결과를 출력

    cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
    overlay
    br_netfilter
    EOF

overlay와 br_netfilter라는 커널모듈을 입력한다.

> overlay : 파일 시스템 드라이버, 격리된 환경을 제공하고 컨테이너가 호스트와 격리되어서 컨테이너의 변화가 호스트에 변화를 주지 않도록 해주는 커널 모듈이다.

> br_netfilter : 데이터 패킷의 흐름을 제어하는 브릿지 관련 기능 제어 모듈, k8s에서 방화벽 등에 따른 패킷흐름을 제어하고, 브릿지에 연결된 인터페이스들을 하나의 네트워크로 인식하게 해준다.

리눅스에서 /etc/modules-load.d/k8s.conf 이 경로에 모듈을 올리면 부팅시에 자동으로 OS가 해당 모듈들을 load한다.

<br>

    sudo modprobe overlay
    sudo modprobe br_netfilter

모듈을 부팅시에 자동으로 실행되게 경로에 선언해뒀지만 지금 현재는 load되지 않았기 때문에, modprobe command로 모듈들을 load해준다.

<br>

    cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
    net.bridge.bridge-nf-call-iptables  = 1 
    net.bridge.bridge-nf-call-ip6tables = 1
    net.ipv4.ip_forward                 = 1
    EOF

net.bridge.bridge-nf-call-iptables  = 1 : network bridge가 iptables을 호출할 수 있도록 true로 설정한다.

> iptables : 클러스터 내부에서 패킷을 필터링 시켜주는 역할을 한다.

> network bridge : 

net.bridge.bridge-nf-call-ip6tables = 1 : ipv6 패킷에 대한 ip6tables 호출도 가능하도록 true로 설정한다.

net.ipv4.ip_forward = 1 : 패킷 포워딩 활성화, 받은 패킷을 다른 노드로 전달할 수 있게하기 위해 필요하다.

<br>

    sudo sysctl --system

위에서 입력한 sysctl 파라미터들을 적용시킨다.

<br>

    sudo swapoff -a
    (crontab -l 2>/dev/null; echo "@reboot /sbin/swapoff -a") | crontab - || true

swap(메모리가 부족해지면 메모리에 올라가는 있지만 잘 안쓰이고있는 프로세스를 디스크에 넣어두거나 넣어둔 프로세스를 필요할 때 꺼내는 것)을 비활성화한다. 노드에서 스왑이 가능해지면 여러 문제점이 생길 수 있다.

1. 성능 저하 : 디스크에 넣고 빼는 비용(cpu, storage, time)이 크다.
2. 격리 : 컨테이너는 호스트와 격리된 환경이어야하는데 디스크에 스왑하다가 호스트와 충돌할 수 있다.

이런 이유로 스왑을 통해 메모리 부족을 해결하기보다는 그냥 수평적 확장(스케일아웃)을 통해 하는 것이 낫다.

sudo swapoff -a : 현재 스왑되어서 디스크에 있는 모든 작업을 다시 다 메모리로 불러오고, 스왑을 비활성화시킨다.

> crontable : 특정 시간마다 주기적으로 실행되는 작업

crontab -l : crontab 목록을 가져온다.

2>/dev/null : 만약 현재 running중인 crontab 목록이 없어도 오류로 인식하지 않도록 설정한다. 오류가 발생하면 스크립트가 중단되기 때문이다.

echo "@reboot /sbin/swapoff -a" : 노드가 부팅될 때마다 swapoff를 자동 실행하도록 설정한다.

crontab - : 적용

true : 이미 swapoff가 설정되어있더라도 오류로 인식하지 않게 한다. 

---

## 컨테이너 런타임 containerd 설치

> container runtime : 컨테이너를 구현하는 소프트웨어이다. docker가 가장 유명한데, 22년에 k8s에서 docker지원을 공식적으로 종료했다.

containerd는 k8s에서 공식적으로 개발한 소프트웨어이고, k8s에서 containerd등의 컨테이너 런타임은 컨테이너를 실질적으로 실행, 생성, 모니터링, 리소스할당, 네트워크 설정 등의 기능을 한다.

<br>

    wget https://github.com/containerd/containerd/releases/download/v1.7.3/containerd-1.7.3-linux-amd64.tar.gz
    sudo tar Czxvf /usr/local containerd-1.7.3-linux-amd64.tar.gz

containerd의 바이너리 파일을 다운로드받는다. 버전은 아래 링크에서 최신 버전을 확인할 수 있다.

> tar OPTION PATH FILE : 압축파일을 해제해서 PATH에 파일을 푼다. tar 뒤에 옵션 설정이 가능하다.

1. C : 디렉토리로 이동, PATH로 이동시키는 역할
2. z : gzip(.gz)으로 압축된 파일을 생성/추출, .gz를 압축해제하는 역할
3. x : 압축 해제, .tar을 압축해제 하는 역할
4. v : 압축 해제 과정을 출력
5. f : 압축 해제한 파일 이름을 지정

[Containerd](https://github.com/containerd/containerd/releases)

    wget https://raw.githubusercontent.com/containerd/containerd/main/containerd.service
    sudo mv containerd.service /usr/lib/systemd/system/

containerd.service는 systemd service unit file이다. containerd를 시스템이 인식, 실행, 관리할 수 있도록, 또 부팅시마다 서비스되도록 하는 등 여러 기능들을 수행한다.

containerd.service를 /usr/lib/systemd/system/ 경로로 옮겨야 리눅스에서 해당 파일을 systemd service unit file로 인식할 수 있다.

<br>

    sudo systemctl daemon-reload
    sudo systemctl enable --now containerd

> daemon : systemd service들을 실행시키는 디폴트 프로그램

systemctl daemon-reload : systemd service 수정사항을 반영시킨다.
systemctl enable --now containerd : 바로 containerd 서비스를 시작시킨다.

<br>

    (선택)
    sudo systemctl status containerd

이 명령어를 통해서 containerd가 active(실행중) 상태인지 확인할 수 있다.

<br>

    wget https://github.com/opencontainers/runc/releases/download/v1.1.8/runc.amd64
    sudo install -m 755 runc.amd64 /usr/local/sbin/runc

runc 바이너리 파일을 다운로드받는다. 버전은 아래 링크에서 최신 버전을 확인할 수 있다.

[Runc](https://github.com/opencontainers/runc/releases)

> runc : 실질적으로 container 관련 기능을 수행하는 오픈소스, containerd는 내부적으로 runc를 이용해 container를 관리한다.

containerd와 runc의 관계는 AWS EKS와 terraform 관계같은 느낌인 것 같다. containerd는 High Level, runc는 Low Level !

<br>

    sudo mkdir -p /etc/containerd/
    containerd config default | sudo tee /etc/containerd/config.toml

config.toml파일을 위한 디렉토리를 생성한다.

default config를 config.toml에 입력한다.

> config.toml : kubeadm이 클러스터를 구축하기 위해 사용할 구성파일

<br>

    sudo sed -i 's/SystemdCgroup \= false/SystemdCgroup \= true/g' /etc/containerd/config.toml

> sed : 파일에서 항목을 찾고, 항목을 수정시키는 command, 여기서는 config.toml파일에서 SystemdCgroup = false를 SystemdCgroup = true로 변경시킨다.

> cgroup : Control Group, 리눅스에서 프로세스의 리소스를 제한한다. Mem, Cpu, Disk 등

> kubelet : 워커노드에서 마스터노드의 요청에 따라 pod를 관리하고 응답하며, 런타임과 통신해서 컨테이너를 생성/종료/리소스제한 등의 기능을 하는 소프트웨어

cgroup driver는 systemd cgroup, cgroupfs 두 가지가 있다. kubelet은 기본적으로 cgroupfs 드라이버를 사용한다.

만약 리눅스 초기 cgroup 설정이 systemd로 지정되어있다면, kubelet과 일치하지 않게 된다. 런타임과 kubelet이 리소스 제한을 위해서 통신할 때 문제가 생길 수 있다. 그래서 둘을 일치시켜주기 위해서 SystemdCgroup driver를 사용하겠다고 설정해주는 것이다.

<br>

    sudo systemctl restart containerd

설정을 적용하기 위해 containerd를 재시작한다.

---

## kubectl, kubelet, kubeadm 설치

    sudo apt-get update && sudo apt-get install -y apt-transport-https ca-certificates curl

> apt : 우분투에서 사용되는 패키지 관리자. 프로그램 설치 및 관리를 간편하게 해준다.

sudo apt-get update : 현재 사용중인 apt pkg들을 최신화시킨다.

sudo apt-get install -y apt-transport-https ca-certificates curl : 세 개의 패키지를 설치한다.

1. apt-transport-https : 패키지를 다운로드할 때 https 사용할 수 있게한다. 보안 강화
2. ca-certificates : ssl/tls 검증 패키지, 노드간 통신시 tls 인증을 거쳐 통신한다.
3. curl

<br>

    curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -

> GPG(GnuPG) key : 패키지를 설치할 때, 패키지를 제작한 곳에서 해당 패키지가 신뢰할 수 있는 패키지인지를 검증할 수 있는 공개키를 만든다. 사용자는 패키지 설치 전에 키를 받아서 패키지가 설치될 때 키와 대조해보고, 시스템은 키를 바탕으로 신뢰여부를 판단하여 패키지를 안전하게 설치할 수 있다.

curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg : 구글 클라우드에서 GPG(GnuPG)키를 다운로드한다. -s 옵션은 다운로드 과정을 출력하지 않겠다는 옵션이다.

sudo apt-key add - : 시스템의 키링에 추가한다.

<br>

    cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list
    deb https://apt.kubernetes.io/ kubernetes-xenial main
    EOF

> deb : Debian 기반(우분투)에서 사용할 수 있는 패키지임을 명시

/etc/apt/sources.list.d/kubernetes.list 파일에는 https://apt.kubernetes.io/의 kubernetes-xenial의 main에 속한 모든 패키지들이 설치되어야한다는 걸 명시해준다.

    sudo apt-get update

update를 통해 실질적으로 시스템에 명시된 경로에 있는 패키지들을 설치한다.

<br>

    sudo apt-get install -y kubelet kubeadm kubectl

> kubeadm : 쿠버네티스/클러스터 설치 툴

> kubectl : 클러스터 CLI 툴

kubelet, kubeadm, kubectl을 설치한다.  

    sudo apt-mark hold kubelet kubeadm kubectl containerd

클러스터 다 구성하고 잘 사용하던 도중에 자동으로 kubectl, kubeadm, kubelet 패키지가 업데이트되면, 버전 간 호환성 문제가 생길 수 있기 때문에, 자동 업데이트를 중지시킨다.

---

## 마스터노드 설정

    IPADDR=$(curl ifconfig.me && echo "")
    NODENAME=$(hostname -s)
    POD_CIDR="192.168.0.0/16"

IPADDR : 마스터노드의 public ip를 클러스터가 알 수 있도록 변수로 설정한다. echo ""는 뒤에 개행문자를 입력시켜준다.

NODENAME : hostname을 노드명으로 설정하기 위한 변수이다. -s는 호스트명 중 순수한 앞부분의 호스트명만을 불러오는 옵션이다.

POD_CIDR : pod들에 할당할 네트워크 범위를 지정한다.

<br>

    sudo kubeadm init --control-plane-endpoint=$IPADDR  --apiserver-cert-extra-sans=$IPADDR  --pod-network-cidr=$POD_CIDR --node-name $NODENAME --ignore-preflight-errors Swap

정의한 변수들을 바탕으로 클러스터를 초기구성(init)한다.

--ignore-preflight-errors Swap : 스왑을 비활성화한다.

<br>

    sudo mkdir -p $HOME/.kube
    sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
    sudo chown $(id -u):$(id -g) $HOME/.kube/config

k8s api서버와 통신할 수 있도록 kubectl을 사용하게 해주는 kubeconfig를 마스터노드가 읽을 수 있는 경로로 복사시키고, 권한을 부여한다.

kubeconfig는 기본적으로 /etc/kubernetes/admin.conf에 작성되어있다.

> $(id -u) : 현재 user의 id

> $(id -g) : 현재 user의 그룹 id

---

여기까지 하고 "kubectl get pods -n kube-system" 으로 kube-system namespace의 pod를 확인하면 etcd, apiserver, controller, scheduler, proxy 총 5개의 pod가 생성되어야한다.

작동은 아직 안할 수 있고, 일부 pod가 crush나 pending 상태여도 아직 네트워크 플러그인을 설치 안해서 그런거니 당황하지 않아도 된다.

---

## Calico 네트워크 플러그인 설치

    kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/calico.yaml

kubeadm은 네트워크 플러그인을 지원하지 않는다. pod끼리 통신하기 위해서는 별도의 네트워크 플러그인을 사용해야한다. 플란넬 등 여러 다양한 플러그인이 있지만, 가장 널리 사용되는 Calico를 설치하겠다.

이 Calico pod를 클러스터에 apply하면, 이전 5개의 kube pod들이 모두 running 상태로 바뀌며 정상기동한다. coredns pod 2개와, calico pod 2개가 추가되어서 총 9개의 포드가 running 상태로 정상 작동해야한다.

---

## K8S 메트릭 서버 설정

    kubectl apply -f https://raw.githubusercontent.com/techiescamp/kubeadm-scripts/main/manifests/metrics-server.yaml

> 매트릭 : 리소스 사용량 및 성능 수치

메트릭 서버를 설정해야 효과적인 모니터링이나 메트릭 기반 오토스케일링(ex : MEM의 80% 이상 사용되면 노드 스케일아웃), 로드밸런싱을 할 수 있다. 

---

## 워커노드 설정

kubeadm init의 결과물로 나온 출력 중 "kubeadm join ~ --token ~" 코드는 후에 worker node에서 실행해야하는 명령어이고, 이 명령어를 워커노드에서 실행해야 마스터노드가 토큰을 바탕으로 워커노드의 합류를 인식할 수 있다.

---

## 참고사항

클러스터 구성에 매일 약 10시간씩 2주를 공부하면서 겪었던 문제들과 해결방법들을 소개한다.

1. kubectl 오류

kubectl을 하면 호스트를 찾을 수 없다고 호스트나 포트가 정확한지 확인해보라는 메시지가 출력된 적이 있다.

그럴 땐 
* /etc/kubernetes/admin.conf 파일이 잘 작성되어있는지
* 이 파일이 $HOME/.kube/config 파일에 잘 copy 되었는지
* $HOME/.kube/config 파일에 대한 권한이 잘 설정되어있는지 (chmod, chown)
* SystemdCgroup 설정이 false로 되어있진 않은지

를 확인해보자 !

만약 디버깅(?)을 하기 위해 로그를 확인하고 싶은데, kubectl logs는 사용할 수가 없어서 고민이라면, /var/log/ 경로에서 kube-system pod들의 로그를 확인할 수 있다. k8s는 이 경로에 pod들의 로그를 모아둔다.

2. DNS서비스 오류

정확한 서비스 이름으로 해당 pod를 호출했는데 검색된 서비스가 없다는 오류가 생길 때가 있다. 

분명 coredns도 running status로 잘 작동하고있고, 서비스 이름으로 호출 시 dns서버의 ip로 잘 접근이 되는데도 서비스를 검색하지 못했다. 이 문제와 클러스터 직접 구성하는 문제로 가장 많이 고생했다.

원인은 나는 AWS 상에서 쿠버네티스를 설치했는데, 방화벽인 VPC 보안그룹의 인바운드 규칙으로 인해 통신이 안됐었다.

물론 포트들을 열어줘야한다는 건 기본으로 알고있었기 때문에 인바운드 규칙에서 모든 ip에 대한 모든 **TCP**를 허용하도록 설정해두었다. 그러나 이건 반만 알고있던 것이었다.

클러스터 내부에서 pod끼리 클러스터ip나 DNS서비스를 이용한 서비스이름으로 통신할 때는 TCP가 아닌 UDP를 사용한다. 때문에 **모든 TCP**에서 **모든 트래픽**으로 방화벽 설정을 변경해주고 나서야 정상적으로 DNS서비스를 이용할 수 있었다.

아웃바운드 규칙은 default로 모든 ip에 대한 모든 트래픽을 허용하기 때문에 변경하지 않아도 된다.

물론 인바운드 규칙을 모든 트래픽으로 설정하는 것은 보안의 위험이 있고, K8S에서 사용하는 포트들을 확인하고 해당 범위의 포트들만 적절히 열어주는 것이 좋다.
[이 글의 kubeadm port requirements 파트 참조](https://devopscube.com/setup-kubernetes-cluster-kubeadm/)

---

## 참조

[How to Install Containerd on Ubuntu 22.04 / Ubuntu 20.04](https://www.itzgeek.com/how-tos/linux/ubuntu-how-tos/install-containerd-on-ubuntu-22-04.html)


[How To Setup Kubernetes Cluster Using Kubeadm](https://devopscube.com/setup-kubernetes-cluster-kubeadm/)

[containerd를 런타임으로 사용한 Kubernetes 설치](https://docmoa.github.io/02-Private%20Platform/Kubernetes/02-Config/kubernetes_with_out_docker.html)

[github.com/saiyam1814/kubeadm + containerd 1.23](https://gist.github.com/saiyam1814/ba56e5fb09d712501d5a4b51e8ad85a5#file-kubeadm-containerd-1-23-L33)

[Install Docker Engine on Ubuntu](https://docs.docker.com/engine/install/ubuntu/#install-from-a-package)

[K8S 공식문서 : kubeadm 설치하기](https://kubernetes.io/ko/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)

<br>

이 중 쿠버네티스 클러스터 설치를 학습하면서

How to Install Containerd on Ubuntu 22.04 / Ubuntu 20.04

How To Setup Kubernetes Cluster Using Kubeadm

github.com/saiyam1814/kubeadm + containerd 1.23

이 세 글의 도움을 가장 많이 받고 참고했다.