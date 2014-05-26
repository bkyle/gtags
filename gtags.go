//
// gtags will produce an emacs ETAGS file for a given directory of go code. 
//
package main;

import ( 
	"fmt"
	"os"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"flag"
);

var outputFilename string;
var shouldAppend bool;
var verbose bool;
var showHelp bool;

func init() {
	flag.StringVar(&outputFilename, "out", "TAGS", "Name of the output file.");
	flag.BoolVar(&shouldAppend, "append", false, "Append to the output file instead of completely writing it.  If the file doesn't exist it will be crated.");
	flag.BoolVar(&verbose, "verbose", false, "Verbose messages");
	flag.BoolVar(&showHelp, "help", false, "Prints this message");
}

func main() {

	flag.Parse();

	if showHelp {
		fmt.Printf("");
		flag.PrintDefaults();
		os.Exit(0);
	}

	flags := os.O_WRONLY;

	if info, _ := os.Stat(outputFilename); shouldAppend && info != nil {
		flags = flags | os.O_APPEND;
	} else {
		flags = flags | os.O_CREATE;
	}

	tagsFile, err := os.OpenFile(outputFilename, flags, 0666);

	if err != nil {
		log.Print(err);
		os.Exit(1);
	}
	
	fileSet := token.NewFileSet();

	var args []string;

	if flag.NArg() == 0 {
		args = make([]string, 1, 1);
		args[0] = ".";
	} else {
		args = flag.Args();
	}

	for _, path := range args {
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if matched, _ := filepath.Match("*.go", path); matched {
				writeSection(tagsFile, fileSet, path);
			}
			return nil;
			});
	}
}

func parseFlags() {

}

// Writes a new section to the passed tagsFile for the given source file
func writeSection(tagsFile *os.File, fileSet *token.FileSet, srcFile string) {

	if verbose {
		fmt.Printf("Processing %s\n", srcFile);
	}

	// Contains all of the tags data for the file.  This is needed because the header for the
	// section must contain the number of bytes in the section.
	buffer := "";
	
	file, err := parser.ParseFile(fileSet, srcFile, nil, parser.AllErrors);
	if err != nil {
		log.Printf("%s", err);
	}

	var decl interface{};
	for _, decl = range file.Decls {
		switch decl := decl.(type) {
			case *ast.FuncDecl:
			fn := ast.FuncDecl(*decl);
			position := fileSet.Position(fn.Pos());
			buffer = buffer + fmt.Sprintf("%s\x7f%s\x01%d,%d\n", "func " + fn.Name.Name, fn.Name.Name, position.Line, position.Offset);
		}
	}

	sectionHeader := "\x0c\n";
	sectionHeader = sectionHeader + fmt.Sprintf("%s,%d\n", srcFile, len(buffer));
	tagsFile.WriteString(sectionHeader);
	tagsFile.WriteString(buffer);

}

