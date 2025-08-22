package main

import (
	"os"
	"path/filepath"
	"time"
)

func main() {
	loadConfigWithCallbackWhenKeyIsTrue()
	u := NewUsecase()
	u.printConf()
	<-make(chan bool)
}

// writeJSONAtomic menulis ke file secara (relatif) atomik
func writeJSONAtomic(path, content string) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	tmp := filepath.Join(dir, "."+base+".tmp")

	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		panic(err)
	}
	// rename biasanya memicu event fsnotify (rename/create)
	if err := os.Rename(tmp, path); err != nil {
		panic(err)
	}
	// sentuh mtime (optional; beberapa FS sudah memicu event saat rename)
	_ = os.Chtimes(path, time.Now(), time.Now())
}
