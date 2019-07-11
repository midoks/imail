package cmd

type Select struct {
}

// func (this *Select) cmdSelect(input string) bool {
// 	inputN := strings.SplitN(input, " ", 3)

// 	if len(inputN) == 3 && this.cmdCompare(inputN[1], CMD_SELECT) {
// 		this.selectBox = strings.Trim(inputN[2], "\"")
// 		msgCount, _ := models.BoxUserMessageCountByClassName(this.userID, this.selectBox)
// 		this.writeArgs("* %d EXISTS", msgCount)
// 		this.writeArgs("* 0 RECENT")
// 		this.writeArgs("* OK [UIDVALIDITY 1] UIDs valid")
// 		this.writeArgs("* FLAGS (\\Answered \\Seen \\Deleted \\Draft \\Flagged)")
// 		this.writeArgs("* OK [PERMANENTFLAGS (\\Answered \\Seen \\Deleted \\Draft \\Flagged)] Limited")
// 		this.writeArgs("%s OK [READ-WRITE] %s completed", inputN[0], inputN[1])
// 		return true
// 	}
// 	return false
// }
