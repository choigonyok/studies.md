Istio는 서비스메시

이상적인 istio 아키텍처는 istio가 서비스에 아무 영향을 끼치지 않는 것이다.
마치 스위치를 켜고 끄듯이, 키든 끄든 오류 없이 서비스에는 변동사항이 없어야한다.

하나의 파드 안에는 일반적으로 하나의 컨테이너가 배포되는데, istio를 사용하면 각 파드에 proxy를 넣어서 컨테이너간 통신을 하면 이 프록시를 통해 전달되도록 한다.

이 프록시는 자체적으로 로직을 수행할 수 있는데, 이 프록시는 로직을 통해 istio-system이라는 namespace의 istiod라는 이스티오 데몬과 통신하며 istiod에 로그/메트릭을 기록하거나 할 수 있다.
istio의 모든 기능은 이 istiod안에 들어있고, 모든 파드에 들어있는 proxy가 이 istiod와 통신하며 기능을 한다.

참고로 istiod는 pod이다.


istio = 컨트롤플레인/데이터플레인

컨트롤 플레인 : containerd 등 모든 기능 파드 모음 논리적 집합

데이터플레인 : 각 파드에 들어가있는 사이드가 프록시의 논리적 집합

사이드카/Envoy/프록시 다 같은 의미로 사용됨

컨테이너에 프록시를 주입할 때 클러스터 구성 파일에 코드를 추가해도 되지만 그럼 분리성이 떨어지니까
sidecar injection기능을 이용해 프록시 컨테이너를 파드에 주입할 수 있음
이 sidecar injection은 네임스페이스 단위로 이루어짐, 그러려면 namespace에 label을 정해줘야함
예전에 노드에 레이블 정해서 노드셀렉터 이용한 적 있는데 그런 것처럼

    kubeclt label namespace default istio-injection=enabled

default namespace에 istio-injection=enabled라는 레이블을 붙여줌

hook도 사용해서 차트가 클러스터 오브젝트가 되는 과정중에 무언가를 수행할 수 있음
hookpod.yml파일을 통해
과정은
pre-install
post-install
pre-delete
post-delete
pre-upgrade
post-upgrade
pre-rollback
post-rollback
test
중에 훅 사용 가능

훅이 여러개면 weight를 두어서 뭐가 먼저 실행될지 결정할 수 있음

훅과 릴리즈의 관계도 설정 가능
1. before-hook-creation : 훅이 생성되기 전에 이미 릴리즈되어있는 훅이 있으면 삭제. 디폴트임 얘가
2. hook-successed : 훅이 성공한 이후에 훅을 지운다
3. hook-failed : 훅이 실패하면 훅을 지운다