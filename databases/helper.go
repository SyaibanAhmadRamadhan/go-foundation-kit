package databases

import "fmt"

func QuestionToDollar(sql string) string {
	var (
		idx int = 1
		out     = make([]rune, 0, len(sql))
	)

	for i := 0; i < len(sql); i++ {
		if sql[i] == '?' {
			out = append(out, []rune(fmt.Sprintf("$%d", idx))...)
			idx++
		} else {
			out = append(out, rune(sql[i]))
		}
	}

	return string(out)
}
