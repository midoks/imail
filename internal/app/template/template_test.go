package template

import (
	// "fmt"
	"testing"
	"time"
)

var DateFmtMailData = []struct {
	in1 string
	in2 string
	out string
	ok  bool
}{

	{"2021-10-10 13:49:01", "2021-10-10 13:49:01", "13:49", true},
	{"2021-10-09 13:49:01", "2021-10-10 13:49:01", "昨天", true},
	{"2021-10-08 13:49:01", "2021-10-10 13:49:01", "2021-10-08", true},
	{"2021-10-07 13:49:01", "2021-10-10 13:49:01", "2021-10-07", true},
}

func ForDateFmtMail(t time.Time, n time.Time) string {
	in := t.Format("2006-01-02")
	now := n.Format("2006-01-02")

	if in == now {
		return t.Format("15:04")
	}
	in2, _ := time.Parse("2006-01-02 15:04:05", in+" 00:00:00")
	now2, _ := time.Parse("2006-01-02 15:04:05", now+" 00:00:00")
	if in2.Unix()+86400 == now2.Unix() {
		return "昨天"
	} else {
		return t.Format("2006-01-02")
	}
}

// go test -v ./internal/app/template -run TestDateFmtMail
func TestDateFmtMail(t *testing.T) {

	for _, test := range DateFmtMailData {
		in1, _ := time.Parse("2006-01-02 15:04:05", test.in1)
		in2, _ := time.Parse("2006-01-02 15:04:05", test.in2)
		out := ForDateFmtMail(in1, in2)
		if out != test.out {
			t.Errorf("ForDateFmtMail(%+q,%+q) expected %+q; got %+q", test.in1, test.in2, test.out, out)
		}
	}

}
