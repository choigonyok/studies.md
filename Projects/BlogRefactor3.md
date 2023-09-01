# #9. 테라폼 인프라 프로비저닝 & 트러블슈팅
# project blog terraform

---

## 개요

운영환경용 도커파일은 작성했으니, 쿠버네티스를 올릴 인프라를 구성해야한다.

테라폼으로 블로그 서비스용 쿠버네티스 인프라를 구성하면서 공부한 내용들을 공유한다. 이 글에는 블로그 서비스의 인프라를 구축하기위해 테라폼으로 작성한 내용들과, 나의 얄팍한 네트워크 지식을 덧붙인 각 내용들의 설명이 포함되어있다.

---

## provider

테라폼에서 provider는 리소스를 제공하는 업체를 의미한다. 유명한 클라우드 컴퓨팅 서비스인 AWS, GCP, Azure부터 네이버클라우드도 오피셜 provider로 등록되어있다. 테라폼으로 인프라 코드를 짜면 테라폼이 코드를 바탕으로 해당 provider에 API요청을 보내 인프라가 실질적으로 구성되는 방식이다.

테라폼의 오피셜 레지스트리에서 수많은 오피셜 provider를 확인해볼 수 있다. provider는 오피셜, 파트너, 커뮤니티로 종류가 나뉜다. 오피셜은 테라폼에서 직접 관리하는 provider, 파트너는 해당 파트너 기업이 직접 관리하는 provider, 커뮤니티는 개인이나 단체 등이 관리하는 provider를 의미한다.

내가 선언한 provider 코드는 아래와 같다.

```
provider "aws" {
  region = "ap-northeast-2"
}
```

만약 provider가 official provider가 아닌 partner/community provider라면 위 문법은 적용되지 않는다.

```
terraform {
  required_provider "PROVIDER" {
    ...
  }
}
```

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

  provisioner "remote-exec" {
    inline = [
      "sh ./master.sh"
    ]
  }
}
```

ccs-master라는 변수명의 instance를 생성한다. 

### ami

ami는 Amazon Machine Image의 약자로, AWS에서 제공하는 다양한 가상 OS 이미지이다. 이 ami 설정에 따라 인스턴스에 설치될 OS가 결정된다. 이 ami는 종류와 버전이 너무 많아서, ami- 뒤의 id로 구분하는데, 어떤 id가 어떤 OS인지는 aws에서 확인이 가능하다. 내가 선택한 ami는 UBUNTU 22.04TLS이다.

### instance_type

인스턴스 타입 속성은 인스턴스의 리소스, 특히 CPU와 RAM을 결정한다. 프리티어에서는 t2.micro를 1년 무료로 사용할 수 있게 제공해주지만, 실 서비스를 배포하기에 t2.micro는 너무나도 작은 용량이다. 그래서 비용을 지불하고 t3.small 타입을 사용했다.

참고로 만약 t2.micro도 충분하다! 할지라도 테라폼에서는 t2 계열의 인스턴스 프로비저닝을 지원하지 않는다. t2는 아마존 초기 인스턴스 타입으로 depricated 상태이다. 그리고 t2와 t3는 비용이 같은데 비해 t3가 t2보다 RAM 용량이 두 배 크다. 

t2.micro도 간단한 SPA 배포정도 경험하기엔 나쁘지 않은 것 같다. 하지만 실서비스를 운영할 예정이라면 최소 t3 이상을 선택하는 것이 옳아보인다.

### vpc_security_group_ids

vpc에 연결해두었던 보안그룹을 인스턴스와도 연결해준다. 실제 콘솔에서 리소스를 생성할 때는 이 부분 떄문에 vpc와 instance의 생성 순서가 아주 중요하다.

인스턴스를 한 번 생성하면 ip는 바뀌지 않는다. publicIP야 종료 후 재시작할 때마다 바뀐다지만, privateIP는 바뀌지 않는다. vpc를 먼저 생성하고, 인스턴스를 생성할 때 그 vpc에 속해야한다는 걸 알려줘야 vpc 네트워크 대역에 맞는 ip가 할당될 수 있다.

다행인 것은, 테라폼에서는 이런 순서들은 알아서 다 처리해준다. instance 리소스 생성 코드에 [aws_security_group.cluster_sg.id] 이렇게 aws_security_group를 참조해야한다는 걸 명시해줬기 때문에 알아서 의존성 여부를 판단해서 리소스를 생성한다.

### subnet_id

vpc와 마찬가지로 vpc 안에 있는 여러 개(일 수 있는) 서브넷 중에 어디에 인스턴스를 포함시킬 건지 명시해줘야한다.

### key_name

인스턴스의 키페어를 지정한다. 이 인스턴스에 접속할 때 인증할 키 페어를 어떤 것으로 설정할지 지정하는 부분이다.

나는 기존 Choigonyok이라는 이름의 pem키를 생성해둔 상태였기 때문에 이렇게 지정했다. 

### tags

마스터노드에서도 tags를 통해 실제 인스턴스 이름을 지정한다.

### connection

커넥션은 뒷부분의 remote-exec을 실행하기 위해 필요한 부분이다. 미리 간단하게 말하자면, remote-exec는 인스턴스를 생성한 후에 그 인스턴스에 무언가 추가작업을 할 때 사용된다.

내가 원하는 건 인스턴스 생성 후에 인스턴스에 원격으로 접속해서 마스터 노드를 스크립트로 구성하는 것이다. 그러려면 원격으로 접속하기 위한 설정을 해줘야하는데 이 부분이 connection이다.

ssh로 접속할 건데, 접속할 때 사용되는 user name은 ubuntu이고(우분투에서 기본 사용자 이름은 ubuntu이다.) 접속하기 위한 pemKey는 로컬의 "../../../PEMKEY/Choigonyok.pem"에 위치해있으며, 접속하려는 호스트는 지금 connection이 위치한 이 인스턴스(self)의 publicIP이다! 라는 것을 의미한다.

### remote-exec

위에서 언급했듯이, remote-exec은 리소스 생성 이후 할 작업을 의미한다.

inline 안에 적은 문자열은, 인스턴스 생성 이후 connection대로 인스턴스에 접속한 이후 실행될 스크립트를 선언할 수 있다.

이 스크립트 안에는 마스터노드를 구성하는 내용이 작성되어있다. 이 내용은 이전 게시글에 작성해두었다. Kubeadmd으로 쿠버네티스 클러스터 구성하기에서 확인할 수 있다.

---

## Instance for WorkerNodes

이번엔 워커노드들을 위한 인스턴스이다. 아래 부분을 제외하고는 마스터노드와 동일한 설정을 해주면된다. 

테라폼에서 같은 리소스를 여러 개 만들 때는 count를 이용해서 간편하게 할 수 있다. 나의 경우에 마스터노드 1개와 워커노드 2개를 사용했다. count = 3을 통해 같은 인스턴스 3개를 만들지 않고, 각 리소스를 따로 생성했다. 이유가 뭘까?

쿠버네티스 클러스터를 설치할 때 마스터노드와 워커노드에 해줘야하는 구성 설정 스크립트가 다르기 때문에 마스터노드/워커노드 리소스를 따로 생성했다.

아래 부분만 마스터노드와 다르고, 이외 부분은 동일하게 설정해주었다.

```
instance_type = "t3.micro"
count = 2
tags = {
    Name = "worker_node2"
}
provisioner "remote-exec" {
  inline = [
    "sh ./worker.sh"
  ]
}
```

---

## Network LoadBalancer

```
resource "aws_lb" "nlb" {
  name               = "blog-nlb"
  internal           = false
  load_balancer_type = "network"
  subnet_mapping {
    subnet_id = aws_subnet.public_subnet.id  # VPC1의 서브넷 ID
  }
  enable_deletion_protection = false
}
```

### internal

internal은 로드밸런서의 체계를 설정하는 부분이다. 콘솔에서 로드밸런서를 생성하면, **내부**와 **인터넷 경계** 중 하나를 선택하는 체계 섹션이 있다. 

내부로 설정하게 되면 nlb가 트래픽을 라우팅할 네트워크 안에서 들어오는 요청만 라우팅하는 것이다. 예를 들어 하나의 어플리케이션을 구성하는 여러개의 쿠버네티스 클러스터가 있다고 가정하자. 각 클러스터들은 각각 다른 VPC에 속해있다. 만약 클러스터 간 통신이 필요하다면, NLB를 통해 서로 다른 네트워크(여기서는 VPC)에 속해있는 클러스터끼리 통신할 수 있게 되는 것이다. 대신 브라우저 등 외부에서는 접근이 불가능하다.

인터넷 경계로 설정하게 되면 외부에서 들어오는 요청을 라우팅하겠다는 것이다. 외부에서 들어오는 TCP/UDP 등의 요청을 받아서 내부 서비스로 라우팅한다. 예시로는 클라이언트가 브라우저를 통해서 클러스터에 배포되어있는 어플리케이션에 접근할 수 있게 해주는 것이다. 

이렇게 역할에 따라 체계를 설정할 수 있다. internal이 false로 설정되면 체계가 인터넷경계로, true로 설정되면 내부로 정해진다. 나의 경우에는 외부 로드밸런서 역할로 사용하는 것이기 때문에 예시 코드의 internal이 false로 설정되어있다. 참고로 internal 속성은 default가 false이기 때문에, false로 지정할 거라면 굳이 해당 코드를 작성하지 않아도 된다.

### load_balancer_type

load_balancer_type은 ELB의 타입을 설정한다. "network"로 설정하면 NLB, "application"으로 설정하면 ALB를 생성하게 된다.

### subnet_mapping

쿠버네티스 문서에는 서브넷 지정 예시가 아래와 같이 나와있다.

    subnets = [for subnet in aws_subnet.public : subnet.id]

직접 실행을 해보니 해당 코드는 작동하지 않는다. in을 인식하지 못하는 것 같다. 이 코드 대신에 

    subnet_mapping {
        subnet_id = aws_subnet.<SUBNET NAME>.id  # VPC1의 서브넷 ID
    }

이 코들르 사용하면 서브넷을 지정할 수 있다.

서브넷을 지정한다는 것은 라우팅할 서브넷을 지정하는 것이다. 서브넷으로 하나의 VPC 안에서 네트워크가 논리적으로 분리될 수 있고, 논리적이지만 분리된 것이기 때문에 원하는 특정 서브넷에만 라우팅을 할 수 있다. 나의 경우에는 서브넷 하나만 사용했지만, 만약 외부에서 접근해도 되는 서브넷과 보안상 접근하면 안되는 서브넷으로 나눠져있다면, 접근해도 되는 서브넷만 지정해서 라우팅하도록 할 수 있다.

### enable_deletion_protection

 enable_deletion_protection는 문자 그대로 삭제보호 기능을 사용할 것인지를 결정하면 된다. 외부 로드밸런서는 서비스와 클라이언트를 잇는 마지막 엔드포인트와도 같기 때문에, 이 로드밸런서가 사라지면 나머지 모든 게 다 갖추어져있어도 서비스에 접근할 수 없게 된다. 실서비스를 유지보수 하던 중 실수로 로드밸런서가 삭제되어서 모든 기능이 마비되는 건 너무 끔찍한 일일 것이다. 
 
 근데 나는 수많은 시행착오 과정 속에서 인프라 스트럭처를 파괴하고 생성하는 일을 꽤 많이 반복하게 되는데, 로드밸런서가 같이 깔끔하게 삭제되지 않는 것이 불편해서 false로 설정해두었다.

---

## LoadBalancer TargetGroup

```
resource "aws_lb_target_group" "tg" {
  name     = "blog-tg"
  port     = 80
  protocol = "TCP"
  target_type = "ip"
  vpc_id   = aws_vpc.mainvpc.id  
}
```

로드밸런서를 생성했다면, 로드밸런서가 트래픽을 전달할 대상그룹을 설정해야한다. 실제 콘솔에서는 대상그룹을 먼저 생성해야 로드밸런서를 생성할 수 있다.

port는 **대상그룹이 리스닝할 포트**라고 할 수 있다. 흐름은 아래와 같다.

    클라이언트 접근 -> 로드밸런서 리스너 -> 로드밸런서 라우팅 -> 대상그룹 리스너 -> 대상그룹 라우팅 -> 서비스로 전달

따라서 이 port는 listener의 리스너/라우팅 포트와 일치해야 로드밸런서로 온 트래픽이 대상그룹에 잘 전달될 수 있게된다. 

protocol도 마찬가지로 로드밸런서에서 정의한 프로토콜과 일치해야한다.

**target type**은 인스턴스, ip, lambda 함수, ALB 총 네 가지가 있다.

### Target Type : 인스턴스

VPC 내의 특정 인스턴스로 타겟을 설정한다.

NLB보다는 ALB에서 많이 사용될 것이다. 예시로 마이크로서비스가 있다고 가정하면, 이 인스턴스 target_type을 통해 여러 노드에 나눠져있는 백엔드 서비스마다 알맞은 트래픽을 전달할 수 있을 것이다.

인스턴스는 "instance"라고 타입을 표기하지 않고, target_type 속성이 없으면 디폴트로 인스턴스가 지정된다.

### Target Type : ip

내가 적용한 타겟타입인데, 특정 인스턴스가 아니라 VPC 전체로 타겟을 정하는 방식이다. 이건 NLB에서 많이 사용될 수 있는데, NLB에서 VPC로 전달한 트래픽을 VPC 내부의 L7 로드밸런서에서 받아서 필요한 클러스터 서비스로 트래픽을 뿌려줄 수 있을 것이다. 이름처럼 타겟은 ip를 통해 지정되는데, 이 ip는 해당 ip로만 트래픽을 보내는 것이 아니라, 해당 ip를 가지고있는 VPC 전체로 트래픽을 보낸다고 볼 수 있다.

### Target Type : lambda

AWS의 람다에 대해선 아직 지식이 거의 없지만, 람다함수를 사용해서 따로 서버를 구축하지 않고 서버리스로 어플리케이션을 배포하는 경우에 로드밸런서를 사용한다면 이 타입을 사용하면 될 것 같다.

### Target Type : ALB

ALB는 로드밸런서이지만 동시에 NLB의 대상 그룹이 될 수 있다. NLB를 거쳐서 ALB로 간 뒤 서비스로 라우팅 되는 방식이 될 수 있다. 나의 경우에는 L7 로드밸런서로 HAProxy를 클러스터 내부에서 사용하기로 했기 때문에, ALB없이 NLB만 사용했다. ALB, NLB 모두 유용하고 좋은 서비스지만 비용이 청구되기 때문에 최대한 유료 서비스를 안쓰는 방향으로 배포했다. 

마지막으로 로드밸런서가 올라갈 VPC를 지정해주면 대상그룹 설정이 완료된다.

---

## LoadBalancer Target Groupo Attachment

```
resource "aws_lb_target_group_attachment" "tg_ip" {
  target_group_arn = aws_lb_target_group.tg.arn
  target_id        = aws_instance.ccs-worker.private_ip
  port             = 32665
}
```

대상 그룹은 그냥 그룹을 만든 것이고, 이 그룹에 들어갈 대상을 넣어줘야한다. 이게 target_group_attachment 리소스의 역할이다.

attachment가 대상그룹과 연결되지 않으면, 이 대상이 어느 대상그룹에 attachment를 해야할지 알 수가 없다. 때문에 arn으로 대상그룹을 지정해준다.

target.id는 대상을 지정하는 부분인데, 대상그룹의 타입 4가지에 따라서 값을 다르게 넣어주어야한다. 나의 경우엔 아까 ip 타입으로 대상그룹을 만들어뒀기 때문에, 대상그룹에 넣을 인스턴스의 private_ip를 target_id로 지정해준다. public_ip는 인식하지 못하고, vpc는 private_ip 기반으로 작동하기 때문에 private_ip를 넣어줘야한다.

인스턴스가 vpc에 속하는 과정을 보면, vpc가 먼저 만들어지고 cidr_addr 블록이 생긴 뒤, 인스턴스를 생성할 때 vpc를 지정하면 해당 vpc의 네트워크 범위 안에 맞는 private_ip를 할당받은 인스턴스가 생성된다. 이게 vpc과 private_ip의 연관성이다.

만약 대상그룹 타입이 lambda function이라면 target.id로 람다함수의 arn을 지정해줘야하고, ALB 타입이라면 ALB의 arn을 지정해줘야한다. 인스턴스 타입이라면 인스턴스의 .id로 지정해주면 된다. 그러니까 곧 target.id의 id는 실제 id라기 보다는 구별할 수 있는 수단으로서의 추상적 id라고 볼 수 있다.

port는 트래픽이 라우팅될 port를 지정해주면 된다.

---

## LoadBalancer Listener

```
resource "aws_lb_listener" "nlb_listner" {
  load_balancer_arn = aws_lb.nlb.arn
  port              = "80"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.tg.arn
  }
}
```

리스너는 로드밸런서가 리스닝할 포트 설정을 하는 부분이다.

arn은 Amazon Resource Name의 약자로, 리소스들을 구별할 수 있게 해주는 AWS 제공 기능이다. 테라폼에서는 이 arn을 통해서 리소스 간 연결이 가능하다.

리스너는 로드밸런서와 대상그룹 사이에서 둘을 연결해주는 역할을 한다. 리스너에 로드밸런서 arn이 정의되지 않는다면, 이 리스너는 리스닝으로 받은 트래픽을 어떤 로드밸런서에게 주어야하는지 알 수가 없고, 리스너가 트래픽을 어느 대상 그룹으로 전달할지도 알 수 없다. 

port는 리스너가 리슨할 포트를 의미한다. 브라우저 등의 외부에서 HTTP 요청을 받아서 라우팅할 것이기 때문에 80으로 설정했다. 또 80으로 받아서 전달해야 NLB 뒷단에 있는 ALB나 HAProxy나 NginX 등의 L7 로드밸런서가 HTTP 프로토콜로 트래픽을 받아서 처리할 수 있다.

protocol은 NLB이기 때문에 HTTP/HTTPS로는 설정이 불가능하고, TCP나 UDP등의 프로토콜만 지정이 가능하다. IP와 포트기반으로만 라우팅하기 때문이다. 

### action

액션은 로드밸런서가 어떤 형태로 라우팅할 것인지를 설정하는 부분이다.

forward, redirect, fixed-response, authenticate-cognito, authenticate-oidc

forward : 하나 이상의 대상그룹으로 트래픽을 분산
fixed-response : HTTP 요청을 지정해서 라우팅할 때
redirect : 말 그대로 리다이렉팅

cognito는 보안을 위해 Amazon에서 제공하는 사용자 인증 서비스이고, OIDC는 사용자 인증 표준 프로토콜이다.
authenticate-cognito와 authenticate-oidc는 사용자 인증을 위한 타입이다. 클라이언트가 로드밸런서로 요청을 보내면 우선 요청을 cognito 또는 OIDC로 보내고, 인증이 완료되면 토큰을 발급받는데, 이 토큰이 있어야지만 서비스로 요청을 라우팅하는 방식을 사용한다.

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

Output은 리소스가 정상적으로 생성된 이후에 특정 값을 출력해주는 역할을 한다. 나는 각 인스턴스의 publicIP와 NLB의 DNS name이 출력되도록 해서, 굳이 콘솔로 가서 확인하는 불편함을 겪지 않고, 바로 확인할 수 있게했다.

```
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
```

---

## TroubleShooting

### RAM limit

처음 쿠버네티스로 서비스를 이전할 때, 마스터노드는 t3.small 타입 인스턴스로, 워커노드 2개는 t3.micro 타입 인스턴스로 배포했다. 

위에서 언급했던 것처럼 마스터/워커노드는 구성하기위한 스크립트가 달랐기 때문이기도 하고, 마스터노드는 기본적은 클러스터 구성 컴포넌트들이 많이 포함되기 때문에 워커노드들보다 상대적으로 큰 리소스를 가지는 타입으로 생성했다. 이전하기 전에도 t3.micro보다 더 작은 t2.micro에 모든 서비스가 배포되어서 간당간당하게 돌아가긴 했었기 때문에, 워커노드는 t3.micro 2개면 충분히 돌아가고도 여유롭겠다는 판단이 있었다.

그래서 워커노드는 t3.micro 타입 인스턴스로, count를 2로 설정해서 사용 중이었는데 문제가 발생했다. 백엔드, 프론트엔드, 데이터베이스 파드를 배치하고 나니 파드들이 계속 재시작을 반복하거나, 때로는 pending 상태에서 멈춰있는 상황이 발생했다. 원인을 찾기 위해서 모든 서비스와 디플로이먼트를 다 삭제한 후

    kubectl top nodes

커맨드로 노드들의 리소스 사용량을 확인해보니, 아무 서비스가 배포되지 않은 상태에서도 각 워커노드가 이미 RAM을 60% 이상 기본적으로 사용하고 있다는 것을 알게됐다. 쿠버네티스에서 워커노드로 구성하기 위해 기본적으로 설치하는 kubelet, kube-proxy가 큰 리소스를 차지하고 있는 것으로 판단되었다. 그래도 남은 리소스 내에서 서비스가 잘 돌아갈 수도 있는 것이니까 서비스를 재배포한 뒤에 또 리소스 사용량을 확인해보았다.

확인해보니 특히 데이터베이스 pod가 위치한 노드는 RAM 사용률이 95% 내외에서 움직이는 걸 확인할 수 있었다. 간당간당하게 노드가 버티다가 DB파드가 조금 더 일하는 상황이 생기자마자 노드가 못이기고 종료되어 재시작되고, 그와 연관되어서 DB와 통신이 불가해진 BE파드도 함께 재시작해버리는 일이 반복되는 것이었다.

### 해결

그래서 워커노드를 배포하는 인스턴스 타입을 기존 t3.mirco 2개에서 t3.micro 1개와 t3.small 1개로 변경했다. 데이터베이스가 위치한 파드만 유독 RAM을 많이 필요로 하기에, t3.small 워커노드는 데이터베이스 파드 전용 노드로 설정해서 데이터베이스 파드가 여유롭게 작동하도록 했고, 나머지 백엔드와 프론트엔드는 t3.micro 하나에 함께 배치되도록 했다.

이 기능을 구현하기 위해서 nodeSelector를 사용했다. 노드 셀렉터는 노드에 label을 설정해두고, 파드를 생성할 때 nodeSelector 속성을 선언해주어서 해당 노드에만 배치되도록 지정해줄 수 있다.

워커노드의 label은 마스터노드에서 아래 커맨드를 실행해서 설정하면 된다.

    kubectl label nodes <your-node-name> <label key>=<label-value>

잘 레이블이 설정됐는지 확인하고 싶으면, 

    kubectl get nodes --show-labels

이 명령어로 현재 노드들의 모든 레이블을 확인할 수 있다. 내 경우로 예시를 들자면, kubectl label nodes workerNode1 instance=small 이렇게 설정했다. 데이터베이스 deployment yaml 파일에서 노드 셀렉터는 아래와 같이 선언했다.
        
    nodeSelector:
      instance: small

이렇게 설정하고 배포한 후, 파드를 삭제하고 재생성되는 과정에서 데이터베이스 파드가 정상적으로 t3.small 타입의 워커노드에만 생성되는 것을 확인할 수 있었다.

그런데 문제는 아직 다 끝나지 않았다.

### Disk Pressure

그래도 백엔드 파드가 계속 오류로 종료되었다가 재생성되는 과정이 반복되었다. 원인을 찾기 위해 

    kubectl logs backend-pod

로 백엔드 파드의 로그를 확인해보려했지만 아무것도 출력되지 않았다. 파드가 생성되고 종료되는 과정을 보기 위해

    kubectl describe pod backend-pod

로 확인해봤다. 그러자 disk pressure로 인해 에러가 발생하는 것을 확인할 수 있었다. 

### 해결

AWS EC2 인스턴스의 디스크에 대해 이야기 하기 전에, 먼저 설명해야할 것이 있다.

아마존은 아마존 서비스가 미래에 계속해서 규모가 커질 것을 대비해 전세계적으로 서버팜을 엄청 많이 지어뒀다. 그런데 아직은 아마존에서 그걸 다 쓰고있지도 않고, 그렇다고 쓰지도 않는 서버들을 그냥 놀게하기가 너무 아까웠다. 유지하는데 드는 전기세만 해도 천문학적일 것이다. 그래서 가상화를 통해 서버를 여러 용량들로 나누어서 이걸 돈을 받고 외부 사용자들이 쓸 수 있게 하기 시작했다. 이것이 AWS의 시작이다.

결국 우리가 사용하는 이 AWS는 아마존에서 구축해둔 서버를 빌려다 쓰는 것이다. 그러니까 우리가 아는 일반적인 데스크탑처럼 CPU와 RAM와 디스크가 연결되어서 하나인게 아니라, 아마존 데이터센터의 CPU끼리 왕창 모아둔 곳에서 일부, RAM끼리 왕창 모아둔 곳에서 일부, 이런 식으로 가져다가 쓰게된다.

디스크도 마찬가지이다. 그래서 디스크를 원하는 만큼 사이즈를 조절할 수 있을 뿐만 아니라, 디스크 자체를 더 가져다가 붙일 수 있다. 30GiB SSD 하나를 쓸 수도 있지만, 15GiB SSD 두 개를 붙여서 쓸 수 있다는 말이다.

이 방식의 장점은 언제든지 특정 디스크를 연결/해제 시킬 수 있고, 다른 곳에다 갖다 붙일 수도 있게 된다는 것이다. 이게 EBS의 개념이고, 테라폼에서 EC2 인스턴스를 생성할 때 EBS를 따로 설정해두지 않으면 자동으로 8GiB의 gp2 타입이라는 ssd가 생성되어 인스턴스에 할당된다. 

default로 백엔드의 8GiB 할당되어있는 EBS의 옹량을 증가시키기 위해 t3.micro 인스턴스를 생성하는 테라폼 코드 안에 EBS 설정을 추가로 작성해주었다.

```
root_block_device {
    volume_size    = 14
    volume_type    = "gp2"
}
```

EBS가 무료는 아니고, 프리티어 계정은 달마다 최대 30GiB의 EBS를 지원하기 때문에, 마스터 노드 8, 데이터베이스 워커노드 8, 백/프론트엔드 워커노드 14로 설정해서 최대한 비용이 발생하지 않도록 구성했다.

~~이렇게 구성하고나니 모든 파드가 정상적으로 운영될 수 있었다.~~

> 이렇게 구성해도 해결이 완료되지 않는다. 문제점을 추가적으로 발견하고 해결방법을 추가적으로 **#13. AWS Root Volume in EC2** 게시글에 작성해두었다.
