package testutils

import (
	"github.com/greenplum-db/gpbackup/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("testutils/functions", func() {
	var buffer *gbytes.Buffer
	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
	})
	Describe("SliceBufferByEntries()", func() {
		It("returns a one item slice", func() {
			bufferLen := utils.MustPrintf(buffer, "CREATE TABLE foo (i int);")
			entries := []utils.MetadataEntry{{Name: "name", Schema: "schema", ObjectType: "TABLE", StartByte: 0, EndByte: bufferLen}}
			results, remaining := SliceBufferByEntries(entries, buffer)
			Expect(remaining).To(Equal(""))
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(Equal("CREATE TABLE foo (i int);"))
		})
		It("returns a multi-item slice with spaces and newlines", func() {
			table1Len := utils.MustPrintf(buffer, "CREATE TABLE foo (i int);")
			table2Len := utils.MustPrintf(buffer, "CREATE TABLE bar (j int);")
			entries := []utils.MetadataEntry{{Name: "name", Schema: "schema", ObjectType: "TABLE", StartByte: 0, EndByte: table1Len}, {Name: "name", Schema: "schema", ObjectType: "TABLE", StartByte: table1Len, EndByte: table1Len + table2Len}}
			results, remaining := SliceBufferByEntries(entries, buffer)
			Expect(remaining).To(Equal(""))
			Expect(len(results)).To(Equal(2))
			Expect(results[0]).To(Equal("CREATE TABLE foo (i int);"))
			Expect(results[1]).To(Equal("CREATE TABLE bar (j int);"))
		})
		It("returns a single slice with start within buffer, end outside buffer", func() {
			bufferLen := utils.MustPrintf(buffer, "CREATE TABLE foo (i int);")
			entries := []utils.MetadataEntry{{Name: "name", Schema: "schema", ObjectType: "TABLE", StartByte: 0, EndByte: bufferLen + 10}}
			results, remaining := SliceBufferByEntries(entries, buffer)
			Expect(remaining).To(Equal(""))
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(Equal("CREATE TABLE foo (i int);"))
		})
		It("returns multiple slices with start outside buffer, end outside buffer", func() {
			bufferLen := utils.MustPrintf(buffer, "CREATE TABLE foo (i int);")
			entries := []utils.MetadataEntry{{Name: "name", Schema: "schema", ObjectType: "TABLE", StartByte: 0, EndByte: bufferLen + 10}, {Name: "name", Schema: "schema", ObjectType: "TABLE", StartByte: bufferLen + 10, EndByte: bufferLen + 40}}
			results, remaining := SliceBufferByEntries(entries, buffer)
			Expect(remaining).To(Equal(""))
			Expect(len(results)).To(Equal(2))
			Expect(results[0]).To(Equal("CREATE TABLE foo (i int);"))
			Expect(results[1]).To(Equal(""))
		})
		It("returns a single slice with extra buffer contents", func() {
			bufferLen := utils.MustPrintf(buffer, "CREATE TABLE foo (i int);")
			utils.MustPrintf(buffer, "More extra stuff")
			entries := []utils.MetadataEntry{{Name: "name", Schema: "schema", ObjectType: "TABLE", StartByte: 0, EndByte: bufferLen}}
			_, remaining := SliceBufferByEntries(entries, buffer)
			Expect(remaining).To(Equal("More extra stuff"))
		})
	})
	Describe("CompareSlicesIgnoringWhitespace()", func() {
		It("returns true when slices and buffer are equal", func() {
			actual := []string{"CREATE TABLE foo (i int);"}
			expected := []string{"CREATE TABLE foo (i int);"}
			Expect(CompareSlicesIgnoringWhitespace(actual, expected)).To(BeTrue())
		})
		It("returns true when slices and buffer are equal other than whitespace", func() {
			actual := []string{"\n\t  CREATE TABLE foo (i int);\n"}
			expected := []string{"CREATE TABLE foo (i int);"}
			Expect(CompareSlicesIgnoringWhitespace(actual, expected)).To(BeTrue())
		})
		It("returns false when slices are of different lengths", func() {
			actual := []string{}
			expected := []string{"CREATE TABLE foo (i int);"}
			Expect(CompareSlicesIgnoringWhitespace(actual, expected)).To(BeFalse())
		})
	})
})
