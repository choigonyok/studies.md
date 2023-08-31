# #10. 백엔드 리팩토링하기
# project blog golang

---

## 개요

리팩토링 이전의 코드는 아래와 같다.

[github: 리팩토링 이전 레거시 코드](https://github.com/choigonyok/blog-project/blob/0c2a248ad7b5ba027516afe034e8c77e60b541f6/src/main.go)

아래 내용들을 적용하기로 했다.

    1. GO Convention 적용
    2. MVC패턴 적용
    3. API 경로 재작성
    4. 함수/변수명 정리
    5. 코드 재사용성 높이기
    6. ' 심볼 DB 저장 불가 이슈 해결
    7. DB 트랜잭션 구현
    8. today/total 기능 오류 해결
    9. recent posts 오류 해결

---

## GO Convention 적용

하나의 레거시 코드를 여러개의 디렉토리로 나눴다.

[기존 코드](https://github.com/choigonyok/blog-project/blob/0c2a248ad7b5ba027516afe034e8c77e60b541f6/src/main.go)

### cmd

main package가 위치하는 곳이다.

### development

인프라 프로비저닝을 위한 테라폼 tf파일과 쿠버네티스 manifest가 위치한다.

### internal

외부에서 사용하지 못하는 내부 패키지들을 모아두는 곳이다. controller와 model이 위치한다.

### build

빌드를 위한 파일들이 위치한다. Dockerfile과 Dockerfile.dev가 위치한다.

도커파일이 기존에는 루트 디렉토리에 있었는데, build 디렉토리 안으로 이동하게 되면서 도커파일로 이미지 빌드시 커맨드에 수정이 생겼다.

    docker build -t USER/REPO .

이렇게 빌드했었는데, 이대로 빌드 명령을 실행하면 도커파일이 위치한 build 디렉토리 내부만 COPY가 되게된다.

    docker build -t USER/REPO -f build/Dockerfile .

-f 옵션으로 도커파일이 build 디렉토리 안에 있음을 명시해주어서 context를 현재 루트디렉토리로 설정해 루트부터 전체가 빌드될 수 있게했다.

---

## MVC패턴 적용

기존 코드는 하나의 소스파일에 레거시하게 작성되어있었다.

앞으로도 운영될 실서비스이기 때문에 유지보수 비용을 줄이기 위해서 MVC패턴으로 리팩토링했다.

---

## API 경로 재작성

이전에 작성해둔 API 경로를 다시보니, 경로만 보고는 어떤 요청을 하는 건지 쉽게 이해할 수 없게 작성되어있었다. 예를들어,

    eg.GET("/api/post/comments/:postid", func (c *gin.Context){})

post안의 comments를 GET하는 요청인 건 알겠는데, postid는 왜 뒤에 붙어있지? 이런 상황이었다. 그래서 API 경로만 보고도 어떤 요청을 하는지 한눈에 파악할 수 있게 하기 위해 API 경로를 REST API convention에 맞게 재작성했다.

### API 수정

    POST : 데이터 저장
    GET : 데이터 가져오기
    PUT : 데이터 수정
    DELETE : 데이터 삭제

이 컨벤션을 지키지 않았다. 예시로,

    답글 작성 : eg.PUT("/api/reply/:commentid", func (c *gin.Context){})

PUT이 넣다라는 뜻이니까 작성을 위해 PUT을 썼다.

    댓글 삭제 : eg.POST("/api/post/comments", func (c *gin.Context){})

댓글을 삭제할 때, 사용자가 입력한 비밀번호를 전송해야하는데, DELETE는 요청과 함께 데이터 전송이 안된다는 걸 알고, 데이터 전송을 위해서 DELETE를 사용하지 않고 POST를 사용했다.

이런 API 요청을 알맞게 수정했다.

### 리소스 표현은 동사가 아닌 명사로

URL 경로에서 리소스의 표현은 동사가 들어가서는 안된다. 기존에는

    글 수정 : eg.POST("/api/mod/:param", func (c *gin.Context){})

수정을 한다는 것을 명시하고 싶어서 mod를 넣거나,

    로그인 : eg.POST("/api/login/:param", func (c *gin.Context){})

로그인 한다는 것을 명시하고 싶어서 login을 넣거나,

    글 삭제 : eg.DELETE("/api/post/delete:deleteid", func (c *gin.Context){})

글을 삭제한다는 것을 명시하고 싶어서 delete를 경로에 넣었었다. 이런 경로들을 전부 수정했다.

### 리소스는 복수형으로 사용

리소스 중 컬렉션은 복수형으로 사용한다. 기존 경로에는 단수로 작성된 부분이 많았다.

    태그와 일치하는 글 가져오기 : eg.POST("/api/tag", func (c *gin.Context){})
    글 보기 eg.GET("/api/post/:postid", func (c *gin.Context){})

tag는 tags로, post는 posts로 수정해주었다.

    답글 작성 : eg.POST("/api/reply", func (c *gin.Context){})

이런 경우는 특정 답글 하나를 작성하는 것이기 때문에 단수로 작성된 그대로 두었다.

---

## 함수/변수명 정리

### camelCase

기존 변수명들은 모두 lower로 작성되어있거나, 여러 단어가 합쳐진 변수명일 때는 _ 를 이용해서 단어를 구분했다. 

```go
imgfile, err := c.MultipartForm()
no_space_thumbnail := strings.ReplaceAll(wholeimg[0].Filename," ", "")
postdata := []SendData{}
```

Go에서는 변수명을 지을 때 camelCase로 짓는다. 그리고 ID, IMG, UUID등의 짧은 단어들은 의미를 확실히 하기위해 UPPER 케이스로 작성한다. 이 컨벤션에 맞게 전체적으로 수정했다.

```go
imgFile, err := c.MultipartForm()
noSpaceThumbnail := strings.ReplaceAll(wholeIMG[0].FileName," ", "")
postData := []SendData{}
```

### 변수명에 타입이 들어가지 않게

Go 컨벤션에서 변수명에는 slice나 map, string, int 등의 타입을 표기하지 않는다. Go는 최대한 짧게 작명하는 선호가 있다고 한다.

```go
commentSlice, err := model.SelectNotAdminWriterComment(postID)
replySlice, err := model.SelectReplyByCommentID(commentID)
```

위와 같은 코드들을

```go
comments, err := model.SelectNotAdminWriterComment(postID)
replies, err := model.SelectReplyByCommentID(commentID)
```

이렇게 수정했다.

### empty slice는 var 사용

Go에서는 변수를 선언할 때 최대한 var 사용을 안한다. 선언과 정의를 같이 하는 경우라면 var a = 2 대신, a := 2 로 표기한다.

대신 JSON을 binding 하는 등 empty 변수가 필요한 경우에는 var를 사용한다고 한다.

```go
posts := []model.Post{}
post := model.Post{}
```

위 코드는 JSON을 바인딩 하기위해 빈 struct를 선언하는 부분이었다. 이 부분을 아래와 같이 수정했다.

```go
var posts []model.Post
var post model.Post
```

---

## 코드 재사용성 높이기

### 쿠키 확인 함수

많은 handler function에서 admin 쿠키를 확인하는 같은 코드가 6번 중복되는 것을 확인했고, 재사용성을 높이고 관리가 편해지도록 따로 내부 함수로 분리했다.

```go
_, err := c.Cookie("admin")
if err == http.ErrNoCookie {
    fmt.Println("ERROR MESSAGE : ", err)
} else {
    ...
}
```

기존 이렇게 작성되어있던 6개의 코드를

```go
func isCookieAdmin(c *gin.Context) bool {
	inputValue, CookieErr := c.Cookie("admin")
	cookieValue, err := model.GetCookieValue(inputValue)
	if err != nil {
		return false
	}
	if CookieErr != nil || cookieValue != inputValue {
		return false
	}
	return true
}
```

쿠키를 확인해 bool로 여부를 알려주는 함수로 분리하고, 필요한 곳에서는 함수 호출을 통해 사용하도록 변경했다.

---

## ' 심볼 DB 저장 불가 이슈 해결

게시글의 제목, 태그, 본문 등의 텍스트에는 ' 심볼을 넣을 수 없었다. Go 백엔드 서버에서 DB에 게시글 관련 데이터를 저장하기 위해 아래와 같은 쿼리문을 작성했는데,

```go
_, err := db.Exec(`INSERT INTO post (tag, writetime, title, text) values ('` + tag + `', '` + writetime + `' ,'` + title + `','` + text + `')`)
```

mysql에서 문자열을 표현할 때, ' 심볼로 값을 감싸다보니 '이 값으로 들어가면 충돌이 생겨 글이 정상적으로 저장되지 않는 문제가 있었다.

이전에는 쿼리문을 통해 데이터를 저장하기 이전에, 텍스트 데이터들에 ' 심볼이 있는지 유효성 검사를 먼저하고, 심볼이 있으면 사용자에게 오류 피드백을 하는 방식으로 구현했었다.

' 심볼을 다른 특수기호로 바꾸고, 글을 불러올 때 그 특수기호를 다시 '로 바꾸는 방법도 가능했지만, 그럼 그 변경되는 특수기호를 사용하지 못하게 된다.

이 ' 심볼을 사용하지 못하는 것이 너무 불편해서 이 기능을 리팩토링했다.

해결방법은 의외로 아주 간단했다. \ 심볼 뒤에 오는 특수문자는 문자열로 인식된다. 그래서 'hello ' world' 는 오류가 생기지만 'hello \' world' 는 정상적으로 저장된다는 것을 알게되었다.

그래서 데이터를 저장하기 이전에 ' 심볼을 확인하고, \'로 변경해준 뒤에, 게시글을 불러올 때, \'를 다시 '로 변경하는 로직을 구현했다.

```go
data.Text = strings.ReplaceAll(data.Text, `'`, `\'`)
```

본문 데이터가 DB에 저장될 때는 위와 같이 '앞에 \를 붙이고,

```go
datas[0].Text = strings.ReplaceAll(datas[0].Text, `\'`, `'`)
```

불러올 때는 \를 다시 삭제해준다.

이렇게 원하는 기능을 구현할 수 있었다.

---

## DB 트랜잭션 구현

내가 처음 발견한 문제점은 이러하다. 아이디와 패스워드를 확인해서, 인증된 사용자(블로그이기 때문에 나)만 글을 작성, 수정, 삭제할 수 있도록 하기 위해 작성했던 model.DeleteCookieRecord()와 model.InsertCookieRecord()를 리팩토링 하고있었다. 리팩토링 이후의 코드는 아래와 같다.

    func DeleteCookieRecord() error {
        _, err :=  db.Query("DELETE * FROM cookie")
        return err
    }

    func InsertCookieRecord(cookieValue string) error {
        _, err := db.Query(`INSERT INTO cookie (value) VALUES ("`+cookieValue+`")`)
        return err
    }

흐름은 다음과 같다.

1. 익명구조체를 만들어서 프론트엔드의 요청을 JSON형식으로 바인딩한다.
2. 환경변수에 정의된 ID, PW와 input된 ID, PW를 비교한다.
3. 올바른 ID, PW이면 **DB cookie table의 레코드를 삭제**한다. 
4. uuid를 생성하고, 쿠키의 값으로 설정해 프론트엔드(브라우저)에 응답한다.
5. 잘 쿠키가 보내졌으면 나중에 비교할 수 있도록 **cookie 테이블에 레코드를 추가**한다.

2번의 ID, PW를 비교하는 로직을 토큰, 세션등을 사용하지 않고 환경변수와의 비교로 설정한 것은 이 블로그는 나만 사용하는 블로그이기 떄문에, 굳이 불필요한 로직을 추가하는 것이 옳지 않다고 판단했다.

3번의 cookie table은 value column 하나만을 가지고있으며 로그인시마다 table의 레코드를 초기화하도록 했는데, 나만 사용하지만 그래도 여러 다른 기기에서의 접속을 막기 위해서 이런 방식으로 구현했다.

이렇게 잘 리팩토링하고 넘어가려고 하다가 문제점을 발견했다. 만약 레코드를 삭제한 이후 추가 중 오류가 발생하면, 레코드는 삭제만 되고 새롭게 추가되지 못한 상태로 남아있게 된다. 내가 원하는 건 삭제가 될 거면 추가도 되고, 둘 중 하나가 안될거면 둘 다 안되기를 원했다. 이걸 가능하게 하는 것을 트랜잭션 처리라고 한다.

이 핸들러 뿐만 아니라 전체적으로 DB의 트랜잭션이 필요한 부분들이 많은 것을 발견하게 되었다. 리팩토링 이전에 배포하고 운영할 때 INSERT 쿼리문에 오류가 생긴 적은 다행히 없다. 그러나 문제가 언제든 발생할 수 있는 여지가 있다는 것을 깨달았고, 이 부분을 리팩토링 하기로 했다.

```go
func UpdateCookieRecord() (uuid.UUID, error) {
  tx, err := db.Begin()
  if err != nil {
      return uuid.Nil, err
  }
  _, err =  db.Exec("DELETE * FROM cookie")
  if err != nil {
      tx.Rollback()
      return uuid.Nil, err
  }
  cookieValue := uuid.New()
  _, err = db.Exec(`INSERT INTO cookie (value) VALUES ("`+cookieValue.String()+`")`)
  if err != nil {
      tx.Rollback()
      return uuid.Nil, err
  }
  err = tx.Commit()
  if err != nil {
      return uuid.Nil, err
  }
  return cookieValue, nil
}
```

기존에는 이전 레코드를 삭제하고 새롭게 생긴 레코드를 추가하는 model 함수가 서로 분리되어있었다. model에서는 함수마다 하나의 작업만 하도록 해서 controller에서 다양하게 model 함수들을 조합해서 사용할 수 있게, 재사용성을 높히기 위해서였다. 이걸 UpdateCookieRecord() 함수 하나로 묶으면서 트랜잭션 처리가 가능해지도록 했다. 대신 코드의 재사용성은 조금 떨어질 수 있다.

db.Begin()은 트랜잭션 시작을 선언하는 함수이다.

시작 이후에는 평소 쿼리문을 구현하듯이 db.Query()나 db.Exec()를 이용해서 처리를 해주면된다. 다만 기존 쿼리문 구현과 다른 점은 err 발생 시 tx.Rollback() 메서드로 Begin()전 상태로 되돌릴 수 있다는 것이다. 트랜잭션이 정상적으로 처리가 되면 tx.Commit() 메서드로 처리되도록 한다. 마치 git의 commit처럼 Commit까지 해야 최종적으로 원하는 결과가 처리되는 것이다.

### Go의 nil

참고로, uuid.Nil은 타입을 의미한다. 그냥 nil을 반환하면 오류가 생긴다. 이건 nil이 uuid.UUID 타입을 지원하지 않기 때문인데, go에서의 nil은 아래와 같은 타입들을 지원한다.

* 포인터
* 맵
* 슬라이스
* 채널
* 인터페이스

만약 string을 반환하는 함수의 에러처리를 할 때, nil과 같은 zero-value를 표현하고 싶다면 nil이 아닌 ""로 표현해야한다.

    if err != "" {
        ...
    }

또 다른 예로, int 역시 지원하지 않아서 nil대신 0으로 zero-value를 표현해야한다. go에서 해당 타입들 이외에 어떤 zero-value를 써야하는지 아래 참고 링크를 통해 확인할 수 있다.

### Exec()와 Query()의 차이?

이번에 트랜잭션 구현에 대해 공부하며넛 Exec() 메서드를 처음 알게 되었다. 기존에 나는 Query() 메서드만 이용해서 DB의 조회, 수정, 추가, 삭제를 모두 구현했었는데, 둘의 차이를 한 번 알아보자.

Query()와 Exec() 메서드는 쿼리문을 실행한다는 공통점이 있다. 그리고 차이점으로는 사용 목적에 있고, 그 이유는 상당히 합리적이다.

Query() 메서드는 DB를 조회할 때 사용된다. 조회를 하면 *sql.Rows 타입이 반환되고, 이 타입은 *sql.Rows.Next()와 *sql.Rows.Scan() 메서드를 통해 쿼리문의 실행 결과값을 가져올 수 있다.

반대로 Exec() 메서드는 DB에 값을 추가하거나, 수정하거나, 삭제하는 등의 조회 이외 작업에 사용된다. Exec() 메서드는 반환하는 값이 없다. 만약 레코드를 삭제하는 쿼리문일 경우에 그냥 삭제만 하면 되지 굳이 반환 값을 보내서 리소스를 잡아먹을 필요가 없기 때문이다.

더 리소스를 아끼고 효율적으로 사용하기 위해 굳이 반환할 결과값이 없는 경우에는 Exec() 메서드를 쓰고, 조회의 결과값을 받을 포인터타입이 필요하면 Query() 메서드를 사용하면 된다. 물론 Query() 메서드로도 조회 이외 작업들이 가능하다. 리팩토링 이전까지의 내가 그렇게 코드를 작성했었다. 그러나 계속 말하는 것처럼 **굳이** 결과값이 필요없는 상황에서도 Query()를 통해 결과값을 위한 리소스를 할당하는 것이 좋은 코드는 아니기 때문에, 상황에 맞는 메서드를 사용하는 것이 권장된다.

---

## total/today 기능 오류 해결

기존 당일 방문자와 총 방문자 수를 구하는 로직은 아래와 같았다.

```
visitnum := 0
totalnum := 0
_, err := c.Cookie("visitor")
if err == http.ErrNoCookie {
    visitnum += 1
    totalnum += 1
    c.SetCookie("visitor", "OK",60*60*24,"/","",false,true)	
} 
```

클라리언트가 visitor cookie를 가지고있는지 확인한 후 쿠키가 없으면 constant인 today, total을 1씩 증가시키고 클라이언트에게 24시간 기한의 쿠키를 전송한다.

이 로직은 정확히 내가 원하는 방식대로는 동작하지 않았다. 내가 원하는 것은 매일 자정에 today가 초기화되며 클라이언트들의 today 기록도 초기화되는 것이었다. 

예를 들어 클라이언트 A가 23:59에 한 번 접속하고, 2분 후인 00:01에 접속하면 2분 전과 2분 후는 자정이 지나 다른 날짜이기 때문에, 두 방문 모두 today에 반영되도록 하는 것이었다.

그러나 기존 방식은 24시간이라는 limitation이 있어서, 23:59에 방문한 사용자는 다음날 23:59이 넘어야지만 새로운 방문으로 인정이된다.

게다가 매일 자정에 total을 초기화하기 위해서 타이머를 사용했는데, 아래와 같이 구현했다.

```
t := time.NewTicker(time.Minute * 60 * 24)
go func() {
    for range t.C {
        visitnum = 0
    }
    }()
```    

이 로직은 내가 원하는대로 동작할 수 있긴 하지만, 그러려면 어플리케이션의 배포를 정확히 정각에 시작해야한다. 그래야 정각에서 24시가 지날때마다(정각마다) today를 초기화시킬 수 있기 때문이다.

가능이야하지만, 즉각적인 업데이트가 어려워진다. 당장 새로운 업데이트를 적용해서 배포하고싶은데, 실제 배포를 하려면 그날 자정까지 기다려야한다. 혹시라도 1분이라도 늦게되면 다음날 자정까지 기다려야한다.

이런 부분들에서 리팩토링이 필요하다고 판단했다.

리팩토링 후

```
cookieNeed := false
today := getTimeNow().Format("2006-01-02")
cookieValue, err := c.Cookie("visitTime")
if err == http.ErrNoCookie {
    cookieNeed = true
} else {
    isValid := strings.Contains(cookieValue, today)
    if !isValid {
        cookieNeed = true
    }
}
if cookieNeed {
    visitor, _ := model.GetVisitorCount()
    model.AddVisitorCount(visitor)
    c.SetCookie("visitTime", getTimeNow().String(), 0, "/", os.Getenv("HOST"), false, true)
}
todayRecord, _ := model.GetTodayRecord()
if today != todayRecord {
    model.ResetTodayVisitorNum(today)
}
visitor, _ := model.GetVisitorCount()
```

로직을 구현하기 위해 DB에 visitor table을 생성했다. 컬럼으로는 today, total 그리고 오늘 날짜를 저장하고있는 date 컬럼을 생성했다.

우선 서버에서 오늘 날짜를 getTimeNow()를 통해 가져와 today를 초기화한다. getTimeNow()는 서버의 시간을 서울 표준 시간대로 일치시킨 후, 현재 시간을 가져오도록 작성된 내부 함수이다.

그리고 클라이언트로부터 방문 날짜가 기록되어있는 쿠키를 받아온다. 방문 쿠키가 없으면 cookieNeed를 true로 설정한다. 있더라도 쿠키 값인 방문날짜가 today와 같지 않다면 cookieNeed를 true로 설정한다.

cookieNeed가 true라면, 오늘의 새로운 방문자이고, 오늘 날짜에 맞는 새로운 쿠기 발급이 필요하다는 의미이다. true인 클라이언트는 현재 visitor table의 today 값을 가져와 1을 더해 업데이트한다. 그리고 쿠키를 발급한다.

만약 cookieNeed가 false라면 오늘 이미 한 번 이상 방문한 클라이언트라는 의미이기에 today 추가나, 쿠키발급 등의 작업은 이루어지지 않는다.

그리고 클라이언트가 블로그에 접속할 때마다, 현재 visitor 레코드의 date 값을 가져와서 todayRecord를 초기화하고, 실제 오늘 날짜와 todayRecord 값이 다르면 todayRecord를 오늘로 변경해주면서 today를 1로 초기화시킨다.

전체적인 흐름은, 사용자 방문시 DB에 저장되어있는 날짜와 실제 현재 날짜 비교, 만약 값이 다르면 하루가 넘어갔다는 거니까 today 1로 리셋하고 날짜 오늘로 최신화한다.

사용자 쿠키를 확인하고, 쿠키가 없거나 쿠키 날짜가 오늘과 다르면 쿠키 재발급하고, today 1 증가시킨다.

이런 로직으로 매 자정마다 today를 초기화시킬 수 있다. 엄밀히 말하면 자정이 지난 이후 첫 방문자가 생겼을 때 today가 초기화된다. 마치 슈뢰딩거의 고양이. 그러나 사용자 경험 측면에서 사용자는 자정마다 today가 초기화되는 것과 동일한 경험을 하게된다.

---

## recent posts 오류 해결

배포 이후 recent posts에 같은 게시물이 중복해서 나오는 버그를 확인했다.

원인을 분석해보니, recent posts를 프론트엔드에서 요청할 때 현재 클라이언트가 보고있는 글의 태그와 동일한 태그를 하나라도 가지고있는 글들을 응답했다.

그런데 만약 태그가 2개 이상 동일하다면, 현재 보고있는 게시글의 첫 번째 태그와 같아서 한 번 응답되고, 두 번째 태그와 같아서 또 한 번 더 응답되는 일이 발생했던 것이었다.

그래서 모든 태그에 부합하는 글들을 slice에 모은 후, 게시글들의 id를 기반으로 중복 검사를 실시해서 중복되는 게시글 없이 응답하도록 아래 코드를 추가했다.

var realdata []model.Post
allKeys := make(map[int]bool)
for _, v := range datas {
    if _, value := allKeys[v.ID]; !value {
        realdata = append(realdata, v)
        allKeys[v.ID] = true            
    } 
}

이후 정상적으로 기능 구현되는 것을 확인할 수 있었다.

---

## 참고

[GoLang Naming Rules and Conventions](https://medium.com/@kdnotes/golang-naming-rules-and-conventions-8efeecd23b68)
[Go Naming Conventions](https://www.smarty.com/blog/go-naming-tutorial)
[Practical Go: Real world advice for writing maintainable Go programs](https://dave.cheney.net/practical-go/presentations/qcon-china.html#_choose_identifiers_for_clarity_not_brevity)
[Go : zero-value](https://www.scaler.com/topics/golang/golang-zero-values/)