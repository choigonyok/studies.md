# #6. 커플 채팅 서비스 시연영상 및 회고
# project couple-chat

---

## 개요

약 한 달간의 커플 채팅 서비스 개발이 마무리되었다.

서비스 시연영상과 함께, 처음 서비스 개발을 기획하며 적용하고자 했던 것들의 후기를 돌아보려고 한다.

---

## 시연 영상

<iframe width="900" height="600" src="https://www.youtube.com/embed/eciM1M9p2E4?si=rCm5wBqr4AZiUp-c" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

---

## 후기

처음 개발을 기획하면서 프로젝트를 얻고자했던 것들을 돌아보자.

```
1. Terraform을 활용한 인프라스트럭처 프로비저닝
2. Kubeadm과 Containerd를 활용한 쿠버네티스 클러스터 프로비저닝
3. Kubernetes를 활용한 어플리케이션 배포
4. Gorilla Websocket 활용한 실시간 채팅
```

### 1. Terraform을 활용한 인프라스트럭처 프로비저닝

테라폼에 대한 지식은 정말 많이 늘었다. AWS 한정이지만, 테라폼을 공부하면서 AWS와 네트워크에 대한 이해도 함께 할 수 있어서 좋았다.


### 2. Kubeadm과 Containerd를 활용한 쿠버네티스 클러스터 프로비저닝

가장 힘든 부분이었는 클러스터 구성이다. 그래도 포기하지 않고 끝까지 물고 늘어지니까 해결이 됐고, 생각보다 단시간 내에 쿠버네티스의 개념적인 부분에 대해서는 상당부분 이해한 것 같다.

다음번 클러스터를 구성할 때는 kubeadm같은 베어메탈 툴 말고, 가장 널리 쓰이는 AWS EKS로 구성해봐야겠다. 분명이 kubeadm으로 스크래치 구성을 해봤기 때문에 훨씬 쉽게 적응하고 사용할 수 있지 않을까 생각된다.

### 3. Kubernetes를 활용한 어플리케이션 배포

정말 어렵고 복잡한 툴이지만, 한 번 개념이 정리되고 큰 그림이 이해되니까 참 재밌는 도구인 것 같다. 다만 학부생 수준에서 쿠버네티스를 적극적으로 활용하기엔 리소스를 구성하기 위한 비용적 부담이 크다. 이후에 Istio나, Jenkins 등의 여러 쿠버네티스 연관된 툴들을 사용하려면 그 파드들을 배포할 노드를 추가적으로 생성하기도 해야하고, etcd 분리와 고가용성 운영을 위한 노드들도 구성해야한다.

그래도 분명 장점이 많고 재미도 있는 도구임은 확실히 느꼈다.

### 4. Gorilla Websocket 활용한 실시간 채팅

사실 웹소켓도 만만치 않게 어려운 개념이었다. 아예 들어본 적도 없는 분야였기 때문에 초반에 고전했지만, 공식문서 번역을 하며 그나마 좀 이해할 수 있었다. 다만 웹소켓이 훨씬 더 다양한 방식으로 응용될 수 있을 것 같다는 생각이 들어서, 기회가 되면 웹소켓을 이용한 고급 기능들을 구현해봐야겠다고 생각했다.

---

## 정리

남들이 "저는 HTML로 코딩합니다" 라는 개발자식 유머에 웃을 때, 나는 HTML이 뭔지 몰라서 웃을 수 없었다. 그런 내가 이젠 혼자서 두 번째 프로젝트를 마무리하고 있다는 사실이 참으로 감격스럽다.

지금 학부생 수준에도 못미치는 실력으로 혼자 개발한 이 프로젝트는 개발 연차가 늘어갈수록 리팩토링에 리팩토링을 거듭해야겠지만, 일단 뭐가됐든 마무리까지 잘 해냈다는 것에 만족한다.

이 프로젝트를 개발하면서, 또 개발하고 배포하기 위해 공부하면서, 앞으로 어떤 방향으로 학습하고 나아가야할지 조금은 감이 잡혔다.

추가적으로, 프로젝트를 진행하면서 개발 과정이나, 문서화를 습관적으로 자주 해놓지는 못했다. 한 달을 온전히 쏟으면서 개발에 매진했는데, 그 과정에서 많이 배우고 느낀 것들을 다 글로 정리해두지 않았다는 것이 아쉽다.

다음 프로젝트부터는 문서화에 더 신경써서, 블로그 게시글도 완성도 있는 높은 수준으로 작성할 수 있도록 해야겠다.

---

## 프로젝트 관련 레포지토리

[Github](https://github.com/choigonyok/couple-chat-service-project)