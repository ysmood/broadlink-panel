package main

import (
	"os"

	g "github.com/ysmood/gokit"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app      = kingpin.New("dev", "dev tool for gokit")
	cmdBuild = app.Command("build", "cross build project")
	cmdDev   = app.Command("dev", "").Default()
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case cmdDev.FullCommand():
		g.Guard([]string{
			"go", "run", "./lib", "@app.conf",
		}, []string{"lib/**"}, nil)

	case cmdBuild.FullCommand():
		g.Remove("dist/**")

		g.E(g.Exec([]string{
			"go", "build",
			"-ldflags=-w -s",
			"-o", "dist/broadlink-panel",
			"./lib",
		}, nil))

		g.E(g.Copy("web", "dist/web"))
	}
}
