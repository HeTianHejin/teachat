package route

import (
	"html/template"
	"path/filepath"
	"testing"
)

// TestTemplatesParse 校验所有页面模板的语法与自定义函数引用。
// 模板在运行期才解析，语法错误若不在测试中暴露，会变成 500 页面。
func TestTemplatesParse(t *testing.T) {
	files, err := filepath.Glob("../templates/*.go.html")
	if err != nil {
		t.Fatalf("glob templates: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no template files found")
	}

	for _, file := range files {
		if _, err := template.New("probe").Funcs(templateFuncs).ParseFiles(file); err != nil {
			t.Errorf("%s: %v", file, err)
		}
	}
}
