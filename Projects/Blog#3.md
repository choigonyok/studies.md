# [BLOG #3] 관련 / 전체 게시글 기능 구현

---

## 개요

현재 게시글과 같은 태그의 글들이 게시글의 맨 밑에 related post 섹션에 나타나도록 구현하고 싶었다. 클라이언트가 홈페이지로 번거롭게 돌아가지 않고도 관련 게시글을 볼 수 있게하는 것이 사용자 경험 측면에서 필요하다고 생각했다.

또 메인 페이지에서 카테고리(태그) 버튼을 눌러야만 해당하는 게시글들이 나타나고, 처음 접속 시에는 모든 태그의 게시글들이 나타나지 않는 상황이었는데, 이 두 가지 구현의 과정을 적은 글이다.

---

## 관련 게시글 보기 구현

구현은 현재 게시글을 필터링하기 위해 리액트의 filter와 includes를 이용했다.

```js
console.log(jsonArray.filter((post) => post.Id.includes(postid)));
```

구현하고나니 게시글을 태그 별로 분리해서 불러오면 현재 보고있는 글도 같이 related posts에 표시되는 문제가 생겼다.

### 타입 문제

jsonArray는 서버로부터 받아온 같은 카테고리를 가진 게시글 데이터 배열이고, postid는 DB에서 기본키로 설정된 게시글의 id이다. 이 코드를 실행하면 오류가 발생한다. 오류 내용을 살펴보면,

![1](/assets/8-2.png)

post.Id.includes가 funcion이 아니라는 메시지가 콘솔에 출력된다. includes 대신 jsonArray의 element들 중에 post[0].Id가 postid와 같지 않은 것들만 남기는 방식으로 다시 구현했다.

```js
console.log(jsonArray.filter((post) => post[0].Id !== postid));
```

이렇게되면 postid(현재 보고있는 글)과 같은 id의 글만 related posts에서 필터링될 수 있을 것이다.

![1](/assets/8-4.png)

그랬더니 콘솔엔 postid와 post[0].Id가 둘 다 10으로 잘 출력되는데, 오류가 발생하는 걸 볼 수 있었다. 10 = 10이 틀렸다는 우리의 리액트

> 가만보니 콘솔창의 두 10의 색상이 다르다.

![1](/assets/8-5.png)

이 보라색 10이 처음 서버에서 넘어올 때 클라이언트가 콘솔에 출력하는 게시글 데이터를 확인해봤다. 살펴보니 여기도 Id의 10이 보라색으로 출력된다. 그래서 백엔드 서버에서 데이터를 클라이언트에 전송하는 구조체를 어떻게 정의했었는지 확인했다.

```go
type SendData struct {
    Id int
    Tag string
    Title string
    Body string
    DateTime string
    ImagePath string
}
```

Id가 int로 선언된 걸 확인할 수 있었다. postid는 URL 파라미터를 통해 받아온 string 10이었고, post[0].Id는 서버에서 int로 받아온 10이었기 때문에 두 10은 타입이 달라 서로 비교될 수가 없어서 에러가 발생했던 것이다.

## 해결

```js
console.log(String(jsonArray[0].Id));
```

이후에 int 10을 string 타입으로 형변환 했더니

![1](/assets/8-8.png)

두 10 모두 흰색으로 잘 출력되는 걸 볼 수 있었다!

```js
setRelatedPostData(jsonArray.filter((post) => String(post.Id) !== postid));
```

필터링을 해서 보고있는 게시글 데이터만 빠진 배열로 UseState를 초기화해주니 컴포넌트가 리렌더링 되면서 Card 컴포넌트에는 필터링된 게시글들의 데이터가 잘 전달되면서 원하는 기능을 구현할 수 있었다. 

---

## 전체 게시글 보기 구현

> 태그(카테고리) 버튼을 아무것도 클릭하지 않은 첫 접속 상태에서는 전체 게시글을 보여주는 기능

처음 클라이언트가 블로그에 접근하면 루트 경로로 라우팅이 되고, 아직 어떤 태그 버튼에도 클릭 이벤트가 발생하지 않은 상태이기 때문에 아무 게시글도 확인할 수가 없었다.

사용자가 태그버튼을 클릭한 이후에서야 태그가 일치하는 게시글들이 나타났다.

![1](/assets/8-10.png)

그 이유는 위의 코드처럼 버튼 클릭 핸들러에만 POST요청을 보내는 **useEffect를** 구현해뒀기 때문이다.
POST 요청은 버튼을 클릭하면 ClickHandler를 통해 useState 함수로 PostData를 JSON 형식으로 초기화하고,

    {tagname : value}

이 PostData를 본문에 넣어 서버로 요청하는 방식을 사용했다.

### useEffect에 대해서

이 기능을 구현하면서 **useEffect**를 처음 알게되고, 또 사용해보았다.

기본적으로 리액트의 컴포넌트는 처음 마운트가 되면 리렌더링하는 방식으로 화면을 수정한다. 근데 만약 한 컴포넌트에서 특정 이벤트가 발생할 때마다 서버에 요청을 보내고 싶다면? 그럼 클라이언트가 매번 새로고침을 누르거나 재접근을 해야만한다. 이 불편을 줄여줄 수 있는 게 useEffect 훅이다.

컴포넌트가 처음 실행(마운트)되면 useEffect 안의 내용도 우선 한 번 실행된다. 이 컴포넌트는 사용자가 처음 접속하면 실행되기 때문에, POST요청은 PostData의 디폴트 값으로 요청이 된다. 이후에는 (다른 useState의 값 변경으로 인한)리렌더링과 관계없이 useEffect의 배열 []에 들어있는 POSTDATA 값에 변화가 있을 때만 useEffect가 재실행 된다. 이게 useEfffect의 특징인데, 내용을 정리해보자면 다음과 같다.

1. useEffect []가 없는 경우

2. useEffect에 []가 있지만 비어있는 경우

3. useEffect에 []가 있고, 내용이 있는 경우

1번 경우엔 컴포넌트가 리렌더링이 될 때마다 useEffect가 실행된다.

2번 경우에 useEffect는 컴포넌트와 동일한 생명주기를 갖는다. 처음 컴포넌트가 마운트될 때 한 번만 실행된다.

3번 경우엔 리렌더링과는 별개로 [] 안에 있는 값이 변화할 때마다 독자적으로 useEffect만 리렌더링된다. 이 기능 덕분에 useEffect가 API요청과 자주 함께 사용되는 것 같다.

---

# useState default로 동적인 기능을 구현하려면

돌아와서, useState함수인 setPostData() 가 클릭 핸들러에서만 동작했기 때문에 태그 버튼을 클릭하는 이벤트가 발생해야만 POST요청이 서버에 전송되었던 것이고, 내가 원하는 기능은 처음 접속했을 때 아무것도 누르지 않아도 전체 POST들을 볼 수 있게 하는 것이었다.

이 기능을 구현하기 위해서는 PostData에 default로 무엇이 들어있는지가 중요하다. useEffect의 []가 있든 없든, 뭐가 들어있는 우선 컴포넌트가 처음 실행될 때 한 번은 모두 동일하게 같이 실행된다. 그럼 첫 접속 때 useEffect가 실행되는 것이고, 전체 게시글을 사용자에게 보여주려면 PostData의 default 값으로 전체 게시글의 태그가 들어있어야한다. 그럼 서버는 DB에서 모든 모든 게시글의 데이터를 받아와 클라이언트에게 전달해줄 수 있을 것이다.

이렇게 setPostData()를 통해서 전체 게시글의 tag를 전달하게되면 한 가지 문제가 생긴다.

### 태그가 동적으로 변할 수 있다는 문제

게시글을 작성하면서 새로운 tag가 생길 수도 있고, 게시글을 삭제하게 되면 있던 tag가 사라지게 될 수도 있다.
근데 default 값은 미리 정적으로 입력해주어야하는데, 그렇게 동적으로 바뀐 태그 변경사항이 있다면 몇몇 태그는 DEFAULT 값에서 누락되거나 혹은 DB에 없는 태그가 접근 요청될 수 있다.

결국 첫 화면에 모든 POST를 전부 보여주지 못하고 몇 개의 POST는 누락될 수 있는 것이다. 

### 해결

그래서 PostData의 default 값으로 키 => TAGNAME, 값 => "ALL" 이라는 JSON 형식의 데이터를 할당했다.

![1](https://choigonyok.com/api/IMAGES/8-13.png)

첫 접속시에 tagname으로 ALL이 POST 요청 본문에 담겨 서버로 전송될 것이고, 서버에서 tagname이 ALL이면 전체 게시글을 반환하도록 로직을 구현하기로 했다.

GO 서버에서 로직을 수행하기 위해 TagData struct 타입의 변수 data를 선언해주었다.

```go
type TagData struct {
        Tags string `json:"tagname"`
}
```

구조체 TAGDATA는 이렇게 정의했다.

JSON형식의 tagname 키를 보면 Tags가 값을 받아오도록 미리 선언해주었다.

이제 tagname : "ALL" 의 처리 로직을 구현해보자.

![1](https://choigonyok.com/api/IMAGES/8-15.png)

ShouldBindJSON으로 JSON 형태의 데이터를 변수 data에 받아온다. 그리고 data에 저장된 키(tagname)의 값이 "ALL"이면 값을 "ALL"에서 빈 문자열 ""로 바꿔준다. GO는 STRING에서 ""을 0이 아니라 NIL(=NULL)로 취급한다. Tags를 빈 문자열로 반환한 이유는 다음과 같다.

원래 태그에 클릭 이벤트가 발생하면 해당 태그를 찾기위한 쿼리문을 작성해두었는데,

```go
db.query("SELECT ID,TAG,TITLE,BODY,DATETIME,IMGPATH FROM POST WHERE TAG LIKE `%"+data.Tags+"%`")
```

DATA.TAGS가 비어있는 상태이기 때문에 쿼리문 후반부의 조건절은 **WHERE TAG LIKE %%** 이렇게 설정되고, 이는 사실상 조건문이 없는 것과 마찬가지인 상태가 된다. 즉, 모든 게시글의 data를 전부 읽어올 수 있게 되는 것이다. 물론 키값이 "ALL"이 아니라 다른 특정 태그가 들어오면, 이 쿼리를 통해 해당 태그가 들어간 POST들을 찾게된다. 이 과정을 통해서 따로 전체 게시물을 위한 쿼리를 따로 작성하지 않고도 코드를 재사용할 수 있게 했다.

```go
response, err := JSON.marshal(PostData)
```

받아온 전체 게시글의 데이터를 marshal(JSON 형식으로 인코딩해주는 함수)을 이용해서 JSON으로 변환하고 response에 할당한다.

```go
c.wirter.HEADER().set("CONTENT-TYPE", "application/json")
C.wirter.write(response)
```
응답이 JSON 형식으로 간다고 헤더에 content-type을 명시해서 클라이언트에게 알려주고, 본문에 전체 게시글의 데이터가 담긴 response를 담아 응답하면 사용자가 홈페이지 루트로 접속했을 때, 전체 게시글을 다 확인할 수 있게된다. 

ALL 태그를 클릭해서 전체 게시글이 다 보이는 것과, 아무것도 클릭하지 않은 상태여서 전체 게시글이 다 보이는 것을 구분하기 위해서 

![1](https://choigonyok.com/api/IMAGES/8-16.png)

홈페이지로 라우팅됐을 때는 아무 태그가 안눌린 상태이기 때문에 태그버튼 위의 태그 표시를 CHOIGONYOK으로 지정해두었고,

![1](https://choigonyok.com/api/IMAGES/8-KakaoTalk_Photo_2023-06-20-19-40-23.png)

ALL 버튼을 누르면 제목이 ALL로 되도록 구현했다. 이를 통해서 사용자가 처음 접속하면 CHOIGONYOK을 보게되고, 다른 태그들 클릭하다가 전체 게시물이 보고싶어서 ALL 태그를 누르면 ALL을 볼 수 있도록 구현했다. |
