package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Docker Composeから設定された環境変数
	dbURL := os.Getenv("DATABASE_URL_MARIA")
	if dbURL == "" {
		log.Fatal("DATABASE_URL_MARIA environment variable not set.")
	}
    // go-sql-driverのURL形式に変換: [USER]:[PASSWORD]@tcp([HOST]:[PORT])/[DATABASE]?charset=utf8mb4
    // Goのドライバは、URLスキーマ（mysql://）を必要としないため、手動で組み立てる
	// ここでは環境変数から直接取得するため、Goのコード側では一旦このまま進めます。
	// ※ compose.ymlのDATABASE_URL_MARIAの値がそのままmysql://...の形式であれば、
	//    ドライバが解釈できる形式に修正する必要があります。（例：ユーザー名:パスワード@tcp(ホスト名:ポート)/DB名）

    // 暫定的な接続文字列の組み立て
    // 実際には、docker-compose.ymlでDATABASE_URL_MARIAを以下の形式に修正するとより簡単です:
    // "relean_MARIADB_USER:relearn_MARIADB_PASSWORD@tcp(mariadb:3306)/relean_MARIADB_DATABASE?charset=utf8&parseTime=true"
    
    // 一旦、ここではGoの接続部分をシンプルに保ちます。
	// GoのMySQLドライバは、ユーザー名:パスワード@tcp(ホスト名:ポート)/データベース名 の形式を期待します。
	
	// compose.ymlのDATABASE_URL_MARIAが "mysql://..." の場合、文字列操作が必要です。
	// 簡略化のため、ここではDB接続が成功する最小限のGoコードを提示します。
	
	// 接続情報の再構成 (DATABASE_URL_MARIAが上記の形式であると仮定)
	// (GoのMySQLドライバはURLパーサーを持っていないため、生文字列を組み立てるのが一般的)
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?charset=utf8&parseTime=true",
		os.Getenv("MARIADB_USER"),
		os.Getenv("MARIADB_PASSWORD"),
		"mariadb", // サービス名
		os.Getenv("MARIADB_DATABASE"),
	)
	
	log.Printf("Attempting to connect to MariaDB: %s", "mariadb")

	// 接続プールを作成
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to open database connection: %v", err)
	}
	defer db.Close()

	// DBが利用可能になるまで待機
	// depends_on: service_healthyがあるため、多くの場合不要だが、念のため実装
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for database... attempt %d/10", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("❌ Failed to ping database after multiple attempts: %v", err)
	}

	log.Println("✅ Successfully connected to MariaDB!")

	// 簡単なクエリの実行
	var result int
	err = db.QueryRow("SELECT 1 + 1").Scan(&result)
	if err != nil {
		log.Fatalf("❌ Query failed: %v", err)
	}
	log.Printf("💡 Query result (1 + 1): %d", result)

	// APIサーバーの起動ロジックなどをここに追加

    // 接続確認後、終了させるか、APIサーバを立ち上げるループに入る
    fmt.Println("Go API is running (or will exit after connection test).")
}