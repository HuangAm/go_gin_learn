package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/testdata/protoexample"
	"log"
	"net/http"
	"time"
)

type UserInfo struct {
	// binding:"required" 修饰的字段，若接收为空值，则报错，是必须字段
	Username string `form:"username" json:"user" uri:"username" xml:"username" binding:"required"`
	Password string `form:"password" json:"passwd" uri:"password" xml:"password" binding:"required"`
}

func main() {
	// 1. 创建路由
	r := gin.Default()

	// 2. 绑定路由规则，执行的函数
	// api 参数之 `:`
	r.GET("/user1/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(http.StatusOK, name)
	})
	// api 参数之 `*`
	r.GET("/user2/:name/*action", func(c *gin.Context) {
		name := c.Param("name")
		action := c.Param("action")
		c.String(http.StatusOK, name+"  "+action)
	})
	// url 参数
	r.GET("/user3/:name", func(c *gin.Context) {
		// DefaultQuery 第二个参数为默认值
		name := c.DefaultQuery("name", "Bard Wu")
		c.String(http.StatusOK, fmt.Sprintf("Hello %s", name))
	})
	// form 表单
	r.POST("/form", login)
	r.POST("/upload", uploadFile)   // 上传单文件
	r.POST("/uploads", uploadFiles) // 上传多个文件
	// 路由组v1, 处理GET请求
	v1 := r.Group("/v1")
	{
		v1.GET("/login", loginV1)
		v1.GET("/register", registerV1)
	}
	// 路由组v2, 处理POST请求
	v2 := r.Group("/v2")
	{
		v2.POST("/login/:name", loginV2)
		v2.POST("/register", registerV2)
	}
	// json 数据解析和绑定
	r.POST("/jsonDemo", jsonDemo)
	// 表单数据解析和绑定
	r.POST("/formDemo", formDemo)
	// uri数据解析与绑定
	r.GET("/uriDemo/:username/:password", uriDemo)
	// 限制传输数据大小为8M，默认为32M
	r.MaxMultipartMemory = 8 << 20

	// 多种响应方式
	// 响应json
	r.GET("/someJson", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "someJson", "status": 200})
	})
	// 响应struct
	r.GET("/someStruct", func(c *gin.Context) {
		type msg struct {
			Name string
			Age  int
		}
		m := msg{Name: "张三", Age: 18}
		c.JSON(http.StatusOK, m)
	})
	// 响应xml
	r.GET("/someXML", func(c *gin.Context) {
		c.XML(http.StatusOK, gin.H{"message": "abc"})
	})
	// 响应YAML
	r.GET("/someYAML", func(c *gin.Context) {
		c.YAML(http.StatusOK, gin.H{"name": "Yu Hao"})
	})
	// 响应protoBuf格式
	r.GET("/someProtoBuf", func(c *gin.Context) {
		reps := []int64{int64(1), int64(2)}
		// 定义数据
		label := "label"
		// 传 protobuf 格式数据
		data := &protoexample.Test{
			Label: &label,
			Reps:  reps,
		}
		c.ProtoBuf(http.StatusOK, data)
	})

	// 模板渲染
	// 加载模板文件
	//r.LoadHTMLGlob("template/*") // 匹配该目录下的所有文件
	r.LoadHTMLFiles("template/index.html") // 匹配该目录下对应的文件
	r.GET("/index", func(c *gin.Context) {
		// 根据文件名渲染
		c.HTML(http.StatusOK, "index.html", gin.H{"Title": "爱情是奢侈品"})
	})

	// 页面跳转
	r.GET("/redirect", func(c *gin.Context) {
		// Redirect 支持内部和外部重定向
		c.Redirect(http.StatusMovedPermanently, "http://www.baidu.com")
	})

	// 同步异步
	// 异步
	r.GET("/long_async", func(c *gin.Context) {
		// 需要搞一个副本
		copyContext := c.Copy()
		// 异步处理
		go func() {
			time.Sleep(3 * time.Second)
			log.Println("异步执行：" + copyContext.Request.URL.Path)
		}()
	})
	// 同步
	r.GET("/long_sync", func(c *gin.Context) {
		time.Sleep(3 * time.Second)
		log.Println("异步执行：" + c.Request.URL.Path)
	})

	// 中间件
	// 注册中间件
	r.Use(MiddleWare()) // 此处属于全局中间件
	// {} 是中间件的一种代码规范
	{
		r.GET("/middleware", func(c *gin.Context) {
			// 取值
			req, _ := c.Get("request")
			fmt.Println("request:", req)
			// 页面接收
			c.JSON(http.StatusOK, gin.H{"request": req})
		})
		// 局部中间件
		r.GET("/middleware2", MiddleWare(), func(c *gin.Context) {
			fmt.Println("开始执行handlerFunc")
			// 取值
			req, _ := c.Get("request")
			fmt.Println("request:", req)
			// 页面接收
			c.JSON(http.StatusOK, gin.H{"request": req})
			fmt.Println("handlerFunc执行结束")
		})

	}

	// Cookie
	r.GET("/cookie", func(c *gin.Context) {
		// 获取客户端是否携带cookie
		cookie, err := c.Cookie("key_cookie")
		if err != nil {
			fmt.Println(err.Error())
			cookie = "NotSet"
			// 给客户端设置cookie
			// maxAge int: 存活时间，单位为秒
			// path string: cookie 存放的目录
			// domain string: 域名（一定要注意设置localhost就要用localhost访问, 不要用127.0.0.1, 会出问题）
			// secure bool: 是否只能通过 https 访问
			// httpOnly bool: 是否允许别人通过js获取自己的cookie
			c.SetCookie("key_cookie", "value_cookie", 3600, "/", "localhost", false, true)
		}
		fmt.Println("cookie的值是", cookie)
	})

	// 认证: Cookie + 中间件
	r.GET("/login_mw_test", func(c *gin.Context) {
		// 登陆成功，设置cookie
		c.SetCookie("abc", "123", 30, "/", "localhost", false, true)
		c.String(http.StatusOK, "Login success!")
	})
	r.GET("/home_mw_test", AuthMiddleWare(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "home"})
	})

	// 3. 监听端口，默认在8000
	r.Run(":8000")
}

func login(c *gin.Context) {
	// DefaultPostForm:接受不一定存在的值
	age := c.DefaultPostForm("age", "18")
	// PostForm:接收正常值
	username := c.PostForm("username")
	password := c.PostForm("password")
	// 多选框
	hobby := c.PostFormArray("hobby")
	c.String(http.StatusOK, fmt.Sprintf("age is %s, username is %s, password is %s, hobbys is %#v", age, username, password, hobby))
	fmt.Println(c.Request.Method)
}

func uploadFile(c *gin.Context) {
	// 表单取文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		fmt.Printf("c.FormFile err:%v", err)
		return
	}
	log.Println(fileHeader.Filename, fileHeader.Size, fileHeader.Header)
	// 传到项目根目录
	dst := fmt.Sprintf("./upload/%s", fileHeader.Filename)
	err = c.SaveUploadedFile(fileHeader, dst)
	if err != nil {
		fmt.Printf("c.SaveUploadedFile err: %v", err)
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("file [%s] upload success!", fileHeader.Filename))
}

func uploadFiles(c *gin.Context) {
	// 表单获取
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get err %s", err.Error()))
		return
	}
	// 获取所有图片
	files := form.File["files"]
	// 遍历所有图片
	for _, file := range files {
		// 逐个保存
		dst := fmt.Sprintf("./upload/%s", file.Filename)
		err = c.SaveUploadedFile(file, dst)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload err %s", err.Error()))
			return
		}
	}
	c.String(http.StatusOK, fmt.Sprintf("upload %d files success!", len(files)))
}

func loginV1(c *gin.Context) {
	name := c.DefaultPostForm("name", "Yu Hao")
	c.String(http.StatusOK, fmt.Sprintf("Welcome %s", name))
}

func registerV1(c *gin.Context) {
	name := c.DefaultPostForm("name", "Fang Shao")
	c.String(http.StatusOK, fmt.Sprintf("hello %s", name))
}

func loginV2(c *gin.Context) {
	name := c.DefaultQuery("name", "Yu Hao")
	c.String(http.StatusOK, fmt.Sprintf("Welcome %s", name))
}

func registerV2(c *gin.Context) {
	name := c.DefaultQuery("name", "Fang Shao")
	c.String(http.StatusOK, fmt.Sprintf("hello %s", name))
}

func jsonDemo(c *gin.Context) {
	var userInfo UserInfo
	err := c.ShouldBindJSON(&userInfo)
	if err != nil {
		// 返回错误信息
		// gin.H 封装了生成 json 数据的工具
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 判断数据是否正确
	if userInfo.Username != "root" || userInfo.Password != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "304", "message": "用户名或密码错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "200", "message": "登陆成功"})
}

func formDemo(c *gin.Context) {
	var userInfo UserInfo
	// Bind() 默认绑定 form 格式
	// 根据请求头中 content-type 自动推断
	err := c.Bind(&userInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 判断数据是否正确
	if userInfo.Username != "root" || userInfo.Password != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "304", "message": "用户名或密码错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "200", "message": "登陆成功"})
}

func uriDemo(c *gin.Context) {
	var userInfo UserInfo
	err := c.ShouldBindUri(&userInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 判断数据是否正确
	if userInfo.Username != "root" || userInfo.Password != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "304", "message": "用户名或密码错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "200", "message": "登陆成功"})
}

func MiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		fmt.Println("中间件开始执行...")
		// 设置变量到Context的key中，可以通过 Get() 取
		c.Set("request", "中间件")
		// 执行中间件
		c.Next()
		status := c.Writer.Status()
		fmt.Println("中间件执行完毕", status)
		t2 := time.Since(t)
		fmt.Println("time", t2)
	}
}

func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端cookie并校验
		cookie, err := c.Cookie("abc")
		if err != nil {
			// 返回错误
			c.JSON(http.StatusUnauthorized, gin.H{"error": "need login"})
			// 终止执行后面的handler
			c.Abort()
			return
		}
		if cookie != "123" {
			return
		}
		c.Next()
		return
	}
}

/*
1. gin.Default() 返回的引擎实例中包含 Logger 和 Recovery 中间件;
2. gin.Context 封装了 request 和 response;
3. api 参数中`:`与`*`一个是接一个参数，一个是接一断参数，接收方式 c.Param();
4. url 参数接受方式 c.DefaultQuery;
5. form 表单接受方式 c.PostForm; 多选框接受方式 c.PostFormArray;
6. 限制传输数据大小 r.MaxMultipartMemory, 默认为 32M
7. 单文件上传要点
       - form 表单 enctype="multipart/form-data"
       - 用 c.FormFile() 接收，保存文件用 c.SaveUploadedFile()
8. 多文件上传要点
       - form 表单 enctype="multipart/form-data", input 标签要写有 multiple 属性
       - 接收表单用 c.MultipartForm(), 然后再循环得到每个图片对象, 逐一处理
9. 路由组, 在 python 中称为路由分发, r.Group()
10. httprouter 会将所有路由规则构造成一颗前缀树
11. shouldbind 系列方法与 bind 系列方法的不同之处，及使用
12. 中间件注意点：
	  1. 注册 r.Use, 注册完后面应该有
      2. 规范（用{}包起路由），Next 方法会挂起中间件，去执行逻辑函数
13. 中间件认证示例
*/
