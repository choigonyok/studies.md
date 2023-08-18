# [BLOG #4] 여러 파일을 한 번에 업로드 하는 법

---

## 개요

게시물 작성 시 이미지를 첨부하려면 여러 이미지 파일을 한 번의 요청으로 전달받아야한다. 이미지가 썸네일에만 사용된다면 단일 파일만 업로드하면 되지만, 게시글이 렌더링 될 때마다 게시글에서 사용되는 이미지도 함께 출력해야해기 때문에 모든 이미지를 업로드가 가능해야한다.

또 게시글 작성 버튼을 누를 때 글의 제목, 내용, 카테고리 부터 이미지까지 모든 데이터가 전송되어야하는데, 글 등은 text/plain 형식이고, 이미지는 multipart/form-data 형식이기 때문에 한 번에 같은 요청에 담아서 보낼 수가 없다.

단일 파일을 업로드하는 건 이전 블로그 프로젝트 Ver.1에서 multipart/form-data 형식으로 전달하고, Go로 작성된 백엔드에서는 PostForm 메서드를 이용해 업로드를 구현했다.

이 글은 Golang과 Golang으로 작성된 웹 어플리케이션 프레임워크인 Gin, 그리고 리액트를 사용해 구현한 글이다.

---

## 전체 흐름

이런 로직을 가지고 구현해보았다.

1. 사용자가 글과 이미지를 선택한 후 작성버튼에 클릭이벤트가 발생하면 json형식으로 변환할 수 있는 데이터(본문, 제목 등)만 가지고 먼저 DB에 데이터 저장한다.
2. json형식 데이터가 전송 완료된 이후에, 이미지 파일을 multipart/form-data형식으로 따로 전송하고, DB에 이미지경로 없이 저장되어있는 레코드에 이미지경로를 업데이트한다.
3. 만약 이미지기 전송되는 도중 오류가 생기면, 기존 이미지경로 없이 DB에 저장되었던 레코드를 삭제한다.

---

## 로직 1번

* 사용자가 글과 이미지를 선택한 후 작성버튼에 클릭이벤트가 발생하면 json형식으로 변환할 수 있는 데이터(본문, 제목 등)만 가지고 먼저 DB에 데이터 저장한다.

```js
<div>
    <input type="button" value="POST" onClick={postHandler}/>
</div>
```

태그, 제목, 본문, 작성일의 데이터는 usestate와 e.target.value로 업데이트 해주고, onClick 이벤트가 발생하면 postHandler 호출한다.

```js
const postHandler = () => {
    const postdata = {
        title: titleText,
        tag: tagText,
        datetime: datetimeText,
        body: bodyText,
    };
    axios
        .post("HOST+URL PATH", postdata)
        .then((response)=>{
            setUnlock(!unlock);
        })
        .catch((error)=>{
            console.log(error);
        })
}
```

postHandler에서는 json형식의 postdata를 만들어서 클릭 이벤트가 발생할 당시의 태그, 제목, 본문, 작성일 데이터를 정의해준다. 그리고 postdata 데이터를 본문에 넣어서 POST API 요청을 보낸다.

setUnlock() useState와 이미지 전송 axios 요청을 useEffect()에 넣으면서 동기적 실행을 구현했다. 텍스트데이터 전송을 한 후 이미지파일이 동기적으로 전송되어야하기 때문이다.

이미지파일 경로는 가장 최신에 저장된 글의 이미지 경로를 업데이트하는 쿼리로 DB에 저장된다. 만약 이미지 경로가 텍스트 데이터보다 DB에 먼저 저장되면 이미지는 엉뚱하게 방금 작성된 게시글 이전의 최신 게시글 이미지에 저장되고, 방금 작성한 글은 썸네일이 없는 NULL 상태로 존재하게 되며, 2번째로 최근에 작성된 게시글의 썸네일이 방금 업로드한 이미지로 변경되게 된다.

처음 코드를 짤 때는 발생할 수 있는 에러 처리를 아래와 같이 클라이언트에서 작성했다.

```js
if (titleText === "" || tagText === "" || dateText === "" || bodyText === "") {
        alert("작성되지 않은 항목이 존재합니다.");
} else {
        ...POST API 요청...
}
```

이런 식으로 작성하니 유효성 검사 로직이 클라이언트에 있기도 하고, 코드의 가독성도 떨어진다고 판단해서, 오류 처리는 백엔드에서 실행하고 프론트에서는 응답받은 상태코드를 바탕으로 사용자에게 피드백만 하도록 코드를 수정했다.

글 작성 시 따로 전송되는 이미지와 텍스트 데이터를 하나의 라우터로 처리하기 위해서 Go 백엔드에서는 param 메서드를 이용했다.

```go
eg.POST("/api/post/:type", func (c *gin.Context){
    ...
})
```

c.Param("type")으로 받은 string이 post면 텍스트데이터, img면 이미지파일로 구분해서 코드 중복을 피할 수 있게 했다.

전송받은 텍스트 데이터를 DB에 저장하기 위해 우선 가장 최근 포스팅된 게시글의 id를 조회해서 recentID에 저장한다.

```go
r, _ := db.Query("SELECT id FROM post order by id desc LIMIT 1")
var recentID int
r.Next()
r.Scan(&recentID)
```

이후에 recentID에 1을 더한 값을 id로 갖는 새로운 게시글 레코드를 저장한다.

> 이후 리팩토링 과정에서 id를 조회해서 1을 더하는 방식 대신 DB 테이블의 컬럼 설정 시 AUTO_INCREMENT를 사용하도록 리팩토링 했다.

```go
if strings.Contains(data.Body, `'`) {
    c.Writer.WriteHeader(400)
    return
} else {
    ...DB에 레코드 저장 로직...
}
```

Go의 sql package에서는 쿼리문을 수행할 때 문자열로 감싸야한다. 게시글을 '로 감싸든 "로 감싸든 `로 감싸든 어쨌든 뭔가로 감싸야하는데, 이 때 텍스트 데이터의 내용 중에 감싼 심볼이 들어가 있으면 쿼리문을 수행하다 오류가 생기게 된다. 그래서 본문에 해당 심볼이 있는지 확인하는 로직을 구현했다.

> 이후 리팩토링 과정에서 본문의 심볼을 찾아서 심볼 앞에 \를 더해주도록 변경한 후 DB에 저장하고, 게시글 데이터를 클라이언트로 전송할 땐 백슬래시 \와 붙어있는 '를 '로 변경해서 전송하는 방식으로 변경했다. 리팩토링 이후에 특정 심볼을 게시글에 사용하지 못하는 문제는 해결되었다.

---

## 로직 2번

* json형식 데이터가 전송 완료된 이후에, 이미지 파일을 multipart/form-data형식으로 따로 전송하고, DB에 이미지경로 없이 저장되어있는 레코드에 이미지경로를 업데이트한다.

위에서 언급했듯이, 클라이언트에서는 텍스트 전송 후 setUnlock useState를 통해 동기적으로 이미지 전송 요청을 보낸다.

```js
useEffect(()=>{
    if(!mounted.current) {
        mounted.current = true;
    } else {
        const formData = new FormData();
        for (let i = 0; i < img.length; i++) {
            formData.append("file", img[i])
        }
        axios
            .post("HOST+URL PATH", formData, {
                "Content-type": "multipart/form-data",
            })
            .then((response)=>{
                navigator("/");
            })
            .catch((error)=>{
                deleteWrongWrittenPost();
            })
    }
},[unlock])
```

POST 요청은 mounted.current가 true일 때만 실행된다. useEffect는 우선 기본적으로 컴포넌트가 마운트될 때 한 번 실행되는데, 이 때 요청이 전송되어버리면 사용자가 Writepage에서 아직 작성을 시작하지도 않았는데 빈 작성요청이 전송되게 된다. 이걸 막기 위해 처음 마운트 될 때 실행이 되지 않게 할 수 있다.

이를 통해서 이 useEffect 블록은 오직 unlock의 값이 변할 때만(텍스트데이터가 성공적으로 전송된 이후에만) 동기적으로 실행되게 된다.

else문 안을 자세히 보면, 

```js
const formData = new FormData();
for (let i = 0; i < img.length; i++) {
    formData.append("file", img[i])
}
```

여러 이미지 파일을 한 번에 전송해야하기 때문에, 사용자가 입력한 이미지의 수만큼 반복문을 돌면서 파일 데이터를 formData에 넣는다.

처음엔 for문 안의 이 "file"을 잘못설정해서 한참을 헤맸다. 이 file은 말하자면 키이다. 이 키를 통해서 서버에서 파일리스트를 인식할 수 있게된다. 밑의 Golang 코드에서 이 "file"이 어디서 쓰이는지 확인할 수 있다.

그리고 formData와 함께 POST 요청을 보낸다. content type이 일반적인 text/plain 이나 json형식이 아니기 때문에, 아래와 같이 요청할 때 content type을 multipart/form-data로 명시해줘야한다.

```js
.post("HOST+URL PATH", formData, {
    "Content-type": "multipart/form-data",
    withCredentials: true,
})
```

### CORS 오류와 OPTIONS 요청

그리고 요청 헤더에 withCredentials를 설정해주는데, 프론트엔드는 백엔드와는 다른 Origin이기 때문이다. 또 파일을 전송하는 건 일반적인 요청(json 등)이 아니기 때문에 크리덴셜을 true로 해주지 않으면 CORS 정책으로 오류가 발생한다. 이거 때문에도 정말 한참을 고생했다. 

서로 다른 오리진끼리 통신을 위해서는 서버에서 ORIGIN 설정을 해야한다. 프론트는 아무 오리진에나 요청을 보낼 수 있지만, 그 요청을 서버가 받으려면 서버에서 오리진 설정을 해야한다는 것이다. 서버에서는 보안을 위해 ORIGIN 설정, 허용할 API설정, 크리덴셜 설정 등을 할 수 있다. 원하는 ORIGIN에서만 요청을 받고, 원하는 API(PUT, DELETE, POST, GET 등)만 받고, 원하는 헤더 내용(Cookie, Content-type)만 받고, WithCredential을 허용하고 말고 하는 것은 다 보안을 위한 서버의 권한이다. 

앞서 말했듯 파일 전송은 일반적인 요청이 아니기 때문에, 서버에서 withCredential 설정을 true로 미리 해줘야하고, 브라우저에서는 실제 파일을 서버로 전송하기 전, 서버가 이 요청을 받을 수 있는 보안 설정 상태인지를 확인하기 위해 OPTIONS라는 API요청을 서버에 먼저 보낸다. 서버는 이에 응답하고, 만약 서버가 해당 요청을 받기에 알맞은 설정이 되어있지 않은 상태면 오류가 생기는데 이 오류가 CORS오류이다. 

위 코드의 withCredential: true는 백엔드로 파일을 전달하기 위한 것이고, 서버에서도 withCredential 설정 + Origin 설정 + POST 요청 승인 설정 등을 같이 해줘야 올바르게 파일이 백엔드로 전송될 수 있다.

---

이어서, 서버에서는 MultipartForm() 메서드로 파일을 받는다. MultipartForm 메서드는 *multipart.Form 타입을 리턴한다. 이 리턴값을 imgfile이라는 변수에 정의한다.

그리고, 가장 최근에 이미지파일 경로 없이(NULL인 채로) 작성되어있는 레코드의 id를 찾아서 해당 레코드의 이미지파일 경로 컬럼을 업데이트해준다.

이제 서버에 이미지파일들을 저장하기 위해서 파일을 읽어야한다.

```go
imgs := imgfile.File["file"]
```

여기서 아까 프론트엔드에서 설정한 키가 쓰인다. 파일들을 읽어서 imgs로 리턴한다. 이후에는 for문과 range 문법을 통해 파일을 하나씩 읽고 저장하는 로직을 구현할 수 있다.

```go
for i, v := range imgs {
    err := c.SaveUploadedFile(v, "IMAGES/"+v.Filename)
}
```

이렇게 하면 IMAGES 폴더 안에 파일들을 저장할 수 있다. v.Filename은 파일 객체의 이름을 가지고있는 필드이다.

---

## 로직 3번

만약 전송 중 오류가 발생해서 이미지 없이 요청이 전송된다면 서버에서는 MultipartForm() 메서드로 파일을 파싱할 때, 파일이 없으니까 에러가 리턴하게 된다. 클라이언트에서는 500 상태코드를 받으면 

```js
.catch((error)=>{
    deleteWrongWrittenPost();
})
```

위에서도 봤듯이 deleteWronglyWrittenPost 함수를 호출한다.

이 함수는 delete API 요청을 하는데, 이 요청을 받은 서버는 가장 최근에 게시된 게시글을 삭제하게 된다.