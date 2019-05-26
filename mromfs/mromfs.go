// Package mromfs provides function for reading files from in-memory Mromfs.
//
package mromfs

import (
    . "github.com/dkoby/go/mromfs/internal"
    "bytes"
    "encoding/binary"
    "errors"
    "strings"
)

//
//
type Mromfs struct {
    Label string
    data []byte
    firstFile uint32
}
// New creates new filesystem from byte slice.
//
func New(data []byte) (*Mromfs, error) {
    var mromfs *Mromfs = new(Mromfs)
    var err error
    var offset int

    if len(data) < HeadSize {
        return nil, errors.New("Image data too small")
    }
    offset = 0
    if string(data[0: len(MromfsTag)]) != MromfsTag {
        return nil, errors.New("Image data has not valid tag")
    }
    offset += len(MromfsTag)
    mromfs.Label = strings.TrimSuffix(string(data[offset: offset + LabelSize]), "\x00")
    mromfs.data = data
    if len(data) >= HeadSize + Align {
        mromfs.firstFile = HeadSize
    } else {
        mromfs.firstFile = 0
    }
    return mromfs, err
}
// Open opens file for reading. Opened file can be used as "bytes.Buffer", but
// all operations done with that buffer stays localy - filesystem not touched.
func (s *Mromfs) Open(name string) (*File, error){
    var file *File = new(File)

    if (s.firstFile == 0) {
        return nil, errors.New("Failed to open file - no such file")
    }

    var offset uint32 = s.firstFile 
    for {
        var next uint32
        buffer := bytes.NewBuffer(s.data[offset:])

        // Next file.
        if err := binary.Read(buffer, binary.LittleEndian, &next); err != nil {
            return nil, errors.New("Failed to open (1)")
        }
        // File size.
        if err := binary.Read(buffer, binary.LittleEndian, &file.Size); err != nil {
            return nil, errors.New("Failed to open (2)")
        }
        // Data offset.
        var baseOffset uint32
        if err := binary.Read(buffer, binary.LittleEndian, &baseOffset); err != nil {
            return nil, errors.New("Failed to open (3)")
        }
        // Name
        if name, err := buffer.ReadString(0x00); err != nil {
            return nil, errors.New("Failed to open (4)")
        } else {
            file.Name = name[0: len(name) - 1]
        }
        if (name == file.Name) {
            var fileData []byte = make([]byte, file.Size)
            copy(fileData, s.data[baseOffset:baseOffset + file.Size])
            file.Buffer = bytes.NewBuffer(fileData)
            return file, nil
        }

        if (next == 0) {
            break
        }
        offset = next
    }

    return nil, errors.New("Failed to open file - no such file")
}
// 
type File struct {
    Name string
    Size uint32
   *bytes.Buffer
}

