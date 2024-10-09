package pretty

func Text(t string) Document {
	return text{Text: t}
}

func Concat(docs ...Document) Document {
	return concat{Documents: docs}
}

func Empty() Document {
	return concat{}
}

// l is the number of empty lines, not the number of the newlines
// therefore there are l + 1 newlines
func Lines(l uint) Document {
	return lines{Lines: l}
}

func Nest(d Document) Document {
	return nest{Document: d}
}

func SpaceBreak() Document {
	return break_{Unbroken: " "}
}

func Group(d Document) Document {
	return group{Document: d}
}
