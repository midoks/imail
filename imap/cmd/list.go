package cmd

type List struct {
}

// func (this *List) cmdList(input string) bool {
// 	inputN := strings.SplitN(input, " ", 4)
// 	if len(inputN) == 4 {
// 		if this.cmdCompare(inputN[1], CMD_LIST) {
// 			list, err := models.ClassGetByUid(this.userID)
// 			if err == nil {
// 				for i := 1; i <= len(list); i++ {
// 					fmt.Println(list[i-1]["flags"], list[i-1]["name"])
// 					mailbox, _ := utf7.Encoding.NewEncoder().String(list[i-1]["name"].(string))
// 					fmt.Println(mailbox)
// 					this.writeArgs("* LIST (\\%s) \"/\" \"%s\"", list[i-1]["flags"], mailbox)
// 				}
// 				this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }
