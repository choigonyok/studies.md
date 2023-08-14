# TRANSACTION IN GOLANG

---

## 개요

첫 프로젝트였던 블로그 프로젝트를 리팩토링하는 과정 중에 있다. 불과 3개월 전 진행한 프로젝트였지만 너무 부족한 점들이 많이 보였다. 리팩토링 관련한 글은 따로 작성하기로 하고, 리팩토링 과정 중 데이터베이스 트랙잭션을 Go로 구현하는 방법에 대해 알아보겠다.

---

## 문제점

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

이 핸들러 뿐만 아니라 전체적으로 DB의 트랜잭션이 필요한 부분들이 많은 것을 발견하게 되었다. 리팩토링 이전에 배포하고 운영할 때 INSERT 쿼리문에 오류가 생긴 적은 다행히 없다. 그러나 문제가 언제든 발생할 수 있는 여지가 있다는 것을 깨달았고, 백엔드에 작성된 Go에서 어떻게 트랜잭션 처리를 지원하는지 알아보겠다.

---

## Golang에서 트랜잭션의 활용

트랜잭션의 특징(ACID)등의 이론적인 내용보다는 글의 제목에 맞게, 어떻게 Golang에서 트랜잭션을 처리할 수 있는지 알아보겠다.

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

기존에는 이전 레코드를 삭제하고 새롭게 생긴 레코드를 추가하는 model 함수가 서로 분리되어있었다. model에서는 함수마다 하나의 작업만 하도록 해서 controller에서 다양하게 model 함수들을 조합해서 사용할 수 있게, 재사용성을 높히기 위해서였다. 이걸 UpdateCookieRecord() 함수 하나로 묶으면서 트랜잭션 처리가 가능해지도록 했다. 대신 코드의 재사용성은 조금 떨어질 수 있다.

db.Begin()은 트랜잭션 시작을 선언하는 함수이다.

시작 이후에는 평소 쿼리문을 구현하듯이 db.Query()나 db.Exec()를 이용해서 처리를 해주면된다. 다만 기존 쿼리문 구현과 다른 점은 err 발생 시 tx.Rollback() 메서드로 Begin()전 상태로 되돌릴 수 있다는 것이다. 트랜잭션이 정상적으로 처리가 되면 tx.Commit() 메서드로 처리되도록 한다. 마치 git의 commit처럼 Commit까지 해야 최종적으로 원하는 결과가 처리되는 것이다.

---

## Go의 nil

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

---

## Exec()와 Query()의 차이?

이번에 트랜잭션 구현에 대해 공부하며넛 Exec() 메서드를 처음 알게 되었다. 기존에 나는 Query() 메서드만 이용해서 DB의 조회, 수정, 추가, 삭제를 모두 구현했었는데, 둘의 차이를 한 번 알아보자.

Query()와 Exec() 메서드는 쿼리문을 실행한다는 공통점이 있다. 그리고 차이점으로는 사용 목적에 있고, 그 이유는 상당히 합리적이다.

Query() 메서드는 DB를 조회할 때 사용된다. 조회를 하면 *sql.Rows 타입이 반환되고, 이 타입은 *sql.Rows.Next()와 *sql.Rows.Scan() 메서드를 통해 쿼리문의 실행 결과값을 가져올 수 있다.

반대로 Exec() 메서드는 DB에 값을 추가하거나, 수정하거나, 삭제하는 등의 조회 이외 작업에 사용된다. Exec() 메서드는 반환하는 값이 없다. 만약 레코드를 삭제하는 쿼리문일 경우에 그냥 삭제만 하면 되지 굳이 반환 값을 보내서 리소스를 잡아먹을 필요가 없기 때문이다.

더 리소스를 아끼고 효율적으로 사용하기 위해 굳이 반환할 결과값이 없는 경우에는 Exec() 메서드를 쓰고, 조회의 결과값을 받을 포인터타입이 필요하면 Query() 메서드를 사용하면 된다. 물론 Query() 메서드로도 조회 이외 작업들이 가능하다. 리팩토링 이전까지의 내가 그렇게 코드를 작성했었다. 그러나 계속 말하는 것처럼 **굳이** 결과값이 필요없는 상황에서도 Query()를 통해 결과값을 위한 리소스를 할당하는 것이 좋은 코드는 아니기 때문에, 상황에 맞는 메서드를 사용하는 것이 권장된다.

---

## 참고

[GO의 zero-value](https://www.scaler.com/topics/golang/golang-zero-values/)