package cmd

type Uid struct {
}

// func (this *Uid) cmdUid(input string) bool {
// 	inputN := strings.SplitN(input, " ", 5)

// 	if len(inputN) == 5 && this.cmdCompare(inputN[1], CMD_UID) {

// 		if strings.EqualFold(inputN[2], "fetch") {

// 			if strings.Index(inputN[3], ":") > 0 {

// 				se := strings.SplitN(inputN[3], ":", 2)
// 				fmt.Println(se)

// 				start, _ := strconv.ParseInt(se[0], 10, 64)
// 				end, _ := strconv.ParseInt(se[1], 10, 64)
// 				list, err := models.BoxListSE(this.userID, this.selectBox, start, end)
// 				fmt.Println(list)
// 				if err == nil {
// 					for i := 1; i <= len(list); i++ {
// 						c := this.parseArgsConent(inputN[4], list[i-1]["mid"].(string))
// 						// fmt.Println(c)
// 						this.writeArgs("* %d FETCH "+c, i)
// 						// this.writeArgs("* %d FETCH (UID %s)", i, list[i-1]["mid"].(string))
// 					}
// 				}
// 			} else {

// 				list, err := models.BoxAllByClassName(this.userID, this.selectBox)
// 				if err == nil {
// 					for i := 1; i <= len(list); i++ {
// 						c := this.parseArgsConent(inputN[4], list[i-1]["mid"].(string))
// 						// fmt.Println(c)
// 						// d := fmt.Sprintf("")
// 						this.writeArgs("* %d FETCH "+c, i)
// 						// this.writeArgs("* %d FETCH (UID %s)", i, list[i-1]["mid"].(string))
// 					}
// 				}
// 			}

// 			fmt.Println(inputN)
// 			// this.w("* 1 FETCH (UID 1320476750)\r\n")
// 			// this.w("* 2 FETCH (UID 1320476751)\r\n")

// 			this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
// 			return true
// 		}
// 	}
// 	return false
// }
