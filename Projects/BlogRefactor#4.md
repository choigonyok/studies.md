# #9 Terraform InfraStructure Provisioning & Trouble Shooting
# project blog refactor terraform

---

## 개요

운영환경용 도커파일은 작성했으니, 쿠버네티스를 올릴 인프라를 구성해야한다.

테라폼으로 블로그 서비스 용 쿠버네티스 인프라를 구성하면서 공부한 내용들을 공유한다.

이 글에는 블로그 서비스의 인프라를 구축하기위해 테라폼으로 작성한 내용들과, 나의 얄팍한 네트워크 지식을 덧붙인 각 내용들의 설명이 포함되어있다.

---

## provider

provider는 리소스를 제공하는 업체이다.

유명한 클라우드 컴퓨팅 서비스인 AWS, GCP, Azure부터 네이버클라우드도 오피셜 provider로 등록되어있다. 테라폼으로 인프라 코드를 짜면 테라폼이 코드를 바탕으로 해당 provider에 API요청을 보내 인프라가 실질적으로 구성되는 방식이다.

테라폼의 오피셜 레지스트리에서 수많은 오피셜 provider를 확인해볼 수 있다. provider는 오피셜, 파트너, 커뮤니티로 종류가 나뉜다.

오피셜은 테라폼에서 직접 관리하는 provider, 파트너는 해당 파트너 기업이 직접 관리하는 provider, 커뮤니티는 개인이나 단체 등이 관리하는 provider를 의미한다.

내가 선언한 provider 코드는 아래와 같다.

```
provider "aws" {
  region = "ap-northeast-2"
}
```

만약 provider가 official provider가 아닌 partner/community provider라면 위 문법은 적용되지 않는다.

    terraform {
        required_provider "PROVIDER" {
            ...
        }
    }

으로 선언해주어야 한다.

---

## 네트워크 설정

### VPC

VPC는 aws에서 제공하는 가상의 네트워크이다. 같은 vpc안에 있는 인스턴스들은 같은 네트워크 대역의 ip를 할당받고, 같은 네트워크 내에서 통신할 수 있다.

같은 네트워크 내에서 통신할 수 있다는 말의 의미는 L3 스위치나 라우터없이 서로간의 통신이 가능하다는 이야기다.

```
resource "aws_vpc" "mainvpc" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name : "ccs-vpc"
  }
}
```

**mainvpc** 자리에는 각자 정한 이름을 넣어주면 된다. 이 이름은 실제 인프라가 프로비저닝 됐을 떄 인프라의 이름이 아니라, 테라폼(.tf파일) 내에서 사용될 변수명이라고 보면 된다.

cidr_block은 이 vpc가 가지는 네트워크 범위이다. 네트워크 범위의 첫 ip는 네트워크 주소로 쓰이기 때문에, 10.0.0.0/16은 10.0.0.0 ~ 10.0.255.255까지의 ip 대역을 의미하는 것이라고 할 수 있다.

실제 aws 상에서 vpc 이름으로 쓰일 부분은 tags를 통해 지정해줄 수 있다.

### Subnet

> 256 * 256 - 2(네트워크/호스트ip) = 65534

10.0.0.0/16 네트워크 범위 안에서 총 65534개의 ip를 사용할 수 있다. 효율적으로 ip를 사용하기위해 여러 서브넷을 생성해서 논리적으로 여러 개의 네트워크로 나눠 사용할 수 있다.

VPC안에 최소한 하나의 서브넷은 존재해야하고, 만약 ALB를 통해 로드밸런싱을 한다면 서브넷이 최소 두 개 이상 필요하다.

```
resource "aws_subnet" "public_subnet" {
  vpc_id     = aws_vpc.mainvpc.id
  tags = {
    Name : "ccs_subnet"
  }
  map_public_ip_on_launch = true
  cidr_block = "10.0.1.0/24"
}
```

public_subnet은 마찬가지로 지정할 수 있는 변수이다.

vpc_id 필드를 설정해주어서 이 서브넷 리소스가 어느 vpc의 서브넷인지를 알 수 있게 해준다. 테라폼에서는 이런식으로 .(dot)을 이용해서 리소스를 구분하는데, 보통

```
RESOURCE.NAME.ATTRIBUTE
```

이런 구조로 지정된다. aws_vpc.mainvpc.id는 mainvpc라는 이름의 aws_vpc리소스의 id라는 의미이다.

마찬가지로 이름은 tags를 통해 지정해줄 수 있다.

```
map_public_ip_on_launch = true
```

서브넷 유형을 public/private 중 선택한다. public은 외부에서 접근 가능한 서브넷이고 private은 반대이다.

이 서브넷에 배포될 블로그 서비스는 클라이언트가 외부에서 접근 가능해야하기에 public으로 설정해야한다. 그래서 설정이 true로 선언했다.

```
cidr_block = "10.0.1.0/24"
```

아까는 vpc의 네트워크범위를 지정했다면, 이번엔 그 안의 서브넷의 네트워크 범위를 설정한다. 여기에 모든 쿠버네티스 노드들이 배포될 예정이고, 이 서브넷의 네트워크 범위가 10.0.1.0/24이기 때문에, 모든 노드들의 privateIP는 10.0.1.1 ~ 10.0.1.254 사이에 할당될 것이다.

### Security Group

```
resource "aws_security_group" "cluster_sg" {
  vpc_id = aws_vpc.mainvpc.id
  name = "ccs_sg"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol  = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

cluster_sg라는 변수명을 가진 보안그룹을 생성한다.

이 보안그룹은 VPC와 연결된다. 외부에서 접근될 때, 이 보안그룹의 인바운드/아웃바운드 규칙에 따라 접근이 가능/불가능해지는 방화벽의 역할을 한다.

tags대신 name 속성을 통해서 실제 보안그룹 이름을 지정할 수 있다.

ingress는 인바운드 규칙이다. AWS에서 왜 인바운드 규칙이라는 용어를 쓰는지 잘 모르겠는데, 그냥 인그레스 룰이다.

from_port와 to_port는 포트의 범위를 지정한다. 포트가 하나라면 아래와 같이 설정할 수 있다.

```
port = 8080
```

protocol은 어떤 유형의 접근인지를 결정한다. -1은 모든 트래픽에 대한 것이고, "TCP", "UDP", "HTTP", "SSH" 이런 식으로 설정해줄 수 있다.

ingress/egress 내부의 cidr_blocks는 어떤 ip에서 오는 접근을 허용할 것인지에 대한 부분이다.

모든 클라이언트가 이 서비스에 접근할 수 있도록 하기위해 0.0.0.0/0으로 설정했다. 만약 "SSH" 접근을 내 로컬에서만 하고싶다면, 해당 블록을 ["192.1.1.1"] 이런식으로 설정해줄 수 있다.

egress는 아웃바운드 규칙이고, 내부 구성은 동일하다.

---

## Instance for MasterNode

마스터노드 용 인스턴스를 생성한다.

```
resource "aws_instance" "ccs-master" {
  ami           = "ami-0c9c942bd7bf113a2"
  instance_type = "t3.small"
  vpc_security_group_ids = [aws_security_group.cluster_sg.id]
  subnet_id = aws_subnet.public_subnet.id
  key_name = "Choigonyok"

  tags = {
    Name = "master_node"
  }

  connection {
    type = "ssh"
    user = "ubuntu"
    private_key = file("../../../PEMKEY/Choigonyok.pem")    
    host = self.public_ip
  }

  root_block_device {
    volume_size    = 8
    volume_type    = "gp2"
  }

  provisioner "remote-exec" {

    inline = [
      "sh ./master.sh"
    ]
  }
}
```

ccs-master라는 변수명의 instance를 생성한다. 

### ami

ami는 Amazon Machine Image의 약자로, AWS에서 제공하는 다양한 가상 OS 이미지이다. 이 ami 설정에 따라 인스턴스에 설치될 OS가 결정된다. 이 ami는 종류와 버전이 너무 많아서, ami- 뒤의 id로 구분하는데, 어떤 id가 어떤 OS인지는 aws에서 확인이 가능하다.

내가 선택한 ami는 UBUNTU 22.04TLS이다.

### instance_type

인스턴스 타입 속성은 인스턴스의 리소스, 특히 CPU와 RAM을 결정한다. 프리티어에서는 t2.micro를 1년 무료로 사용할 수 있게 제공해주지만, 실 서비스를 배포하기에 t2.micro는 너무나도 작은 용량이다. 그래서 비용을 지불하고 t3.small 타입을 사용했다.

참고로 만약 t2.micro도 충분하다! 할지라도 테라폼에서는 t2 계열의 인스턴스 프로비저닝을 지원하지 않는다. t2는 아마존 초기 인스턴스 타입으로 depricated 상태이다. 그리고 t2와 t3는 비용이 같은데 비해 t3가 t2보다 RAM 용량이 두 배 크다. 

t2.micro도 간단한 SPA 배포정도 경험하기엔 나쁘지 않은 것 같다. 하지만 실서비스를 운영할 예정이라면 최소 t3 이상을 선택하는 것이 옳아보인다.

### vpc_security_group_ids

vpc에 연결해두었던 보안그룹을 인스턴스와도 연결해준다. 실제 콘솔에서 리소스를 생성할 때는 이 부분 떄문에 vpc와 instance의 생성 순서가 아주 중요하다.

인스턴스를 한 번 생성하면 ip는 바뀌지 않는다. publicIP야 종료 후 재시작할 때마다 바뀐다지만, privateIP는 바뀌지 않는다. vpc를 먼저 생성하고, 인스턴스를 생성할 때 그 vpc에 속해야한다는 걸 알려줘야 vpc 네트워크 대역에 맞는 ip가 할당될 수 있다.

다행인 것은, 테라폼에서는 이런 순서들은 알아서 다 처리해준다. instance 리소스 생성 코드에 [aws_security_group.cluster_sg.id] 이렇게 aws_security_group를 참조해야한다는 걸 명시해줬기 때문에 알아서 의존성 여부를 판단해서 리소스를 생성한다.

## subnet_id

vpc와 마찬가지로 vpc 안에 있는 여러 개(일 수 있는) 서브넷 중에 어디에 인스턴스를 포함시킬 건지 명시해줘야한다.

## key_name

인스턴스의 키페어를 지정한다. 이 인스턴스에 접속할 때 인증할 키 페어를 어떤 것으로 설정할지 지정하는 부분이다.

나는 기존 Choigonyok이라는 이름의 pem키를 생성해둔 상태였기 때문에 이렇게 지정했다. 

## tags

마스터노드에서도 tags를 통해 실제 인스턴스 이름을 지정한다.

## root_block_device

EBS 스토리지 볼륨을 설정하는 부분이다. 

여기서 AWS의 인스턴스의 개념에 대해 짚고 넘어갈 필요가 있다. AWS 등의 클라우드 서비스의 시초는 이러하다. 아마존에서 미래를 대비해서 서버 센터를 엄청 많이 지어뒀는데, 아직 그걸 다 쓰고있지도 않고, 쓰지도 않는 걸 그냥 놀게하기가 너무 아까웠다. 그래서 가상화를 통해 서버를 여러 타입들로 나누면서 이걸 돈을 받고 외부 사용자들이 쓸 수 있게 하기 시작한 것이다. 

결국 이 AWS는 아마존의 서버를 빌려다 쓰는 것이다. 그러니까 우리가 아는 일반적인 데스크탑처럼 CPU와 RAM와 디스크가 연결되어서 하나인게 아니라, 아마존 데이터센터의 CPU끼리 왕창 모아둔 곳에서 일부, RAM끼리 왕창 모아둔 곳에서 일부, 이런 식으로 가져다가 쓰게된다.

디스크도 마찬가지이다. 그래서 디스크를 원하는 만큼 사이즈를 조절할 수 있을 뿐만 아니라, 디스크 자체를 더 가져다가 붙일 수 있다. 30GiB SSD 하나를 쓸 수도 있고, 15GiB SSD 두 개를 붙여서 쓸 수 있다는 말이다.

이 방식의 장점은 언제든지 디스크를 연결/해제 시킬 수 있고, 다른 곳에다 갖다 붙일 수도 있게 된다는 것이다. 이게 EBS의 개념이고, 기본적으로 인스턴스에는 8GiB의 EBS가 default로 붙는다.

이 EBS의 용량과 타입을 정하는 부분이 volume_size와 volume_type이다.

한 번 정한 EBS 볼륨 크기는 늘릴 순 있어도 줄일 순 없다. 물론 연결을 해제하고 작은 걸 생성해서 갖다붙이면 가능하긴하다.

### connection

커넥션은 뒷부분의 remote-exec을 실행하기 위해 필요한 부분이다. 미리 간단하게 말하자면, remote-exec는 인스턴스를 생성한 후에 그 인스턴스에 무언가 추가작업을 할 때 사용된다.

내가 원하는 건 인스턴스 생성 후에 인스턴스에 원격으로 접속해서 마스터 노드를 스크립트로 구성하는 것이다. 그러려면 원격으로 접속하기 위한 설정을 해줘야하는데 이 부분이 connection이다.

ssh로 접속할 건데, 접속할 때 사용되는 user name은 ubuntu이고(우분투에서 기본 사용자 이름은 ubuntu이다.) 접속하기 위한 pemKey는 로컬의 "../../../PEMKEY/Choigonyok.pem"에 위치해있으며, 접속하려는 호스트는 지금 connection이 위치한 이 인스턴스(self)의 publicIP이다! 라는 것을 의미한다.

## remote-exec

위에서 언급했듯이, remote-exec은 리소스 생성 이후 할 작업을 의미한다.

inline 안에 적은 문자열은, 인스턴스 생성 이후 connection대로 인스턴스에 접속한 이후 실행될 스크립트를 선언할 수 있다.

이 스크립트 안에는 마스터노드를 구성하는 내용이 작성되어있다. 이 내용은 이전 게시글에 작성해두었다. Kubeadmd으로 쿠버네티스 클러스터 구성하기에서 확인할 수 있다.

---

## Instance for WorkerNode1

이번엔 워커노드1를 위한 인스턴스이다. 테라폼에서 같은 리소스를 여러 개 만들 때는 count를 이용해서 간편하게 할 수 있다. 

그러나 마스터노드/워커노드 리소스를 따로 생성하는 이유는 쿠버네티스 클러스터를 설치할 때, 마스터노드와 워커노드에 해줘야하는 구성 설정 스크립트가 다르기 때문이다. 그리고 워커노드는 총 2개인데 각자 분리해서 리소스가 생성된다. 이유는 데이터베이스가 배포될 노드는 RAM과 디스크를 많이 잡아먹어서 t3.small로는 한계가 있기 때문이다.

처음엔 마스터노드는 t3.small, 워커노드 2개는 t3.micro로 구성해서 배포하려고했다. 마스터노드는 기본적으로 컨트롤 플레인을 구성하기 위한 컴포넌트들의 파드가 여럿 들어가기에 비교적 큰 리소스를 할당했다. 나머지 BE/FE/DB는 이전 배포에서 힘겹지만 t2.micro로도 세 서비스가 모두 동작했었기에, t3.micro는 t2.micro보다 RAM도 두 배 크니까, t3.small 2개면 충분하고도 남겠다는 생각이었다.

그러나 쿠버네티스에서 워커노드로 구성하기 위해 기본적으로 설치하는 kubelet, kube-proxy 등을 구성하고나서 kubectl top nodes로 확인해보면 이미 RAM을 60% 이상 기본적으로 사용하고 있는 것을 알아차렸다. 배포해보니 특히 데이터베이스 pod가 위치한 노드는 RAM 사용률이 95% 내외에서 움직였고, 버티다가 DB파드가 못이기고 재시작하고, DB와 연결된 BE파드도 함께 재시작해버리는 일이 반복됐다.

또 반대로 BE는 디스크가 부족해서 파드가 종료되는 일이 계속 발생했다. 사실 디스크떄문인지 모르고있다가, kubectl top nodes를 해도 BE가 배포된 노드는 리소스 사용량이 적은데, 뭐가 문제일까 고민하다가 kubectl describe pod로 상태를 확인했더니, disk pressure로 강제종료되었다는 오류 메시지를 확인하게 되었다.

나도 그 이유를 알고싶진 않았다..

데이터베이스를 위한 인스턴스 생성에서 스크립트를 제외한 나머지는 마스터 노드의 것과 동일하고, 아래 부분들만 다르다.

```
instance_type = "t3.small"
root_block_device {
    volume_size    = 8
    volume_type    = "gp2"
}
tags = {
    Name = "worker_node1"
}
```

```
instance_type = "t3.micro"
root_block_device {
    volume_size    = 16
    volume_type    = "gp2"
}
tags = {
    Name = "worker_node2"
}
```

---

## Network LoadBalancer


resource "aws_lb" "nlb" {
  name               = "blog-nlb"
  internal           = false // # 체계 설정 내부로 할지 인터넷경계로 할지
  load_balancer_type = "network" # for NLB or "application" for ALB
  
  subnet_mapping {
    subnet_id = aws_subnet.public_subnet.id  # VPC1의 서브넷 ID
  }

  enable_deletion_protection = false # true이면 terraform이 LB 삭제하는 걸 막아줌, 디폴트가 false라 false면 굳이 안써도 되긴 함
  
  tags = {
    Environment = "production"
  }
}

resource "aws_lb_target_group" "tg" {
  name     = "blog-tg"
  port     = 80
  protocol = "TCP"
  target_type = "ip" # 인스턴스면 타켓타입 미표시. 람다, alb면 각각 "lambda", "alb"로 타겟 타입을 선언해줘야함
  vpc_id   = aws_vpc.mainvpc.id  
}

resource "aws_lb_target_group_attachment" "tg_ip" {
  target_group_arn = aws_lb_target_group.tg.arn
  target_id        = aws_instance.ccs-worker.private_ip
  port             = 32665
}


resource "aws_lb_listener" "nlb_listner" {
  load_balancer_arn = aws_lb.nlb.arn
  port              = "80"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.tg.arn
  }
}

---

## Gateway

resource "aws_internet_gateway" "IGW" {
    vpc_id =  aws_vpc.mainvpc.id
}

resource "aws_route_table" "PublicRT" {
    vpc_id =  aws_vpc.mainvpc.id
    route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.IGW.id
    }
}

resource "aws_route_table_association" "PublicRTassociation" {
    subnet_id = aws_subnet.public_subnet.id
    route_table_id = aws_route_table.PublicRT.id
}

---

## Output


output "master-ip" {
  value = "${aws_instance.ccs-master.public_ip}"
}

output "worker-database-ip" {
  value = "${aws_instance.ccs-worker-database.public_ip}"
}

output "worker-ip" {
  value = "${aws_instance.ccs-worker.public_ip}"
}

output "lb-dnsname" {
  value = "${aws_lb.nlb.dns_name}"
}