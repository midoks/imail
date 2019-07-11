package cmd

type Fetch struct {
}

func (r *Fetch) Handle(resp Resp) error {
	return nil
}

// func (this *Fetch) cmdFetch(input string) bool {
// 	inputN := strings.SplitN(input, " ", 4)

// 	if len(inputN) == 4 && this.cmdCompare(inputN[1], CMD_FETCH) {
// 		// fmt.Println("fetch:%s", input)

// 		list, err := models.BoxAllByClassName(this.userID, this.selectBox)
// 		// fmt.Println(list)
// 		if err == nil {
// 			for i := 1; i <= len(list); i++ {
// 				this.writeArgs("* %d FETCH (UID %s)", i, list[i-1]["mid"].(string))
// 			}
// 		}
// 		this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
// 		return true
// 	}
// 	return false
// }

// func (this *Fetch) parseArgs(input string) string {
// 	input = strings.TrimSpace(input)
// 	input = strings.Trim(input, "()")

// 	inputN := strings.Split(input, " ")
// 	list := make(map[string]int64)

// 	for i := 0; i < len(inputN); i++ {
// 		if strings.EqualFold(inputN[i], "messages") {
// 			count, _ := models.BoxUserMessageCountByClassName(this.userID, this.selectBox)
// 			list[inputN[i]] = count
// 		}
// 		if strings.EqualFold(inputN[i], "recent") {
// 			list[inputN[i]] = 0
// 		}

// 		if strings.EqualFold(inputN[i], "unseen") {
// 			list[inputN[i]] = 0
// 		}

// 	}

// 	out := ""
// 	for i := 0; i < len(inputN); i++ {
// 		// fmt.Println(i, inputN[i], list[inputN[i]])
// 		out += fmt.Sprintf("%s %d ", inputN[i], list[inputN[i]])
// 	}

// 	out = fmt.Sprintf("( %s )", out)
// 	return out
// }

// func (this *Fetch) parseArgsConent(format string, mid string) string {
// 	format = strings.TrimSpace(format)
// 	format = strings.Trim(format, "()")

// 	inputN := strings.Split(format, " ")
// 	list := make(map[string]interface{})

// 	midInt64, _ := strconv.ParseInt(mid, 10, 64)
// 	s, _ := models.MailById(midInt64)
// 	content := s["content"].(string)

// 	bufferedBody := bufio.NewReader(strings.NewReader(content))
// 	header, err := ReadHeader(bufferedBody)

// 	// fmt.Println("headerString:", headerString)
// 	if err != nil {
// 		fmt.Errorf("Expected no error while reading mail, got:", err)
// 	}
// 	bs, err := FetchBodyStructure(header, bufferedBody, true)
// 	if err == nil {
// 		bs.ToString()
// 	} else {
// 		fmt.Println(err)
// 	}

// 	// contentN := strings.Split(content, "\n\n")
// 	// contentL := strings.Split(content, "\n")
// 	// fmt.Println("cmdCompare::--------\r\n", contentN, len(contentN))
// 	// fmt.Println("cmdCompare::--------\r\n")

// 	for i := 0; i < len(inputN); i++ {

// 		if strings.EqualFold(inputN[i], "uid") {
// 			// list[inputN[i]] = libs.Md5str(mid)
// 			list[inputN[i]] = mid
// 		}

// 		if strings.EqualFold(inputN[i], "flags") {
// 			list[inputN[i]] = "(\\Flagged)"
// 		}

// 		if strings.EqualFold(inputN[i], "rfc822.size") {
// 			nnn := fmt.Sprintf("%d", len(content))
// 			list[inputN[i]] = nnn
// 		}

// 		if strings.EqualFold(inputN[i], "bodystructure") {
// 			// cccc := fmt.Sprintf("(\"text\" \"html\" (\"charset\" \"UTF-8\") NIL NIL \"8bit\" %d %d NIL NIL NIL)", 12225, 229)
// 			// cccc := "(  )"

// 			// ccc2 := "((\"text\" \"plain\" (\"charset\" \"UTF-8\") NIL NIL \"quoted-printable\" 4542 104 NIL NIL NIL)(\"text\" \"html\" (\"charset\" \"UTF-8\") NIL NIL \"quoted-printable\" 43308 574 NIL NIL NIL) \"alternative\" (\"boundary\" \"--==_mimepart_5d09e7387efec_127483fd2fc2449c43048322e7\" \"charset\" \"UTF-8\") NIL NIL)"
// 			// list[inputN[i]] = cccc

// 			list[inputN[i]] = bs.ToString()
// 		}

// 		if strings.EqualFold(inputN[i], "body.peek[header]") {
// 			headerString, _ := ReadHeaderString(bufio.NewReader(strings.NewReader(content)))
// 			list["body[header]"] = fmt.Sprintf("{%d}\r\n%s", len(headerString), headerString) //len(headerString)
// 			// list[inputN[i]] = "{1218} \r\nTo: \"midoks@163.com\" <midoks@163.com> \r\nFrom:  <report-noreply@jiankongbao.com>\r\nSubject: 123123\r\nMessage-ID: <80d0b8ee122340ceb665ad1bf5220a42@localhost.localdomain>"
// 		}

// 		if strings.EqualFold(inputN[i], "body.peek[]") {
// 			list[inputN[i]] = fmt.Sprintf("{%d}\r\n%s", len(content), content) //len(content)
// 			// list[inputN[i]] = "{1218} \r\nTo: \"midoks@163.com\" <midoks@163.com> \r\nFrom:  <report-noreply@jiankongbao.com>\r\nSubject: 123123\r\nMessage-ID: <80d0b8ee122340ceb665ad1bf5220a42@localhost.localdomain>"
// 		}
// 	}

// 	out := ""
// 	for i := 0; i < len(inputN); i++ {
// 		fmt.Println(i, inputN[i], list[inputN[i]])

// 		if strings.EqualFold(inputN[i], "body.peek[header]") {
// 			out += fmt.Sprintf("%s %s", strings.ToUpper("body[header]"), list["body[header]"].(string))
// 		} else if strings.EqualFold(inputN[i], "body[]") {
// 			out += fmt.Sprintf("%s %s ", strings.ToUpper("body[]"), list["body[]"].(string))
// 		} else {
// 			out += fmt.Sprintf("%s %s ", strings.ToUpper(inputN[i]), list[inputN[i]].(string))
// 		}
// 	}

// 	out = fmt.Sprintf("(%s)", out)
// 	// fmt.Println(out)
// 	return out
// }
