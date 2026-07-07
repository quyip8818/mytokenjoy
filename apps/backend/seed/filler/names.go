package filler

const emailDomain = "example.com"

var givenNames = []string{
	"伟", "芳", "娜", "敏", "静", "丽", "强", "磊", "军", "洋",
	"勇", "艳", "杰", "涛", "明", "超", "秀英", "霞", "平", "刚",
	"桂英", "建华", "文", "华", "金凤", "玉兰", "桂兰", "志强", "秀兰", "建国",
}

var surnames = []string{
	"王", "李", "张", "刘", "陈", "杨", "赵", "黄", "周", "吴",
	"徐", "孙", "胡", "朱", "高", "林", "何", "郭", "马", "罗",
	"梁", "宋", "郑", "谢", "韩", "唐", "冯", "于", "董", "萧",
}

func buildChineseName(index int) string {
	surname := surnames[index%len(surnames)]
	given := givenNames[(index/len(surnames))%len(givenNames)]
	return surname + given
}

func buildEmail(index int) string {
	return "user" + itoa(index) + "@" + emailDomain
}

func buildPhone(index int) string {
	suffix := itoa(10000000 + index)
	if len(suffix) > 8 {
		suffix = suffix[len(suffix)-8:]
	}
	for len(suffix) < 8 {
		suffix = "0" + suffix
	}
	return "13" + suffix
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 12)
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
