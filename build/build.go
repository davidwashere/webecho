package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	releaseDir   = "bin/release"
	buildDir     = "bin/build"
	buildTargets = []BuildTarget{
		{"linux", "amd64", "webecho"},
		{"darwin", "amd64", "webecho"},
		{"windows", "amd64", "webecho.exe"},
	}
	listFlag bool
)

type BuildTarget struct {
	OS       string
	Arch     string
	Filename string
}

func (b BuildTarget) String() string {
	return fmt.Sprintf("{os: %v, arch: %v, filename: %v}", b.OS, b.Arch, b.Filename)
}

func (b BuildTarget) BuildDir() string {
	return filepath.Join(buildDir, fmt.Sprintf("%s-%s", b.OS, b.Arch))
}

func (b BuildTarget) BuildFileName() string {
	return filepath.Join(b.BuildDir(), b.Filename)
}

func (b BuildTarget) ReleaseDir() string {
	return filepath.Join(releaseDir)
}

func (b BuildTarget) ReleaseFileName() string {
	var ext string
	switch b.OS {
	case "windows":
		ext = ".zip"
	default:
		ext = ".tgz"
	}

	// webecho-linux-amd64.zip
	return filepath.Join(b.ReleaseDir(), fmt.Sprintf("webecho-%s-%s%s", b.OS, b.Arch, ext))
}

func init() {
	flag.BoolVar(&listFlag, "list", false, "lists files built")
}

func main() {
	flag.Parse()
	if listFlag {
		listReleaseFiles()
		return
	}
	buildall()
}

func buildall() {
	for _, target := range buildTargets {
		log.Printf(" Building: %s", target.BuildFileName())
		build(target)

		log.Printf("Archiving: %s", target.ReleaseFileName())
		archive(target)
	}
}

func build(t BuildTarget) {
	cmd := exec.Command("go", "build", "-o", t.BuildFileName(), ".")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOOS=%s", t.OS))
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOARCH=%s", t.Arch))
	cmd.Env = append(cmd.Env, "CGO_ENABLED=0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func archive(t BuildTarget) {
	outFileName := t.ReleaseFileName()
	files := []string{t.BuildFileName()}

	if err := os.MkdirAll(filepath.Dir(outFileName), 0755); err != nil {
		log.Fatal(err)
	}

	if strings.HasSuffix(outFileName, ".zip") {
		if err := ZipFiles(outFileName, files); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := CreateTarball(outFileName, files); err != nil {
			log.Fatal(err)
		}
	}
}

func listReleaseFiles() {
	sb := strings.Builder{}
	for _, t := range buildTargets {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(filepath.Base(t.ReleaseFileName()))
	}
	fmt.Print(sb.String())
}
