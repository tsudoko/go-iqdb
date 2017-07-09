package iqdb

const (
	OpSketch = 1 << iota
	OpGrayscale
	_
	OpWidthID
	OpDiscardCommon
)

const (
	ResInfo = 100 + iota
	ResKeyValue
	ResDBList
)

const (
	ResQuery = 200 + iota
	ResMultiQuery
	ResDuplicate
)

const (
	ResErrGeneric = 300 + iota
	ResErrNonFatal
	ResErrFatal
)
