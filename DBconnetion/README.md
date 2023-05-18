##Problem

> problem.go로 DB connection을 공부하던 도중,
defer문을 적용에 대해 궁금한 점이 생겼다.
defer로 db.Close()를 먼저 호출하고, 나중에 r.Close()도 호출하는데,
main이 끝나기 전에 둘 중 어느 것이 먼저 호출될까

##Solution

> 테스트한 solution.go에서 
defer로 호출된 after가 먼저 출력되었다.

##Learn

> GO에서 defer가 쌓이는 방식은 stack과 같은 LIFO라고 볼 수 있겠다.

