# Helm

---

Helm은 쿠버네티스를 위한 패키지 매니저. 차트를 통해 소프트웨어를 쿠버네티스 리소스로 만들어준다.

설치, 업그레이드, 롤백, 삭제가 가능하다.

장점

1. yaml 파일은 정적인데, values.yaml파일을 통해 쿠버네티스 리소스를 동적으로 생성할 수 있다.
2. 코드로서 쿠버네티스 리소스의 형상이 관리되기 때문에, 형상관리가 쉽다.



--keep history
--reuse-values
--values ~
--dry-run : 실제 차트로 쿠버네티스 리소스를 생성하지는 않고, 생성하면 어떤 yaml파일을 통해 뭐가 만들어지는지, 오류는 없는지 알려준다. 실제 쿠버네티스 클러스터가 실행중이어야 확인이 가능하다.

helm template CHART
: 템플릿은 dry-run과는 다르게 유효성을 검사하지 않고 그냥 템플릿 파일로 만들어둔다. 클러스터가 작동중이지 않아도 되고, 이걸 미리 만들어뒀다가 나중에 사용하는 식으로 이용할 수 있음

helm get values CHART (--revision NUM)
: 차트에 추가한 values.yml의 내용 확인. helm status CHART와 비슷한데 status는 note 포함한 더 많은 정보를 보여줌. NUM은 버전 수 (1,2,3 ...)

helm get notes CHART
: CHART를 install했을 때 나오는 Notes 확인

helm get manifest CHART (--revision NUM)
: 실제 이 차트 install을 통해 쿠버네티스에 전달되는 리소스 생성을 위한 yml파일 확인

helm history CHART
: 이 차트가 언제 설치되고, 언제 업그레이드 됐는지 기록 확인. 설치/업그레이드 뿐만 아니라 업그레이드 중 오류가 발생했다면 오류도 기록하고 롤백하면 그것도 기록함, 이 기록은 secret에 저장되고, helm uninstall 을 하면 이 history로 삭제되기 때문에, 히스토리를 보존하고싶으면 uninstall에 --keep-history 옵션 지정해줘야함

helm rollback CHART NUM
: NUM은 revision number. 차트가 uninstall 상태여도 history가 남아있다면, 해당 히스토리로 롤백할 수 있음

helm install ~ --namespace NAMESPACE (--create-namespace)
: 차트를 특정 ns에 설치하고, 만약 ns가 없다면 ns생성까지 가능.

위와 비슷하게,
helm upgrade --install CHART
: 업그레이드하는데, 차트가 아예 없으면 생성을 하는. if문 느낌 

helm install CHART --wait (--timeout 0m0s)
: 쿠버네티스 리소스를 먼저 생성해서 running이 되면 차트를 생성. 만약 제한시간동안 쿠버네티스 리소스가 생성되지 않으면 에러리턴.
기본 타임아웃은 5분. 차트만 생성되고 리소스가 생성되지 않는 상황을 막을 수 있음

위와 비슷하게
helm install CHART --atomic (--timeout 0m0s)
: 쿠버네티스 리소스 구성 중 오류가 발생하면 이전 버전의 revision으로 자동 롤백시킴

helm install CHART --force
: 이전 상태와 그대로여도 전체를 삭제하고 다시 실행시킴

helm install CHART --cleanup-on-failure
: 생성중 실패한 오브젝트가 있으면 해당 오브젝트를 삭제. 차트 전체를 다 삭제하는 건 아님

helm create NAME 
: 커스텀 차트 생성

create을 하면 폴더가 하나 생성되고, 안에는 template(폴더), values.yml, charts.yml, 

template
: 여러 쿠버네티스 매니페스트(deployment.yml, service.yml...)과 helpers.tpl, NOTES.txt파일이 있음
쿠버네티스 오브젝트 얌 파일은 실제로 차트가 install될 때, 쿠버네티스 클러스터에 리소스를 생성하기 위해 전달될 yml파일들임.
근데 내부는 다 파라마티로 이루어져있는데, 이 값은 /Values.yml파일을 어떻게 개발자가 지정하느냐에 따라 커스텀할 수 있음. replicas라던가 name이라던가 volumes라던가 등등

NOTES.txt파일도 파라미터와 지정되어있는데, 이 파일에서 파드이름과, 컨테이너의 포트를 계산해서 포트포워딩을 하고, 이 NOTES.txt는 디폴트로 차트 install후에 콘솔창에 출력됨

helpers.tpl은 재사용 가능한 함수가 정의되어있는데, 모든 오브젝트 생성 yml파일에서 리소스의 이름을 받아오거나, labels를 받아오거나 하는 과정이 동일하기 때문에, 모든 오브젝트에서 재사용가능한 부분을 함수로 빼내서 따로 관리하는 것. 개발자는 건드릴 필요 없는 것 같다. 필요한 상황이 아니면!

values.yml은 이렇게 오브젝트를 생성할 때 개발자가 커스텀해야하는 부분인데, 
helm install CHART --set-~ 
을 통해서 values.yml파일을 수정하지 않고, 오버라이딩해서 커스텀시킬 수도 있고, values.yml파일을 직접 수정해서
helm install CHART --values (VALUES.YML FILE LOCATION)
으로 적용할 수도 있다. 근데 values.yml파일로 수정하는게 코드로 상태가 저장되니까 관리하기 쉽지 않을까 싶다.

values.yml에 있는 필드이지만 이해안가는 필드 모음
- image.pullPolicy : 
1. Always: 컨테이너 이미지를 항상 새로 가져옵니다. 이 설정은 새로운 이미지로 업데이트되는지 여부와 상관없이 항상 새로운 이미지를 가져옵니다.
2. IfNotPresent: 로컬에 이미지가 존재하지 않는 경우에만 가져옵니다. 이미지가 로컬에 있는 경우에는 가져오지 않습니다.
3. Never: 컨테이너 이미지를 가져오지 않습니다. 로컬 이미지만 사용합니다. 새로운 이미지를 가져오려고 할 때 에러가 발생합니다.

- imagePullSecrets: [] :
컨테이너 이미지가 비공개 컨테이너 이미지 레포지토리에 있을 때, 인증정보를 담고있는 secret 오브젝트의 이름을 정하면, 참고해서 이미지를 pull한다.

- nameOverride: "" :
차트의 이름을 덮어씌운다. helm ls에서 나오는 차트의 name을 변경하는 것.

- fullnameOverride: "" :
차트 안에있는 모든 리소스의 이름을 하나로 변경. 만약 abc로 변경하면, 나중에 차트 설치 후 클러스터에서 kubectl get deployment abc나, kubectl get svc abc이런식으로 리소스들의 이름이 하나로 통일되는 것.

- podSecurityContext: {}
  \# fsGroup: 2000
기본적으로 컨테이너는 각자 다른 컨텍스트를 부여받아서, 볼륨등의 파일 시스템 접근을 할 때, 해당 컨텍스트만 접근할 수 있게 함으로써 격리성을 유지시킨다.
안그래도 호스트 볼륨과 통신하는게 격리성이 떨어지는데, 만약 컨테이너1도 A볼륨에 접근가능하고, 컨테이너2도 A볼륨에 접근 가능하다면, 1이 변경한 A의 정보로 인해 2의 데이터가 변경되게 되는, 이런 격리성 떨어지는 상황을 막기위한 것이다.
근데 그걸 개발자가 알고, 그렇게 의도하고 싶다면, 이 podSecurityContext를 통해서 컨테이너1과 2를 같은 파일시스템그룹(fsGroup)에 할당함으로써 파일시스템에 1이 접근하든 2가 접근하든 접근이 가능해지도록 설정한다.
설정하지 않으면 처음 볼륨을 마운트한 컨테이너가 파일시스템권한을 가져가고, 나중에 접근하는 컨테이너는 권한이 없어서 접근이 불가능해진다.

근데 만약 임의로 fsGroup: 2000 설정했는데, 알고보니 다른 컨테이너가 먼저 쓰고있던 fsGroup이라면? 확률은 적지만 어쨌든 그 이름모를 컨테이너도 해당 파일시스템에 접근권한이 생기게 되는것이다.

그럼 어떻게 해야할까? 모든 컨테이너는 생성될 때 임의로 fsGroup을 지정해주면 된다. 그럼 각 컨테이너의 모든 fsGroup을 알고있으니, 파일시스템 접근을 공유하고 싶은 컨테이너끼리 현재 안쓰고있는 fsGroup number로 같게 지정해서 권한을 부여해주면 된다.

fsGroup의 범위는 리눅스에서는 0~65535 사이이다. 이 사이의 숫자로 설정하면 된다.

이 podSecurityContext는 헬름차트 values.yml설정시의 필드이고, 실제 쿠버네티스 매니페스트에서는 

spec:
 securityContext:
    fsGroup: NUMBER

이런식으로 설정할 수 있다.

---

helm package DIR
: helm create 로 생성한 DIR을 패키징한다. 커스터마이징이 끝나면 패키징을 해야 .tgz로 차트가 압축되어서 생긴다.

helm lint DIR
: 지금까지 작성된 차트 내용에 문법적 등 기타 오류가 없는지 검증한다. 없으면 0리턴, 있으면 0말고 다른거 리턴

