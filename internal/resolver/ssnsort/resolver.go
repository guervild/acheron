package ssnsort

import (
	"errors"
	"sort"
	"strings"
	"unsafe"

	"github.com/f1zm0/acheron/internal/resolver"
	wt "github.com/f1zm0/acheron/internal/types"
	"github.com/f1zm0/acheron/pkg/hashing"
	rrd "github.com/f1zm0/acheron/pkg/rawreader"

	"github.com/Binject/debug/pe"
)

type ssnSortResolver struct {
	// hashing provider
	hasher hashing.Hasher

	// map of Zw* InMemProc structs indexed by their name's hash
	zwStubs map[int64]wt.InMemProc

	// slice with addresses of clean gadgets
	safeGates []uintptr
}

var _ resolver.Resolver = (*ssnSortResolver)(nil)

func NewResolver(h hashing.Hasher) (resolver.Resolver, error) {
	r := &ssnSortResolver{}
	if err := r.init(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *ssnSortResolver) init() error {
	var zwStubs []wt.InMemProc

	hNtdll, err := r.getNtdllModuleHandle()
	if err != nil {
		return err
	}

	ex, err := hNtdll.File.Exports()
	if err != nil {
		return err
	}
	for _, exp := range ex {
		memAddr := int64(hNtdll.BaseAddr) + int64(exp.VirtualAddress)
		// TODO: check if stub has syscall;ret gadget, search adiacent ones if not
		// add clean ones to safeGates

		if strings.HasPrefix(exp.Name, "Zw") {
			zwStubs = append(zwStubs, wt.InMemProc{
				Name:     exp.Name,
				BaseAddr: uintptr(memAddr),
			})
		}
	}

	sort.Slice(zwStubs, func(i, j int) bool {
		return zwStubs[i].BaseAddr < zwStubs[j].BaseAddr
	})

	for idx := range zwStubs {
		zwStubs[idx].SSN = idx
		r.zwStubs[r.hasher.HashString(zwStubs[idx].Name)] = zwStubs[idx]
	}

	return nil
}

func (r *ssnSortResolver) getNtdllModuleHandle() (*wt.PEModule, error) {
	entries := resolver.GetLdrTableEntries()
	ntdllHash := r.hasher.HashByteString(
		[]byte{0x6e, 0x74, 0x64, 0x6c, 0x6c, 0x2e, 0x64, 0x6c, 0x6c, 0x00}, // ntdll.dll
	)
	for _, entry := range entries {
		if r.hasher.HashString(entry.BaseDllName.String()) == ntdllHash {
			modBaseAddr := uintptr(unsafe.Pointer(entry.DllBase))
			modSize := int(uintptr(unsafe.Pointer(entry.SizeOfImage)))
			rr := rrd.NewRawReader(modBaseAddr, modSize)

			p, err := pe.NewFileFromMemory(rr)
			if err != nil {
				return nil, errors.New("error reading module from memory")
			}

			return &wt.PEModule{
				BaseAddr: modBaseAddr,
				File:     p,
			}, nil
		}
	}
	return nil, errors.New("module not found, probably not loaded.")
}

// GetSyscallSSN returns the syscall ID of a native API function by its djb2 hash.
// If the function is not found, 0 is returned.
func (r *ssnSortResolver) GetSyscallSSN(fnHash int64) (uint16, error) {
	if v, ok := r.zwStubs[fnHash]; ok {
		return uint16(v.SSN), nil
	}
	return 0, errors.New("could not find SSN")
}

func (r *ssnSortResolver) GetSafeGate() uintptr {
	// FIXME: this panics as safeGates is empty
	return r.safeGates[0]
}
