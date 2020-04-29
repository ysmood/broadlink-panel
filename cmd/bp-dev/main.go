package main

import (
	"os"

	. "github.com/ysmood/kit"
)

func main() {
	os.Chdir(ThisDirPath() + "/../..")

	Tasks().Add(
		Task("build", "cross build project").Run(func() {
			E(Remove("dist/**"))

			Exec(
				"go", "build",
				"-ldflags=-w -s",
				"-o", "dist/broadlink-panel",
				"./lib",
			).MustDo()

			E(Copy("web", "dist/web"))

		}),
		Task("dev", "").Init(func(cmd TaskCmd) func() {
			cmd.Default()

			return func() {
				Guard("go", "run", "./lib", "@app.conf").
					Patterns("lib/**", "*.conf", "go.*").
					MustDo()
			}
		}),
	).Do()
}
