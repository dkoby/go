//
//
//
package main

import (
    "os"
    "fmt"
    "log"
    "path"
    "io"
    "bufio"
)
//
//
//
func main() {
    log.SetFlags(0);

    if (len(os.Args) < (3 + 1)) {
        usage();
    }

    inputFileName  := os.Args[1]
    outputFileName := os.Args[2]
    packageName    := os.Args[3]
    var varName string
    if (len(os.Args) >= (4 + 1)) {
        varName = os.Args[4]
    } else {
        varName = "BinaryData"
    }

    var inputReader *bufio.Reader
    if (inputFileName == "-") {
        inputReader = bufio.NewReader(os.Stdin)
    } else {
        file, err := os.Open(inputFileName)
        if (err != nil) {
            log.Fatalln("Failed to open input file", inputFileName)
        }
        defer file.Close()
        inputReader = bufio.NewReader(file)
    }

    var outputWriter *bufio.Writer
    if (outputFileName == "-") {
        outputWriter = bufio.NewWriter(os.Stdout)
    } else {
        file, err := os.Create(outputFileName)
        if (err != nil) {
            log.Fatalln("Failed to create output file", outputFileName)
        }
        defer file.Close()
        outputWriter = bufio.NewWriter(file)
    }

    if _, err := outputWriter.WriteString("package " + packageName + "\n"); err != nil {
        log.Fatal("Failed to write output file")
    }
    if _, err := outputWriter.WriteString("var " + varName + " = []byte{"); err != nil {
        log.Fatal("Failed to write output file")
    }

    if (true) {
        n := -1
        for {
            var err error
            var b byte
            if b, err = inputReader.ReadByte(); err == nil {
                if n++; n % 8 == 0 {
                    if _, err := outputWriter.WriteString("\n    "); err != nil {
                        log.Fatal("Failed to write output file")
                    }
                }
                if _, err := fmt.Fprintf(outputWriter, "0x%02X, ", b); err != nil {
                    log.Fatal("Failed to write output file")
                }
                continue
            }
            if (err != io.EOF) {
                log.Fatal("Failed to read input file")
            }
            break
        }
    }
    if _, err := outputWriter.WriteString("\n}\n"); err != nil {
        log.Fatal("Failed to write output file")
    }
    outputWriter.Flush()
}
func usage() {
    log.Println("Usage: " + path.Base(os.Args[0]) + " [INPUT] [OUTPUT] [PACKAGE_NAME] <NAME>")
    log.Println("    INPUT        Input file name. \"-\" for standard input.")
    log.Println("    OUTPUT       Output file name. \"-\" for standard output.")
    log.Println("    PACKAGE_NAME Name of package in output file.")
    log.Println("    NAME         Optional name of variable. Default name is \"BinaryData\".")
    log.Fatalln()
}

