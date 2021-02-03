// +build go1.12
// +build !go1.17

package goloader

import (
	"fmt"
	"unsafe"
)

const stackObjectRecordSize = unsafe.Sizeof(stackObjectRecord{})

// A stackObjectRecord is generated by the compiler for each stack object in a stack frame.
// This record must match the generator code in cmd/compile/internal/gc/ssa.go:emitStackObjects.
type stackObjectRecord struct {
	// offset in frame
	// if negative, offset from varp
	// if non-negative, offset from argp
	off int
	typ *_type
}

func addr2stackObjectRecords(addr unsafe.Pointer) *[]stackObjectRecord {
	n := int(*(*uintptr)(addr))
	slice := sliceHeader{
		Data: uintptr(add(addr, uintptr(PtrSize))),
		Len:  n,
		Cap:  n,
	}
	return (*[]stackObjectRecord)(unsafe.Pointer(&slice))
}

func (linker *Linker) _addStackObject(funcname string, symbolMap map[string]uintptr) (err error) {
	Func := linker.symMap[funcname].Func
	if Func != nil && len(Func.FuncData) > _FUNCDATA_StackObjects &&
		Func.FuncData[_FUNCDATA_StackObjects] != 0 {
		objects := addr2stackObjectRecords(adduintptr(Func.FuncData[_FUNCDATA_StackObjects], 0))
		for i := range *objects {
			name := EmptyString
			stkobjName := funcname + StkobjSuffix
			if symbol := linker.symMap[stkobjName]; symbol != nil {
				name = symbol.Reloc[i].Sym.Name
			}
			if ptr, ok := symbolMap[name]; ok {
				(*objects)[i].typ = (*_type)(adduintptr(ptr, 0))
			} else {
				return fmt.Errorf("unresolve external Var! Function name:%s index:%d, name:%s", funcname, i, name)

			}
		}
	}
	return nil
}
