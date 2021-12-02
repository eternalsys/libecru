package main

import (
	"context"
	"database/sql"
	"fmt"
	"libecru/models"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

/*type FreeLanceInfo struct {
	LastName      string `form:"lastName"`
	FirstName     string `form:"firstName"`
	LastNameKana  string `form:"lastNameKana"`
	FirstNameKana string `form:"firstNameKana"`
	Age           string `form:"age"`
	Prefectures   string `form:"prefectures"`
	MailAddress   string `form:"mailAddress"`
	TelNumber     string `form:"telNumber"`
}*/

func CreateDb() (db *sql.DB) {
	db, err := sql.Open("mysql", "admin:admin@tcp(localhost:3306)/libecru?parseTime=true&loc=Asia%2FTokyo")
	if err != nil {
		log.Fatal("Cannot connect database: %v", err)
	}

	return db
}

// フリーランス登録処理
func GetDataFreeRegister(ctx context.Context, c *gin.Context) {

	var b models.Freelance
	// 電話番号の設定
	b.TelNumber = null.StringFrom(c.Request.FormValue("tel_number"))

	if err := c.Bind(&b); err != nil {
		fmt.Errorf("%#v", err)
	}

	// DB生成
	var db = CreateDb()
	boil.SetDB(db)

	err := b.Insert(ctx, db, boil.Infer())

	if err != nil {
		fmt.Errorf("Get freelance error: %v", err)
	}

	c.HTML(http.StatusOK, "registercompletion.html", map[string]interface{}{})
}

// ログインユーザー認証処理
func GetDataUser(ctx context.Context, c *gin.Context) (userConfFlg bool) {

	userConfFlg = false

	// DB生成
	var db = CreateDb()
	boil.SetDB(db)

	// ユーザー情報
	user, err := models.Users(models.UserWhere.ID.EQ(c.Request.FormValue("id")),
		models.UserWhere.Password.EQ(c.Request.FormValue("password"))).One(ctx, db)

	if err != nil {
		fmt.Errorf("Get user error: %v", err)
	}

	if user != nil {
		userConfFlg = true
	}

	return userConfFlg
}

func GetDataTodo(ctx context.Context, c *gin.Context) {
	var b models.Todo
	if err := c.Bind(&b); err != nil {
		fmt.Errorf("%#v", err)
	}

	// DB生成
	var db = CreateDb()
	boil.SetDB(db)

	b.Status = 0
	err := b.Insert(ctx, db, boil.Infer())

	todos, err := models.Todos().All(ctx, db)
	println(todos)
	if err != nil {
		fmt.Errorf("Get todo error: %v", err)
	}

	c.HTML(http.StatusOK, "index.html", map[string]interface{}{
		"todo": todos,
	})
}

func GetDoneTodo(ctx context.Context, c *gin.Context) {
	var b models.Todo
	if err := c.Bind(&b); err != nil {
		fmt.Errorf("%#v", err)
	}

	// DB生成
	var db = CreateDb()
	boil.SetDB(db)

	if b.Status == 0 {
		b.Status = 1
	} else {
		b.Status = 0
	}

	_, err := b.Update(ctx, db, boil.Whitelist("status", "updated_at"))
	if err != nil {
		fmt.Errorf("Get todo error : %v", err)
	}

	todos, err := models.Todos().All(ctx, db)
	if err != nil {
		fmt.Errorf("Get todo error : %v", err)
	}
	c.HTML(http.StatusOK, "index.html", map[string]interface{}{
		"todo": todos,
	})
}

func main() {

	// DB接続記述
	ctx := context.Background()
	var db = CreateDb()

	boil.SetDB(db)

	r := gin.Default()
	r.LoadHTMLFiles("./templates/index.html")
	r.GET("/libecru", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "world",
		})
	})
	r.GET("/todo", func(c *gin.Context) {

		r.LoadHTMLFiles("./templates/index.html")

		todos, _ := models.Todos(qm.OrderBy("updated_at desc")).All(ctx, db)

		c.HTML(http.StatusOK, "index.html", map[string]interface{}{
			"todo": todos,
		})
	})

	// フリーランス登録画面
	r.GET("/freelanceregister", func(c *gin.Context) {

		r.LoadHTMLFiles("./templates/freelanceregister.html")

		c.HTML(http.StatusOK, "freelanceregister.html", map[string]interface{}{})
	})

	// フリーランス登録完了画面
	r.GET("/free_register", func(c *gin.Context) {
		r.LoadHTMLFiles("./templates/registercompletion.html")
		GetDataFreeRegister(ctx, c)
	})

	// ログイン画面
	r.GET("/login", func(c *gin.Context) {

		r.LoadHTMLFiles("./templates/login.html")
		c.HTML(http.StatusOK, "login.html", map[string]interface{}{})
	})

	// ログイン確認画面
	r.GET("/login_verific", func(c *gin.Context) {

		var userConfFlg bool = false

		//r.LoadHTMLFiles("./templates/freelanceverific.html")
		userConfFlg = GetDataUser(ctx, c)

		if userConfFlg {
			// フリーランス管理画面
			r.LoadHTMLFiles("./templates/freelanceManage.html")
			c.HTML(http.StatusOK, "freelanceManage.html", map[string]interface{}{})
		} else {
			// ログイン画面（エラー表示）
			r.LoadHTMLFiles("./templates/login.html")
			c.HTML(http.StatusOK, "login.html", map[string]interface{}{})
		}
	})

	// メールテスト
	r.GET("/mailtest", func(c *gin.Context) {

		// メール送信処理
		from := mail.NewEmail("Example User", "mk.foreverlove.music@icloud.com")
		subject := "Sending with Twilio SendGrid is Fun"
		to := mail.NewEmail("Example User", "k-hamasaki@eternalsys.co.jp")
		plainTextContent := "and easy to do anywhere, even with Go"
		htmlContent := "<strong>and easy to do anywhere, even with Go</strong>"
		message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
		client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
		response, err := client.Send(message)

		fmt.Println("from: ", from)
		fmt.Println("subject: ", subject)
		fmt.Println("to: ", to)
		fmt.Println("plainTextContent: ", plainTextContent)
		fmt.Println("htmlContent: ", htmlContent)
		fmt.Println("message: ", message)
		fmt.Println("client: ", client)

		if err != nil {
			println("エラー")
			log.Println(err)
		} else {
			println("正常")
			fmt.Println(response.StatusCode)
			fmt.Println(response.Body)
			fmt.Println(response.Headers)
		}
	})
	r.GET("/yaru", func(c *gin.Context) {
		GetDataTodo(ctx, c)
	})
	r.GET("/done", func(c *gin.Context) {
		GetDoneTodo(ctx, c)
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
