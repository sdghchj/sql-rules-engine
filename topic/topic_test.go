package topic

import "testing"

func TestTopicMatch(t *testing.T) {
	text, match := "aaa/bbb/ccc/ddd/eee/fff", "aaa/+/ccc/*/fff"
	if !Match(&text, &match) {
		t.Fail()
	}

	text, match = "aaa/bbb/ccc/ddd/eee/fff", "aaa/+/ccc/*/ggg"
	if Match(&text, &match) {
		t.Fail()
	}
}
