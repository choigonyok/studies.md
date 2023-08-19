# Gorilla Websocket 공식문서 번역 #1
# study golang

---

## 개요

블로그 프로젝트의 다음 프로젝트로 커플 채팅 서비스를 기획하던 중 HTTP 프로토콜의 단점과 이를 개선한 웹소켓 프로토콜의 실시간 통신에 대해 알게 되었다.

Go에서는 Gorilla 웹 툴킷을 이용해 웹소켓을 구현하는 것이 가장 일반적이다.

웹소켓을 어떻게 학습하면 좋을까 고민하다가, 가장 기본이 되는 공식 문서를 직접 번역하면서 전체적인 그림을 이해하고 프로젝트를 하면서 실전에 적용해야겠다고 생각했다.

> 개인적으로 공부하기 위해 번역한 글이라 의역이 다수 포함되어 있다.

> 붉은 박스 안의 내용은 나의 개인적인 해석이다.

---

## 설치

```go
go get github.com/gorilla/websocket
```

---

## Overview

웹소켓에서 메시지를 주고받는 방법은 크게 두가지로 나뉜다. *Conn 객체를 활용하거나, io.WriterCloser, io.Reader를 활용할 수 있다.

### *Conn 활용하기

Conn type은 웹소켓의 커넥션을 의미한다. 서버는 request 핸들러에서 Upgrader 객체의 Upgrade 메서드를 호출하면 *Conn 객체를 얻을 수 있다.

![img](http://www.choigonyok.com/api/assets/27-1.png)

요청 응답 메시지를 보내거나 받기 위해서 conn의 WriteMessage/ReadMessage 메서드를 사용하면 메시지를 **[]bytes 타입**으로 주고받을 수 있다.

![img](http://www.choigonyok.com/api/assets/27-2.png)

ReadMessage 메서드는 websocket.BinaryMessage나 websocket.TextMessage 값(아래 data message 섹션에서 자세히 설명)을 가지는 int 타입의 MessageType, []bytes 타입의 p, err를 리턴한다.

### io.WriterCloser, io.Reader 활용하기

메시지를 보내려면 Conn.NextWriter 메서드를 사용하면 io.Writecloser 인터페이스를 얻어서 writer에 Write하고 다 끝나면 close하는 식으로 메시지를 보낼 수 있고,

메시지를 받으려면 conn.NextReader 메서드로 io.Reader 인터페이스를 얻어서 io.EOF가 리턴될 때까지 읽으면 메시지를 받을 수 있다.

![img](http://www.choigonyok.com/api/assets/27-3.png)

---

웹소켓의 메시지는 크게 두 종류가 있다. Data message와 Control message이다. Data message 두 종류, Controll message 세 종류로, 총 다섯가지의 메시지 타입이 있다.

## Data Message

웹소켓 프로토콜은 text 메시지 / 바이너리 데이터 메시지를 구분하고, 텍스트 메시지는 UTF-8로 인코딩된다.

이 패키지는 **TextMessage**와 **BinaryMessage** 라는 const 상수를 사용하는데, 두 메시지 타입을 구분하기 위해 사용된다.

위 Overview에서 언급한 ReadMessage와 NextReader 메서드는 수신한 메시지의 타입을 리턴한다.

ReadMessage 메서드는 websocket.BinaryMessage나 websocket.TextMessage 값을 가지는 int 타입의 MessageType, []bytes 타입의 p, err를 리턴한다.

WriteMessage 와 NextWriter 메서드의 messageType argument는 보내는 메시지의 타입을 특정하는 것이다.

> 보낼 땐 뭘 보내는지 타입을 특정해서 보내고, 받을 땐 뭘 받았는지 타입을 리턴한다고 이해하면 될 것 같다.

둘 중에 텍스트 메시지 타입은 꼭 UTF-8로 인코딩되도록 프로그래머가 잘 짜야한다.

---

## Control Message

웹소켓 프로토콜은 3가지의 control message 타입을 정의한다.

    close, ping, pong

이 컨트롤 메시지를 보내려면 conn의 WriteControl, WriteMessage, NextWriter 메서드 중 하나를 사용해야한다.

컨트롤 메시지 전용인 WriteControl 메서드로도 보낼 수 있지만 컨트롤 메시지가 아니더라도 메시지를 보낼 수 있는 WriteMessage와 NextWriter 메서드로도 컨트롤 메시지를 보낼 수 있다.

### Close Message Handling

conn은 

1. SetCloseHandler메소드가 있는 핸들러 함수를 호출하고
2. NextReader/ReadMessage/message.Read 메서드에서 *CloseError를 리턴함으로써
3. 받은 close 메시지를 핸들링할 수 있다.

### Ping Message Handling

conn은 SetPingHandler 메서드가 있는 핸들러 함수를 호출해서 ping 메시지를 수신할 수 있다. default ping 핸들러는 pong 메시지를 보낸다.

### Pong Message Handling

conn은 SetPongHandler 메서드가 있는 핸들러 함수를 호출해서 pong 메시지를 수신할 수 있다. default pong 핸들러는 아무것도 안한다.

**그래도 ping 메시지를 보내면 일치하는 pong 메시지를 수신하기 위해서 꼭 pong 핸들러를 set 해야한다.**

> ping 메시지는 받으면 pong 응답을 해줘야하니까 ping handler는 수행하는 작업이 있지만, pong 메시지는 받으면 그냥 잘 연결이 되어있구나/안되어있구나를 확인하는 용도이기 때문에 pong handler의 역할이 없는 것 같다.

컨트롤 메시지 핸들러 함수는 NextReader/ReadMessage/r.Read 메서드에서 호출된다.

---

## Concurrency

conn은 하나의 reader와 하나의 writer의 동시성 작동을 지원한다.

> 논리적으로 writer와 reader가 동시에 작동하는 것처럼 만든다는 것이다.

개발자는 

1. write 메서드(NextWriter/setWriteDeadline/WriteMessage/WriteJSON/EnableWriteCompression/SetCompressionLevel)를 하나 이상의 고루틴이 호출하지 않도록 관리해야하고
2. read메서드  (NextReader, SetReadDeadline, ReadMessage, ReadJSON, SetPongHandler, SetPingHandler) 를 하나 이상의 고루틴이 호출하지 않도록 관리해야함

> 동시성 지원은 writer, reader당 하나씩만 지원하기 때문에!

* 참고로 Close/WriteControl 메서드는 다른 메서드들과 concurrent하게 실행 가능하다.

> 이 둘은 Read/Write 메서드들과 함께 고루틴으로 호출해도 상관 없다는 뜻

---

## Origin Considerations

웹브라우저는 JS 어플리케이션이 모든 host에 웹소켓 커넥션을 열도록 허용한다. 그래서 다른 오리진에 대한 접근을 막고 보안을 유지하는 건 서버쪽에서 **origin request header를 브라우저에 보냄**으로써 관리하고 책임져야한다.

Upgrader는 Origin을 체크하기 위해 CheckOrigin 필드값을 설정하는 함수를 호출한다.

1. CheckOrigin이 false면 Upgrader 메서드는 403 상태코드와 함께 커넥션에 실패한다.
2. CheckOrigin이 nil이면 Upgrader는 safe default를 사용한다.

Safe Default는 Origin Request Header가 존재하고 + Origin Host가 Host Request Header 와 같지 않으면 실패하게 하는 로직이다.

deprecated package-level의 Upgrade function은 오리진 체킹을 안한다. Upgrade function이 호출되기 전에만 Origin header를 체크하는 것이다.

---

## Buffers

커넥션 버퍼 네트워크는 메시지를 읽고 쓸 때 시스템 콜의 수를 줄이기 위해서 input/output한다.

추가로 Write buffer는 웹소켓 프레임을 구축하기 위해 사용되기도 한다.

> - RFC 6455, Section 5의 메시지 프레이밍을 참조
 
write 버퍼가 네트워크에 올라갈 때마다 네트워크에 웹소켓 프레임 헤더가 작성된다. write 버퍼의 사이즈를 줄이는 건 커넥션의 프레이밍 오버헤드를 증가시킬 수 있다.

버퍼의 Byte 사이즈는 Dialer와 Upgrader의 ReadBufferSize/WriteBufferSize 필드로 설정된다.

버퍼사이즈 필드가 0으로 설정되어있으면

1. Dialer는 4096바이트를 default로 사용한다.
2. Upgrader는 HTTP 서버에서 생성된 버퍼를 재사용한다. HTTP 서버의 버퍼는 작성될 때 4096 사이즈를 가지고 있다.

버퍼 사이즈는 커넥션에 의해 읽히거나 쓰여지는 메시지 사이즈를 제한시키지 않는다.

버퍼는 커넥션이 살아있는 동안 default로 열리다가 Dialer나 Upgrader의 WriteBufferPool 필드가 set되면 커넥션은 메시지를 쓸 때만 write 버퍼를 유지한다.

개발자는 메모리 사용량과 퍼포먼스를 위해서 버퍼사이즈를 잘 조정해야한다.

버퍼사이즈를 적절히 올리면 네트워크에 읽고 쓰는 시스템콜 수를 줄이는 이득을 볼 수 있다.

writing을 예로 들면, 버퍼 사이즈를 올리면 네트워크에 쓰여지는 프레임 헤더 수가 줄어들 수 있다.

> 매 데이터마다 바로바로 네트워크에 데이터를 보내는 게 아니라, 버퍼를 만들어두고 버퍼가 다 차면 한 번에 데이터를 보낸다. 데이터를 보낼 때마다 헤더가 포함된다.

> 이 버퍼 사이즈가 작으면 더 빠릿빠릿하게 데이터를 보내 실시간에 더 가까운 빠른 전송이 가능한 대신, 자주 데이터를 보내니까 네트워크에 오버헤드가 생길 수 있다.

> 그렇다고 크면 버퍼가 다 찰 때까지 기다려야하기 때문에, 극단적으로 예를 들면 채팅을 보냈는데 버퍼 차기를 기다리느라 2초 뒤에 채팅이 보내지게 되는 이런 상황이 생길 수 있다.

### Config Buffer Guideline

* 예상되는 최대 메세지 사이즈로 버퍼 사이즈를 제한하기

가장 큰 메시지 사이즈보다 큰 버퍼는 아무 장점이 없다.

메시지 사이즈 분산(중앙값이나 평균)을 생각해서 버퍼 사이즈를 예상되는 최대 메시지 사이즈보다 작게 잡는 건 퍼포먼스 감소는 적은데 반해서 메모리를 획기적으로 아낄 수 있다.

예를들면, 99%의 메시지는 256바이트를 넘지 않고, 가장 큰 메시지 사이즈가 512바이트면, 버퍼사이즈를 256바이트로 설정하는 건 버퍼사이즈를 512바이트로 설정하는 것보다 1.01배의 시스템콜을 일으킨다. 메모리 사용량은 50%인데!

> 1.01배 손해보면서 메모리 사용량은 반이나 줄이면 이득이란 뜻

write 버퍼 pool은 커넥션 수는 많은데 커넥션 안에서 write의 수는 그렇게 많지 않을 때 유용하게 사용할 수 있다.

버퍼가 pooled 상태면 큰 버퍼 사이즈는 전체 메모리 사용량에 대한 영향을 줄이고 시스템콜과 프레임 오버헤드를 줄이는 이득이 있을 수 있다.

> 버퍼를 pool에 넣어놓고 필요할 때 가져가서 write할 수 있게, 커넥션은 많아도 어차피 실질적으로 write하는 수는 많지 않으니까 굳이 버퍼를 많이 만드느라 시스템콜 많이 할 필요가 없는 것임!

---

## Compression Experimental

각 메시지를 압축하는 확장자는 제한된 범위 안에서 시범지원된다.

![img](http://www.choigonyok.com/api/assets/27-4.png)

커넥션 상대방과 메시지가 압축되도록 둘 다 설정이 잘 되면, 모든 메시지는 압축된 형태로 받아지고 자동으로 압축해제된다.

모든 read 메서드는 압축되지 않은 bytes를 리턴할 것이다.

> 자동으로 압축 해제 되니까!

conn에 write된 각 메시지의 압축은 

![img](http://www.choigonyok.com/api/assets/27-5.png)

이 메서드를 통해서 설정 취소할 수 있다.

---

## 정리

* 다음 글에서는 이번 글에 나온 주요한 메서드/함수/인터페이스 들에 대해 다룰 예정이다.

---

## 참고

[원문](https://pkg.go.dev/github.com/gorilla/websocket#section-readme)