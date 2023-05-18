##Problem

defer문을 적용하다 궁금한 점이 생겼다.
defer로 db.Close()를 먼저 호출하고, 나중에 r.Close()도 호출하는데,
main이 끝나기 전에 둘 중 어느 것이 먼저 호출될까? 그리고 왜 그럴까?

##Solution

