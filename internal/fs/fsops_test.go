package dir

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/types"
)

func getFullPath(relPath ...string) string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, filepath.Join(relPath...))
}

type testMtaYamlStr struct {
	fullpath string
	path     string
	err      error
}

func (t *testMtaYamlStr) GetMtaYamlFilename() string {
	return t.fullpath
}

func (t *testMtaYamlStr) GetMtaYamlPath() string {
	return t.path
}

func (t *testMtaYamlStr) GetMtaExtYamlPath(platform string) string {
	return t.fullpath
}

var _ = Describe("FSOPS", func() {

	var _ = Describe("CreateDir", func() {

		AfterEach(func() {
			os.RemoveAll(getFullPath("testdata", "level2", "result"))
		})

		var _ = DescribeTable("CreateDir", func(dirPath string) {

			Ω(CreateDirIfNotExist(dirPath)).Should(Succeed())
		},
			Entry("Sanity", getFullPath("testdata", "level2", "result")),
			Entry("DirectoryExists", getFullPath("testdata", "level2", "level3")),
		)
	})

	var _ = Describe("Archive", func() {
		var targetFilePath = getFullPath("testdata", "arch.mbt")

		AfterEach(func() {
			os.RemoveAll(targetFilePath)
		})

		var _ = DescribeTable("Archive", func(sourceFolderPath string, matcher GomegaMatcher, created bool) {

			Ω(Archive(sourceFolderPath, targetFilePath)).Should(matcher)
			if created {
				Ω(targetFilePath).Should(BeAnExistingFile())
			} else {
				Ω(targetFilePath).ShouldNot(BeAnExistingFile())
			}
		},
			Entry("Sanity", getFullPath("testdata", "mtahtml5"), Succeed(), true),
			Entry("SourceIsNotFolder", getFullPath("testdata", "level2", "level2_one.txt"), Succeed(), true),
			Entry("SourceNotExists", getFullPath("testdata", "level3"), HaveOccurred(), false),
		)
	})

	var _ = Describe("Create File", func() {
		AfterEach(func() {
			os.RemoveAll(getFullPath("testdata", "result.txt"))
		})
		It("Sanity", func() {
			file, err := CreateFile(getFullPath("testdata", "result.txt"))
			Ω(getFullPath("testdata", "result.txt")).Should(BeAnExistingFile())
			file.Close()
			Ω(err).Should(BeNil())
		})
	})

	var _ = Describe("CopyDir", func() {
		var targetPath = getFullPath("testdata", "result")
		AfterEach(func() {
			os.RemoveAll(targetPath)
		})

		It("Sanity", func() {
			sourcePath := getFullPath("testdata", "level2")
			Ω(CopyDir(sourcePath, targetPath)).Should(Succeed())
			Ω(countFilesInDir(targetPath)).Should(Equal(countFilesInDir(sourcePath)))
		})

		It("TargetFileLocked", func() {
			f, _ := os.Create(targetPath)
			sourcePath := getFullPath("testdata", "level2")
			Ω(CopyDir(sourcePath, targetPath)).Should(HaveOccurred())
			f.Close()
		})

		var _ = DescribeTable("Invalid cases", func(source, target string) {
			Ω(CopyDir(source, targetPath)).Should(HaveOccurred())
		},
			Entry("SourceDirectoryDoesNotExist", getFullPath("testdata", "level5"), targetPath),
			Entry("SourceIsNotDirectory", getFullPath("testdata", "level2", "level2_one.txt"), targetPath),
			Entry("DstDirectoryNotValid", getFullPath("level2"), ":"),
		)

		var _ = DescribeTable("Copy File - Invalid", func(source, target string, matcher GomegaMatcher) {
			Ω(CopyFile(source, target)).Should(matcher)
		},
			Entry("SourceNotExists", getFullPath("testdata", "fileSrc"), targetPath, HaveOccurred()),
			Entry("SourceIsDirectory", getFullPath("testdata", "level2"), targetPath, HaveOccurred()),
			Entry("WrongDestinationName", getFullPath("testdata", "level2", "level2_one.txt"), getFullPath("testdata", "level2", "/"), HaveOccurred()),
			Entry("DestinationExists", getFullPath("testdata", "level2", "level3", "level3_one.txt"), getFullPath("testdata", "level2", "level3", "level3_two.txt"), Succeed()),
		)
	})

	var _ = Describe("Copy Entries", func() {

		AfterEach(func() {
			os.RemoveAll(getFullPath("testdata", "result"))
		})

		It("Sanity", func() {
			sourcePath := getFullPath("testdata", "level2", "level3")
			targetPath := getFullPath("testdata", "result")
			os.MkdirAll(targetPath, os.ModePerm)
			files, _ := ioutil.ReadDir(sourcePath)
			// Files wrapped to overwrite their methods
			var filesWrapped [3]os.FileInfo
			for i, file := range files {
				filesWrapped[i] = testFile{file: file}
			}
			Ω(copyEntries(filesWrapped[:], sourcePath, targetPath)).Should(Succeed())
			Ω(countFilesInDir(sourcePath) - 1).Should(Equal(countFilesInDir(targetPath)))
			os.RemoveAll(targetPath)

			targetPath = getFullPath("testdata", "//")
			Ω(copyEntries(filesWrapped[:], getFullPath("testdata", "level2", "levelx"), targetPath)).Should(HaveOccurred())
		})
	})

	var _ = Describe("Copy By Patterns", func() {

		AfterEach(func() {
			os.RemoveAll(getFullPath("testdata", "result"))
		})

		var _ = DescribeTable("Valid Cases", func(modulePath string, patterns, expectedFiles []string) {
			sourcePath := getFullPath("testdata", "testbuildparams", modulePath)
			targetPath := getFullPath("testdata", "result")
			Ω(CopyByPatterns(sourcePath, targetPath, patterns)).Should(Succeed())
			for _, file := range expectedFiles {
				Ω(file).Should(BeAnExistingFile())
			}
		},
			Entry("Single file", "ui2",
				[]string{"deep/folder/inui2/anotherfile.txt"},
				[]string{getFullPath("testdata", "result", "anotherfile.txt")}),
			Entry("Wildcard for 2 files", "ui2",
				[]string{"deep/*/inui2/another*"},
				[]string{getFullPath("testdata", "result", "anotherfile.txt"),
					getFullPath("testdata", "result", "anotherfile2.txt")}),
			Entry("Wildcard for 2 files - dot start", "ui2",
				[]string{"./deep/*/inui2/another*"},
				[]string{getFullPath("testdata", "result", "anotherfile.txt"),
					getFullPath("testdata", "result", "anotherfile2.txt")}),
			Entry("Specific folder of second level", "ui2",
				[]string{"*/folder/*"},
				[]string{
					getFullPath("testdata", "result", "inui2", "anotherfile.txt"),
					getFullPath("testdata", "result", "inui2", "anotherfile2.txt")}),
			Entry("All", "ui1",
				[]string{"*"},
				[]string{getFullPath("testdata", "result", "webapp", "Component.js")}),
			Entry("Dot", "ui1",
				[]string{"."},
				[]string{getFullPath("testdata", "result", "ui1", "webapp", "Component.js")}),
			Entry("Multiple patterns", "ui2", //
				[]string{"deep/folder/inui2/anotherfile.txt", "*/folder/"},
				[]string{
					getFullPath("testdata", "result", "folder", "inui2", "anotherfile.txt"),
					getFullPath("testdata", "result", "anotherfile.txt")}),
			Entry("Empty patterns", "ui2",
				[]string{},
				[]string{}),
		)

		var _ = DescribeTable("Invalid Cases", func(targetPath, modulePath string, patterns []string) {
			sourcePath := getFullPath("testdata", "testbuildparams", modulePath)
			err := CopyByPatterns(sourcePath, targetPath, patterns)
			Ω(err).Should(HaveOccurred())
		},
			Entry("Target path relates to file ", getFullPath("testdata", "testbuildparams", "mta.yaml"), "ui2",
				[]string{"deep/folder/inui2/somefile.txt"}),
			Entry("Wrong pattern ", getFullPath("testdata", "result"), "ui2",
				[]string{"[a,b"}),
		)
	})

	It("getRelativePath", func() {
		Ω(getRelativePath(getFullPath("abc", "xyz", "fff"),
			filepath.Join(getFullPath()))).Should(Equal(string(filepath.Separator) + filepath.Join("abc", "xyz", "fff")))
	})

	var _ = Describe("Read", func() {
		It("Sanity", func() {
			test := testMtaYamlStr{
				fullpath: getFullPath("testdata", "testproject", "mta.yaml"),
				path:     getFullPath("testdata", "testproject", "mta.yaml"),
				err:      nil,
			}
			res, resErr := Read(&test)
			Ω(res).ShouldNot(BeNil())
			Ω(resErr).Should(BeNil())
		})
	})

	var _ = Describe("ReadExt", func() {
		It("Sanity", func() {
			test := testMtaYamlStr{
				fullpath: getFullPath("testdata", "testproject", "mta.yaml"),
				path:     getFullPath("testdata", "testproject", "mta.yaml"),
				err:      nil,
			}
			res, resErr := ReadExt(&test, "cf")
			Ω(res).ShouldNot(BeNil())
			Ω(resErr).Should(BeNil())
		})
	})
})

func countFilesInDir(name string) int {
	files, _ := ioutil.ReadDir(name)
	return len(files)
}

type testFile struct {
	file os.FileInfo
}

func (file testFile) Name() string {
	return file.file.Name()
}

func (file testFile) Size() int64 {
	return file.file.Size()
}

func (file testFile) Mode() os.FileMode {
	if strings.Contains(file.file.Name(), "level3_one.txt") {
		return os.ModeSymlink
	}
	return file.file.Mode()
}

func (file testFile) ModTime() time.Time {
	return file.file.ModTime()
}

func (file testFile) IsDir() bool {
	return file.file.IsDir()
}

func (file testFile) Sys() interface{} {
	return nil
}