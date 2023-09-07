# Golang 유닛, 통합, E2E 테스트
# study golang

---

## 개요

CI/CD 파이프라인은 일반적으로 빌드 -> 테스트 -> 배포의 순서로 이루어진다. 코드가 마스터 브랜치에 푸시되면, 자동으로 푸시된 코드를 빌드하고 빌드된 코드를 테스트하고, 테스트에 통과하면 운영서버에 배포한다. 이 파이프라인이 정상적으로 동작하는지 확인하기 위해 테스트 작업을 수행해야하는데, 아직 나는 테스트에 대한 기본 지식이 없었다.

TDD와 테스트 코드의 중요성에 대해서는 많이 들어봤지만 그게 정확히 무엇인지, 어떻게 작성하는 것인지에 대한 감이 잘 안잡혀서 테스트 코드의 종류와, Golang에서 테스트 코드를 작성하는 방법에 대해 알아보려고한다.

---

## 테스트 코드의 필요성

테스트 코드는 말그대로 코드를 테스트하기위한 코드이다. 테스트 코드는 왜 필요할까?

1. 어플리케이션에 대한 신뢰도 증가

개인적인 경험으로는, A기능이 잘 동작하는 걸 확인하고 B기능의 개발을 시작했다. B기능이 잘 동작하는 걸 확인하고 배포했는데, 배포하고보니 A기능에 문제가 생겼다. 분명 A기능이 잘 동작하는 것을 확인하고 B기능 개발로 넘어왔는데, 언제인지 모르게 A기능에 문제가 생긴 것이다.

이런 경험이 있다보니, 여러 수많은 기능들을 개발하면서도 "아 그 때 그 기능이 아직도 잘 동작할까?" 하는 불안함이 계속 있었다. 배포하고도 문제가 많았다. 모든 기능에 대해 직접 하나하나 테스트를 하려하다보니 불필요하게 너무 많은 시간을 낭비하고 있다는 생각을 했다.

테스트 코드를 통한 주기적인 테스트는 어플리케이션의 신뢰도를 높인다. 배포할 때도 상대적으로 자신있게 배포할 수 있게된다. 테스트 코드에 대해 더 빨리 알았더라면 개발 비용을 더 줄일 수 있었을 것이다.

2. 쉬운 디버깅

신뢰도 증가에 이어서, 테스트 코드는 디버깅 비용을 줄일 수 있다. A기능에 문제가 다시 생겼다는 것을 배포 이후에 알게된 나는, A기능을 수정하느라 A기능과 연관된 모든 코드들을 수정해야만 했다. 다행히 아니었지만, 그렇게 연관된 코드들를 수정하다가 생뚱맞게 다른 기능에 문제가 충분히 생길 수도 있는 상황이었다.

테스트 코드를 작성했다면 문제가 발생한 즉시 해결할 수 있었을 것이다. 그럼 A기능과 관련된 코드들을 수정할 필요 또한 없었을 것이다.

![img](http://www.choigonyok.com/api/assets/67-1.png)

위 그래프처럼 프로그램이 개발 완료 시점에 다다를수록 테스트 코드를 작성하지 않은 프로젝트는 그 비용이 기하급수적으로 증가하게된다.

많은 테스트 코드의 장점이 있지만 이 두 가지 이유만으로도 테스트 코드를 도입하기에 충분한 이유가 된다고 생각한다.

---

## 테스트의 종류

테스트는 크게 유닛 테스트, 통합 테스트, e2e 테스트로 나뉘고, 유닛 -> 통합 -> e2e의 순서로 테스트가 진행된다.

![img](http://www.choigonyok.com/api/assets/67-2.jpeg)

### 유닛 테스트

유닛 테스트는 소스 코드의 모듈(함수 등)이 의도된대로 작동하는가를 검증하는 테스트이다. 위 다이어그램처럼 세 종류의 테스트 중 가장 많은 양을 차지하지만 개별적인 테스트의 길이나 테스트 내용은 작아서 테스트의 실행시간이 빠르다.

브라우저의 쿠키 여부를 확인하는 함수가 있다면, 이 함수가 쿠키 여부에 따라 true/false를 잘 반환하는지를 확인하게 된다. **유닛테스트는 입력과 결과만을 테스트한다.** 쿠키가 있을 때(입력) true를 반환하는가(출력), 또는 쿠기가 없으면(입력) false를 반환하는가(출력)만을 확인한다는 것이다.

그 중간 로직의 복잡성이나 유효성은 상관하지 않고, 그냥 입력이 들어갔을 때 의도한대로 출력이 나오는가만을 테스트한다.

### 통합 테스트

통합 테스트는 유닛 테스트를 통과한 모듈들을 서로 연관있는 모듈들끼리 **통합**한다. 특정 기능을 수행하는 모듈 집합(통합된 모듈들)이 정상적으로 의도한대로 동작하는지를 테스트한다.

예를 들어, 게시글을 작성하면 DB에 저장되는 기능을 테스트한다고 가정하자. 게시글 작성 및 저장 모듈이 있다.

작성모듈은 사용자의 타이핑을 입력으로 받아서 입력값 그대로 출력하는 유닛테스트를 통과했다. 저장모듈은 문자열을 전달받아 DB에 저장하는 유닛테스트를 통과했다. 

만약 DB에서 필요한 문자열 형식이 10글자 미만인데, 게시글 작성 모듈이 20글자를 사용자로부터 입력받았다면, 두 모듈 모두 유닛테스트에는 통과했지만, 두 모듈은 통합한 기능은 의도한대로 작동하지 않는다.

이렇게 유닛 테스트를 통과한 모듈들을 묶어 더 큰 범위에서 테스트하며 추가적으로 발생될 수 있는 문제들을 테스트한다.

통합테스트를 수행할 때는 시나리오를 작성한다.

![img](http://www.choigonyok.com/api/assets/67-3.png)

앞서 언급한 예시대로라면, Test Case Objective는 **사용자가 입력한 게시글이 정상적으로 잘 저장되는지 확인**이 되고, Test Case Description은 **작성페이지에서 글을 입력 후 작성완료 버튼을 클릭**이 되고, Expected Result는 **post table에 제목, 내용, 작성자가 정상적으로 삽입**이 될 수 있겠다.


### E2E 테스트

E2E는 End-to-End를 의미한다. 어플리케이션을 시작하는 순간부터 종료까지 전체를 테스트하는 것이다.

쇼핑몰을 예로 들자면 사용자가 회원가입하는 것부터, 로그인, 검색, 장바구니 담기, 주문, 결제, 배송, 후기작성까지의 모든 과정을 테스트한다.

위의 피라미드 그래프처럼 테스트의 수는 적지만 테스트 시간이 오래걸리고, 문제가 생기더라도 전체적으로 테스트했기 때문에 어느 부분에서 문제가 발생한 것인지 디버깅하는데 상대적으로 오랜 시간이 걸린다.

장점으로는 사용자 관점에서 진행하는 전체적인 테스트이기 때문에, 통합테스트가 다 통과되었더라도 사용자 입장에서 불편을 느낄 수 있는 것들을 체크할 수 있다.

예를 들어, 쿠버네티스 클러스터로 분산 시스템을 운영하고 있다고 가정하자. 5개의 백엔드 중 4개가 문제가 생겼다면 사용자는 개발자가 예상했던 것보다 5배 느린 서비스 속도를 경험할 것이다. 이런 것들을 e2e를 통해 확인하고 환경적인 문제들을 수정해나갈 수 있다.

---

## UnitTest in GO

Go는 테스트를 하기 위한 기본적인 몇 가지 규칙이 있다. 테스트 파일은 테스트 하려는 소스파일 뒤에 _test를 붙여야한다. 컨벤션 개념이 아니라 문법같은 개념이다.

예를 들어, calculate.go 파일 안의 함수들을 유닛테스트 하고싶다면, 테스트 파일의 이름은 calculate_test.go로 지정해야 go가 인식할 수 있다.

유닛 테스트 예시를 살펴보자. 참고로 Go는 테스트를 위한 기본적인 testing 패키지를 내장하고있다.

```go
package hello

import "fmt"

func Hello() string {
	return "hello"
}
func HelloWorld(s string) string {
	n := 123
	fmt.Println(n)
	return "hello, " + s
}
```

Hello()는 hello를 리턴하는 함수, HelloWorld는 n에 저장한 123을 출력하고, 문자열 s를 받아서 hello, 뒤에 s를 붙여 리턴하는 함수이다.

이 함수를 테스트하는 테스트 파일은 아래와 같이 구성할 수 있다.

```go
package hello_test

import (
	"test/hello"
	"testing"
)

func TestHello(t *testing.T){
	s := hello.Hello()
	if s != "hello" {
		t.Error("TEST FAILED1")
	}
}
func TestHelloWorld(t *testing.T) {
	s := hello.HelloWorld("world")
		if s != "hello, world" {
			t.Error("TEST FAILED2")
		}
}
```

소스 코드보다 테스트 코드의 양이 더 많다. 우선 테스트 대상인 hello 패키지를 import하고, testing 패키지도 import한다.

테스트 함수 명명은 "Test+Test하는 함수명"이 컨벤션인데, 이건 지키지 않는다고 오류가 발생하진 않는다.

### TestHello

```go
hello.Hello()
```

hello 패키지의 Hello() 함수를 실행하고, 결과 값이 hello가 아니면 테스트 실패를 반환한다.

### TestHelloWorld

```go
s := hello.HelloWorld("world")
```

이 코드 역시 정상적으로 테스트에 통과할 것이다. 소스코드에 

```go
n := 123
fmt.Println(n)
```

이 부분이 있지만 이 부분은 테스트에 영향을 미치지 않는다. 앞서 말했듯이 유닛 테스트는 입력(world)에 알맞은 출력(hello, world)이 나오는지만 확인하기 때문이다.

### 테스트 실행

이제 소스 코드와 테스트 코드를 작성했으니 테스트를 실행해보자.

Go에서는 테스트 파일이 있는 디렉토리에서 아래 커맨드를 실행하면 테스트를 할 수 있다.

```
go test
```

대상 패키지의 코드 커버리지(전체 소스 코드 중 몇 퍼센트가 테스트 되었는지)를 확인하고싶으면 아래 커맨드를 실행하면 된다.

```
go test -cover .
```

이 커맨드를 실행하면

![img](http://www.choigonyok.com/api/assets/67-4.png)

이런 식으로 테스트의 결과, 테스트에 걸린 시간, 코드의 커버리지를 확인할 수 있다.

---

## IntegrationTest in GO

이번엔 통합테스트 예시를 살펴보자. Handler가 의도한대로 요청에 대한 응답을 수행하는지 테스트하는 코드이다. 테스트 전에 통합테스트 시나리오를 먼저 작성해보자.

Test Case Objective를 **요청을 받으면 핸들러함수가 정상적으로 응답하는지 확인**, Test Case Description을 **모킹 서버에서 요청을 보내가**, Expected Result를 **상태코드 200 응답과 본문 데이터 Hello, World! 응답**으로 설정했다고 가정하자.

```go
package main

import (
	"net/http"
	"test/hello"
)
func main() {
	http.HandleFunc("/", hello.HandlerFunction)
}
```

main.go에서 핸들러를 선언한다. "/" 경로로 오는 요청은 hello 패키지의 HandlerFunction으로 핸들링하겠다는 의미이다. 그럼 hello 패키지의 HandlerFunction을 확인해보자.

```go
package hello

import (
	"fmt"
	"net/http"
)
func HandlerFunction(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello, World!")
}
```

HandlerFunction이 정의되어있다. 요청이 들어오면 writer에 Hello, World!를 응답하는 간단한 핸들러이다. 이 핸들러를 테스트하기 위한 통합테스트 코드는 아래와 같다.

```go
package hello_test

import (
	"net/http"
	"net/http/httptest"
	"test/hello"
	"testing"
)

func TestHandler(t *testing.T) {
    req, err := http.NewRequest("GET", "/hello", nil)
    if err != nil {
        t.Fatal(err)
    }
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(hello.HandlerFunction)
    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("테스트 실패: 예상한 상태 코드 %v, 실제 상태 코드 %v", http.StatusOK, status)
    }
    expectedBody := "Hello, World!"
    if rr.Body.String() != expectedBody {
        t.Errorf("테스트 실패: 예상한 응답 본문 %v, 실제 응답 본문 %v", expectedBody, rr.Body.String())
    }
}
```

테스트를 하기위해 모킹 서버를 구현했다. 모킹은 실제가 아닌 테스트를 위해 구현한 가짜를 의미한다. 핸들러를 테스트하려면 그 핸들러에 요청을 보내고 응답을 받는 서버가 있어야한다. 서버의 요청은 http.NewRequest로, 서버의 응답은 httptest.NewRecorder로 구현했다.

가상의 요청을 보내기 위해 NewRequest로 /hello경로에 GET method 요청을 생성한다. GET 요청이라서 본문데이터 파라미터는 nil이다.

가상의 응답을 받고 응답 내용을 기록하기 위해 httptest 패키지의 NewRecorder 함수가 리턴한 rr을 responseWriter로 사용한다.

```go
handler := http.HandlerFunc(hello.HandlerFunction)
handler.ServeHTTP(rr, req)
```

그리고 HandlerFunction의 핸들러를 생성해서 핸들러가 responseWriter가 rr이고, request가 req인 요청을 처리하도록 ServeHTTP 메서드를 사용한다.

이러면 HandlerFunction이 req 요청을 받고 rr에게 응답하는 로직이 구현된다. 이제 테스트만 하면 된다.

rr.Code로 응답 데이터의 상태코드가 200인지를 확인한다. 만약 다르면 테스트에 통과하지 못한 것이고, 피드백을 출력한다.

그리고 기대하는 응답 본문데이터가 Hello, World!인데, 실제 응답한 본문 데이터가 다르면 또 피드백을 출력한다.

핸들러같이 여러 모듈들이 통합된 기능을 이렇게 통합테스트 시나리오대로 테스트할 수 있다.

---

## E2ETest in GO

E2E 테스트는 도구를 사용한다. 예를 들어 투두리스트 어플리케이션을 테스트한다고 가정하자.

Cypress가 E2E테스트를 진행하는 과정은 다음과 같다.

1. 브라우저에서 "작성하기"라는 버튼 찾기
2. "작성하기" 버튼 누르기
3. placeholder에 "Title"이라고 되있는 텍스트 박스 찾기
4. 그 텍스트박스에 ABC 입력하기
5. "저장하기"라는 버튼 찾기
6. "저장하기" 버튼 누르기
7. "할일 목록"이라는 버튼 찾기
8. "할일 목록" 버튼 누르기
9. 브라우저에서 "ABC" 문자열 찾기

이런 식으로 작성부터 저장, 저장한 글이 목록에 성공적으로 출력되는지까지를 확인할 수 있다. E2E 테스트에서 Go가 따로 해야할 역할은 없고, 이런 E2E 도구가 있다는 것과 작동과정을 설명하고 싶었다.

---

## 정리

개인적인 정리로는 유닛테스트는 입출력, 통합테스트는 모듈 간의 통신, E2E테스트는 UI와 응답시간에 대한 테스트가 주를 이루는 것 같다고 느꼈다.

다음 프로젝트에서 백엔드를 개발할 때, 테스트코드를 꼭 적용해봐야겠다.

---

## 참고

[Go에서 테스트 작성하기](http://golang.site/go/article/115-Go-%EC%9C%A0%EB%8B%9B-%ED%85%8C%EC%8A%A4%ED%8A%B8)

[5 reasons why testing code is great](https://www.educative.io/answers/5-reasons-why-testing-code-is-great)

[유닛 테스트](https://ko.wikipedia.org/wiki/%EC%9C%A0%EB%8B%9B_%ED%85%8C%EC%8A%A4%ED%8A%B8)

[Integration testing](https://www.javatpoint.com/integration-testing)

[What is Integration Testing? Definition, Tools, and Examples](https://www.simform.com/blog/what-is-integration-testing/)