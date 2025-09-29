package main

import (
        "strings"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"gopkg.in/gomail.v2"
)

// 生成一个 6 位数字验证码
func genCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}


func hideCode(code string) string {
    // 简单同形字替换映射（可根据需要扩展）
    homoglyph := map[rune]rune{
        '0': '⓿', // 圆圈 0
        '1': '①',
        '2': '②',
        '3': '③',
        '4': '④',
        '5': '⑤',
        '6': '⑥',
        '7': '⑦',
        '8': '⑧',
        '9': '⑨',
        'a': 'а', // Cyrillic a
        'b': 'Ь', // Cyrillic soft B
        'c': 'с', // Cyrillic c
        // 可根据需要继续添加
    }

    // 转换同形字 + 插入零宽空格
    var result strings.Builder
    zw := '\u200B' // 零宽空格
    for _, ch := range code {
        if rep, ok := homoglyph[ch]; ok {
            result.WriteRune(rep)
        } else {
            result.WriteRune(ch)
        }
        result.WriteRune(zw)
    }
    return result.String()
}

func main() {
	// 推荐：从环境变量读取凭据（更安全）
	// 如果你想直接硬编码（仅用于临时测试），将下面的 os.Getenv 替换为字符串常量即可。
	smtpHost := os.Getenv("SMTP_HOST")   // e.g. "smtp.163.com"
	smtpPortStr := os.Getenv("SMTP_PORT")// e.g. "465"
	smtpUser := os.Getenv("SMTP_USER")   // e.g. "travellingdotgo@163.com"
	smtpPass := os.Getenv("SMTP_PASS")   // e.g. "QPxuBZT9FQyKNiH3"
	to := os.Getenv("TO_EMAIL")          // e.g. "atu.cool@qq.com"

	// 如果环境变量未设置，为了方便测试，这里给出示例默认值（请在真实环境删除或替换）
	if smtpHost == "" {
		smtpHost = "smtp.163.com"
	}
	if smtpPortStr == "" {
		smtpPortStr = "465"
	}
	if smtpUser == "" {
		smtpUser = "travellingdotgo@163.com" // 你提供的账号
	}
	if smtpPass == "" {
		smtpPass = "QPxuBZT9FQyKNiH3" // 你提供的临时密钥（仅示例）
	}
	if to == "" {
		to = "atu.cool@qq.com"
	}

	// 把端口字符串转为整数
	var smtpPort int
	_, err := fmt.Sscanf(smtpPortStr, "%d", &smtpPort)
	if err != nil || smtpPort <= 0 {
		log.Fatalf("无效的 SMTP 端口: %s", smtpPortStr)
	}

	code := genCode()
        hiddenCode := hideCode(code)
	subject := "【验证码】您的验证码为"
	body := fmt.Sprintf("你好，\n\n你的验证码是：%s\n该验证码 5 分钟内有效。\n\n如果不是你本人操作，请忽略本邮件。", hiddenCode)

	// 使用 gomail 发送（支持 SSL/TLS）
	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	// 如果你通过 465 端口（SSL）连接，以下 TLS 设置通常可用：
	d.TLSConfig = &tls.Config{
		ServerName: smtpHost,
	}

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		log.Fatalf("发送邮件失败: %v", err)
	}

	fmt.Printf("邮件已发送到 %s ，验证码: %s\n", to, code)

	// 如果你需要把验证码保存以便后续校验，可以把 code 存到 redis/db 或内存中，并记录过期时间。
}

