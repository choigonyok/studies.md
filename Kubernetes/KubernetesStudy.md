## 쿠버네티스 기본 용어

---

## 개요

쿠버네티스는 컨테이너 오케스트레이션 플랫폼이다.

컨테이너 기반으로 배포되는 어플리케이션은 일반적으로 컨테이너 하나만으로는 배포가 불가능하다. 

컨테이너 오케스트레이션 플랫폼은 수많은 컨테이너들이 서로 통신하는 과정을 관리하고, 자원을 효율적으로 사용하고, 컨테이너의 상태를 모니터링하며, 컨테이너를 생성/변경/삭제하는 컨테이너 관련된 모든 기능들을 도와주는 플랫폼이다.

쿠버네티스의 기능들과 공부하면서 배운 내용들을 정리해본다.

--

어플리케이션이 오류가 생겨도 파드는 문제없이 실행될 수 있음
그래서 어플리케이션이 잘 작동하는지 헬스체크가 필요함
1. url이용 : pod 생성시 livenessProbe, readinessProbe 지정
   1. 컨테이너에 주기적으로 command 실행
헬스체크를 해야 문제가 생긴 파드를 재시작하고, 트래픽을 문제없는 다른 노드로 라우팅할 수가 있기때문에, 운영환경에서는 헬스체크를 꼭 포함시켜야함
livenessProbe : 컨테이너가 실행중인지를 확인, 실패하면 pod 재시작
readinessProbe : 컨테이너가 요청을 처리할 수 있는 상태인지를 확인, pod는 재시작하지 않고 검증이 실패하면 이 pod로 라우팅되지 않게하고, 다시 readiness 확인 후 확인되면 트래픽을 다시 받음
Probe는 컨테이너 단위로 진행되는데, Pod는 내부의 모든 컨테이너가 정상이어야 pod로 트래픽 라우팅을 할 수 있음
예를 들어, 한 파드에 컨테이너 A,B,C가 있는데, B가 livenessProbe에 실패하면 컨테이너B만 재시작하고 A,C는 그대로 있음
그래도 A,B,C는 한 파드에 있기 때문에 A,C는 정상상태이지만 트래픽을 라우팅받지는 못함





daemonset
노드당 하나씩 pod가 존재하게 하는 것
노드가 하나 추가되면 그 노드에 pod가 생성되고,
삭제되면 그 pod도 삭제됨. -> 다른 노드로 스케줄링 되지 않음
노드별 모니터링, 로그수집 등에 사용

request/limit
리소스 모니터링을 통해 일정 cpu 사용량이 넘거나 부족하면 자동으로 노드를 확장/축소하는 기능
1000m이 1코어, 200m은 0.2코어 사용을 의미
4코어 노드에서 리소스를 1000m으로 설정하는 건 전체 사용량의 1/4 = 25%를 사용한다는 뜻
request와 limit이 있음
request를 200m으로 설정했다는 것 = 이 파드는 최소 0.2코어는 있어야만 한다는 것
limit을 200m으로 설정했다는 것 = 이 파드에는 0.2코어까지만 리소스를 주겠다는 것
limit이랑 request 수치를 동일하게 하는게 좋다고 함 (https://www.slideshare.net/try_except_/optimizing-kubernetes-resource-requestslimits-for-costefficiency-and-latency-highload)
ㄴ limit이 너무 작으면 cpu 쓰로틀링이 생기긴하지만 request가 너무 커서 메모리가 뻑나는 것보단 낫기 때문
request/limit 체크는 default로 30초마다하고, 변경 가능

autoscaling
HorizontalPodAutoscaler kind로 수평확장 오토스케일러 설정 가능
기준점은 이런게 있음
targetCPUUtilizationPercentage: 이미 언급한 것처럼, CPU 사용률의 목표치를 설정하여 해당 목표치를 유지하도록 오토스케일링합니다.
targetMemoryUtilizationPercentage: 메모리 사용률의 목표치를 설정하여 해당 목표치를 유지하도록 오토스케일링합니다.
targetCustomMetricValue: 사용자 정의 메트릭 값을 기준으로 오토스케일링할 수 있습니다. 이를 사용하여 특정 지표나 애플리케이션의 특정 동작에 따라 스케일링할 수 있습니다.
targetRequestPerSecond: 초당 요청 수를 기준으로 오토스케일링할 수 있습니다. 주로 HTTP 요청이 많이 발생하는 웹 서비스에서 사용됩니다.
    **pod들의 평균 값을 기준으로 높아지면 평균을 낮추기 위해 확장시키고, 낮아지면 높이기위해 파드 수를 축소시키는 것!!!**

toleration
affinity와는 다르게, pod에 toleration이 **지정되어있으면** 일치하는 노드에만 배포될 수 있는것
예를들어, toleration이 설정되어있지 않은 pod1은 어느 노드에나 배치가 가능함.
근데 toleration이 설정되어있으면 해당 toleration 조건에 맞는 노드에만 배치가 가능함
node affinity는 pod의 affinity rule과 node의 affinity rule이 일치하는 경우에만 배치가능
toleration은 노드엔 taint pod에는 toleration을 설정해서 비교해야하고, node affinity는 pod에서만 지정하면 되는것임
NoSchedule : 스케줄링 시 일치하지않는 node에는 pod를 배포시키지 않음
PreferNoSchedule : 우선적으로는 일치하는 node에는 pod를 배포시키지 않음
NoExecute : toleration이 일치하지 않는 pod가 노드에 배치되면 설정한 time만큼은 노드에서 pod를 실행시키고, 그 이후에는 삭제시킴
noschedule과는 아예 올리지도 못하게 하는거랑, 올려두고 언제 삭제할지를 정하는 것의 차이!
  뭔가 잘못알고있는 것 같음. 다시 공부하기

etcd의 분리
etcd는 리소스 정보 등을 담고있는 데이터베이스
이게 중요하기 때문에, default는 마스터노드에 포함되어있지만
1. 마스터노드는 작업량이 커 cpu가 많이 필요한데 반해, etcd는 그렇지 않고, 
2. etcd는 중요해서 안정성을 높여야하기 때문에
etcd를 마스터노드에서 분리해서 각 etcd를 개별 노드에 배치하는 게 일반적.
보통 3개의 etcd를 쓰고, 많으면 5개?
etcd는 leader, follower, candidate로 나뉘는데
leader는 읽기/쓰기/follower에게 데이터 전달을 담당.
follower는 leader가 준 데이터 동기화/읽기 담당
candidate는 leader가 죽으면 leader역할을 수행
보통 etcd가 3개니까, 1번 : leader 2번 : follower+candidate 3번: follower 이렇게 구성
쓰는 건 leader만 쓸 수 있는데 읽을 땐 아무 etcd나 가서 읽어도됨. (leader, follower 다 읽기 기능을 지원하기 때문)

마스터노드 고가용성(HA : High Availability)
etcd의 안정성을 위해 3개 운영하는 고가용성처럼 마스터노드도 여러개 운영 가능
같은 ap-northeast-2여도
리전마다(ex: ap-northeast-2a, ap-northeast-2b, ap-northeast-2c) 마스터노드를 지정해두면
한 지역에서 장애가 생겨도 (자연재해나, 오류 등) 다른 리전의 마스터가 바톤터치해서 클러스터를 유지시킬 수 있음