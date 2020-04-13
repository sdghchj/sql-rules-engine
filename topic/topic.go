package topic

import (
	"strings"
)

const sep = "/"
const single = "+"
const multi = "*"

func Match(topicText, matchTopic *string) bool {
	topicSecs := strings.Split(*topicText, sep)
	matchSecs := strings.Split(*matchTopic, sep)
	i, j, leni, lenj := 0, 0, len(topicSecs), len(matchSecs)
	for i < leni && j < lenj {
		if matchSecs[j] == single {
			i++
			j++
		} else if matchSecs[j] == multi {
			if j >= lenj-1 {
				return true
			}
			if topicSecs[i] == matchSecs[j+1] {
				i++
				j = j + 2
			} else {
				i++
			}
		} else if topicSecs[i] == matchSecs[j] {
			i++
			j++
		} else {
			return false
		}
	}
	return i == leni && j == lenj
}
