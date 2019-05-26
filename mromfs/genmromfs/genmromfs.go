//
//
//
package main

import (
    . "github.com/dkoby/go/mromfs/internal"
    "os"
    "log"
    "path"
    "io"
    "io/ioutil"
    "strings"
    "encoding/binary"
)

var success bool = false

//
//
//
func main() {
    defer func() {
        recover();
    }()

    log.SetFlags(0)

    if (len(os.Args) < (2 + 1)) {
        usage();
    }

    inputDirName   := os.Args[1]
    outputFileName := os.Args[2]
    var labelName string
    if (len(os.Args) >= (3 + 1)) {
        labelName = os.Args[3]
    } else {
        labelName = "NOLABEL"
    }
    // Truncate label to LabelSize.
    if (len(labelName) > LabelSize) {
        labelName = labelName[0:LabelSize]
    }
    if (len(labelName) < LabelSize) {
        labelName += strings.Repeat("\x00", LabelSize - (len(labelName) % LabelSize))
    }

    log.Println("INPUT DIR      ", inputDirName)
    log.Println("OUTPUT FILE    ", outputFileName)
    log.Println("LABEL          ", labelName)

    var outputFile *os.File
    if of, err := os.Create(outputFileName); err != nil {
        log.Panicln("Failed to create output file", outputFileName, ":", err)
    } else {
        outputFile = of
    }
    defer outputFile.Close()
    defer removeOutputFile(outputFileName, &success)

    if _, err := outputFile.WriteString(MromfsTag); err != nil {
        log.Panicln("Failed to write output file:", err)
    }
    if _, err := outputFile.WriteString(labelName); err != nil {
        log.Panicln("Failed to write output file:", err)
    }

    if _, err := outputFile.Seek(HeadSize, os.SEEK_SET); err != nil {
        log.Panicln("Failed to seek output file:", err)
    }

    var offset uint32 = HeadSize
    var files = sortFiles(getFiles(inputDirName, nil), func(f1, f2 *fileInfo) int {
        return strings.Compare(f1.fullName, f2.fullName)
    })
    //
    //XXX Limit HEAD of file to 512 bytes, so file name must
    //    not exceed Align.
    //
    for index, file := range(files) {
        func() {
            fileName := file.fullName[len(inputDirName):]
            log.Println("Process", fileName)

            var inputFile *os.File
            if inpf, err := os.Open(file.fullName); err != nil {
                log.Panicln("Failed to open file", file.fullName, ":", err)
            } else {
                inputFile = inpf
            }
            defer inputFile.Close()

            var next uint32 = offset
            next += Align
            next += uint32(file.fileInfo.Size())
            if (next % Align != 0) {
                next += Align - (next % Align)
            }
            fileName += strings.Repeat("\x00", NameAlign - (len(fileName) % NameAlign))

            // Next.
            nextWrite := next
            if (index == len(files) - 1) {
                nextWrite = 0
            }
            if err := binary.Write(outputFile, binary.LittleEndian,
                nextWrite); err != nil {
                log.Panicln("Failed to write output file: ", err)
            }
            // File size.
            if err := binary.Write(outputFile, binary.LittleEndian,
                uint32(file.fileInfo.Size())); err != nil {
                log.Panicln("Failed to write output file: ", err)
            }
            // Offset of file data.
            if err := binary.Write(outputFile, binary.LittleEndian,
                uint32(offset + Align)); err != nil {
                log.Panicln("Failed to write output file: ", err)
            }
            // Name.
            if _, err := outputFile.WriteString(fileName); err != nil {
                log.Panicln("Failed to write output file: ", err)
            }
            if cofs, err := outputFile.Seek(0, os.SEEK_CUR); err != nil {
                log.Panicln("Failed to seek output file:", err)
            } else {
                // NOTE Name must have two null terminators, so "-2". 
                if uint32(cofs) - offset > (FileHeadSize - 2) {
                    log.Panicln("File name too big", fileName)
                }
            }

            if _, err := outputFile.Seek(int64(offset + FileHeadSize), os.SEEK_SET); err != nil {
                log.Panicln("Failed to seek output file:", err)
            }

            var chunk = make([]byte, 512)
            for {
                n, err := inputFile.Read(chunk)
                if err == io.EOF {
                    break
                }
                if err != nil {
                    log.Panicln("Failed to read file", file.fullName, ":", err)
                }
                if _, err := outputFile.Write(chunk[0: n]); err != nil {
                    log.Panicln("Failed to write output file: ", err)
                }
            }

            // Align.
            if cofs , err := outputFile.Seek(0, os.SEEK_CUR); err != nil {
                log.Panicln("Failed to seek output file:", err)
            } else {
                if cofs % Align != 0 {
                    var chunk = make([]byte, Align)
                    if _, err := outputFile.Write(chunk[0: Align - (cofs % Align)]); err != nil {
                        log.Panicln("Failed to write output file: ", err)
                    }
                }
            }

            // Seek to next.
            if (nextWrite != 0) {
                if _, err := outputFile.Seek(int64(next), os.SEEK_SET); err != nil {
                    log.Panicln("Failed to seek output file:", err)
                }
            }
            offset = next
        }()
    }

    success = true

}
//
//
//
func usage() {
    log.Println("Usage: " + path.Base(os.Args[0]) + " [INPUT DIR] [OUTPUT FILE] <LABEL>")
    log.Panicln()
}

type sortFunction func(fileInfo, fileInfo) bool
//
// XXX Unoptimal algorithm.
//
func sortFiles(files []fileInfo, swap func(f1, f2 *fileInfo) int) []fileInfo {
    for n := 0; n < len(files) - 1; n++ {
        swapped := false
        for m := n; m < len(files) - 1; m++ {
            d := swap(&files[m], &files[m + 1])
            if (d != 0) {
                swapped = true
                if (d > 0) {
                    files[m], files[m + 1] = files[m + 1], files[m]
                }
            }
        }
        if (!swapped) {
            break
        }
    }
    return files
}
//
//
//
func getFiles(dirName string, files []fileInfo) []fileInfo {
    if (files == nil) {
        files = make([]fileInfo, 0)
    }

    finfo, err := ioutil.ReadDir(dirName)
    if (err != nil) {
        log.Panic("Failed to read directory", dirName, ":", err)
    }
    for _, fi := range(finfo) {
        if (fi.IsDir()) {
            files = getFiles(path.Join(dirName, fi.Name()), files);
        } else {
//            log.Println(path.Join(dirName, fi.Name()))
            files = append(files, fileInfo {
                fullName: path.Join(dirName, fi.Name()),
                fileInfo: fi,
            })
        }
    }

    return files
}
//
//
//
func removeOutputFile(fileName string, success *bool) {
    if (!*success) {
        os.Remove(fileName)
    }
}

type fileInfo struct {
    fullName string
    fileInfo os.FileInfo
}

