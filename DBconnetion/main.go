package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" //DB connection을 위한 driver pkg
)

func main() {
	db, err := sql.Open("mysql", "root:PASSWORD@/classicmodels")
	//password는 보안상 변경, mysql의 root계정의 pw를 통해 classicmodels DB를 Open
	if err != nil {
		fmt.Println("error1")
	}
	defer db.Close() // defer로 main이 끝나기 전에 db를 닫아주기
	r, err := db.Query("SELECT customername, phone FROM customers ORDER BY customername ASC")
	// DB에서 쿼리문 작성. result-row의 ptr을 r에 초기화
	if err != nil {
		fmt.Println("error2")
	}
	defer r.Close()        // defer로 main이 끝나기 전에 r를 닫아주기
	var p1, p2 string      // column이 2개라서 variable을 2개 선언
	ss, err := r.Columns() // column명을 ss(string slice type)에 저장
	if err != nil {
		fmt.Println("error3")
	}
	fmt.Println(ss[0], ss[1])
	for r.Next() { // Next()는 한 row씩 내려가고, 다음에 읽을 row가 없으면 false
		r.Scan(&p1, &p2) // p1, p2의 addr에 DB column 값 저장
		fmt.Printf("%v\n\t%v\n", p1, p2)
	}
}
