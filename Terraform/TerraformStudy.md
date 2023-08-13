

Terraform

Terraform은 **인프라스트럭처 프로비저닝 툴**이다.

공식 홈페이지에선 Terraform을 **Automate infrastructure on any cloud with Terraform** 라고 소개한다.

어떤 클라우드든 Terraform을 통해 인프라를 자동화할 수 있다는 것이다.

provisioning의 사전적 정의는 공급이다.

Ansible의 벤더인 RedHat에서는 Cloud Provisioning을 다음과 같이 정의한다.

> Cloud provisioning includes creating the underlying infrastructure for an organization’s cloud environment, like installing networking elements, services, and more.

클라우드 환경을 위한 인프라 구성을 high level에서 도와주는 툴이 infrastructure provisioning tool이라고 할 수 있겠다.

---

terraform은 코드로 인프라를 구축할 수 있게 해주는 IaC(Infrastructure as Code)의 대표주자이다.

Terraform은 공식 홈페이지에서 바이너리 파일을 다운받아 로컬에 설치하거나, Mac OS, Linux의 경우엔 Terminal을 통해 설치할 수 있다.

설치하면 각각 사용하는 IDE에서 .tf extension의 파일을 생성하면 Terraform을 구성할 수 있다.

Terraform은 볼륨과 CPU 사용량이 상당히 크다.

---

Terraform엔 provider라는 개념이 있다. 

provider는 리소스를 제공하는 업체이다. 유명한 클라우드 컴퓨팅 서비스인 AWS, GCP, Azure부터 우리나라의 네이버클라우드도 official provider로 등록되어있다.

Terraform로 인프라 코드를 짜면 Terraform이 코드를 바탕으로 해당 provider에 API요청을 보내 인프라가 실질적으로 구성되는 방식이다.

Terraform official homepage의 registry에 들어가보면 수많은 provider를 확인해볼 수 있다. provider는 official, partner, community로 나뉘어진다.

official은 Terraform에서 직접 관리하는 provider, partner는 해당 파트너가 직접 관리하는 provider, ㅊ는 개인이나 단체 등이 관리하는 provider를 의미한다.

provider를 tf파일에 선언하는 문법은 다음과 같다.

    provider "PROVIDER NAME" {
        ...
    }

만약 provider가 official provider가 아닌 partner/community provider라면 위 문법은 적용되지 않는다.

    terraform {
        required_provider "PROVIDER NAME" {
            ...
        }
    }

으로 선언해주어야 한다.

---

기본적으로 Terraform은 init -> plan -> apply 의 순서로 작업이 이루어진다.

tf파일에 provider를 지정하고

    terraform init

command를 입력하면 Terraform은 해당 디렉토리에 있는 모든 tf파일의 provider에 인프라 구성 작업을 실행하기 위한 플러그인들을 로컬의 .terraform 디렉토리에 다운로드 받는다.

    terraform plan

을 입력하면 지금 사용자가 짠 코드가 오류는 없는지, 실행되면 어떤 리소스가 생성/수정/삭제되는지에 대한 개요를 보여준다. 내용을 확인해보고 그대로 적용하기를 원하면

    terraform apply

command로 코드를 적용하고, 실질적인 리소스 생성/수정/삭제가 이루어진다.

---

apply로 생성된 인스턴스를 삭제하는 방식은 두 가지가 있다.

1. 코드 수정 후 apply

terraform은 리소스의 상태를 **현상태**로 유지한다. 예를 들어 AWS의 EC2 instance를 하나 생성하는 tf 코드를 작성했다고 가정하자. 이 상황에서 terraform apply를 3번 반복하면 어떻게 될까? 인스턴스가 3개 생성될까? 답은 **그렇지 않다.** 앞서 기술하였듯 terraform은 리소스의 상태를 코드의 현상태(Desired State)로 유지하기 때문이다. 그렇기에 이미 해당 코드로 인스턴스가 하나 생성되어있는 상태에선 다시 apply 한다고 해서 추가적으로 인스턴스가 생성되진 않는다. 추가적으로 인스턴스를 생성하려면 리소스 블록을 하나 더 추가해야한다.

그럼 인스턴스를 생성하는 리소스 코드 블럭을 삭제하고 apply를 한다면, terraform은 현상태를 적용하기 위해 생성되어있던 인스턴스를 destroy할 것이다.

2. terraform destroy

1번 방식처럼 하게되면 두 가지 문제가 발생한다. 첫쨰로 매번 코드를 일일이 수정해야한다는 점과, 둘쨰로 삭제된 코드를 되돌리려면(인스턴스를 재생성하려면) 추가적인 비용이 든다는 것이다.

Terraform은 destroy command를 지원한다. 생성된 리소스들을 destroy하게 해주는 명령어이다.

    terraform destroy

이 command는 현재 디렉토리의 모든 리소스들을 destroy한다.

만약 한 디렉토리에 여러 tf파일과 리소스들이 선언되어있고, 그 중 특정 리소스만 destroy하고 싶다면 어떻게 해야할까?

    terraform destroy -target RESOURCE NAME.LOCAL NAME

이 명령어를 통해 원하는 리소스만 파괴시킬 수 있다.

예시로 AWS의 EC2 인스턴스를 생성하는 리소스 블럭을 보자.

    resource "aws_instance" "chat-service" {
        ami           = "ami-0c9c942bd7bf113a2"
        instance_type = "t2.micro"

        tags = {
            Name = "ChatService"
        }
    }

리소스 블럭의 첫 string은 resource name이다. provider로 aws를 선언해두었기 때문에, instance를 생성하려면 정의된 aws_instance를 이용해야한다. 두번쨰 string인 "chat-service"는 local name이다. terraform 내부적으로 리소스들을 구별하기 위한 id이다.

다른 리소스들은 두고 이 리소스만 파괴하고 싶다면,

    terraform destroy -target aws_instance.chat-service

이런식으로 target flag를 이용해 리소스를 파괴시킬 수 있다.

---

앞서서 terraform은 tf파일을 실행할 때마다 반복해서 리소스를 생성하지 않는다고 말했다. 이건 terraform state를 통해 가능한 것이다.

terraform apply를 처음 하게되면 디렉토리에 .tfstate의 파일이 하나 생성된다.

![img](2.png)

이 파일 내부엔 현재 terraform이 알고있는 리소스의 상태들이 적혀있다.

terraform destroy를 실행한 직후라면 비어있을 것이고, 리소스가 생성되었다면 생성된 리소스에 대한 정보가 담겨있다. 이런식으로 terraform은 자체적으로 변화를 기록하기 때문에, 반복해서 apply를 실행해도 "아, tfstate파일을 보니까 이 리소스는 이미 생성되어있구나" 하고 추가적으로 생성하지 않는 것이다.

때문에 이 tfstate파일을 건들지 않는 것이 좋다고 한다. 건들다 리소스 정보를 지워버리면 이미 있는 줄 모르고 또 리소스를 생성하게 되고, 사용자는 알지못한채 버려진 리소스가 생길 수 있기 때문이다.

---

current/desired state

tf파일에 적혀있는 상태(사용자가 원하는 결과값의 상태)를 desired state, 실제 클라우드 상에 적용되어있는 리소스의 상태를 current state라고 한다.

둘은 일치할 수도 아닐 수도있다. 지금 막 문제없이 apply를 실행한 직후라면 둘은 일치할 것이다. 그러나 이후에 사용자가 수동으로(manually) 리소스에 변화를 준다면 tf파일은 변화된 상태를 알지 못하기 때문에 일치하지 않게 된다.
둘을 일치시키기 위해서

    terraform refresh

명령어를 사용할 수 있다. 이 명령어는 current state를 tfstate파일에 최신화시켜준다.

근데 이 명령어를 함부로 사용하는 것은 위험하다고 한다. 우선 refresh는 terraform plan시에 자동으로 실행되어서 굳이 따로 쓰지 않아도 되기 때문이고, 둘째로는 refresh를 잘못 실행하게되면 모든 state의 정보가 날아갈 수 있기 때문이다.

예를 들어보겠다.

AWS에서 region을 A로 선언하고, 인스턴스 하나를 생성하도록 선언했다고 치자. 그리고 apply를 하면 tfstate는 region A에 인스턴스가 하나 생성된 것으로 기록하고있다. 여기서 region을 B로 변경하고 refresh를 하면 terraform은 tf파일의 코드를 보고 해당 인스턴스를 찾는다.

하지만 우리는 A region의 인스턴스만 생성한 상태이기에 B region에는 아무 인스턴스도 존재하지 않는다.

terraform은 B region에 리소스가 없는 걸 확인하고 tfstate를 빈 상태로 초기화한다.

이후 사용자가 A region으로 변경해서 apply를 다시 실행한다.

terraform은 현재 tfstate가 비어있는 상태이기때문에 region A에 인스턴스를 추가로 생성한다.

이렇게 되면 region을 변경하기 이전에 A region에 생성했던 인스턴스는 잃어버리게 된 것이다. 아무도 사용하지 않지만 가동중인 상태로 그렇게 요금이 꼬박꼬박 청구되게 될 것이다..

이런 이유로 refresh는 사용하지 않는 것을 권장한다.

---

버전관리

terraform에선 provider의 버전을 관리할 수 있다.

기본적으로 버전을 명시하지 않고 apply를 실행하면 가장 최신 버전으로 실행된다. 그러나 중간에 새로운 provider 버전이 업데이트 되는 경우 어떻게 해야할까?

우선 처음 terraform init으로 플러그인을 설치하면 디렉토리에 lock 파일이 생성된다.

![img](3.png)

해당 lock 파일 내부에는 provider의 버전이 명시되어있다. 만약 새로운 버전의 provider를 사용하고 싶다면 lock파일을 지우고 다시 terraform init을 실행해 최신버전의 prodiver plugin을 설치할 수도 있고, 아니면

    terraform init -upgrade

를 실행하면 lock파일을 지우지 않아도 lock파일의 버전 내용이 최신 버전으로 업데이트 된다.

운영 도중에 최신버전의 provider를 도입하는 것은 더 발전된 기술들을 적용하는 장점이 있을 수 있지만, 반대로 버전을 변경하면서 오늘 호환성 문제가 생길 수도 있다. 그렇기 때문에 새로운 버전을 도입할 떄는 테스트를 잘 거친 후에 버전 업데이트를 하는 것이 좋다.

---

terraform은 리소스를 생성한다. AWS같은 서비스를 이용하기 위해선 인증/인가가 필요하다. 리소스를 사용한 만큼 값을 지불해야하는 유료서비스이기 때문이다. 그럼 terraform에서 비밀번호와 같은 인증정보를 코드로 작성해야할텐데, 이 파일이 깃허브같은 공유저장소에 올라가게되면 문제가 생길 수 있다. 악의적인 목적으로 사용자의 인증수단을 탈취해 클라우드 컴퓨팅 서비스를 악한 목적으로 사용할 수 있기 때문이다.

클라우드 계정이 해킹당해서 자신도 모르는 사이 몇천만원의 비용이 청구되는 사례를 심심치않게 찾아볼 수 있다.

원래같으면 

    provider "aws" {
        region = "ap-northeast-2"
        access_key = "..."
        secret_key = "..."
    }

이런식으로 코드를 작성하겠지만, 대신 AWS CLI를 이용하는 방법을 사용할 수 있다.

AWS CLI를 설치하고, console에서

    aws console

을 실행하면 region, access key, secret key 등의 환경변수를 설정할 수 있다. 설정하면 각 key는 파일형태로 변환되어서 특정한 위치에 저장되게 된다. terraform은 provider인 aws가 default로 파일을 저장하는 위치를 알고있기 때문에, access/secret key를 코드상에 입력하지 않아도

> 아 key가 없네? aws default directory 찾아봐야겠다!

이렇게 판단해서 인증을 거칠 수 있게 된다.


<br><br><br><br><br><br><br><br><br><br><br><br><br><br>

 terraform의 의존성문자열 등)을 하고싶다면 "${}" 를 이용해야함
 ex) "${aws_instance.chat-service.public_ip}32" (서브넷마스크)
 이러면 terraform은 알아서 아 이걸 참조하려면 인스턴스가 생성된 후에 이걸 실행해야하는구나 판단해서
 작업을 오류없이 동기적으로 실행할 수 있게 됨 이걸로 의존성 해결 가능

 aws console 이용하지않고 리소스에서 값 가져오기
 output "<local name>" {value = <value>} 코드를 tf파일에 작성하면 콘솔에 출력됨
 ex) output {value = aws_instance.chat-mservice.public_ip}
 output을 통한 출력값은 그냥 콘솔에만 표시하는 게 아니라
 이 output을 통해 다른 프로젝트의 tf파일에서 이 output 값을 참조해서 사용할 수 있음

 terraform의 변수
 반복적으로 사용되는 값이 바뀌게 되면(ex) local ip addreess 등) 매번 여러 곳의 ip를 다 수정해줘야함
 변수를 설정해서 값들이 변수를 참조하도록하면
 값 변경시 변수 값만 변경해주면 됨
 variable "<local name>" { default = "<value>"}로 tf파일에 코드 작성하고
 참조할 때는 [var.<local name>] 으로 참조할 수 있음
 default말고 다른 값으로 즉석? 사용하고싶으면 terraform plan 말고
 terraform plan -var="<local var name> = <value>" 로 command를 실행하면
 해당 varname의 값이 value로 인식되어서 terraform이 적용되게 됨 
 default를 지정하지 않으면 terraform이 실행될 때 알아서 물어봄
 아니면 terraform.tfvars 파일을 만들어서 안에 <local var name> = "<value>"를 적어두면 이걸 value로 인식함
 terraform.tfvars 파일의 이름이나 extension이 다르면 변수 인식을 못함
 예를 들어 custom.tfvars 파일의 내용을 적용시키려면 terraform plan -var-file "custom.tfvars" 로 plan을 시작해야 해당 파일의 변수내용을 적용시킬 수 있음
 변수선언시 type field 초기화를 통해 특정 타입의 값만을 변수로 받을 수 있음
 type을 선언하는게 좋음

 count 변수
 만약 100개의 동일한 instance를 생성해야한다면? 100개의 resource코드를 붙여넣을 순 없다
 resource의 field로 count = <number> 를 설정하면 그 수만큼 resource가 생성된다.
 여기서 resouce의 localname은 array로 동작하고 <local name>[<index>] 로 몇번째 인스턴스에 접근할 수 있다
 근데 이러면 모든 instance의 속성(이름 등)이 다 같아지게되는데, ${count.index} 를 통해 해당인스턴스의 index를 이용해 설정할 수 있다
 근데 또 이러면 각 인스턴스의 역할이 다 다른데 이름이 <~~~><index>로 통일되게 된다.
 이럴 땐 list var을 하나 더 만들어서 리스트 안에 사용할 변수명들을 지정해주고
 <list varname>[count.index]를 통해 반복적으로 생성되는 instance의 이름을 각기 다르게 설정해줄 수 있다 

 조건문
 리액트 조건문이랑 동일
 조건 ? 참 : 거잣
 ex) count = isExist == true ? 1 : 0
 위 코드는 isExist 변수가 참이면 resource를 1번 생성하고 거짓이면 생성하지 않는다

 locals
 variable이랑 코드 재사용성을 높이는 건 비슷한데 
 variable은 default를 정의해둬도 추가적인 정의가 가능한데
 local은 추가 정의가 불가능하고 value로 조건문이나 함수를 넣어서 사용이 가능함

 선언은 
locals {
    <Name> = "aaa"
}

 이런식으로 선언하고
 사용은 local.<Name> 으로 접근


 terraform function
 사용자가 함수를 정의할 순 없고 terraform 내부적으로 정의된 함수 사용 가능
 terraform console command로 함수가 어떻게 동작할 지 확인해볼 수 있음

 data type
 ami는 region에 따라 같은 os더라도 ami가 달라진다
 이를 해결하기 위해서 ami를 하드코딩하는 게 아니라
 참조하는 형식으로 구현
* data "aws_ami" "ami_test" {
    most_recent = true 가장 최신 버전 ami를 사용하겠다는 뜻
    owners = [amazon] 아마존 ami를 사용하겠다는 뜻

    filter {
        name = "name"
        values = ["amzn2-ami-hvm*"] 아마존ami중에서도 뭘 선택할건지 리눅스 기반 등등 선택하는 것
    }
} *

 terraform 파일의 코드를 포맷팅하고 싶으면 terrform fmt command

 terraform validate command
 해당 리소스들이 가지고있는 속성들이 유효한 속성들인지 확인할 수 있음
 aws_instance 리소스에 유효한 속성은 ami, instance_type 등인데
 그 외 정의되지 않은 속성을 작성했다면 validate command로 오류를 확인 가능
 물론 terraform plan에는 유효성 검사가 내장되어있음 refresh처럼!

 dynamic 블록
 dynamic 블록과 for_each로 어떤 var의 element를 돌면서 반복문을 동적으로 실행할 수 있음
 ex)
* dynamic "ingress" {
    for_each = var.<variable name>
    contents {
        port = ingress.value
    }
} *
 기본적으로 iterator는 lable namd (여기선 ingress)로 설정되는데
 iterator name을 변경하고 싶으면
 iterator = <name> 으로 설정해서 사용할 수 있음
 그럼 위 코드에서 ingress.value => <name>.value
 ingress 블럭은 인바운드, egress는 아웃바운드

 -replace (terraform taint)
 만약 인스턴스는 desired와 current가 동일한데
 manually current state의 설정을 수정했다면? (보안그룹, iam, eip 등등)
 그냥 다시 apply를 하면 인스턴스는 그대로이기 때문에 변화가 없음
 그럴 떄 current를 다 지우고 desired로 다시 깔아버리는 법
 terraform apply -replace "<resource name>.<local name>"
 이전에는 terrafrom taint라는 command를 사용했는데 지금은 apply -replace로 사용하는게 더 일반적

 specific index 대신 *을 사용하면 전체를 의미함
 -> splat expression

 terraform에서는 리소스간의 의존성을 graph로 확인할 수 있는 기능을 제공
 terraform graph > <파일명>.dot command로 파일 생성 가능
 이 파일을 graphviz 등으로 변환하면 이미지 파일로 볼 수 있음

 terraform plan to file
 terraform plan -out=path command로 테라폼플랜을 파일 형태로 저장할 수 있음
 이 파일은 terraform apply <파일이름>으로 적용이 바로 가능함

 terraform output <var name> command를 사용하면 매번 output을 보기위해 change도 없는 apply를 반복할 필요가 없이
 output만 확인 가능하다

 terraform의 resource들이 많을 떄 일부를 수정하면
 전체를 다시 새로고침하기 때문에 aws의 api 호출 제한이 걸리는 등의 오류가 생길 수 있다.
 1. terraform plan -refresh=false
 ㄴ 이러면 새로고침 기능이 사라진다. 코드를 수정한 부분의 업데이트는 refresh가 false여도 일어난다.
 1. terraform plan -target=<resource name>.<local name>
 ㄴ 이러면 전체 코드가 아닌 해당부분에만 대해서 planapply를 한다.
 근데 이건 둘 다 운영환경에서는 사용하면 안됨

 [] -> list
 {} -> set
 set은 ordered되지 않아서 index가 정해져있지 않지만, 대신 중복이 불가능하다
 list는 order되어서 index가 정해져있지만, 대신 중복이 가능하다.
toset func로 list를 set으로 변하시킬 수 있음

for each를 쓰면 each.key, each.value 로 값에 접근할 수 있음

provisoner는 리소스가 생성된 이후의 작업들을 하게 해줌(nginx를 인스턴스에 설치 등..)
local exec 는 로컬에서 하는 작업들
remote exec는 원격서버에서 하는 작업들

remote exec는 ssh접속, nginx설치 등 생성한 인스턴스의 서버에서 할 작업들을 의미
local exec는 terraform이 실행되고있는 지금 나의 로컬 서버에서 할 작업들을 의미

provisioner는 크게 creation/destroy time provisioner로 나뉨

creation time은 따로 명시하지 않아도 default이고 리소스가 생성될 때 작동하는 exec임
destroy time은 exec 안에 when = destroy 를 넣어서 리소스가 파괴될 때 실행되도록 설정할 수 있음
creation time provisioner의 특징은 만약 리소스 생성시 오류로 해당 provisioner가 실행되지 못했다면 그 리소스는 taint된 것으로 표시된다는 것
taint된 리소스는 다음에 동일한 terraform apply시에 자동으로 파괴되고 재생성된다
taint된 리소스는 tfstate파일의 state의 값으로 "tainted"를 가진다

provisioner 블럭 안에 on_failure 를 fail/continue로 설정하는 것에 따라서 실패(taint)시 어떻게 terraform이 작동할지를 선언할 수 있다
fail은 default로 provisioner가 실패하면 taint되면서 작업이 중지되고, 만약 continue로 설정해두면 오류가 생겨도 taint되지도 않고 작업이 중지되지도 않는다

여러 프로젝트를 진행할때 동일한 리소스를 계속 생성하게 된다면 코드 중복을 피하기 위해 module 블럭을 이용할 수 있음
module "LOCAL NAME" {
    source = "RELATIVE PATH"
}
이런 식으로 작성하면 상대경로에 있는 파일의 리소스를 참조해서 사용할 수 있음
디렉토리는 Modules, projects 나눠서 사용/참조하는 게 일반적인듯 함
참조할 때 source에는 tf파일이 아니라, tf파일이 위치한 디렉토리를 적는 것 같음
> 이게 가능하려면 기본적으로 모든 프로젝트의 tf파일들은 하나의 루트 디렉토리에서 작성되는 것이 컨벤션인 것 같음
> 깃허브 tf 레포지토리도 하나 만들어서 모든 프로젝트의 tf파일들을 관리할 수 있게 하면 좋을 것 같음

module output
module을 참조하고 output 값을 읽어올 수 있음
module.<local module name>.<output name>
이런 식으로 모듈의 output 값을 불러와서 속성으로 정의할 수 있음

모듈은 꼭 직접 선언한 것 뿐만 아니라 terraform registry 에서 인증된 모듈을 가져다 쓸 수도 있음

terraform workspace
환경 별 워크스페이스를 별도로 설정할 수 있음
dev, staging, prod 등등
terraform workspace -h 로 커맨드 확인 가능
워크스페이스별 instance type을 다르게 주고싶다면 map과 lookup function 이용해서 할 수 있음
terraform은 tfstate파일을 워크스페이스 별로 구분해서 가지고있음

module의 source는 git도 참조할 수 있는데, github의 특정 레포지토리를 참조하기 위해선
source = "github.com/<usename>/<repo name>"
이렇게 지정하면 해당 repo의 tf파일들을 참조한다
비밀번호 등의 보안내용들을 .gitignore 이용해서 public repo에 올라가지 않도록 해야함
.gitignore에 포함시켜야할 것들
- .terraform 파일
- terraform.tfstate (password가 포함되어있음)
- <name>.tfvars (password가 포함되어있을 가능성이 높음)

그런데 tfstate파일이 협업간 공유되지 않으면 문제가 생길 수 있는데
tfstate파일은 보안문제로 공유저장소에 push할 수 없으니 어떻게 해야할까?
중앙 백엔드에 tfstate파일을 저장하는 방식을 사용해야함
중앙 백엔드는 s3 bucket이 될수도, 혹은 쿠버네티스가 될 수도 있음
이를 통해서 terraform을 이용한 협업을 할 때, 공유저장소에서 tf파일들을 가져오고, 작업할 때 중앙 백엔드의 tfstate파일을 참조/수정하며 협업이 가능해지게 됨
백엔드 설정하는 법
terrform {
    backend "<backend name>" {
        ...
    }
}
근데 또 백엔드는 locking을 지원하지 않을 수 있음
locking이 지원되지 않으면 한명이 terraform 을 진행하고 tfstate가 수정중인 상황에서 끼어들어서
tfstate가 엉키고 원하는 결과를 못얻을 수 있음
그래서 tfstate를 위한 백엔드를 구분해서 구축할 때 db를 이용할 수 있음
방식은 terraform이 작업중일때 db테이블에 레코드를 추가하고, 끝나면 삭제하는 방식
다른 사람이 terraform 작업을 실행하면 terraform이 db에 레코드가 있는지 먼저 확인하고, 없으면 레코드를 만들면서 작업을 실행하는 방식으로 lock이 이루어짐

terraform state 관련 명령어들
terraform state list : 현재 terraform으로 생성되어있는 리소스들의 이름과 로컬 이름을 보여줌
terraform state mv <resource name>.<localname1> <resource name>.<localname2> : 리소스를 파괴 후 재생성하지 않고도 로컬이름을 변경
terraform state pull : tfstate파일 내용을 출력. 특히 tfstate파일을 위한 중앙백엔드를 이용하고있을 때 tfstate 상태 확인에 유용함
terraform state remove <resource name>.<localname> : 해당 localname을 가진 resource를 tfstate파일에서 지움. 그럼 결론적으로 리소스 하나를 잃어버리게 되는 셈
terraform state show <resource name>.<localname> : 해당 localname을 가진 resource의 desired state를 확인할 수 있음