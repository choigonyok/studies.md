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
    10. reply 삭제 오류 해결

---

## GO Convention 적용

하나의 레거시 코드를 

[기존 코드](https://github.com/choigonyok/blog-project/blob/0c2a248ad7b5ba027516afe034e8c77e60b541f6/src/main.go)

---

## MVC패턴 적용

기존 코드는 하나의 소스파일에 레거시하게 작성되어있었다.

[기존 코드](https://github.com/choigonyok/blog-project/blob/0c2a248ad7b5ba027516afe034e8c77e60b541f6/src/main.go)

너무 강한 결합으로 묶이고, 유지/보수 관리

---

## API 경로 재작성

---

## 함수/변수명 정리

---

## 코드 재사용성 높이기

---

## ' 심볼 DB 저장 불가 이슈 해결

---

## DB 트랜잭션 구현

---

## total/today 기능 오류 해결

---

## recent posts 오류 해결

---

## reply 삭제 오류 해결

---

## 정리
