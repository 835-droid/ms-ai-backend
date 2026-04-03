package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("🚀 مرحبًا بك في أداة تحميل الصور لتطبيق MS-AI")
	fmt.Println("-------------------------------------------")

	// 1. طلب روابط الصور (يمكنك نسخ قائمة روابط كاملة ولصقها هنا)
	fmt.Println("أدخل روابط الصور (رابط في كل سطر)، ثم اكتب 'done' عند الانتهاء:")

	var urls []string
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "done" {
			break
		}
		if line != "" {
			urls = append(urls)
		}
	}

	// 2. إنشاء مجلد الحفظ
	folder := "MS_AI_Manga"
	os.MkdirAll(folder, os.ModePerm)

	// 3. التحميل المتتابع
	client := &http.Client{}
	for i, url := range urls {
		fileName := fmt.Sprintf("%s/page_%03d.jpg", folder, i+1)
		err := downloadFile(client, url, fileName)
		if err != nil {
			fmt.Printf("❌ فشل تحميل الصفحة %d: %v\n", i+1, err)
		} else {
			fmt.Printf("✅ تم حفظ: %s\n", fileName)
		}
	}
	fmt.Println("-------------------------------------------")
	fmt.Println("✨ انتهت العملية بنجاح!")
}

func downloadFile(client *http.Client, url string, filepath string) error {
	req, _ := http.NewRequest("GET", url, nil)

	// أهم إعدادات لتجاوز حماية السيرفرات
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://asuratoon.com/")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
