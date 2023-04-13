package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	flag.Parse()

	targetDir := flag.Arg(0)
	if targetDir == "" {
		targetDir = "."
	}

	e := adaptDir(targetDir)
	if e != nil {
		log.Println(e)
		return
	}

}

func adaptDir(dir string) error {
	chunkID, e := findChunkID(dir)
	if e != nil {
		log.Println(e)
		return e
	}

	replacements := make(map[string]string)
	// _next => next
	e = os.Rename(path.Join(dir, "_next"), path.Join(dir, "next"))
	if e != nil {
		log.Println(e)
		return e
	}

	replacements["/_next/"] = "/next/"

	// next/static/{chunkID}/
	// _buildManifest.js => buildManifest.js
	relative := "next/static/" + chunkID + "/"
	from := relative + "_buildManifest.js"
	to := relative + "buildManifest.js"
	e = os.Rename(path.Join(dir, from), path.Join(dir, to))
	if e != nil {
		log.Println(e)
		return e
	}

	replacements[from] = to

	// next/static/{chunkID}/
	// _ssgManifest.js => ssgManifest.js
	from = relative + "_ssgManifest.js"
	to = relative + "ssgManifest.js"
	e = os.Rename(path.Join(dir, from), path.Join(dir, to))
	if e != nil {
		log.Println(e)
		return e
	}

	replacements[from] = to

	// next/static/chunks/pages/
	// _app-556f37b589d6ac62.js => app-556f37b589d6ac62.js
	relative = "next/static/chunks/pages/"
	appPage, e := findChildStartsWith(path.Join(dir, relative), "_app-")
	if e != nil {
		log.Println(e)
		return e
	}
	from = relative + appPage
	to = relative + strings.TrimPrefix(appPage, "_")
	e = os.Rename(path.Join(dir, from), path.Join(dir, to))
	if e != nil {
		log.Println(e)
		return e
	}

	replacements[from] = to

	// next/static/chunks/pages/
	// _error-8353112a01355ec2.js => error-8353112a01355ec2.js
	errorPage, e := findChildStartsWith(path.Join(dir, relative), "_error-")
	if e != nil {
		log.Println(e)
		return e
	}
	from = relative + errorPage
	to = relative + strings.TrimPrefix(errorPage, "_")
	e = os.Rename(path.Join(dir, from), path.Join(dir, to))
	if e != nil {
		log.Println(e)
		return e
	}

	replacements[from] = to

	// find and replace
	e = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		switch filepath.Ext(info.Name()) {
		case ".html", ".js":
		default:
			return nil
		}
		bs, e := os.ReadFile(path)
		if e != nil {
			log.Println(e)
			return e
		}
		str := string(bs)
		for k, v := range replacements {
			str = strings.ReplaceAll(str, k, v)
		}
		if str != string(bs) {
			fo, e := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if e != nil {
				log.Println(e)
				return e
			}
			defer fo.Close()
			fo.WriteString(str)
			fmt.Println(path)
		}
		return nil
	})
	if e != nil {
		log.Println(e)
		return e
	}

	return nil
}

func findChunkID(rootDir string) (string, error) {
	dirs, e := ioutil.ReadDir(path.Join(rootDir, "_next"))
	if e != nil {
		log.Println(e)
		return "", e
	}
	for _, v := range dirs {
		if v.Name() == "static" {
			continue
		}
		if len(v.Name()) > 6 {
			return v.Name(), nil
		}
	}
	return "", fmt.Errorf("chunk id not found")
}

func findChildStartsWith(dir, prefix string) (string, error) {
	dirs, e := ioutil.ReadDir(path.Join(dir))
	if e != nil {
		log.Println(e)
		return "", e
	}
	for _, v := range dirs {
		if strings.HasPrefix(v.Name(), prefix) {
			return v.Name(), nil
		}
	}
	return "", fmt.Errorf("child starts with [%s] not found in [%s]", prefix, dir)
}
