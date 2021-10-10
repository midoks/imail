package mail

import (
	"testing"
)

var GetMailFromInContentData = []struct {
	in  string
	out string
	ok  bool
}{
	{"From: =?UTF-8?B?6Zi/6YeM5LqR?= <alc@163.com>", "阿里云", true},
	{"From: The Sourcegraph Team <hi@sourcegraph.com>", "The Sourcegraph Team", true},
	{"From: \"=?utf-8?B?NjI3MjkzMDcy?=\"<627293072@qq.com>", "627293072", true},
	{"From: <midoks@163.com>", "midoks", true},
}

//go test -v ./internal/tools/mail -run TestGetMailFromInContent
func TestGetMailFromInContent(t *testing.T) {

	for _, test := range GetMailFromInContentData {
		out := GetMailFromInContent(test.in)
		if out != test.out {
			t.Errorf("GetMailFromInContent(%+q) expected %+q; got %+q", test.in, test.out, out)
		}
	}
}

var GetMailSubjectData = []struct {
	in  string
	out string
	ok  bool
}{
	{"Subject: [GitHub] A third-party OAuth application has been added to your", "[GitHub] A third-party OAuth application has been added to your", true},
	{"Subject: =?utf-8?B?5rWL6K+V?=", "测试", true},
	{"Subject: =?GBK?B?suLK1A==?=", "测试", true},
}

//go test -v ./internal/tools/mail -run TestGetMailSubject
func TestGetMailSubject(t *testing.T) {

	for _, test := range GetMailSubjectData {
		out := GetMailSubject(test.in)
		if out != test.out {
			t.Errorf("GetMailSubject(%+q) expected %+q; got %+q", test.in, test.out, out)
		}
	}
}
