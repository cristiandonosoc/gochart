package unreal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (p *Project) UBT(args []string) error {
	//cmd := ubtBuildBat(p, args)
	cmd := experimentalDirectCmd(p, args)

	fmt.Println("> Running:", cmd.Args)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting %v: %w", cmd.Args, err)
	}

	exitCode := 0
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			exitCode = exiterr.ExitCode()
		} else {
			return fmt.Errorf("running %v: %w", cmd.Args, err)
		}
	}

	if exitCode != 0 {
		return fmt.Errorf("UBT exited with error code %d", exitCode)
	}

	return nil
}

func ubtBuildBat(p *Project, args []string) *exec.Cmd {
	editor := p.Config.EditorConfig

	buildBat := filepath.Join(editor.EditorDir, "Engine", "Build", "BatchFiles", "Build.bat")
	cmd := exec.Command(buildBat, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

func experimentalDirectCmd(p *Project, args []string) *exec.Cmd {
	editor := p.Config.EditorConfig

	var cmdargs []string
	cmdargs = append(cmdargs, editor.UBTDll)
	cmdargs = append(cmdargs, args...)
	cmdargs = append(cmdargs, "-Project", p.Config.UProject)

	cmd := exec.Command(editor.Dotnet, cmdargs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Join(editor.EditorDir, "Engine", "Source")

	return cmd


}
