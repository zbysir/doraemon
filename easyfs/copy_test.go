package easyfs

import (
	"fmt"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/zbysir/doraemon/db"
	"github.com/zbysir/doraemon/gobilly"
	"testing"
)

func TestCopy(t *testing.T) {
	d := db.NewKvDb("./editor/database")

	st, err := d.Open(fmt.Sprintf("project_1"), "theme")
	if err != nil {
		t.Fatal(err)
	}
	fsTheme := gobilly.NewDbFs(st)
	if err != nil {
		t.Fatal(err)
	}

	err = CopyDir("/", "", gobilly.NewStdFs(fsTheme), osfs.New("../.cached"))
	if err != nil {
		t.Fatal(err)
	}
}
