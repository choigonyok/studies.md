# Terraform으로 AWS NLB 구성하기

데이터베이스 파드는 노드셀렉터로 t3.small 좀 큰데로 고정
-> t3.micro로 하고 kubectl top nodes와 kubectl get pods -o wide로 확인해보니 ram의 95%를 기본적으로 먹고 들어가고, 작업이 들어가면 뻑남
그리고 백엔드랑 프런트엔드는 물론 트래픽이 적어서겠지만 리소스 사용량이 적어서 t3.micro 하나에 넣기로 결정
넣는 방식은 데이터베이스와 안티어피니티 설정을 통해!
어차피 워커 노드가 총 2개라 데이터베이스가 있는 t3.small 아니면 마스터 노드 제외한 t3.micro 워커노드밖에 안남음
t2.micro는 프리티어인데 t3로 선택한 이유는
사실 t2 와 t3의 차이는 버전의 차이와 cpu가 t3가 두배 많다는 것, 근데 가격은 같음
그래도 프리티어니까 t2를 쓸 수도 있지만 t2는 어어엄청 옛날 버전이라 곧 지원 중단될 예정이고, 이로인해 테라폼에서 aws로 인프라 프로비저닝시 t2 계열 인스턴스들은 지원을 안함
그래서 t3 사용중

노드셀렉터 사용하려면 데이터베이스가 사용될 노드에 label을 지정해줘야 셀렉트를 할 수 있겠지?

    kubectl label nodes <노드명> instance=small

ip-10-0-1-57

---.elb.ap-northeast-2.amazonaws.com 으로 끝나는 nlb의 dns이름과
이 로드밸런서를 A타입으로 호스팅영역에서 라우팅한 도메인
프론트엔드에서 이 api요청을 어디로 하느냐에 따라 접근이 가능하거나 가능해지지 않는다.
원인예상
1. 둘은 다른 메커니즘
2. 백엔드에서 오리진으로 서비스디스커버리를 사용했기 때문에 요청하는 곳에만 가능한 것
3. 

---

## 개요

기존 블로그 프로젝트는 AWS 프리티어 인스턴스 하나에 DB, FE, BE가 전부 배포되어있었다. 리팩토링하며 쿠버네티스 클러스터에 서비스를 재배포하기로 했다.

Terraform으로 AWS 인프라 프로비저닝은 인스턴스, 보안그룹, VPC 정도만 경험해본 상태여서 로드밸런서, 대상그룹은 따로 콘솔에서 작업하고 있었고, 이에 따른 불편이 갈수록 크게 느껴졌다. 테라폼 리소스를 생성/삭제할 때마다 로드밸런서와 대상그룹을 생성/삭제주어야 했고, 로드밸런서는 프로비저닝된 특정 한 인스턴스와 연결되어있기 때문에 리소스 삭제 시 먼저 로드밸런서와 대상그룹을 삭제해주어야만 정상적으로 테라폼에서 destroy가 이루어지는 점도 상당히 불편했다. 

이런 이유로 AWS provider의 resource documentation을 참고해 쿠버네티스 클러스터 구성에 필요한 ELB를 테라폼으로 구성해보기로 했다.

이 글은 기본적인 테라폼 문법을 알고있다는 가정하에 작성된 글이다.

---

## aws_lb

AWS의 로드밸런서는 ALB와 NLB로 나뉘어진다. 더 넓게 보면 CLB도 있는데, 이건 이전 버전 로드밸런서라 이 글의 내용에서는 제외시켰다. 공식문서에는 ALB와 NLB끼리, CLB는 따로 설명이 되어있다.

ALB와 NLB의 차이는 IngressController.md 글에서 간단히 확인할 수 있다. NLB 생성 코드를 보자.

    resource "aws_lb" "<NLB VARIABLE NAME>" {
        name               = "<NLB NAME>"
        internal           = false
        load_balancer_type = "network"
        
        subnet_mapping {
            subnet_id = aws_subnet.<SUBNET NAME>.id  # VPC1의 서브넷 ID
        }

        enable_deletion_protection = false
    
        tags = {
            Environment = "production"
        }
    }

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

## listener

리스너는 로드밸런서가 리스닝할 포

    resource "aws_lb_listener" "<LISTENER VARIABLE NAME>" {
        load_balancer_arn = aws_lb.<LOADBALANCER VARIABLE NAME>.arn
        port              = "80"
        protocol          = "TCP"

        default_action {
            type             = "forward"
            target_group_arn = aws_lb_target_group.tg.arn
        }
    }

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

## target group

로드밸런서를 생성했다면, 로드밸런서가 트래픽을 전달할 대상그룹을 설정해야한다. 실제 콘솔에서는 대상그룹을 먼저 생성해야 로드밸런서를 생성할 수 있다. 예시 코드를 보자.

    resource "aws_lb_target_group" "<TARGET GROUP VARIABLE NAME>" {
        name     = "<TARGET GROUP NAME>"
        port     = 80
        protocol = "TCP"
        target_type = "ip" # 인스턴스면 타켓타입 미표시. 람다, alb면 각각 "lambda", "alb"로 타겟 타입을 선언해줘야함
        vpc_id   = aws_vpc.<VPC VARIABLE NAME>.id  
    }

port는 **대상그룹이 리스닝할 포트**라고 할 수 있다. 흐름은 아래와 같다.

    클라이언트 접근 -> 로드밸런서 리스너 -> 로드밸런서 라우팅 -> 대상그룹 리스너 -> 대상그룹 라우팅 -> 서비스로 전달

따라서 이 port는 listener의 리스너/라우팅 포트와 일치해야 로드밸런서로 온 트래픽이 대상그룹에 잘 전달될 수 있게된다. 

protocol도 마찬가지로 로드밸런서에서 정의한 프로토콜과 일치해야한다.

target type은 인스턴스, ip, lambda 함수, ALB 총 네 가지가 있다.

## 인스턴스

VPC 내의 특정 인스턴스로 타겟을 설정한다.

NLB보다는 ALB에서 많이 사용될 것이다. 예시로 마이크로서비스가 있다고 가정하면, 이 인스턴스 target_type을 통해 여러 노드에 나눠져있는 백엔드 서비스마다 알맞은 트래픽을 전달할 수 있을 것이다.

인스턴스는 "instance"라고 타입을 표기하지 않고, target_type 속성이 없으면 디폴트로 인스턴스가 지정된다.

## ip

내가 적용한 타겟타입인데, 특정 인스턴스가 아니라 VPC 전체로 타겟을 정하는 방식이다. 이건 NLB에서 많이 사용될 수 있는데, NLB에서 VPC로 전달한 트래픽을 VPC 내부의 L7 로드밸런서에서 받아서 필요한 클러스터 서비스로 트래픽을 뿌려줄 수 있을 것이다. 이름처럼 타겟은 ip를 통해 지정되는데, 이 ip는 해당 ip로만 트래픽을 보내는 것이 아니라, 해당 ip를 가지고있는 VPC 전체로 트래픽을 보낸다고 볼 수 있다.

## lambda 함수

AWS의 람다에 대해선 아직 지식이 거의 없지만, 람다함수를 사용해서 따로 서버를 구축하지 않고 서버리스로 어플리케이션을 배포하는 경우에 로드밸런서를 사용한다면 이 타입을 사용하면 될 것 같다.

## ALB

ALB는 로드밸런서이지만 동시에 NLB의 대상 그룹이 될 수 있다. NLB를 거쳐서 ALB로 간 뒤 서비스로 라우팅 되는 방식이 될 수 있다. 나의 경우에는 L7 로드밸런서로 HAProxy를 클러스터 내부에서 사용하기로 했기 때문에, ALB없이 NLB만 사용했다. ALB, NLB 모두 유용하고 좋은 서비스지만 비용이 청구되기 때문에 최대한 유료 서비스를 안쓰는 방향으로 배포했다. 

마지막으로 로드밸런서가 올라갈 VPC를 지정해주면 대상그룹 설정이 완료된다.

---

## target group attachment

대상 그룹은 그냥 그룹을 만든 것이고, 이 그룹에 들어갈 대상을 넣어줘야한다. 이게 target_group_attachment 리소스의 역할이다.

    resource "aws_lb_target_group_attachment" "<TARGET GROUP ATTACHMENT VARIABLE NAME>" {
        target_group_arn = aws_lb_target_group.<TARGET GROUP VARIABLE NAME>.arn
        target_id        = aws_instance.<INSTANCE VARIABLE NAME>.private_ip
        port             = <PORT NUM>
    }

attachment가 대상그룹과 연결되지 않으면, 이 대상이 어느 대상그룹에 attachment를 해야할지 알 수가 없다. 때문에 arn으로 대상그룹을 지정해준다.

target.id는 대상을 지정하는 부분인데, 대상그룹의 타입 4가지에 따라서 값을 다르게 넣어주어야한다. 나의 경우엔 아까 ip 타입으로 대상그룹을 만들어뒀기 때문에, 대상그룹에 넣을 인스턴스의 private_ip를 target_id로 지정해준다. public_ip는 인식하지 못하고, vpc는 private_ip 기반으로 작동하기 때문에 private_ip를 넣어줘야한다.

인스턴스가 vpc에 속하는 과정을 보면, vpc가 먼저 만들어지고 cidr_addr 블록이 생긴 뒤, 인스턴스를 생성할 때 vpc를 지정하면 해당 vpc의 네트워크 범위 안에 맞는 private_ip를 할당받은 인스턴스가 생성된다. 이게 vpc과 private_ip의 연관성이다.

만약 대상그룹 타입이 lambda function이라면 target.id로 람다함수의 arn을 지정해줘야하고, ALB 타입이라면 ALB의 arn을 지정해줘야한다. 인스턴스 타입이라면 인스턴스의 .id로 지정해주면 된다. 그러니까 곧 target.id의 id는 실제 id라기 보다는 구별할 수 있는 수단으로서의 추상적 id라고 볼 수 있다.

port는 트래픽이 라우팅될 port를 지정해주면 된다.

---

## 오개념

나는 이제껏 로드밸런서를 통해 트래픽이 전달되면 포트가 변한다고 생각하고 있었다. 대상그룹의 대상에서도 포트를 지정하면 80번 포트로 받은 요청이 

---

## 정리

테라폼에서는 코드를 다 짜두고 실행시키면 테라폼 내부적으로 의존성을 따져서 동기적으로 순서에 맞게 인프라를 생성/파괴하지만, 콘솔 상 흐름을 정리하자면 아래와 같다.

    대상그룹 생성 -> 대상그룹에 대상 추가  -> 로드밸런서 생성 -> 리스너 생성 -> 리스너에 대상그룹과 로드밸런서 연결

이번은 HAProxy를 사용해서 클러스터 내부 L7 로드밸런서를 구성했지만, 다음 번엔 ALB를 사용해서 로드밸런싱을 하는 과정도 소개해보겠다.

---

## 참고

[테라폼 공식문서](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb)

