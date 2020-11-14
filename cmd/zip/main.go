package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	help = `Usage:
    zip -l zipfile.zip         # Show listing of a zipfile
    zip -e zipfile.zip target  # Extract zipfile into target dir
    zip -c zipfile.zip src ... # Create zipfile from sources`
)

var (
	flCreate  = flag.Bool("c", false, "create zipfile from sources")
	flExtract = flag.Bool("e", false, "extract zipfile into target dir")
	flList    = flag.Bool("l", false, "show listing of a zipfile")
)

func mainCreate(src string, dst string) error {
	fw, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fw.Close()

	zw := zip.NewWriter(fw)
	defer zw.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		var (
			p string
			h *zip.FileHeader
			r *os.File
			w io.Writer
		)
		if err != nil {
			return err
		}
		h, err = zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		p, err = filepath.Rel(filepath.Dir(src), path)
		if err != nil {
			return err
		}
		p = filepath.ToSlash(p)
		if info.IsDir() {
			p = p + "/"
		}
		fmt.Println(p)
		h.Name = p
		h.Method = zip.Deflate

		w, err = zw.CreateHeader(h)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		r, err = os.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()

		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}
		return nil
	})
}

func mainExtract(src string, dst string) error {
	zr, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, zf := range zr.File {
		fmt.Println(zf.Name)
		p := filepath.Join(dst, zf.Name)

		if zf.FileInfo().IsDir() {
			if err := os.Mkdir(p, zf.Mode()); err != nil {
				return err
			}
			continue
		}

		r, err := zf.Open()
		if err != nil {
			return err
		}
		w, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR|os.O_TRUNC, zf.Mode())
		if err != nil {
			return err
		}
		n, err := io.Copy(w, r)
		if err != nil {
			return err
		}
		_ = n
		w.Close()
		r.Close()
	}
	return nil
}

func mainList(src string) error {
	zr, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, zf := range zr.File {
		fmt.Println(zf.Name)
	}
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), help)
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
	}
	flag.Parse()

	if !*flCreate && !*flExtract && !*flList {
		flag.Usage()
		return
	}

	var err error
	switch {
	case *flCreate:
		err = mainCreate(flag.Arg(1), flag.Arg(0))
	case *flExtract:
		err = mainExtract(flag.Arg(0), flag.Arg(1))
	case *flList:
		err = mainList(flag.Arg(0))
	}
	if err != nil {
		fmt.Println(err)
	}
}
