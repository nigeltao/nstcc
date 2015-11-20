package nstcc

import (
	"debug/elf"
)

func appendELFSym64(b []byte, s elf.Sym64) []byte {
	b = appendU32LE(b, s.Name)
	b = append(b, s.Info)
	b = append(b, s.Other)
	b = appendU16LE(b, s.Shndx)
	b = appendU64LE(b, s.Value)
	b = appendU64LE(b, s.Size)
	return b
}

func appendU16LE(b []byte, u uint16) []byte {
	return append(b,
		uint8(u>>0),
		uint8(u>>8),
	)
}

func appendU32LE(b []byte, u uint32) []byte {
	return append(b,
		uint8(u>>0),
		uint8(u>>8),
		uint8(u>>16),
		uint8(u>>24),
	)
}

func appendU64LE(b []byte, u uint64) []byte {
	return append(b,
		uint8(u>>0),
		uint8(u>>8),
		uint8(u>>16),
		uint8(u>>24),
		uint8(u>>32),
		uint8(u>>40),
		uint8(u>>48),
		uint8(u>>56),
	)
}

func putU32LE(b []byte, u uint32) {
	b[0] = uint8(u >> 0)
	b[1] = uint8(u >> 8)
	b[2] = uint8(u >> 16)
	b[3] = uint8(u >> 24)
}

func elfHash(b []byte) (h uint32) {
	for _, c := range b {
		h = (h << 4) + uint32(c)
		g := h & 0xf0000000
		if g != 0 {
			h ^= g >> 24
		}
		h &^= g
	}
	return h
}

type section struct {
	data []byte

	shName      int32
	shNum       int32
	shType      int32
	shFlags     uint32
	shInfo      int32
	shAddrAlign int32
	shEntSize   int32
	shSize      uint32
	shAddr      uint64
	shOffset    uint32

	nbHashedSyms int32

	link  *section
	reloc *section
	hash  *section
	next  *section

	name []byte
}

func (s *section) putELFStr(sym []byte) int {
	ret := len(s.data)
	s.data = append(s.data, sym...)
	s.data = append(s.data, 0)
	return ret
}

func (s *section) putELFSym(arch Arch, value uint64, size uint64, info uint8, other uint8, shndx uint16, name []byte) int {
	ret := len(s.data)

	nameOffset := 0
	switch arch {
	case ArchAMD64:
		nameOffset = len(s.data)
		s.data = appendELFSym64(s.data, elf.Sym64{
			Name:  0, // Placeholder.
			Info:  info,
			Other: other,
			Shndx: shndx,
			Value: value,
			Size:  size,
		})
	default:
		panic("TODO: implement this architecture")
	}

	if name != nil {
		putU32LE(s.data[nameOffset:], uint32(s.putELFStr(name)))
	}

	if hs := s.hash; hs != nil {
		// TODO.
	}

	return ret
}
