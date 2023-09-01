# AWS EBS As Root Volume in EC2
# study kubernetes aws

---

## 개요

블로그 리팩토링 게시글 중 하나인 **#9. Terraform Infrastructure Provisioning & Trouble Shooting**에서 Disk Pressure로 인해 볼륨을 확장하는 방법에 대해 알아보았다.

이 게시글 바로 직전 글인 **#12. Config AWS EBS PersistentVolume in K8S**에서는 K8S에서 노드의 hostPath Volume을 사용하지 않고 PersistenceVolume을 연결해 사용하는 방법을 알아보았다.

잘못 알고있던 사실이 있다. EBS는 EC2를 생성할 때 기본적으로 루트 볼륨으로 설정될 수 있다는 것과 콘솔에서 볼륨을 확장한다고 디스크 용량이 확장되는 것이 아니라는 것이다.

#9 글에서 진행했던 내용만으로는 실질적인 디스크 확장이 이루어지지 않고, #12 글에서 진행한 PV 설정은 굳이 하지 않아도 되는 일이었다.

---

## #9 내용 정정

#9 글애서 내가 작성한대로 볼륨을 확장했다. 글에서처럼 처음부터 큰 EBS 볼륨 사이즈로 생성한 것이 아니라, 기존 8GiB로 볼륨이 생성되어있었기 때문에, 콘솔에서 16GiB로 확장을 했었다.

그렇게 서비스들을 재배포하니 Disk Pressure가 생기지 않아서 잘 설정된 줄 알았다. 그러다가 PV 설정을 하면서 Disk Pressure가 다시 발생한 것을 확인하게 되었다.

### 리눅스 디스크 관련 Command

그래서 리눅스에서 가용 Disk 옹량, 현재 사용중인 Disk 용량을 확인할 수 있는 커맨드를 알아봤다.

    lsblk

이 커맨드는 리눅스에서 여러 저장장치 정보를 보여주는 명령어이다. 출력 내용은 아래와 같다.

![img](http://www.choigonyok.com/api/assets/61-1.png)

분명 disk 타입의 nvme0n1이 16GiB로 설정이 되어있는데, 뭐가 문제인걸까? 16GiB도 부족해서 그런걸까?

    df -Th
  
이 커맨드는 디스크의 용량과 현재 사용량을 보여준다. 출력 내용은 아래와 같다.

![img](http://www.choigonyok.com/api/assets/61-2.png)

ext4 타입의 /dev/root 디스크의 용량이 7.6GiB, 사용량이 5.8GiB였다. 어째서? 왜 lsblk와 df -Th의 결과가 다른거지?

### Partition

답은 **파티션**에 있었다.

디스크는 "파티션"으로 논리적 구분이 되어있다. 전체 용량은 16GiB로 늘어났지만 / (루트) 파티션에는 늘어난 용량이 할당되지 않았기 때문에, 계속해서 Disk Pressure가 발생했던 것이다.

다시 lsblk로 출력된 내용을 보면, nvme0n1 디스크의 하위항목으로 nvme0n1p1, nvme0n1p14, nvme0n1p15가 있었다.

![img](http://www.choigonyok.com/api/assets/61-3.png)

이 p가 파티션을 의미하는 것이었고, nvme0n1 디스크는 이 세 개의 파티션으로 이루어져있음을 의미하는 것이었다.

그 중 nvme0n1p1은 MOUNTPOINTS가 **part /** 로 설정되어있다. 이 파티션이 루트 파티션인 것이다. 파티션의 용량을 키워주기 위해 아래 커맨드를 입력했다.

    sudo growpart /dev/nvme0n1 1

nvme0n1 디스크의 1번 파티션을 확장시키는 명령어이다. 실행하고 lsblk의 출력 결과를 확인했다.

![img](http://www.choigonyok.com/api/assets/61-4.png)

nvme0n1의 루트 파티션인 nvme0n1p1이 아까 8G와 다르게 16G로 잘 확장이 된 걸 확인할 수 있었다.

그래서 신나서 마스터노드로 접속해서 Pod를 재배포했다. 그런데 또 Disk Pressure가 발생하는 것을 확인할 수 있었다. 혹시나해서 배치되는 워커노드의 taint도 확인하고, disk-pressure taint도 삭제한 후 재배포했다. 그러나 Disk Pressure는 계속 일어나고 있었다.

다시 해당 워커노드에 접속했다.

lsblk는 16G로 확장되었는데, df -Th로 확인한 출력에는 확장이 되어있지 않은 걸 확인할 수 있었다.

알아보니 

    sudo resize2fs /dev/root

이 커맨드를 추가로 사용해줘야한다고 한다. 확장시킬 파일시스템을 알려주는 것이다. 근데 확장은 아까 growpart 커맨드로 했는데 왜 또 확장을 해야하지?

### 파티션 확장 VS 파일시스템 확장

이유는 growpart는 파티션을 확장하는 명령어이고, resize2fs는 파일시스템을 확장하는 명령어이기 때문이다.

파티션 안에는 하나의 파일시스템이 생성된다.

근데 왜 굳이 둘을 나눠놨을까?

파티션은 디스크를 **물리적**으로 나눈다. 하나의 디스크여도 여러개의 디스크로 나눠서 각 파티션의 격리로 인한 보안성도 올라가고, OS도 별개로 설치할 수 있다.

중요한 것은 물리적이라는 것이다. 그래서 파일시스템은 파티션이 확장되어도 확장여부를 알지못한다. 그래서 명시적으로 파일시스템도 별도로 확장을 해주어야하는 것이다.

---

결론적으로 sudo resize2fs /dev/root 커맨드를 사용하고나니,

![img](http://www.choigonyok.com/api/assets/61-5.png)

df -Th 출력결과도 아래와 같이 16G로 확장된 것을 확인할 수 있었다.

![img](http://www.choigonyok.com/api/assets/61-6.png)

그리고 다시 백엔드와 프론트엔드 Pod를 배포하니 문제없이 잘 실행되는 것을 확인할 수 있었다.

---

## #12 내용 정정

#12 글에서는 내가 잘못 이해하고 있던 개념이 있다.

AWS에는 EC2의 기본 스토리지가 존재하고, EC2 인스턴스를 생성하면 그 기본 스토리지가 사용되며, 확장이 필요할 경우 EBS를 붙여서 사용하는 것인 줄 알았다.

그래서 쿠버네티스에서 EBS를 사용하기 위해 PV 설정을 한 것이었는데, 사실 EC2의 기본 스토리지라는 건 없고, 인스턴스 생성에 필요한 스토리지는 EBS가 루트볼륨으로 기본적으로 사용된다는 것을 알게되었다.

이후에 원한다면 루트 볼륨을 다른 스토리지로 변경할 수도 있지만 우선 default는 EBS가 사용된다.

#12에서 다룬 EBS CSI Driver를 활용한 EBS 볼륨 설정은 기본 스토리지 이외에 특정 데이터를 영구보존하고 싶을 떄 유용하게 사용할 수 있을 것 같다.

EC2의 기본 루트 볼륨이 EBS이기 때문에, Pod에서 hostPath 볼륨으로 설정하고 마운트해도 결국 EBS에 데이터가 저장되는 셈이니까 안전하게 데이터가 보존될 수는 있겠지만, hostPath는 영구 보존을 원하는 데이터 이외에도 OS부터 모든 인스턴스의 데이터가 다 들어가있기 때문에, 데이터를 영구보존하고, 언제든지 붙였다 뗐다 하면서 유용하게 사용하기엔 무리가 있을 것이다.

결론적으로 EBS CSI Driver를 활용한 쿠버네티스 PV설정도 유용한 기능이고 공부해두길 잘 했지만, 잘못된 정보를 알고있었던 것은 수정이 필요할 것 같아 글로 정리하게 되었다.

---

## 정리

채팅 서비스를 개발할 때는 웹소켓, 쿠버네티스를 공부할 때는 쿠버네티스, 이런 식으로 필요한 딱 그것만 공부하면 될 것 같은 막연한 느낌이 있었다. 사실 많은 것들이 서로 연관되어있고, 쿠버네티스 같은 고급 툴을 다루고 배워가면서도 리눅스, 네트워크 같은 기본기에 대한 지식을 꾸준히 쌓아가고 있는 것 같아서 뿌듯하다.

---

## 참고

[리눅스 커맨드](https://blog.desdelinux.net/ko/HDD-%EB%98%90%EB%8A%94-%ED%8C%8C%ED%8B%B0%EC%85%98%EC%9D%98-%EB%8D%B0%EC%9D%B4%ED%84%B0%EB%A5%BC-%EC%95%8C%EA%B8%B0%EC%9C%84%ED%95%9C-4-%EA%B0%80%EC%A7%80-%EB%AA%85%EB%A0%B9/)

[Taint 삭제](https://cloud.google.com/kubernetes-engine/docs/how-to/node-taints?hl=ko#remove_a_taint_from_a_node)

[EBS 확장](https://velog.io/@harvey/AWS-EC2-%EC%9D%B8%EC%8A%A4%ED%84%B4%EC%8A%A4-%EC%9A%A9%EB%9F%89-%ED%99%95%EC%9E%A5)