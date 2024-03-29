# #1. 나만의 테크블로그 개발
# project blog

---

## 개요

5월 중순에 첫 프로젝트로 2주간 블로그를 개발해 배포까지 했었다. 

![img](http://www.choigonyok.com/api/assets/29-1.png)

백엔드와 프론트엔드라는 단어조차도 얼핏 들어보기만한 정도의 수준에서 그냥 부딪히며 프로젝트를 진행했다. 마무리하고 보니 아쉬운 점들이 많았다. 지난번엔 개발과정의 전반적인 경험과 이해를 위한 목적으로 개발했다면, 이번엔 실제로 사용할 목적을 가지고 다시 블로그를 개발하기로 했다.

---

## 이전 프로젝트에서 구현한 것

Go와 Go의 웹 프레임워크인 Gin으로 블로그를 개발했다. 구현한 기능들은 다음과 같다.

### AdminPage에 접속하기 위한 LoginPage

블로그 특성 상 로그인이 필요한 사용자는 나 혼자였기 때문에, 사전에 서버에 ID/PW를 정의해두고, 입력된 문자열과 비교하는 로직을 구현했다. 확인되면 로그인 없이 글을 쓰거나 지우는 등의 admin 작업을 할 수 있도록 쿠키를 설정했다. 게시글 수정/삭제/작성 이벤트가 발생하면 쿠키를 확인하고, 쿠키가 없으면 로그인 페이지로 리다이렉트 시켰다.

### 게시글 작성/삭제/수정

게시글의 작성 및 렌더링은 깃허브의 마크다운 API를 활용했다. 마크다운 API로 사용자가 입력한 텍스트를 보내 마크다운 형식으로 변환했다. 이렇게 외부API를 사용해보면서 API에 대한 감을 조금 잡았다. 추가적으로, 변환된 본문/제목/작성날짜/카테고리 데이터를 DB에 저장하는 로직을 구현했다.

게시글 수정은 DB에 저장되어있던 본문을 가져와 작성 칸의 value로 입력되게 했다. 사용자가 글을 수정하고 저장하면 기존 DB 레코드에 덮어쓰는 방식으로 구현했다.

* ### Today/Total 방문자 수 체크

방문자 수를 세기위해 쿠키를 사용했다. 블로그에 접속하면 방문자의 쿠키 여부를 확인하고, 쿠기가 없으면 Today와 Total 수를 1 증가시키면서 방문자에게 쿠키를 발급하는 로직을 구현했다. 물론 쿠키가 이미 존재하면 Today는 증가하지 않는다. Total은 1초 간격으로 계속 반복실행되는 함수를 통해 구현했다. 매 초마다 현재 시각과 00:00:00(자정)을 비교하고, 같아지면 Today를 0으로 초기화하는 로직을 구현했다.

* ### 최근 포스트 게시

DB에 각 게시판 별로 테이블을 만들었고, 게시판에 접속하면 해당 테이블의 레코드들을 뿌려주는 방식이었다. 최근 포스트는 게시판을 구분하지 않고 전체 글 중 최신 작성 글을 보여줘야하기 때문에 각 게시판 테이블의 기본키를 외래키로 갖는 whole 테이블을 만들어서 작성된 게시판과 전체 게시글 id가 저장되도록 구현했다. 이 whole 테이블을 통해서 메인 페이지 하단에 최근 작성된 6개의 게시글을 카테고리 구분 없이 볼 수 있게 했다.

* ### 검색 기능

SQL의 LIKE 문법을 이용해서 본문에 검색어가 있으면 최신 작성 순으로 보여주도록 구현했다. 

검색된 결과가 없으면 사용자에게 검색 게시물 없음을 나타내도록 구현했다.

* ### 이전/다음 포스트 전환 버튼

게시글에서 이전/다음 버튼을 누르면 현재 보고있는 게시글 id를 기준으로 바로 앞 id를 가진 게시글, 바로 다음 id를 가진 게시글을 볼 수 있게 구현했다.

<br>

핵심 기능은 이 정도였고, AWS의 EC2 우분투에 빌드해 배포했다. 배포는 nohup command를 이용해서 24/7 블로그에 접속할 수 있도록 했다.

---

## 이전 프로젝트에서 아쉬웠던 것

### 1. 프런트엔드에 대한 지식 부족

클라이언트 사이드에 대한 기본적인 지식이 아무것도 없었다. 언어적인 부분 뿐만 아니라 클라이언트는 뭘 하는 건지에 대한 큰 그림이 없었던 것이다. 그래서 개발을 위해 HTML 템플릿 2,3장을 다운받아서 미리 구현되어 있던 버튼 같은 컴포넌트들을 조금씩만 수정하며 개발했다.

JS는 커녕 HTML 태그나 CSS 속성에 대한 기본적인 지식도 없었기 때문에 내가 원하는 디자인이나 사용자 경험을 생각하기보단 기능적인 로직을 구현하는 것에만 집중할 수 밖에 없었다.

### 2. 3-tier 아키텍처에 대한 이해 부족

웹 개발에 대한 깊은 이해 없이 계란으로 바위치기 심정으로 진행한 첫 프로젝트기도 했고 위에 말했듯 클라이언트 사이드에 대해 모르다보니, 기본적인 HTML 렌더링부터 라우팅, 인증/보안, DB 커넥팅, 게시글 렌더링 등 모든 기능을 GO로 서버 사이드에서 작성하게 되었다.

프런트엔드와 서버-클라이언트 아키텍처에 대해 좀 더 공부해서 완성도 있는 블로그를 다시 개발하고 싶다는 생각을 했다.

이런 이유들을 바탕으로 클라이언트 사이드의 대표 언어라고 할 수 있는 react.js를 이론적으로 공부하면서 동시에 몸으로 부딪혀가며 배우는 "리액트로 바위치기" 블로그 개발 프로젝트 ver.2를 기획하게 되었다.

### 3. 쿠키 보안 문제

쿠키값을 클라이언트가 임의로 변경할 수 있는지 전혀 모르고있었다. admin을 확인하기 위해 쿠키의 키는 admin, 값은 ok로 설정하고, admin기능을 수행할 때마다 쿠키의 키가 admin이면서 값이 ok인지를 확인했다. 누구든 쿠키를 통해 admin이 작동한다는 걸 알았다면, 관리자권한으로 내 블로그에 접근해 글을 삭제하거나 작성할 수 있었을 것이다.

### 4. 방문자 수 체크 한계

위에서 말한 것처럼, 전체 방문자 수를 구하려면 매일 자정마다(날찌가 바뀔 때마다) Total 값을 0으로 초기화해줘야한다. 매일 자정을 확인하기 위해 1초마다 시간을 확인하는 로직을 구현했다. 현재 시간이 00시 00분 00초가 되면 초기화하도록 하는 것이다.

문제가 두 가지 있다.

1. 나머지 23시간 59분 59초동안을 전혀 쓸데없는데, 날짜가 넘어가는 단 1초를 위해서 그 시간동안 리소스를 사용하게 된다. 상당히 비효율적이다.
2. 1초의 타임슬립을 지정해두고 반복문을 돌렸는데, 만약 1초 + 약간의 실행 딜레이가 발생해서 마침 딱 정각을 못보고 지나치게 된다면? 그럼 다음날짜에도 전날의 today 수가 이어지게 된다.

---

## 왜 하필 블로그를 개발할까?

유튜브에 "블로그 개발"이라고 검색하면

> 첫 프로젝트로 블로그 개발하지 마세요

> 클론 코딩 하지 마세요

이런 영상들이 나온다.

첫 번째, 두 번째 프로젝트로 남들이 다 하지말라는 블로그를 개발하게 된 이유가 세 가지 있다.

### 1. 쉬워서

블로그 개발은 너무 쉽다고들 한다. 오히려 잘 됐다고 생각했다. 첫 프로젝트인 만큼 과분하게 어려운 목표를 정하면 쉽게 지칠 수 있을 뿐더러 종국엔 프로젝트를 완성하지 못하는 상황이 생길 수 있다.

또 쉽다는 이유로 많은 개발자들이 블로그 개발에 도전하기 때문에, 진행하면서 겪는 어려움이나 문제들을 금방 해결할 수 있을 것이다.

### 2. 클론 코딩 하지 않을거라서

클론 코딩의 정의가 사람들마다 나뉘는 것 같다.

1. 인프런, 유튜브 등 개발강의를 보면서 A-Z까지 하라는대로 따라서 만드는 것

2. 한 레퍼런스를 정하고 어떻게 기능들을 구현하였을까 고민하면서 같은 기능을 구현하는 것

둘 중 클론코딩의 정의가 뭐가 됐던 난 둘 다 하지 않을 것이다.

직접 들이박고 해결해내며 얻는 지식의 가치나 기쁨을 많이 느껴봤기 때문이다. 물론 문제에 직면했을 때, 빠르게 답을 찾아내 문제를 해결하는 것도 개발자의 중요한 덕목 중 하나이다.

그러나 지금 개발에 막 뛰어든 나로써는 혼자 고민하고 탐구하며 기본기를 탄탄히 쌓아가는 것이 더 중요하다.

### 3. 앞으로 유용하게 쓸 거라서

> 내가 직접 개발한 이 세상 하나뿐인 나만의 블로그! 

완성도있게 개발해두면 블로그에 대한 애정이 넘쳐서, 혹은 개발한 게 아까워서라도 블로그에 좋은 내용의 글을 더 담게 될 것이다. 블로그 개발이 학습에 대한 동기부여로 작용할 수 있다.

---

## 블로그 개발을 통해 얻고싶은 것

이전 블로그 프로젝트의 가장 큰 목표는 그 동안 나름 전공이라고 배운 것들이 실제 개발에 얼마나 유익한지 직접 경험해보는 것이었다. 그 목표는 이뤘다.

이번 블로그 개발 Ver.2를 통해 얻고 싶은 것은 크게 세 가지가 있다.

* ### 클라이언트 사이드 언어(리액트)에 대한 이해와 학습

-> 리액트를 프론트엔드 개발자만큼 깊게 파야겠다는 생각은 없다. 그래도 다양한 프로젝트를 진행할 때 클라이언트를 구현하기에 큰 문제가 없을 정도의 지식을 쌓고싶다.

* ### 서버-클라이언트 아키텍처에 대한 이해와 경험

-> 처음오로 BE/FE를 나눠서 개발하게 될텐데, 서로 다른 언어와 서버 간에 어떻게 서로 통신하는지, 왜 그렇게 해야하는 건지, 다른 방법은 없는지에 대한 이해를 하고싶다.

* ### RESTful API 개발에 대한 지식과 경험

-> 그 유명한 RESTful API. 실제로 개발해보고, 기회가 된다면 GraphQL에 대해서도 공부해보려고 한다.