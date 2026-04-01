package tools

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main1() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "الاستخدام: %s <المسار_الأساسي> [ملف_الإخراج]\n", os.Args[0])
		os.Exit(1)
	}

	rootDir := os.Args[1]
	outputFile := "combined_go_files.txt"
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	}

	// فتح ملف الإخراج (إنشاء أو استبدال)
	out, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "خطأ في إنشاء ملف الإخراج: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	writer := bufio.NewWriter(out)
	defer writer.Flush()

	// تعداد الملفات التي تمت معالجتها
	count := 0

	// المسح التكراري للمجلدات
	err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "خطأ في الوصول إلى %s: %v\n", path, err)
			return nil // متابعة المسح حتى في حال وجود أخطاء
		}
		// تجاهل المجلدات
		if d.IsDir() {
			return nil
		}
		// التحقق من امتداد الملف
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}

		// قراءة محتويات الملف
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "خطأ في قراءة الملف %s: %v\n", path, err)
			return nil // تخطي الملف ومواصلة المعالجة
		}

		// كتابة الفاصل العلوي مع المسار
		separator := fmt.Sprintf("// ----- START OF FILE: %s -----\n", path)
		if _, err := writer.WriteString(separator); err != nil {
			return err
		}
		// كتابة محتوى الملف
		if _, err := writer.Write(content); err != nil {
			return err
		}
		// التأكد من وجود سطر جديد في نهاية المحتوى (لتجنب التصاق الفاصل السفلي)
		if len(content) > 0 && content[len(content)-1] != '\n' {
			writer.WriteString("\n")
		}
		// كتابة الفاصل السفلي
		endSeparator := fmt.Sprintf("// ----- END OF FILE: %s -----\n\n", path)
		if _, err := writer.WriteString(endSeparator); err != nil {
			return err
		}

		count++
		fmt.Printf("تمت معالجة: %s\n", path)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "خطأ أثناء المسح: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("تم إنشاء الملف %s بنجاح. عدد الملفات المعالجة: %d\n", outputFile, count)
}
