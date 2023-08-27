# #5. 채팅 기능 구현
# project couple-chat

---

## 개요

채팅 기능을 구현했다. 코드의 오류처리는 글의 가독성을 위해 글에서는 삭제시켰다.

---

## 웹소켓 업그레이드

실시간 통신에 이용되는 웹소켓 프로토콜은 HTTP를 기반으로한 프로토콜이다. 채팅을 하기 위해선 HTTP 프로토콜을 Websocket 프로토콜로 업그레이드 해주어야한다.

```go
var upgrader  = websocket.Upgrader{
		WriteBufferSize: 1024,
		ReadBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == os.Getenv("ORIGIN")
		    },
	}
```

Write/ReadBufferSize와 CheckOrigin은 이 블로그의 게시글인 **"웹소켓 공식문서 번역 #2"**에서 확인할 수 있다.

여기서 간단하게 말하자면 버퍼의 크기를 적절히 설정해줘야 웹소켓 통신의 성능이 보장되고, 웹소켓 통신에서 Origin을 체크하는 건 서버가 담당하는 역할이기에 올바론 요청인지 서버에서 유효성을 검사한다.

## 커넥션 생성

```go
conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
defer conn.Close()
defer func(){
  mutex.Lock()
  conns[uuid] = nil
  mutex.Unlock()
}()
```

websocket.Upgrader로 리턴받은 upgrader의 Upgrade메서드를 사용해서 conn 객체를 얻는다.

이 conn 객체를 통해 메시지를 읽고 쓰는 것 관련한 작업들이 모두 이루어진다.

### conns map

mutex lock으로 둘러쌓여있는 conns는 string 타입인 uuid를 키, *websocket.Conn을 값으로 갖는 map 배열이다.

내가 헷갈렸던 웹소켓에서 중요한 개념은 **커넥션은 상대방과 맺는 것이 아니라, 클라이언트와 서버 간에 맺는 것이라는 것이다.** 2인 채팅서비스면 채팅 이용자들끼리 커넥션을 맺는게 아니라 각각 서버와 커넥션을 맺고, 서버가 실시간으로 알맞은 사용자에게 채팅을 전달해주는 방식이다.

따라서 내가 보낸 채팅이 다른 사람이 아닌 나와 상대방에게만 도착하도록 하려면, 나의 커넥션과 상대방의 커넥션이 관련이 있다는 걸 서버가 알 수 있게 해야한다.

이 때 conns map이 사용된다. 커넥션이 종료될 때 defer문으로 conns map의 value를 nil로 변경하는 과정을 뮤텍스 락으로 감싼 이유는 아래 메시지 **송/수신 섹션의 target_conn**에서 설명하겠다.

```go
var conns = make(map[string]*websocket.Conn)
```

conns map은 사용자 별 커넥션 확인을 위해 필요하다.

A가 채팅을 보냈을 때, 

* A와 연인이 누구인지(B라고 가정)
* A와 B의 커넥션이 무엇인지

를 알아야한다. 그래야 올바르게 채팅을 전달해줄 수 있다.

conns map이 없다면, 연인 A가 B에세 채팅을 쳐도, 서버가 A로부터 받은 채팅을 어느 커넥션으로 전송해야할 지 알 수가 없다.

이 conns[uuid] = nil을 defer로 설정해둔 이유는 커넥션이 종료되면 conns map에도 그 상태를 반영해줘야 상대방에게도 채팅을 보낼지, 나한테만 보낼지를 판단할 수 있게된다.

만약 상대방과의 커넥션이 끊겼는 줄 모르고 채팅을 보내게 된다면, 서버는 존재하지 않는 커넥션에 채팅을 보내려고 한다며 오류를 피드백할 것이다.

### 커넥션이 끊긴 상태에서 채팅을 보내면?

그럼 상대방의 웹소켓 연결이 끊긴 상태에서 내가 채팅을 보내면 그 채팅은 나만 보고 사라지는 것인가?

그렇지 않다.

모든 채팅은 웹소켓 연결을 통해 실시간으로 전송된 이후에 연인 별 DB에 모두 저장이 된다. 

사용자가 웹소켓 연결이 끊어진 후 다시 웹소켓에 연결되면, default로 사용자가 보낸 채팅, 또 상대방이 보낸 채팅이 DB로부터 로딩되어서 화면에 표시하게된다.

## 채팅 작성자 구분

```go
jsonUUID := struct {
  UUID string `json:"uuid"`
}{
  uuid,
}
_ := conn.WriteJSON(jsonUUID)
```

이후에 커넥션에 uuid를 전달해서 나의 uuid가 무엇인지 파악할 수 있게 한다. 다행인 것은 2인 채팅 서비스이기 때문에, 내 uuid로부터 온게 아니면 상대방으로부터 온 것이라고 판단하고,

내 채팅은 오른쪽에 하늘색 바탕으로, 상대방의 채팅은 왼쪽에 분홍색 바탕으로, 작성자를 구분하여 표시할 수 있게 된다.

---

## 기존 채팅 불러오기

```go
first_uuid, second_uuid, conn_id, _ := model.GetConnectionByUsrsUUID(uuid)
initialChats, _ := model.SelectChatByUsrsUUID(first_uuid, second_uuid)
```

GetConnectionByUsrsUUID(uuid) 함수는 나의 uuid를 바탕으로 상대방의 uuid, conn_id를 리턴하는 함수이다.

나의 uuid는 이미 알고있는데 굳이 first_uuid, second_uuid 모두를 리턴하는 이유는 DB 레코드의 재사용성을 위함이다.

만약 내 기준으로 상대방의 uuid만 알게하려면, 각 사용자마다 기준이 되는 레코드가 따로 있어야한다. 그 대신 양 쪽의 uuid를 다 가지고있는 레코드 하나로 통합함으로써, 내가 조회해도 이 레코드가 나오고, 상대가 상대의 uuid로 조회해도 이 레코드를 리턴할 수 있다.

이후에 first_uuid와 second_uuid 중 어느 것이 나/상대방인지는 string 비교를 통해 쉽게 알 수 있다.

그렇게 상대의 uuid를 알게 된 후, 

```go
SelectChatByUsrsUUID(first_uuid, second_uuid)
```

를 통해서 DB에 저장되어있던 나와 상대방이 작성한 채팅을 slice 형태로 initialChats 변수에 받아온다.

---

## 답변 안한 question 확인

내가 답변하지 않은채로 커넥션이 종료된 question이 있는지를 확인한다.

이 채팅 서비스의 핵심인 question 기능은, 채팅 중 나온 특정 단어를 기반으로 제공되는 공통 질문에 대한 답변을 하지 않으면 상대방과의 채팅을 이어갈 수 없게 한다.

이 강제성을 통해 질문에 대한 답변을 하게하고, 나와 상대방의 답변 내용은 따로 저장되어서 서로의 생각이나 가치관을 언제든지 확인할 수 있게 도와준다.

만약 question이 생긴채로 커넥션이 종료되더라도, 추후에 다시 커넥션에 연결(채팅방에 입장)했을 때, 답변하지 않은 question부터 답변을 하도록 하는 기능이다.

```go
order, _ := model.GetUsrOrderByUUID(uuid)
question_id, _ := model.QuestionIDOfEmptyAnswerByOrder(order, conn_id)
if question_id != 0 {
  _, questionContents, _ := model.GetQuestionByQuestionID(question_id)
  questiondata := model.ChatData{
    Text_body: questionContents,
    Writer_id: "question",
    Write_time: time.Now().Format("2006/01/02 03:04"),
    Is_answer: 1,
    Is_deleted: 0,
    Is_file: 0,
    Chat_id: 0,
    Question_id: question_id,
  }
  questiondatas := []model.ChatData{}
  questiondatas = append(questiondatas, questiondata)
	_ := conn.WriteJSON(questiondatas)
}
```
GetUsrOrderByUUID(uuid)로 내가 first_user인지, second_user인지 확인한다.

QuestionIDOfEmptyAnswerByOrder(order, conn_id)로 answer 테이블에 내가 대답하지 않은 question이 있는지 확인한다.

대답하지 않은 question이 없으면 question_id는 0을 리턴하고, 0이 아닌 경우에 GetQuestionByQuestionID(question_id)로 questionID에 맞는 question의 내용을 찾아서 사용자에게 메시지를 전송한다.

대답하지 않은 question이 여러 개일 수도 있다. 예를 들어 내가 커넥션이 종료된 사이, 상대방이 혼자 채팅을 많이 치다가 여러 키워드들을 언급했을 경우.

그런 경우를 위해 questiondatas는 slice 형태로 전송된다.

---

## 커넥션 연결 유지시키기

커넥션 연결은 일정 시간동안 송/수신이 되는 메시지가 없으면 자동으로 종료된다. 그렇다고 사용자에게 1분마다 꼭 한 개씩은 채팅을 입력하라고 강제할 순 없는 노릇이다.

다행히 웹소켓 커넥션에 ping메시지를 보내고 응답받는 것으로 웹소켓 연결이 유지될 수 있다.

브라우저마다 다르지만 일반적으로 웹소켓 연결 자동 종료는 3분 내외의 timeout을 가지고있다. 이 시간보다 적은 시간(나의 경우엔 30초로 설정했다)으로 주기적으로 ping 메시지를 보낸다.

```go
go func(){
	ticker := time.NewTicker(30 * time.Second) 
	defer ticker.Stop()
	for range ticker.C {
    _ := conn.WriteMessage(websocket.PingMessage, nil);
		if  err != nil {
			break
		}
	}
}()
```

ping메시지조차 보내지지 않으면 이미 웹소켓 연결이 종료된 것이기 때문에 break를 통해 루프를 빠져나오고 defer ticker.Stop()이 실행되면서 타이머가 종료된다.

---

## 메시지 READ/WRITE

메시지를 읽고 쓰는 건, 커넥션이 끊기기 전까지 무한루프를 돈다.

클라이언트가 서버로 채팅을 보내면, 서버에서 **읽고**, 알맞은 커플들에서 **쓰고** 를 반복한다.

for문 안의 내용을 살펴보자.

```go
var chatData []model.ChatData
  err := conn.ReadJSON(&chatData)
  if err != nil {
	  break;
  }
```

클라이언트로부터 값을 읽는다. ReadJSON()은 값이 읽어질 때까지 기다린다. err는 오래 기다려서가 아니고, 클라이언트로부터 읽을 수 없는 상태. 즉 커넥션이 끊긴 상태일 경우 발생한다. 

이런 경우는 break를 통해 루프를 빠져나올 수 있게 한다. 이후에 사용자가 커넥션(채팅방)에 다시 연결되면 위에서 설명한 모든 과정을 처음부터 다시 수행하며 커넥션을 다시 맺게 된다.

```go
if chatData[0].Is_answer == 1 {
	recieveAnswer(uuid, conn_id, chatData, first_uuid)
} else if chatData[0].Is_deleted == 1 {
	if chatData[0].Is_file == 1 {
		filepath.Walk("assets", func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				if strings.Contains(info.Name(), strconv.Itoa(chatData[0].Chat_id)+"-") {
					os.Remove(path)
				}
			}
			return nil
		})
	}
	model.DeleteChatByChatID(chatData[0].Chat_id)
}
```

chatData는 채팅 데이터 struct이다. 이 chatData에는, 이 채팅이 question인지, question에 대한 답변인지, 누가/언제/어느 커넥션으로 쓴 것인지에 대한 정보가 담긴다.

Is.answer가 1이라는 것은 서버가 클라이언트로부터 받은 메시지가 question에 대한 답변이라는 것을 의마한다. 이 경우 recieveAnswer()로 DB answer 테이블에 사용자가 답변한 내용을 저장한다.

recieveAnswer()는 가독성을 위해 따로 question 관련 쓰기/읽기 함수를 분리해두었다. recieveAnswer()에 대한 내용은 글의 마지막에서 확인할 수 있다.

만약 답변이 아니라 Is.deleted가 1이라는 것은 해당 채팅을 삭제하겠다는 것을 의미한다.

만약 그 채팅이 Is_file == 1로, 파일 전송 채팅이라면, 채팅만 지우면 안되고 서버에 저장되어있던 파일도 같이 지워줘야한다. 그래서 파일을 삭제하고, 만약 Is_file이 1이 아니면 DeleteChatByChatID()로 chat table의 채팅만 삭제한다.

```go
else if chatData[0].Is_file != 1 {
	chat_id, _ := model.InsertChatAndGetChatID(chatData[0].Text_body, uuid, chatData[0].Write_time, 0, 0)
	
	chatData[0].Chat_id = chat_id
} 
```

이 else if는 바로 위의 if chatData[0].Is_answer == 1 {}에 대한 else if이다.

만약 클라이언트에서 보낸 메시지가 답변이 아닌데, file 형태도 아니라면, 일반 채팅인 것이므로 InsertChatAndGetChatID()로 채팅을 DB chat table에 저장한다.

```go
else {
	chatID, _ := model.GetChatIDFromRecentFileChatByUUID(uuid)
	chatData[0].Chat_id = chatID
	text_body, _ := model.GetTextBodyByChatID(chatID)
	chatData[0].Text_body = text_body
}
```

이 else는 채팅이 답변도 아니고 일반 채팅도 아닌 경우를 의미하고, 이 경우는 파일 전송인 것이다. 파일은 따로 HTTP요청을 통해 먼저 저장되어있기 때문에, 가장 최근 파일의 chatID를 확인해서 

```go
target_conn := []*websocket.Conn{}

mutex.Lock()
if conns[first_uuid] != nil && conns[second_uuid] != nil {
	first_conn := conns[first_uuid]
	second_conn := conns[second_uuid]
	target_conn = append(target_conn, first_conn, second_conn)

} else if conns[first_uuid] != nil {
	first_conn := conns[first_uuid]
	target_conn = append(target_conn, first_conn)

} else {
	second_conn := conns[second_uuid]
	target_conn = append(target_conn, second_conn)	
}
mutex.Unlock()
```

이 부분은 어떤 커넥션들에게 전달해야할지 정해서 커넥션들을 target_conn이라는 커넥션 slice에 append하는 부분이다.

아까 살펴본 conns map을 확인해서 nil이 아닌지(상대방의 커넥션이 종료되었는지 아닌지)를 확인해 알맞게 target_conn에 append한다.

이 부분이 mutex lock이 걸려있는 이유는 맨처음 프로토콜을 웹소켓으로 업그레이드한 이후에 conns map에 uuid를 키로, conn을 값으로 넣어뒀었다. 만약 그 부분과 이 부분이 서로 락이 걸려있지 않은 상태여서 비동기적으로 실행이 가능하다면, 매시지를 상대방에게 보내던 도중 커넥션이 끊겨 오류가 발생할 수 있다.

그래서 연결이 먼저 끊어질 거면 먼저 끊고 target_conn에는 나만 append한다던지, 아니면 먼저 메시지를 보낼거면 target_conn에 둘 다 넣어서 메시지를 보내고 나서 연결을 끊을 것인지를 확실하게 할 수 있다.

```go
if chatData[0].Is_answer != 1 {
	for _, item := range target_conn {
		_ := item.WriteJSON(chatData)
	}
}
sendQuestion(chatData, conn_id, target_conn)
```

마지막으로, 서버가 받은 메시지가 answer이 아니면 target_conn에 메시지를 송신한다.

그리고 sendQuestion()을 호출한다.

---

## sendQuestion()

```go
var target_word, question_contents string
var question_id int
r, _ := model.SelectQuetions()
defer r.Close()
for r.Next() {
r.Scan(&target_word, &question_id, &question_contents)	
if strings.Contains(chatData[0].Text_body, target_word) {
	isExist, _ := model.CheckAnswerByConnIDandQuestionID(conn_id, question_id)
	if !isExist {
		questiondata := model.ChatData{
			Text_body: question_contents,
			Writer_id: "question",
			Write_time: time.Now().Format("2006/01/02 03:04"),
			Is_answer: 1,
			Is_deleted: 0,
			Is_file: 0,
			Chat_id: 0,
			Question_id: question_id,
		}
		questiondatas := []model.ChatData{}
		questiondatas = append(questiondatas, questiondata)
		for _, item := range target_conn {
			item.WriteJSON(questiondatas)
		}
		model.InsertAnswer(chatData[0].Write_time, conn_id, question_id)
	}
}
}
```

---

## recieveAnswer()

팝업된 question에 대해 사용자가 답변을 했을 때, 해당 답변을 처리하는 함수이다. sendQuestion()과 마찬가지로 메시지 송/수신 함수에 너무 많은 기능이 들어가있어서 가독성을 위해 분리했다.

question을 커넥션으로 전송할 때, 빈 answer 레코드를 생성하기 때문에, 사용자로부터 답변을 받으면 해당 answer의 값을 사용자가 답변한 값으로 업데이트한다.

```go
func recieveAnswer(uuid string, conn_id int, chatData []model.ChatData, first_uuid string){
	isExist, _ := model.CheckAnswerByConnIDandQuestionID(conn_id, chatData[0].Question_id)
	if isExist {
		if first_uuid == uuid {
			model.UpdateFirstAnswerByQuestionID(chatData[0].Text_body, chatData[0].Question_id)
		} else {
			model.UpdateSecondAnswerByQuestionID(chatData[0].Text_body, chatData[0].Question_id)
		}
	}
}
```

우선, model.CheckAnswerByConnIDandQuestionID()로 answer 레코드가 정상적으로 생성되어있는지 확인한다.

그리고 존재하면(isExist == true) 답변한 사용자가 커넥션 유저 중 first_uuid인지, second_uuid인지를 parameter로 입력받아서 answer table의 알맞은 column을 업데이트해준다.

나중에는 이 레코드 값을 보고 chatPage의 answer 기능에서 연인들이 한 답변을 서로 볼 수 있는 기능이 제공된다.