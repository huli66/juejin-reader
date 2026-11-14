package crawl

import "testing"

func TestSafeName(t *testing.T) {
	// got := safeName("第 1 章 开篇")
	// want := "第_1_章_开篇"
	// if got != want {
	// 	t.Errorf("safeName() = %q, want %q", got, want)
	// }
	//
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"普通标题", "开篇", "开篇"},
		{"含空格", "第 1 章", "第_1_章"},
		{"含斜杠", "a/b\\c", "a_b_c"},
		{"连续特殊字符", "a: /b", "a_b"},
		{"空字符串", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safeName(tt.in); got != tt.want {
				t.Errorf("safeName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
