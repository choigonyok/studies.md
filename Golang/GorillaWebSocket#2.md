# Gorilla Websocket 공식문서 번역 #2
# study golang

---

## 개요

> 개인적으로 공부하기 위해 번역한 글이라 의역이 다수 포함되어 있다.

> 붉은 박스 안의 내용은 나의 개인적인 해석이다.

---

## Close code

```
const (
	CloseNormalClosure           = 1000
	CloseGoingAway               = 1001
	CloseProtocolError           = 1002
	CloseUnsupportedData         = 1003
	CloseNoStatusReceived        = 1005
	CloseAbnormalClosure         = 1006
	CloseInvalidFramePayloadData = 1007
	ClosePolicyViolation         = 1008
	CloseMessageTooBig           = 1009
	CloseMandatoryExtension      = 1010
	CloseInternalServerErr       = 1011
	CloseServiceRestart          = 1012
	CloseTryAgainLater           = 1013
	CloseTLSHandshake            = 1015
)
```

Close Code가 있다. 마치 HTTP 상태코드 같은 것이다. 정상적으로 close된 건지 아니면 다른 오류에 의해서 close된건지 Close Code를 통해 확인할 수 있다.

---

## Message Type

Close Code처럼 메시지 타입도 상수로 작성할 수 있다. 

> 메시지 타입은 다섯 종류가 있다고 번역 1탄에서 언급했었다.

    TextMessage = 1
    BinaryMessage = 2
    CloseMessage = 8
    PingMessage = 9
    PongMessage = 10

> 이 타입이 1탄에서 언급한 ReadMessage와 NextReader가 리턴하는 메시지 타입값이다.

1. TextMessage = UTF-8로 인코딩된 텍스트 데이터이다.

2. CloseMessage = CloseCode가 숫자 + 텍스트로 들어있다.

Close Message의 형식을 지정하려면 FormatCloseMessage 함수를 이용하면 된다.

3. Ping/PongMessage = 모두 UTF-8로 인코딩 된 텍스트 데이터를 담고있다.

---

## Variables 사전 정의되어있는 변수들

DefaultDialer
: HandShakeTimeout은 45초, Proxy는 http.ProxyFromEnvironment로 설정되어있는 Dialer

---

## 함수

    func FormatCloseMessage(closeCode int, text string) []byte

Close Code를 웹소켓 클로즈 메시지에 맞게 변환해준다. close code가 CloseNoStatusReceived면 ""를 리턴한다.

    func IsCloseError(err error, codes ...int) bool

Close에 에러가 있는지 확인한다.

    func IsUnexpectedCloseError(err error, expectedCodes ...int) bool

예상치 못한 Close 에러가 있는지 확인한다.

    func IsWebSocketUpgrade(r *http.Request) bool

클라이언트에서 웹소켓 프로토콜로 업그레이드를 요청했는지 여부를 확인한다.

    func JoinMessages(c *Conn, term string) io.Reader

받은 메시지들 각각 뒤에 term string을 붙여서 io.Reader가 읽을 수 있게 만들어준다. 이 함수로 만들어진 io.Reader는 Read 메서드를 하나씩만 처리할 수 있다.

> 한 번에 하나씩, 순서대로 읽고 고루틴 사용해서 동시에 읽는 게 불가능하다.

    func Subprotocols(r *http.Request) []string

클라이언트가 요청한 요청메세지의 Sec-Websocket-Protocol 헤더에 있는 서브프로토콜들을 리턴한다.

---

## 타입

    BufferPool interface

버퍼 풀을 의미하고 Get, Put 메서드로 풀에 넣거나 꺼내올 수 있다.

    CloseError struct

close message이고, code과 text 필드로 이루어져있다. Error() 메서드 사용이 가능하다.

    Conn struct

웹소켓 커넥션이다. Conn 구조체는 메서드가 많아서 자세히 알아보겠다.

### Conn의 메서드

    Close() error

close message를 보내거나 기다리지 않고 바로 커넥션을 close한다.

    CloseHandler(code int, text string) error

현재 close handler를 리턴한다.

    EnableWriteCompression(enable bool)

압축해서 쓰는걸 쭉 혀용할지 아니면 허용되있던 걸 끌지 설정한다. 상대방이랑 협의가 되어야한다.

    LocalAddr() net.Addr

로컬 네트워크 주소를 리턴한다.

    NextReader() (messageType int, r io.Reader, err error)
    
상대로부터 받은 다음 메시지를 리턴, messageType은 바이너리 아니면 텍스트이다. 

커넥션에서 open reader는 하나까지만 있을 수 있다. 이전 메시지를 아직 읽지 않았어도 메시지가 새로 오면 이전 메시지는 폐기된다. 

이 메서드가 nil이 아닌 에러를 리턴하면 바로 read loop를 빠져나와야한다. 에러가 한 번 리턴되면 이후에는 무조건 쭉 에러가 리턴되기 때문이다.

    NextWriter(messageType int) (io.WriteCloser, error)

메시지를 보내기 위한 writer를 리턴한다. writer는 커넥션마다 하나씩만 있을 수 있다

> 한 번에 하나씩 쓸 수 있다는 뜻 같다.

그래서 이전 Writer가 아직 다 write 안했는데 새로운 메시지를 보내려고 하면 이전 writer는 closed된다.

writer는 텍스트/바이너리/핑/퐁/클로즈 다 보낼 수 있다.

    PingHandler() func(appData string) error

현재 핑핸들러를 리턴한다.

    PongHandler() func(appData string) error

현재 퐁핸들러를 리턴한다.

    ReadJSON(v interface{}) error

JSON으로 인코딩된 메시지를 커넥션에서 읽고, v로 pointing된 곳에 저장한다.

    ReadMessage() (messageType int, p []byte, err error)

NextReader 메서드를 내부적으로 이용해서 reader 얻는 걸 도와주고, 그 reader에서 버퍼로 읽는 걸 도와준다.

    RemoteAddr() net.Addr

원격 네트워크 주소를 리턴한다.

    SetCompressionLevel(level int) error

이제부터 압축 레벨을 얼마로 할 건지를 설정한다. 물론 상대랑 압축하기로 협의가 되어야한다.

    SetCloseHandler(h func(code int, text string) error)

받은 close message를 핸들링하는 핸들러를 생성한다. default close hander는 받은 close 코드를 상대에게 다시 전달한다.

    SetReadDeadline(t time.Time) error

커넥션에서 읽기 제한시간을 설정한다. deadline이 지난 이후의 모든 read는 에러를 리턴하게 된다. 리턴값이 0면 아직 deadline 안지났다는 뜻이다.

    SetWriteDeadline(t time.Time) error

위와 동일한데 쓰기 시간을 제한한다.

    SetReadLimit(limit int64)

메세지를 읽을 최대 크기를 바이트로 설정한다. 크기를 넘기면 close 메시지를 보내고 ErrReadLimit 에러를 리턴한다.

    Subprotocol() string

커넥션에 협의된 프로토콜을 리턴한다.

    UnderlyingConn() net.Conn

현재 커넥션의 내부 net.Conn 객체를 리턴한다. 커넥션에 특정 플래그를 더 수정하고 싶을 때 사용한다.

    WriteControl(messageType int, data []byte, deadline time.Time) error

deadline과 함께 컨트롤 메시지를 write한다. 가능한 messageType은 핑/퐁/클로즈 메시지이다.

    writeJSON(v interface{}) error

v interface{}를 JSON 형식으로 인코딩해서 메시지를 작성한다.

> golang에서 v interface{}는 모든 타입이 다 가능한 걸 의미한다.

여기까지가 Conn struct의 메서드들이다.

---

다음 타입으로는 Dialer 타입이 있다.

    Dialer

웹소켓 서버와 연결하기 위한 옵션들을 필드로 가지고있다. Dialer은 메서드를 concurrent하게 호출해도 안전하다.

### Dialer의 메서드

    Dial(urlStr string, requestHeader http.Header) (*Conn, *http.Response, error)

내부적으로 DialContext를 호출해서 새로운 클라이언트 커넥션을 생성한다.

> 보통은 서버와 외부 클라이언트가 웹소켓 커넥션을 맺는데, Dialer/Dial을 사용하면 서버 내부적으로 자신과 웹소켓 연결을 맺을 수 있다.

    DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (*Conn, *http.Response, error)

새로운 클라이언트 커넥션을 생성한다. 이 때 requestHeader가 사용되는데,

1. origin을 설정하려면 - ORIGIN
2. 서브 프로토콜을 지정하려면 - Sec-WebSocket-<Protocol>
3. 쿠키를 설정하려면 - Cookie

를 지정해주면 된다.

response.Header도 사용되는데,

1. 선택된 서브프로토콜을 get하려면 - Sec-WebSocket-<Protocol>
2. 쿠키를 가져오려면 - Set-Cookie

를 지정해주면 된다.

context는 request와 Dialer에서 사용될 것이다.

만약 websocket handshake가 실패하면, ErrBadHandshake가 리턴되고 호출한 사람은 이 리턴을 보고 리다이렉트나, 인증/인가 등을 할 수 있다.

응답 본문은 전체 응답을 담고있지 않을 수 있고, close될 필요가 없다.

여기까지가 Dialer의 메서드이다.

---

다시 타입으로 넘어와서,

    PreparedMessage struct

메시지 데이터를 캐싱한다. 여러 커넥션이 존재할 때 PreparedMessage를 사용하면 효과적으로 메시지를 보낼 수 있다. 

특히 압축이 되는 상황일 때 유용한데, cpu나 메모리등의 리소스들이 압축을 한 번에 할 수 있게 하면서 리소스를 아낄 수 있다.

    NewPreparedMessage(messageType int, data []byte) (*PreparedMessage, error)

초기화된 preparedMessage를 리턴한다. 이를 통해 커넥션에 WritePreparedMessage 메서드로 메시지를 보낼 수 있다.

    Upgrader struct

업그레이더는 HTTP 커넥션을 웹소켓 커넥션으로 업그레이드하기 위한 파라미터들을 지정한다. 업그레이터 메서드들을 concurrent하게 써도 된다.

### Upgrader의 메서드

    Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*Conn, error)

HTTP -> 웹소켓 프로토콜로 업그레이드를 시킨다.

> 웹소켓 프로토콜은 HTTP 기반이다. 그래서 업그레이드라는 표현을 사용하는 것

클라이언트의 업그레이드 요청에 대한 응답메시지에는 응답헤더가 포함되어있고, 쿠키를 설정하기 위해 Set-Cookie를 써도 되고, 서버가 지원하는 서브프로토콜을 지정하기 위해서 Upgrader.Subprotocols를 직접 써도 된다.

만약 업그레이드가 실패하면 이 메서드는 클라이언트에 HTTP 에러 응답을 보낸다.

---

## 정리

이렇게 두 번의 글로 Gorilla Websocket 공식 문서 번역을 작성해보았다. 

실질적으로 이번 채팅 어플리케이션 개발에 쓰이지 않을만한 내용들도 많았지만, 전체적으로 번역하고 이해하면서 웹소켓 뿐만 아니라 네트워크에 대한 이해도 조금 더 된 것 같아 하길 잘했다는 생각이 든다.

---

## 참고

## [원문](https://pkg.go.dev/github.com/gorilla/websocket#section-readme)