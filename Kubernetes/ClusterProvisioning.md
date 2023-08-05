## k8s cluster configuration w/ kubeadm in Ubuntu

---

## 개요

github에 알아서 쿠버네티스 클러스터가 구성되도록 잘 만들어진 스크립트 파일이 많이 있는데,
직접 구성해보면서 어떤 과정을 거치는지 알고싶어서 kubeadm으로 클러스터 구성을 해보기로 함
우분투와 우분투 command도 익히고, 클러스터 구성 과정도 익히고, 네트워크 공부도 할 겸!
일주일동안 하루 10시간씩 클러스터 구성만 붙잡고 공부했다.

---

## 초기 구성

* 해당 파일(/etc/modules-load.d/k8s.conf)에 내용을 입력하는 것. overlay와 br_netfilter라는 커널모듈을 입력함
* overlay : 파일 시스템 드라이버, 격리된 환경을 제공, 컨테이너가 호스트와 격리되어서 컨테이너의 변화가 호스트에 변화를 주지 않도록 해주는 커널 모듈
* br_netfilter : 브릿지 관련 기능 제어 모듈, 데이터 패킷의 흐름을 제어, k8s에서 방화벽 등에 따른 패킷흐름을 제어하고, 브릿지에 연결된 인터페이스들을 하나의 네트워크로 인식하게 해줌
  cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
  overlay
  br_netfilter
  EOF

* 해당 모듈들을 modprobe command로 로드
  sudo modprobe overlay
  sudo modprobe br_netfilter

* 파일(/etc/sysctl.d/k8s.conf)에 내용 입력
* net.bridge.bridge-nf-call-iptables  = 1 => 네트워크 브릿지의 iptables 호출을 true 설정
* iptables = 클러스터 내부에서 패킷을 필터링시켜줌
* network bridge = 
* net.bridge.bridge-nf-call-ip6tables = 1 => ipv6 패킷에 대한 ip6tables 호출도 true 설정
* net.ipv4.ip_forward                 = 1 => 패킷 포워딩 활성화, 받은 패킷을 다른 노드로 전달할 수 있게하기 위해 필요
  cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
  net.bridge.bridge-nf-call-iptables  = 1 
  net.bridge.bridge-nf-call-ip6tables = 1
  net.ipv4.ip_forward                 = 1
  EOF

* 위에서 입력한 sysctl 파라미터들을 적용
  sudo sysctl --system

* 스왑(메모리 부족시 잘안쓰이고있는 메모리 작업을 디스크에 넣어두는 작업)되지 않도록 하기. 아래는 노드들이 스왑될 떄의 문제점
* 1. 성능 저하 : 스왑되면 디스크에 넣다 빼는 시간 생김
* 2. 격리 : 컨테이너는 호스트와 격리된 환경이어야하는데 디스크에 스왑하다가 호스트에 문제를 일으킬 수 있음
* 이런 이유로 스왑을 하기 보다는 메모리가 부족하면 수평적 확장(스케일아웃)을 하자
* sudo swapoff -a : 현재 스왑되어서 디스크에 있는 모든 작업을 다 메모리로 불러오고, 스왑불가능하도록 설정
* crontab -l : crontab(crontable : 특정 시간마다 주기적으로 실행되는 작업) 목록을 가져오기
* 2>/dev/null : 만약 crontab목록이 없어도 오류나서 스크립트가 중단되는 일이 없도록 설정
* echo "@reboot /sbin/swapoff -a" : 인스턴스가 부팅될 때마다 swapoff를 자동실행하도록 cron작업 설정
* crontab - : 적용
* true : 이미 스왑이 off 되어있더라도 오류나서 스크립트가 중단되는 일이 없도록 설정
  sudo swapoff -a
  (crontab -l 2>/dev/null; echo "@reboot /sbin/swapoff -a") | crontab - || true


---

## 컨테이너 런타임 containerd 설치

* 컨테이너 런타임 : 컨테이너를 구현하는 소프트웨어, docker가 가장 유명한데, k8s에서 docker engine 지원을 공식적으로 종료함
* k8s에서 containerd(컨테이너 런타임)은 컨테이너를 실질적으로 실행, 생성, 모니터링, 리소스할당, 네트워크 설정 등의 기능을 함

* https://github.com/containerd/containerd/releases
wget https://github.com/containerd/containerd/releases/download/v1.7.3/containerd-1.7.3-linux-amd64.tar.gz
sudo tar Czxvf /usr/local containerd-1.7.3-linux-amd64.tar.gz

wget https://raw.githubusercontent.com/containerd/containerd/main/containerd.service
sudo mv containerd.service /usr/lib/systemd/system/

sudo systemctl daemon-reload
sudo systemctl enable --now containerd

sudo systemctl status containerd

* https://github.com/opencontainers/runc/releases
wget https://github.com/opencontainers/runc/releases/download/v1.1.8/runc.amd64
sudo install -m 755 runc.amd64 /usr/local/sbin/runc

sudo mkdir -p /etc/containerd/
containerd config default | sudo tee /etc/containerd/config.toml

sudo sed -i 's/SystemdCgroup \= false/SystemdCgroup \= true/g' /etc/containerd/config.toml

sudo systemctl restart containerd

---

## kubectl, kubelet, kubeadm 설치

* sudo apt-get update : apt는 우분투에서 쓰이는 패키지 관리자인데, 설치등을 할 때 사용되는 도구, 현재 사용중인 패키지들을 최신화시킴
* sudo apt-get install -y apt-transport-https ca-certificates curl : 세 개의 패키지를 설치
* 1. apt-transport-https : 패키지를 다운로드할 때 https 사용할 수 있도록
* 2. ca-certificates : ssl/tls 검증 패키지
* 3. curl
sudo apt-get update && sudo apt-get install -y apt-transport-https ca-certificates curl

curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF
sudo apt-get update

* 최신버전으로 설치
  sudo apt-get install -y kubelet kubeadm kubectl

* 중간에 자동으로 kubectl, kubeadm, kubelet 패키지를 업데이트 하다가 호환성으로 문제 생기면 안되니까 업데이트 중지
  sudo apt-mark hold kubelet kubeadm kubectl containerd

---

## Initialize Kubeadm On Master Node To Setup Control Plane

* IPADDR : 마스터노드의 public ip를 클러스터가 알 수 있도록, echo ""는 뒤에 개행문자 입력
* NODENAME : 노드명으로 사용하기 위한 hostname 설정, -s는 호스트명 중 순수한 앞부분의 호스트명만을 불러오는 flag
* POD_CIDR : 마스터노드에서 사용할 pod들의 ip주소 블럭을 설정
  IPADDR=$(curl ifconfig.me && echo "")
  NODENAME=$(hostname -s)
  POD_CIDR="192.168.0.0/16"

* 설정한 환경변수대로 kubeadm을 init
  sudo kubeadm init --control-plane-endpoint=$IPADDR  --apiserver-cert-extra-sans=$IPADDR  --pod-network-cidr=$POD_CIDR --node-name $NODENAME --ignore-preflight-errors Swap

* init의 결과물로 나온 값 1 - kubeadm join ~
* 이 명령어를 worker node에 입력해줘야 해당 인스턴스를 클러스터의 node로 인식함

* k8s api서버와 통신할 수 있도록 kubectl을 사용하게 해주는 kubeconfig를 마스터노드에 생성
* sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config : config 파일을 복사해서 .kube/config에 복사 (여기 넣어줘야 kubectl이 클러스터 관리 권한을 얻음)
* sudo chown $(id -u):$(id -g) $HOME/.kube/config : .kube/config의 권한을 현재 사용자에게 부여
* $(id -u) : 현재 user의 id
* $(id -g) : 현재 user의 그룹 id
  sudo mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

> 여기까지 하고 "kubectl get pods -n kube-system" 으로 kube-system namespace의 pod를 확인하면 
> etcd, apiserver, controller, scheduler, proxy 총 5개의 pod가 생성되어야함
> 작동은 아직 안할 수 있음, error나 pending 상태여도 아직 네트워크 플러그인을 설치 안해서 그런거니 당황 X

---

## Install Calico Network Plugin for Pod Networking

kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/calico.yaml

> 이 pod를 apply하면 이전 5개의 kube pod들이 모두 running 상태로 바뀌고, coredns pod 2개와, calico pod 2개 해서 총 9개의 포드가 running 상태로 정상 작동해야함

---

## Setup Kubernetes Metrics Server

kubectl apply -f https://raw.githubusercontent.com/techiescamp/kubeadm-scripts/main/manifests/metrics-server.yaml

---

## 참조

https://www.itzgeek.com/how-tos/linux/ubuntu-how-tos/install-containerd-on-ubuntu-22-04.html
https://devopscube.com/setup-kubernetes-cluster-kubeadm/
https://docmoa.github.io/02-Private%20Platform/Kubernetes/02-Config/kubernetes_with_out_docker.html
https://gist.github.com/saiyam1814/ba56e5fb09d712501d5a4b51e8ad85a5#file-kubeadm-containerd-1-23-L33
https://docs.docker.com/engine/install/ubuntu/#install-from-a-package