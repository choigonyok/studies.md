# #11. 클러스터 보안그룹 인바운드 룰 설정 및 IP-in-IP 프로토콜
# project blog kubernetes aws

---

## 개요

AWS에서는 가상의 방화벽인 Security Group(보안그룹) 기능을 제공한다. 인바운드/아웃바운드 규칙을 설정해서 VPC 또는 인스턴스 등의 다양한 AWS 리소스로 들어오거나 리소스에서 나가는 트래픽을 프로토콜, 포트, IP 기반으로 제어할 수 있는 가상의 방화벽이다.

기존 인바운드 룰은 테스트 편의성을 위해 모든 트래픽과 IPv4에 대해 접근 허용 상태로 되어있었다. 편리하지만 언제든지 공격자가 클러스터로 접근할 수 있게되기 때문에, 필요한 포트만을 open해서 사용하는 것이 보안에 좋다고 판단해서 이 부분을 수정하게 되었다.

이 글은 쿠버네티스 클러스터가 정상적으로 동작할 수 있게 하기위한 보안그룹 인바운드 규칙들과 그 과정에서 겪은 어마무시한 트러블슈팅 과정을 기록한 글이다.

---

## 백엔드 서비스 포트

Go 백엔드 포트 8080, React 프론트엔드 포트 3000, Mysql 데이터베이스 포트 3306을 열어준다.

---

## SSH 포트

마스터노드와 워커노드에 원격으로 접근이 가능하도록 하기 위해 SSH 포트인 22번을 열어준다. 대신 허용 IP는 보안을 위해 나의 로컬 IP로 설정했다.

---

## HTTP/HTTPS 포트

HTTP/HTTPS 접근을 허용하기 위해 80, 443 포트를 열었다.

---

## 컨트롤 플레인 및 워커노드 포트

쿠버네티스 공식 문서에서 제공하는 사용 포트들을 열었다.

---

## HAProxy 포트

로드밸런서인 HAProxy의 HTTP용 노드포트와 HTTPS용 노드포트를 열었다.

---

## DNS 포트

DNS서비스를 사용하기 위해 TCP 53, UDP 53, TCP 9153을 열었다.

---

## troubleShooting

위 항목들을 다 잘 적용했고 빠진 부분이 없게 잘 설정한 것 같은데, 어플리케이션이 동작하지 않았다. 정확히는 프론트엔드만 접근이 되고, 프론트엔드에서 백엔드로 요청을 보내는 과정에서 서버가 상태코드 500을 응답하는 문제가 발생했다. 

마스터노드에 원격으로 접근해서

    kubectl logs pods BACKEND-POD

커맨드로 백엔드 파드의 로그를 확인해봤다. 백엔드 로그에서는 데이터베이스 서비스와 연결이 되지 않는다는 오류를 확인할 수 있었다.

TCP 3306포트 규칙은 인바운드 규칙에 잘 설정해둔 상태였다. 원인을 알 수 없어 우선 모든 규칙을 삭제하고 **모든 TCP**, **모든 UDP** 두 규칙만 적용해보았다. 그래도 여전히 동작하지 않았다. **모든 트래픽**으로 설정하면 어플리케이션이 다시 잘 운영되는 것을 확인할 수 있었다.

mysql 포트는 TCP 3306이고, 모든 TCP 규칙에는 3306 포트도 포함이 될텐데 왜 **모든 트래픽** 규칙에는 작동하고 **모든 TCP** 규칙에는 작동하지 않는걸까?

모든 트래픽에서는 동작하고, 모든 TCP + 모든 UDP에는 동작하지 않는다면, TCP/UDP를 제외한 다른 프로토콜이 사용되는 것이겠구나 판단했다. AWS 보안그룹에서는 기본적으로 TCP, UDP, ICMP를 지원하고, 이외 다른 프로토콜들은 **사용자 지정 프로토콜 유형**으로 직접 지정해야한다.

Protocol number list를 검색해서 0번 프로토콜 HOPOPT부터 142번 프로토콜 ROHC까지 하나씩 다 확인해보기로 했다. 다행히 아주 앞쪽 4번 프로토콜인 IP-in-IP 프로토콜에서 어플리케이션이 동작하는 것을 확인했다.

그래서 위에서 언급한 각 규칙들을 다시 적용하고 추가로 사용자 지정 프로토콜 4번 규칙을 추가해주었다.

### 4번 프로토콜과 VPC의 관계

4번 프로토콜 IP-in-IP가 뭐길래 설정되지 않았다고 문제가 생긴걸까? 4번 프로토콜은 IPv4에서 캡슐화를 통해 서로 다른 네트워크 간 패킷을 주고받을 수 있게해주는 프로토콜이다.

이 프로토콜이 필요한 이유는 VPC에 있다. VPC는 AWS에서 제공하는 가상 네트워크이다. 

AWS에서 EC2 인스턴스1, 2를 생성했다고 가정해보자. 인스턴스 1과 2는 서로 물리적으로 떨어져있고, 실질적으로 다른 네트워크에 속해있을 가능성이 더 많다. 근데 같은 VPC를 지정해서 생성하면 마치 같은 네트워크에 속해있는 것처럼 외부와는 격리되고, 인스턴스 1과 2는 서로 별도 작업 없이 통신이 가능하게 된다.

어떻게 이게 가능할까? VPC의 원리가 뭘까? 어떻게 가상으로 네트워크를 설정할 수 있다는 걸까?

답이 바로 4번 프로토콜 IP-in-IP에 있다. 인스턴스 1이 2로 패킷을 보낼 때, IP-in-IP 프로토콜을 통해서 원래 보낼 패킷이 한 번 캡슐화되고 캡슐에 헤더가 추가적으로 붙는다. 이렇게 목적지에 도착한 패킷은 캡슐화가 풀리고 원래 보내려고 했던 패킷을 인스턴스 2에 전달할 수 있게 된다.

---

## 정리

트러블슈팅 관련해서, 해결하고보니 간단한 문제였지만 답을 알기까지 10시간 가까이 붙잡혀있었다. 이 넓고 넓은 인터넷 세상에서 아무도 IP-in-IP를 언급하는 사람은 없었다. 대부분 AWS 보안그룹과 쿠버네티스를 함께 사용하는 개발자들은 EKS로 클러스터를 구성하고, EKS에서 함께 제공되는 파드별, 노드별 보안그룹을 적용하는 식이었다. 

혹시라도 베어메탈 클러스터를 구성해서 AWS 인스턴스에 설치하고 보안그룹을 생성하던 중 나와 같은 어려움을 겪고있는 사람이 있었다면 이 글이 꼭 도움이 되었기를 바란다.

---

## 참고

https://kubernetes.io/ko/docs/reference/networking/ports-and-protocols/
https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml