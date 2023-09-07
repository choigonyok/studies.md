# AWS 젠킨스 서버 및 쿠버네티스 CI/CD 파이프라인 구축
# ci/cd ops

---

## 개요

혼자 직접 개발한 테크블로그를 운영하고있다. 실제 운영하면서 사용하다보니 사용하면서 생기는 크고 작은 추가적인 요구사항들이 생겼다. 최근에는 모바일웹 환경에서의 CSS와, ID/PW 변경 업데이트를 진행했다.

1. 모바일웹 환경에서의 hover

웹에서 게시글 카드 UI에 마우스 포인터를 가져다대면 확대가 되는 hover 효과를 주었다. 

![11](http://www.choigonyok.com/api/assets/71-1.gif)

웹에서는 정상적으로 동작하는데 반해, 모바일웹에서는 한 번 클릭을 하면 클릭으로 인식되지 않고 hover로 인식되는 오류가 발생했다. 그래서 이 hover가 모바일웹에서는 작동하지 않고, 웹에서만 작동하도록 반응형 css 코드를 수정했다.

```css
@media (min-width: 1100px) {
  .postcard:hover {
    transform: scale(1.1);
    transition: 300ms;
  }
}
```

2. ID/PW 공유저장소 노출

블로그의 ID와 PW가 env로 담긴 manifest가 공유저장소에 그대로 올라갔다. 때문에 시크릿을 따로 마스터노드에서 생성해서 ID/PW를 변경해줄 필요가 있었다.

이 요구사항들을 적용해서 재배포하기 위해 다음과 같은 과정을 거쳤다.

1. 코드 수정
2. 컨테이너 이미지 빌드
3. 컨테이너 레지스트리에 푸시
4. 마스터노드 원격 접속
5. manifest 재적용(apply)

만약 5번까지 완료했는데 1번에서 수정한 코드에 문제가 있었다면 다시 1번으로 돌아가 전 과정을 반복해야한다. 그리 복잡한 일은 아니지만 비효율적으로 같은 일을 반복하면서 시간을 낭비하고있었다.

블로그 운영은 앞으로도 몇 년간 쭉 유지할 어플리케이션이고, 그 긴 시간동안 직접 사용하며 더 많은 요구사항들과 리팩토링을 거치게 될텐데, 장기적인 관점으로 유지보수 편의성과 효율성을 높이기 위해 블로그 어플리케이션을 위한 CI/CD 파이프라인을 구축하게 되었다.

이 글은 젠킨스 클라우드 서버와 CI/CD 파이프라인을 블로그 어플리케이션에 구축하는 과정을 정리한 글이다.

---

## 젠킨스 클라우드 서버 구축

젠킨스 클라우드 서버를 구축했다. 현재 운영중인 블로그는 AWS t3 계열 EC2로 이루어진 쿠버네티스 클러스터 안에서 서비스 중이다. 이 말인 즉슨 AWS에서 프리티어 계정은 1년간 무료로 사용 가능한 t2.micro 인스턴스가 놀고있다는 의미이다.

젠킨스의 minimum resource requirements는 256MB 메모리와 1GB의 스토리지이다. 스토리지는 10GB정도를 추천한다고는 한다.

생각보다 jenkins의 요구 리소스 용량이 적어서, t2.micro를 Jenkins 전용 서버로 사용할 수 있지 않을까? 하는 생각을 했고, 구축해보았다.

### 젠킨스 서버용 인스턴스 생성 

![11](http://www.choigonyok.com/api/assets/71-2.png)

AWS에서 블로그 쿠버네티스 클러스터와 같은 VPC안에 새로운 서브넷을 생성해서 그 안에 인스턴스를 생성했다. 그리고 원격으로 접속하려고 보니

![11](http://www.choigonyok.com/api/assets/71-3.png)

퍼블릭IP가 할당되지 않았다. 젠킨스 인스턴스가 속한 서브넷이 private 서브넷이기 때문이다.

![11](http://www.choigonyok.com/api/assets/71-4.png)

서브넷의 설정을 변경해주면, 서브넷 안에 생성되는 인스턴스들도 publicIP를 할당받게 된다.

이미 생성된 인스턴스에는 IP를 할당할 수 없어서 인스턴스를 종료하게 재생성했다. 재생성 후 jenkins-server 인스턴스를 다시 확인해보니 퍼블릭IP가 잘 할당된 것을 확인할 수 있었다.

![11](http://www.choigonyok.com/api/assets/71-5.png)

물론 이 IP는 EIP가 아니라서 인스턴스가 재부팅될 때마다 새롭게 할당될 것이다.

VPC 속 서브넷 속에 젠킨스 인스턴스가 있다. 새로운 서브넷은 아직 private 서브넷이고, SSH등의 외부에서의 접근이 가능하게 하려면 인터넷 게이트웨이를 생성/연결해주어야한다.

![11](http://www.choigonyok.com/api/assets/71-6.png)

나의 경우에는 쿠버네티스 클러스터가 사용하는 인터넷 게이트웨이가 있어서 새로 생성한 jenkins-subnet을 인터넷 게이트웨이와 연결해주었다. 이제 이 subnet은 public subnet이 되었다.

마지막으로 젠킨스 서버가 속한 VPC의 보안그룹 인바운드 규칙에 SSH 접근을 아래와 같이 허용해주면 된다.

![11](http://www.choigonyok.com/api/assets/71-7.png)

그리고 터미널에서 SSH로 접속하면

![11](http://www.choigonyok.com/api/assets/71-8.png)

이렇게 원격 접속에 성공한다.

### 젠킨스 설치 및 실행

원격으로 접속한 우분투 서버에 젠킨스를 설치해준다. Jenkins는 Java로 개발된 오픈소스여서 젠킨스 설치 전 Java를 먼저 설치해준다.

```bash
sudo apt update
sudo apt install openjdk-17-jre
java -version
openjdk version "17.0.7" 2023-04-18
OpenJDK Runtime Environment (build 17.0.7+7-Debian-1deb11u1)
OpenJDK 64-Bit Server VM (build 17.0.7+7-Debian-1deb11u1, mixed mode, sharing)
```

그리고 Jenkins를 설치해준다.

```bash
curl -fsSL https://pkg.jenkins.io/debian-stable/jenkins.io-2023.key | sudo tee \
  /usr/share/keyrings/jenkins-keyring.asc > /dev/null
echo deb [signed-by=/usr/share/keyrings/jenkins-keyring.asc] \
  https://pkg.jenkins.io/debian-stable binary/ | sudo tee \
  /etc/apt/sources.list.d/jenkins.list > /dev/null
sudo apt-get update
sudo apt-get install jenkins
```

젠킨스가 어느 포트에서 실행되는지 확인하기 위해 네크워크 매핑 및 포트 정보를 출력하는 아래 리눅스 명령어를 실행해준다.

```bash
netstat -lntp
```

![11](http://www.choigonyok.com/api/assets/71-9.png)

이렇게 8080 포트에서 실행되는 걸 확인할 수 있다. 이제 브라우저에서 젠킨스 서버의 IP + 포트로 접속하면

![11](http://www.choigonyok.com/api/assets/71-10.png)

젠킨스가 실행되는 걸 볼 수 있다.

처음 접속하면 admin 계정을 생성해야한다. initialPassword는 이미지에 나온 것처럼, 

```
/var/lib/jenkins/secrets/initialAdminPassword
```

경로에서 확인할 수 있다. 리눅스 원격 접속 터미널에서

```bash
cat /var/lib/jenkins/secrets/initialAdminPassword
```

를 입력해준다. 그럼 긴 문자열 하나가 출력되는데 이걸 젠킨스에 입력해주면 된다.

![11](http://www.choigonyok.com/api/assets/71-11.png)

기본적인 젠킨스 의존성들이 설치되고나면, 

![11](http://www.choigonyok.com/api/assets/71-12.png)

기본 항목들을 입력해주면 젠킨스 Dashboard UI가 정상적으로 실행되는 것을 확인할 수 있다.

![11](http://www.choigonyok.com/api/assets/71-13.png)

### 서버 리소스 체크

```bash
top
```

top은 현재 리소스 사용량을 출력하는 리눅스 명령어이다. t2.micro가 하도 적은 스펙을 갖고있는지라, 실행시키자마자 리소스 사용량을 확인해보았다.

![11](http://www.choigonyok.com/api/assets/71-14.png)

전체 965.7MB의 메모리 중 84.9MB만 남았다. 거의 90%의 메모리를 사용하고있는 셈이다. 메모리 부족으로 언제 서버가 폭파될지 모르지만 우선 최대한 비용을 아끼기 위해 이대로 두기로 했다.

젠킨스의 메모리 사용량 peak/bottom 변화폭이 얼마나 되는지도 확인해 볼 수 있을 것 같다.

---

## 깃허브 웹훅

블로그의 레포지토리는 Github에 있다. 변경사항을 push하면 알아서 테스트하고 쿠버네티스에 배포할 수 있도록 블로그 CI/CD 파이프라인은 GitOps로 구현할 것이다. 

### 깃허브 토큰 발급

우선 젠킨스가 나의 깃허브 레포지토리에 대한 접근권한을 주기 위해 깃허브 토큰을 발급한다. 토큰은 자신의 깃허브 페이지에서 Settings - Developer settings - Personal access tokens - Tokens (classic)에 접속하면 생성할 수 있다.

![11](http://www.choigonyok.com/api/assets/71-15.png)

![11](http://www.choigonyok.com/api/assets/71-16.png)

Generate new token을 눌러주고, repo, admin:repo_hook를 선택한다. 이 토큰에 부여된 권한을 설정하는 것이다.

![11](http://www.choigonyok.com/api/assets/71-17.png)

생성하면 아래와 같은 키가 나오는데, 이 키는 한 번 브라우저를 떠나면 다시는 확인할 수 없기 때문에 지금 잘 복사해둔다.

![11](http://www.choigonyok.com/api/assets/71-18.png)

### 젠킨스 깃허브 연동

이제 젠킨스로 돌아가서, 왼쪽 탭에 위치한 Jenkins 관리 - System Configuration - System에 들어가서 Github - Github server를 추가해준다.

![11](http://www.choigonyok.com/api/assets/71-19.png)

API URL은 깃허브 퍼블릭 서버를 사용한다면 따로 건들지 않아도 된다. 이 부분은 깃허브 엔터프라이즈 등 사내 깃허브 서버를 구축해서 private하게 사용할 때 접근 URL을 입력하는 곳이다.

이제 아까 복사해둔 깃허브 접근 권한 토큰 키를 입력해주어야한다. Credentials에 Add를 눌러서 Kind를 Secret text로 설정하고, Secret에 토큰 키를 복사해서 넣어준다.

![11](http://www.choigonyok.com/api/assets/71-20.png)

### 깃허브 웹훅 생성 및 연결

그리고 저장하고 깃허브로 다시 돌아간다. 파이프라인을 구축하고자하는 레포지토리에 접속해서 Settings를 클릭한다. 이 Settings는 토큰을 세팅한 Settings와 다른 Settings이다.

![11](http://www.choigonyok.com/api/assets/71-21.png)

좌측 탭에 Webhooks에 들어가서 Add webhook을 누른다. 나는 이미 생성해두어서 웹훅이 하나 있지만, 비어있어도 괜찮다.

![11](http://www.choigonyok.com/api/assets/71-22.png)

위 이미지처럼, Content type은 application/json 형식으로 지정하고, 페이로드 URL은 젠킨스 서버의 URL을 입력해준다.

예를 들어 현재 젠킨스 대시보드 UI가 

```
http://127.0.0.1:8080
```

에서 실행된다면, 그걸 그대로 입력해주면된다. 그리고 웹훅을 생성한다.

좀 기다리면 웹훅 연결이 되었는지 안되었는지를 확인할 수 있다. 성공적으로 연결이 되면 아래처럼 체크표시가 나타난다.

![11](http://www.choigonyok.com/api/assets/71-23.png)

### 젠킨스 웹훅 아이템 생성

이제 마스터 브랜치에 푸시가 일어나면 젠킨스에서 웹훅을 감지하는 아이템을 생성하자. 젠킨스 좌측 탭의 새로운 Item에서 Freesytle project를 생성한다.

첫 General 필드의 Github project를 선택하고 레포지토리 URL을 입력한다. 

![11](http://www.choigonyok.com/api/assets/71-24.png)

소스 코드 관리 필드에서 git을 선택하고, 이번엔 레포지토리 URL 말고, git clone 시 쓰이는 URL을 입력해야한다. 이 URL은 아래와 같이 레포지토리에서 확인할 수 있다.

![11](http://www.choigonyok.com/api/assets/71-25.png)

Credentials도 설정해줘야하는데 Add를 눌러 Username with password Kind를 선택하고, 깃허브 아이디와 깃허브 비밀번호를 입력한다.

여기서 깃허브와 젠킨스의 연동 원리를 짚고 넘어가자.

---

## 젠킨스 & 깃허브 연동 원리

1. 젠킨스는 깃허브에서 발급한 토큰을 이용해 레포지토리 접근 권한을 얻는다.
2. 깃허브는 젠킨스가 설정한 브랜치에 코드가 푸시되면 웹훅으로 알림을 보낸다.
3. 젠킨스는 웹훅으로 알림을 받으면 Credentials로 깃허브 API에 접근해서 푸시된 코드를 가져온다.
4. 젠킨스는 이 가져온 코드로 개발자가 선언한 빌드를 실행한다.

---

Credentials를 저장하고, 설정해준다. 그리고 빌드의 트리거(빌드 유발)로는 아래와 같이 GitHub hook trigger for GITScm polling를 설정한다. 이렇게 저장해준다.

![11](http://www.choigonyok.com/api/assets/71-26.png)

---

## 젠킨스 파이프라인 구성

이제 젠킨스와 깃허브의 연동/권한 설정 및 웹훅 생성은 마쳤다. 내가 구성할 파이프라인의 워크플로우는 다음과 같다.

1. 깃허브 레포지토리에서 마스터 브랜치로 푸시 이벤트가 발생
2. 젠킨스가 레포지토리의 코드를 clone 
3. 코드를 테스트
4. 테스트 커버리지 확인
5. 컨테이너 이미지로 빌드
6. 이미지 레지스트리에 푸시
7. 쿠버네티스 클러스터에 접근
8. manifest를 apply해서 쿠버네티스에 배포

1, 2번은 윗 단계에서 웹훅 및 웹훅 감지 젠킨스 item을 생성함으로써 구현했다. 이제 젠킨스 좌측 탭의 새로운 item을 누르고 pipeline을 생성한다.

![11](http://www.choigonyok.com/api/assets/71-24.png)

이전과 마친가지로 첫 필드의 Github project를 선택하고 레포지토리 URL을 입력한다. 그리고 빌드의 트리거로는 이전 아이템 생성과 다르게, Build after other projects are built를 적용한다. 윗단계에서 생성한 jenkins item이 웹훅으로 코드 푸시를 감지하고 코드를 성공적으로 가져오면, 이어서 이 파이프라인이 실행되게 된다.

![11](http://www.choigonyok.com/api/assets/71-27.png)









---