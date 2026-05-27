package dspi_test

import (
	"encoding/binary"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("ParseUF2", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "uf2-test")

		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	writeUF2 := func(family uint32, numBlocks int) string {
		path := filepath.Join(tmpDir, "test.uf2")

		data := make([]byte, numBlocks*dspi.UF2BlockSize)

		for i := range numBlocks {
			off := i * dspi.UF2BlockSize

			binary.LittleEndian.PutUint32(data[off:off+4], 0x0A324655)     // magic start
			binary.LittleEndian.PutUint32(data[off+4:off+8], 0x9E5D5157)   // magic start
			binary.LittleEndian.PutUint32(data[off+8:off+12], 0x00002000)  // flags
			binary.LittleEndian.PutUint32(data[off+12:off+16], 0x10000000) // target addr
			binary.LittleEndian.PutUint32(data[off+16:off+20], 256)        // payload size
			binary.LittleEndian.PutUint32(data[off+20:off+24], uint32(i))  // block no
			binary.LittleEndian.PutUint32(data[off+24:off+28], uint32(numBlocks))
			binary.LittleEndian.PutUint32(data[off+28:off+32], family)       // family
			binary.LittleEndian.PutUint32(data[off+508:off+512], 0x0AB16F30) // magic end
		}

		err := os.WriteFile(path, data, 0o644)
		Expect(err).ToNot(HaveOccurred())

		return path
	}

	It("parses an RP2350 UF2 file (family 0xe48bff57)", func() {
		path := writeUF2(dspi.UF2FamilyRP2350, 3)

		info, err := dspi.ParseUF2(path)
		Expect(err).ToNot(HaveOccurred())
		Expect(info.BoardFamily).To(Equal(uint32(dspi.UF2FamilyRP2350)))
		Expect(info.NumBlocks).To(Equal(uint32(3)))
	})

	It("parses an RP2350 UF2 file (family 0xe48bff59)", func() {
		path := writeUF2(dspi.UF2FamilyRP2350V2, 3)

		info, err := dspi.ParseUF2(path)
		Expect(err).ToNot(HaveOccurred())
		Expect(info.BoardFamily).To(Equal(uint32(dspi.UF2FamilyRP2350V2)))
		Expect(info.NumBlocks).To(Equal(uint32(3)))
	})

	It("parses an RP2040 UF2 file", func() {
		path := writeUF2(dspi.UF2FamilyRP2040, 5)

		info, err := dspi.ParseUF2(path)
		Expect(err).ToNot(HaveOccurred())
		Expect(info.BoardFamily).To(Equal(uint32(dspi.UF2FamilyRP2040)))
		Expect(info.NumBlocks).To(Equal(uint32(5)))
	})

	It("rejects a non-512-byte-multiple file", func() {
		path := filepath.Join(tmpDir, "bad.uf2")
		err := os.WriteFile(path, make([]byte, 1023), 0o644)
		Expect(err).ToNot(HaveOccurred())

		_, err = dspi.ParseUF2(path)
		Expect(err).To(MatchError(ContainSubstring("not a multiple of block size")))
	})

	It("rejects a too-small file", func() {
		path := filepath.Join(tmpDir, "small.uf2")
		err := os.WriteFile(path, make([]byte, 100), 0o644)
		Expect(err).ToNot(HaveOccurred())

		_, err = dspi.ParseUF2(path)
		Expect(err).To(MatchError(ContainSubstring("too small")))
	})

	It("rejects a file with no magic number", func() {
		path := filepath.Join(tmpDir, "bad.uf2")
		err := os.WriteFile(path, make([]byte, 512), 0o644)
		Expect(err).ToNot(HaveOccurred())

		_, err = dspi.ParseUF2(path)
		Expect(err).To(MatchError(ContainSubstring("invalid magic start")))
	})
})

var _ = Describe("PlatformForFamily", func() {
	It("maps RP2040 family", func() {
		p, err := dspi.PlatformForFamily(dspi.UF2FamilyRP2040)
		Expect(err).ToNot(HaveOccurred())
		Expect(p).To(Equal(dspi.PlatformRP2040))
	})

	It("maps RP2350 family 0xe48bff57", func() {
		p, err := dspi.PlatformForFamily(dspi.UF2FamilyRP2350)
		Expect(err).ToNot(HaveOccurred())
		Expect(p).To(Equal(dspi.PlatformRP2350))
	})

	It("maps RP2350 family 0xe48bff59", func() {
		p, err := dspi.PlatformForFamily(dspi.UF2FamilyRP2350V2)
		Expect(err).ToNot(HaveOccurred())
		Expect(p).To(Equal(dspi.PlatformRP2350))
	})

	It("returns an error for unknown families", func() {
		_, err := dspi.PlatformForFamily(0xDEADBEEF)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("FamilyForPlatform", func() {
	It("maps RP2040 platform", func() {
		f, err := dspi.FamilyForPlatform(dspi.PlatformRP2040)
		Expect(err).ToNot(HaveOccurred())
		Expect(f).To(Equal(uint32(dspi.UF2FamilyRP2040)))
	})

	It("maps RP2350 platform to the primary family", func() {
		f, err := dspi.FamilyForPlatform(dspi.PlatformRP2350)
		Expect(err).ToNot(HaveOccurred())
		Expect(f).To(Equal(uint32(dspi.UF2FamilyRP2350V2)))
	})
})
