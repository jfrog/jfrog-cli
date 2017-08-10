package types

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
)

type booleanEnum int

const (
	False booleanEnum = iota
	True
)

type BoolEnum struct {
	boolean booleanEnum
}

func CreateBoolEnum() *BoolEnum {
	return new(BoolEnum)
}

func(b *BoolEnum) SetValue(val bool) {
	b.boolean = booleanEnum(utils.Bool2Int(val))
}

func (b *BoolEnum) GetValue() bool {
	return b.boolean == True
}






