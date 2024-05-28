package models

import (
    "archive/tar"
    "compress/gzip"
    "fmt"
    "io"
    "os"
    "path/filepath"
)

func BuildImage(outputFilename string, sourceDir string) error {
    tarFile, err := os.Create(outputFilename)
    if err != nil {
        return fmt.Errorf("error creating tar file: %v", err)
    }
    defer tarFile.Close()

    gzipWriter := gzip.NewWriter(tarFile)
    defer gzipWriter.Close()

    tarWriter := tar.NewWriter(gzipWriter)
    defer tarWriter.Close()

    err = filepath.Walk(sourceDir, func(file string, fi os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        header, err := tar.FileInfoHeader(fi, fi.Name())
        if err != nil {
            return err
        }

        header.Name, err = filepath.Rel(filepath.Dir(sourceDir), file)
        if err != nil {
            return err
        }

        if err := tarWriter.WriteHeader(header); err != nil {
            return err
        }

        if fi.IsDir() {
            return nil
        }

        f, err := os.Open(file)
        if err != nil {
            return err
        }
        defer f.Close()

        if _, err := io.Copy(tarWriter, f); err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        return fmt.Errorf("error adding files to tar: %v", err)
    }

    fmt.Println("Image tarball created successfully.")
    return nil
}

