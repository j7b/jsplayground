// +build !js

package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/j7b/jsplayground/travail"
)

func init() {
	log.SetFlags(0)
}

var gopath, goroot string

func main() {
	flag.Parse()
	goexec, err := exec.LookPath("go")
	if err != nil {
		log.Fatal(err)
	}
	env := os.Environ()
	for i, s := range env {
		if strings.HasPrefix(s, "GOPATH") {
			env[i] = `OLD` + s
			gopath = s
			break
		}
	}
	td, err := ioutil.TempDir("/tmp", "jsplayground-")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Building in %s", td)
	cmd := exec.Command(goexec, "get", "github.com/gopherjs/gopherjs")
	cmd.Env = append(env, "GOPATH="+td)
	o, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(string(o))
	}
	extras := make(map[string]string)
	if fa := flag.Args(); len(fa) > 0 {
		log.Println("Installing additional packages")
		args := []string{"get", "-m"}
		args = append(args, fa...)
		cmd = exec.Command(filepath.Join(td, "bin", "gopherjs"), args...)
		cmd.Env = append(env, "GOPATH="+td)
		o, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(string(o))
		}
		for _, pk := range fa {
			m, err := travail.Map(pk, filepath.Join(td, "src", filepath.FromSlash(pk)))
			if err != nil {
				log.Printf("Couldn't add %s for imports.json", pk)
			}
			for k, v := range m {
				extras[k] = v
			}
		}
	}
	cmd = exec.Command(goexec, "env", "GOROOT")
	o, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatal(string(o))
	}
	gr := filepath.Join(strings.TrimSpace(string(o)), "src")
	log.Printf("Using GOROOT %s", gr)
	if err = filepath.Walk(gr, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			p = strings.Replace(p, gr, "", 1)
			return os.MkdirAll(filepath.Join(td, "src", p), 0755)
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		d, err := os.Create(filepath.Join(td, "src", strings.Replace(p, gr, "", 1)))
		if err != nil {
			return err
		}
		defer d.Close()
		_, err = io.Copy(d, f)
		return err
	}); err != nil {
		log.Fatal(err)
	}
	for i, s := range env {
		if strings.HasPrefix(s, "GOROOT") {
			env[i] = `OLD` + s
			goroot = s
			break
		}
	}
	args := []string{"install", "-m"}
	args = append(args, (strings.Split(stdlibs, " "))...)
	cmd = exec.Command(filepath.Join(td, "bin", "gopherjs"), args...)
	cmd.Env = append(env, "GOROOT="+td)
	o, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatal(string(o))
	}
	if err := os.Chdir(filepath.Join(td, "pkg")); err != nil {
		log.Fatal(err)
	}
	alldirs, err := filepath.Glob("*")
	if err != nil {
		log.Fatal(err)
	}
	dirs, err := filepath.Glob("*_js_min")
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range dirs {
		f, err := os.Open(d)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		names, err := f.Readdirnames(-1)
		if err != nil {
			log.Fatal(err)
		}
		for _, n := range names {
			from := filepath.Join(d, n)
			if err := os.Rename(from, n); err != nil {
				log.Fatal(err)
			}
		}
	}
	for _, dir := range alldirs {
		os.RemoveAll(dir)
	}
	if err := os.Chdir(td); err != nil {
		log.Fatal(err)
	}
	os.RemoveAll("src")
	os.RemoveAll("bin")
	if len(extras) > 0 {
		f, err := os.Create(filepath.Join("pkg", "imports.json"))
		switch {
		case err != nil:
			log.Println("Couldn't create imports.json")
		default:
			defer f.Close()
			encode := json.NewEncoder(f).Encode
			encode(extras)
		}
	}
	log.Printf("JS Playground packages at %s", filepath.Join(td, "pkg"))
}

const stdlibs = `archive/tar ` +
	`archive/zip ` +
	`bufio ` +
	`bytes ` +
	`compress/bzip2 ` +
	`compress/flate ` +
	`compress/gzip ` +
	`compress/lzw ` +
	`compress/zlib ` +
	`container/heap ` +
	`container/list ` +
	`container/ring ` +
	`crypto/aes ` +
	`crypto/cipher ` +
	`crypto/des ` +
	`crypto/dsa ` +
	`crypto/ecdsa ` +
	`crypto/elliptic ` +
	`crypto/hmac ` +
	`crypto/md5 ` +
	`crypto/rand ` +
	`crypto/rc4 ` +
	`crypto/rsa ` +
	`crypto/sha1 ` +
	`crypto/sha256 ` +
	`crypto/sha512 ` +
	`crypto/subtle ` +
	`database/sql/driver ` +
	`debug/gosym ` +
	`debug/pe ` +
	`encoding/ascii85 ` +
	`encoding/asn1 ` +
	`encoding/base32 ` +
	`encoding/base64 ` +
	`encoding/binary ` +
	`encoding/csv ` +
	`encoding/gob ` +
	`encoding/hex ` +
	`encoding/json ` +
	`encoding/pem ` +
	`encoding/xml ` +
	`errors ` +
	`fmt ` +
	`go/ast ` +
	`go/doc ` +
	`go/format ` +
	`go/printer ` +
	`go/token ` +
	`hash/adler32 ` +
	`hash/crc32 ` +
	`hash/crc64 ` +
	`hash/fnv ` +
	`html ` +
	`html/template ` +
	`image ` +
	`image/color ` +
	`image/draw ` +
	`image/gif ` +
	`image/jpeg ` +
	`image/png ` +
	`index/suffixarray ` +
	`io ` +
	`io/ioutil ` +
	`math ` +
	`math/big ` +
	`math/cmplx ` +
	`math/rand ` +
	`mime ` +
	`net/http/cookiejar ` +
	`net/http/fcgi ` +
	`net/http/httptest ` +
	`net/http/httputil ` +
	`net/mail ` +
	`net/smtp ` +
	`net/textproto ` +
	`net/url ` +
	`path ` +
	`path/filepath ` +
	`reflect ` +
	`regexp ` +
	`regexp/syntax ` +
	`runtime/internal/sys ` +
	`sort ` +
	`strconv ` +
	`strings ` +
	`sync/atomic ` +
	`testing ` +
	`testing/iotest ` +
	`testing/quick ` +
	`text/scanner ` +
	`text/tabwriter ` +
	`text/template ` +
	`text/template/parse ` +
	`time ` +
	`unicode ` +
	`unicode/utf16 ` +
	`unicode/utf8`
