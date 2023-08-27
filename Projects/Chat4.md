# #4. 리액트 캘린더 구현하기
# project couple-chat

---

## 개요

커플 간의 일정을 공유/생성/삭제하기 위해 캘린더 기능이 필요했다. 리액트에 여러 캘린더 라이브러리들이 있지만, 직접 스크래치로 구현했다.

캘린더에 일정을 추가하고 삭제하는 것은 크게 어렵지 않았다. D-Day를 설정하고 며칠 남았는지를 표시하는 것도 어렵지 않았다. 구현에 있어서 어려웠던 점은 아래와 같다.

1. 4년마다 윤년이 돈다는 것
2. 매달 주의 수가 일정하지 않고 4주 ~ 6주로 달마다 다르다는 것
3. 매달 말일이 28, 29, 30, 31로 다양하다는 것
4. 매달 시작일의 요일이 다 다르다는 것

---

## 변수 설정

Calender 컴포넌트 안에 필요한 변수들을 정의했다.

    const date1 = new Date(); // 06/16, 이달 1일 계산
    const date2 = new Date(); // 06/16, 이달 말일 계산
    const thisYear = date1.getFullYear();
    const thisDate = date1.getDate();
    const thisMonth = date1.getMonth(); // thisMonth = 7
    const [month, setMonth] = useState(thisMonth); // 기본값 month = 7
    const [year, setYear] = useState(thisYear); // 기본값 month = 7
    date1.setDate(1); // date1 = 06/01
    date1.setMonth(month); // date1 = 07/01
    date2.setDate(1); // date1 = 06/01
    date2.setMonth(month + 1); // date1 = 07/01
    const [dateArray, setDateArray] = useState([]);
    const [weeksArray, setWeeksArray] = useState([]);

하나하나 살펴보겠다.

    const date1 = new Date();
    const date2 = new Date();

리액트에서 new Date()는 현재 날짜를 불러온다. 이 값을 date1과 date2에 저장했다.

    const thisYear = date1.getFullYear();
    const thisMonth = date1.getMonth();
    const thisDate = date1.getDate();

현재년도, 현재월, 현재일을 저장하는 thisYear, thisMonth, thisDate를 정의했다.

이 변수들은 캘린더에 현재날짜를 표시하는 용도로 사용될 것이다.

주의할 점은, getMonth()는 0부터 11사이의 값을 가진다. 1월을 0으로 보는 셈이고, 개발 당시 8월이었기 때문에, 현재 thisMonth에는 7이 저장되어있는 상태다.

    const [month, setMonth] = useState(thisMonth);
    const [year, setYear] = useState(thisYear);

useState인 month와 year는 유동적으로 값이 변하면서 해당 월의 말일 계산, 요일 계산을 위해 사용될 것이다.

    date1.setDate(1);
    date1.setMonth(month);

오늘이 며칠인지는 thisDate에 이전에 저장해뒀기 때문에, date1의 date를 1로 변경해준다.

이렇게되면 date1은 **"이번달의 첫 일"**를 가리키고 있을 것이다.

    date2.setDate(1);
    date2.setMonth(month + 1);

date2에도 같은 작업을 해준다. 대신 date2에는 month + 1을 해서 다음달 1일을 가리키도록 설정했다.

date2에서 하루를 빼면 전 날짜를 얻을 수 있고, 다음달 1일의 하루 전 날은 이번달 말일이 된다.

    const [dateArray, setDateArray] = useState([]);
    const [weeksArray, setWeeksArray] = useState([]);

dateArray는 해당 월의 주차 수 * 7 크기의 배열에 달력 모양대로 일자를 입력하는 배열이다. 예를 들어, 1월의 1일이 화요일부터 시작되고, 말일인 31일은 목요일에 종료된다고 하면, 이 dateArray는 [0(일),0(월),1(화),2,3,...,29,30,31(목),0(금),0(토)] 이런 식으로 저장될 것이다.

이 배열은 캘린더에 날짜를 표시할 때 사용될 것이다.

weeksArray는 해당 월에 주차 수를 배열에 그대로 집어넣은 배열이다. 이 역시 dateArray와 함께 날짜를 캘린더에 출력할 때 사용된다.

---

## 마운트/렌더링 설정

변수 이후에 리렌더링 될 때마다 값을 계산한다.

```js
let firstWeeksLastDate = 7 - date1.getDay();
let lastDateOfThisMonth = date2.getDate(date2.setDate(date2.getDate() - 1));

let weeksOfThisMonth;
for (let i = 0; firstWeeksLastDate + 7 * i < lastDateOfThisMonth; i++) {
weeksOfThisMonth = i;
}
weeksOfThisMonth += 2;
```

    let firstWeeksLastDate = 7 - date1.getDay();

getDay는 요일을 0부터 6까지의 수로 반환한다. 일요일을 한 주 시작의 기준으로 보기때문에, date1가 일요일이면 0, 토요일이면 6을 반환한다.

그럼 7에서 오늘 요일값을 빼주면, 첫 주의 마지막 날짜가 며칠인지를 알 수 있다.

이 값을 firstWeeksLastDate에 저장한다.

일요일을 주의 시작이라고 정하고, 7에서 이번 달 첫 날짜를 빼면 첫주의 마지막 날짜(토요일)가 나온다.

    let lastDateOfThisMonth = date2.getDate(date2.setDate(date2.getDate() - 1));

date2는 다음 달 1일을 가리키고 있었다. 위에서 언급했듯이 date2.getDate() - 1을 하면 date2는 이번 달 말일을 가리키게 된다.

setDate로 이 값을 date2의 값으로 저장해주고, 다시 한 번 getDate를 하면 이번 달 말일에 대한 데이터를 date2가 가지고있게 된다.

이 값을 lastDateOfThisMonth에 저장해준다. 

lastDateOfThisMonth는 오늘이 몇 년도, 몇 월, 며칠이든 해당 달의 말일을 가리키게 된다.

    let weeksOfThisMonth;
    for (let i = 1; firstWeeksLastDate + 7 * i < lastDateOfThisMonth; i++) {
    weeksOfThisMonth = i;
    }
    weeksOfThisMonth += 1;

weeksOfThisMonth는 이번 달이 몇 주까지 있는지에 대한 데이터가 저장될 것이다.

첫 주의 마지막 날짜에 말일이 넘지 않을 때까지 7일씩 더한 뒤, 더한 횟수에 1을 더하면 그 달이 몇 주차까지 있는지를 알 수 있게된다.

1을 더하는 이유는 예를 들어보면 쉽게 이해할 수 있다.

말일이 30일이고, 7씩 더해서 25일까지 왔다고 가정해보자.

또 7을 더하면 32로 30보다 커지기 때문에 더 더하지 않고 반복문은 종료된다. 그렇다고 다음 주가 없는 것은 아니다. 다음주의 토요일이 없는 것 뿐이지 다음주는 26, 27, 28, 29, 30일이 존재한다.

그래서 마지막 주를 세기위해 1을 더해주는 것이다.

여기까지 해서,

1. 첫 주 토요일의 날짜 (firstWeeksLastDate)
2. 이 달의 말일 (lastDateOfThisMonth)
3. 이 달의 week 수 (weeksOfThisMonth)

컴포넌트 마운트나 렌더링이 이루어질 때마다 최신화된 세 정보를 알 수 있게 되었다.

---

## Prev, Next Month

temp_date는 각 날짜 칸을 이동할 때마다 1씩 더해지게되고, 날짜 칸에는 이 temp_date가 들어가서 모든 날짜에 알맞게 일 수를 표시할 수 있게된다.

그 이후에 루프를 통해 해당 월의 주차 수 * 7만큼 렌더링을 할 건데, 날짜 표시는 첫날부터 마지막날까지만 표시하고 첫 날보다 작거나 말일보다 크면, 이전 달, 다음 달이기 때문에 캘린서에 날짜를 출력하지 않도록 하는 로직이다.

    useEffect(() => {
      const array = [];
      let temp_date = 1;
      for (let i = 0; i < 7 * weeksOfThisMonth; i++) {
        if (
          date1.getDay() <= i &&
          i <= lastDateOfThisMonth + date1.getDay() - 1
        ) {
          array[i] = temp_date;
          temp_date += 1;
        } else {
          array[i] = 0;
        }
      }
      setDateArray(array);

      const temp_weeks = [];
      for (let j = 0; j < weeksOfThisMonth; j++) {
        temp_weeks[j] = j;
      }
      setWeeksArray(temp_weeks);
    }, [month]);



이 useEffect는 next, prev 버튼을 통해 month가 변경될 때마다 month에 맞게 새롭게 계산되면서 해당 월에 맞는 캘린더를 보여줄 수 있게 된다.

---

## TroubleShooting

날짜는 알맞게 구현됐는데, 해당 월이 출력될 때, 월이 음수로 표시되는 문제가 생겼다.

처음 월을 getMonth()로 받아올 땐 양수인데, 계속 이전달로 넘어가다보면 날짜는 잘 표시되는데 월이 -1월, -2월로 넘어가길래 1월 전은 12월이 되도록, 또 12월 다음은 13월이 아니라 1월이 되도록 month를 설정해줄 필요가 있었다.

그렇다고 모듈러를 사용하자기엔 매 년도 같은 캘린더가 생성될 수 있었다.

    month = (month % 12) + 1

이런 식으로하면 올해 1월과 내년 1월을 같은 일정이 될 것이다. 그래서 실질적으로 month는 음수와 양수 자유롭게 크고 작아지도록 두고, 사용자에게 표시할 때 모듈러를 사용해서 표시하는 방법을 선택했다.

    const prevMonthHandler = () => {
      if (month % 12 === 0) {
        setYear(year - 1);
      }
      setAnniversaries([]);
      setMonth(month - 1);
      setDateInfo(0);
      setInputAnniversary("");
    };

    const nextMonthHandler = () => {
      if ((month + 1) % 12 === 0) {
        setYear(year + 1);
      }
      setAnniversaries([]);
      setMonth(month + 1);
      setDateInfo(0);
      setInputAnniversary("");
    };

중요한 부분은 각 핸들러의 if문이다. 이전 버튼을 누르면 전월로 month - 1이 되는데, 이 때 month % 12 === 0이 된다면 이번달이 1월이었고 전월을 눌러서 이제 12월로 가야한다는 이야기다, 그래서 년도를 1 빼주는 식으로 구현했다.

next버튼을 누르면 위와 동일하지만 반대로 작동하도록 했다.

---

## 결과

<iframe width="900" height="600" src="https://www.youtube.com/embed/YrsYGLeEhGI" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>