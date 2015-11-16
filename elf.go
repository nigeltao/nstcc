package nstcc

type section struct {
	dataOffset uint32
	data       []byte

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
