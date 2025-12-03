package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings" 
	"time"
	_ "github.com/go-sql-driver/mysql" 
)

// データベースのスキーマファイル名
const schemaFilePath = "schema.sql"

// データベーススキーマを初期化する関数
func initializeDB(db *sql.DB, schemaFilePath string) error {
	log.Println("Initializing database schema...")

	// 1. schema.sqlファイルを読み込む
	sqlBytes, err := ioutil.ReadFile(schemaFilePath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}
	sqlStr := string(sqlBytes)

    // ★修正箇所: SQL文字列をセミコロンで分割し、個別に実行する
    // MariaDBはdb.Exec()で複数のSQL文を一度に受け付けるとは限らないため
	statements := strings.Split(sqlStr, ";")
    
    successCount := 0
    
	// 2. 各SQL文を個別に実行
	for _, stmt := range statements {
        // 前後の空白と改行を削除
		stmt = strings.TrimSpace(stmt) 
		if stmt == "" {
			continue // 空のステートメントはスキップ
		}
        
        // SQLを実行
		_, err = db.Exec(stmt)
		if err != nil {
            // エラーが発生しても、次のテーブルに進む (テーブルが既に存在する場合などがあるため)
			log.Printf("Warning: Failed to execute SQL statement [%s...]. Error: %v", stmt[:50], err)
		} else {
            successCount++
        }
	}
    
    if successCount > 0 {
        log.Printf("Database schema executed successfully! (%d statements executed)", successCount)
    } else {
        log.Println("No SQL statements were successfully executed.")
    }

	return nil
}


//ユーザーをPOSTで受け取り、DBに保存するハンドラ
func handleUserPost(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var data UserData
        if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
            http.Error(w, "Invalid request body (JSON format error)", http.StatusBadRequest)
            return
        }
        
        // 必須フィールドのチェック
        if data.Username == "" || data.Email == "" || data.PasswordHash == "" {
            http.Error(w, "Missing required fields (username, email, password_hash)", http.StatusBadRequest)
            return
        }

        query := `
            INSERT INTO users 
            (username, email, password_hash) 
            VALUES (?, ?, ?)
        `
        
        result, err := db.Exec(query, 
            data.Username, 
            data.Email, 
            data.PasswordHash,
        )

        if err != nil {
            log.Printf("❌ Database INSERT error: %v", err)
            http.Error(w, "Failed to save user due to database error (e.g., duplicate email)", http.StatusInternalServerError)
            return
        }

        lastID, _ := result.LastInsertId()
        w.WriteHeader(http.StatusCreated)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "message": "User saved successfully", 
            "id": lastID,
        })
    }
}

// 質問をPOSTで受け取り、DBに保存するハンドラ
func handleQuestionPost(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var data QuestionData
        // JSONデータをパース
        if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
            http.Error(w, "Invalid request body (JSON format error)", http.StatusBadRequest)
            return
        }

        // データベースに挿入するためのSQLクエリ
        query := `
            INSERT INTO questions 
            (qualification_id, topic_id, author_user_id, question_data) 
            VALUES (?, ?, ?, ?)
        `
        
        result, err := db.Exec(query, 
            data.QualificationID, 
            data.TopicID, 
            data.AuthorUserID, 
            data.QuestionData,
        )

        if err != nil {
            log.Printf("❌ Database INSERT error: %v", err)
            http.Error(w, "Failed to save data due to database error", http.StatusInternalServerError)
            return
        }

        lastID, _ := result.LastInsertId()
        w.WriteHeader(http.StatusCreated)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "message": "Question saved successfully", 
            "id": lastID,
        })
    }
}

func main() {
	// Docker Composeで設定した接続URLを環境変数から取得
	dbURL := os.Getenv("DATABASE_URL_MARIA")
	if dbURL == "" {
		log.Fatal("FATAL: DATABASE_URL_MARIA environment variable not set.")
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?charset=utf8&parseTime=true",
		os.Getenv("MARIADB_USER"),
		os.Getenv("MARIADB_PASSWORD"),
		"mariadb", 
		os.Getenv("MARIADB_DATABASE"),
	)

	log.Printf("Attempting to connect to MariaDB: %s", "mariadb")

	// MariaDBに接続
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to open database connection: %v", err)
	}
	defer db.Close()
	
	// DBが起動するまでリトライするロジック（depends_onだけでは不十分な場合があるため）
	log.Println("Attempting to connect to MariaDB...")
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("✅ Successfully connected to MariaDB.")
			break
		}
		log.Printf("Waiting for MariaDB... attempt %d/%d (Error: %v)", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
		
		if i == maxRetries-1 {
			log.Fatalf("❌ Failed to ping database after %d attempts: %v", maxRetries, err)
		}
	}

	// データベーススキーマを初期化（テーブル作成）
	if err := initializeDB(db, schemaFilePath); err != nil {
		log.Fatalf("❌ Database initialization failed: %v", err)
	}
	
	// ----------------------------------------------------
	// API エンドポイントの設定
	// ----------------------------------------------------
	
	//ユーザー登録エンドポイント
	http.HandleFunc("/api/data/user", handleUserPost(db))
	//問題登録エンドポイント
	http.HandleFunc("/api/data/question", handleQuestionPost(db))
	
	// ヘルスチェックエンドポイント (コンテナが動作しているか確認)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	log.Println("Go API Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}