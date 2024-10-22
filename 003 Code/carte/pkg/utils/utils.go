package utils

import (
    "archive/tar"
    "compress/gzip"
    //"fmt"
    "io"
    "os"
    "path/filepath"
)

// ExtractTar는 tar 파일을 추출합니다.
func ExtractTar(src, dest string) error {
    if err := os.MkdirAll(dest, 0755); err != nil {
        return err
    }

    file, err := os.Open(src)
    if err != nil {
        return err
    }
    defer file.Close()

    gzipReader, err := gzip.NewReader(file)
    if err != nil {
        return err
    }
    defer gzipReader.Close()

    tarReader := tar.NewReader(gzipReader)

    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        target := filepath.Join(dest, header.Name)

        switch header.Typeflag {
        case tar.TypeDir:
            os.MkdirAll(target, header.FileInfo().Mode())
        case tar.TypeReg:
            outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, header.FileInfo().Mode())
            if err != nil {
                return err
            }
            io.Copy(outFile, tarReader)
            outFile.Close()
        }
    }

    return nil
}
