package ff

import (
	"errors"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/functions/sc"
	"skeyevss/core/tps"
)

type (
	FFMpeg struct {
		conf tps.YamlFFMpeg
	}
	MediaInfo struct {
		Duration     int
		VideoDecodec string
		AudioDecodec string
		Aspect       string
		Rotate       int
	}
)

var execOnce sync.Once

func NewFFMpeg(conf tps.YamlFFMpeg) *FFMpeg {
	return &FFMpeg{conf: conf}
}

func (f *FFMpeg) logError(content ...interface{}) {
	functions.LogError(content...)
}

func (f *FFMpeg) logInfo(content ...interface{}) {
	functions.LogInfo(content...)
}

func (f *FFMpeg) binName() string {
	switch runtime.GOOS {
	case "windows":
		return path.Join(f.conf.Home, "ffmpeg.exe")

	default:
		return path.Join(f.conf.Home, "ffmpeg")
	}
}

func (f *FFMpeg) exec(args ...string) error {
	var binPath = f.binName()
	execOnce.Do(func() {
		if err := sc.ExecPermission(binPath); err != nil {
			panic(err)
		}
	})

	var cmd = exec.Command(binPath, args...)
	cmd.Dir = filepath.ToSlash(f.conf.Home)
	cmd.Stderr = nil
	cmd.Stdout = nil

	f.logInfo("exec command: ", cmd.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg exec failed \n command: %s \n err: %v", cmd.String(), err)
	}

	return nil
}

func (f *FFMpeg) SnapFile(file, dest string) error {
	if file == "" || dest == "" {
		return errors.New("file or dest is empty")
	}

	var ext = filepath.Ext(file)
	var params = []string{"-hide_banner", "-i", file, "-y", "-f", "image2", dest}
	if strings.ToLower(ext) == ".mp3" {
		params = []string{"-hide_banner", "-i", file, "-y", "-an", "-vcodec", "copy", dest}
	}

	return f.exec(params...)
}
