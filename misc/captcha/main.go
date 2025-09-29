package main

import (
	"fmt"
	"net/http"
	"github.com/dchest/captcha"
)

// 显示验证码页面
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// 生成一个新的验证码ID
	captchaId := captcha.New()
	
	html := fmt.Sprintf(`
	<html>
	<head><title>验证码示例</title></head>
	<body>
	<form method="POST" action="/verify">
		<p>请输入验证码:</p>
		<img src="/captcha/%s.png" alt="captcha"><br>
		<input type="hidden" name="captchaId" value="%s">
		<input type="text" name="captchaSolution">
		<input type="submit" value="提交">
	</form>
	</body>
	</html>
	`, captchaId, captchaId)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// 验证验证码
func verifyHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	captchaId := r.FormValue("captchaId")
	captchaSolution := r.FormValue("captchaSolution")

	if captcha.VerifyString(captchaId, captchaSolution) {
		w.Write([]byte("验证成功 ✅"))
	} else {
		w.Write([]byte("验证失败 ❌"))
	}
}

func main() {
	// 提供验证码图片
	http.Handle("/captcha/", captcha.Server(captcha.StdWidth, captcha.StdHeight))
	// 页面路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/verify", verifyHandler)

	fmt.Println("服务启动在 http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

