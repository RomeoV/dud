package stage

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kevin-hanselman/duc/src/artifact"
	"github.com/kevin-hanselman/duc/src/fsutil"
)

// A Stage holds all information required to reproduce data. It is the primary
// building block of Duc pipelines.
type Stage struct {
	// The string to be evaluated and executed by a shell.
	Command string
	// WorkingDir is a directory path relative to the Duc root directory. An
	// empty value means the Stage's working directory _is_ the Duc root
	// directory. All outputs and dependencies of the Stage are themselves
	// relative to WorkingDir. The Stage's Command is executed in WorkingDir.
	WorkingDir string
	// Dependencies is a set of Artifacts which the Stage's Command needs to
	// operate. The Artifacts are keyed by their Path for faster lookup.
	Dependencies map[string]*artifact.Artifact
	// Outputs is a set of Artifacts which are owned by the Stage. The
	// Artifacts are keyed by their Path for faster lookup.
	Outputs map[string]*artifact.Artifact
}

type stageFileFormat struct {
	Command      string               `yaml:",omitempty"`
	WorkingDir   string               `yaml:",omitempty"`
	Dependencies []*artifact.Artifact `yaml:",omitempty"`
	Outputs      []*artifact.Artifact
}

// Status is a map of Artifact Paths to statuses
type Status map[string]artifact.ArtifactWithStatus

func (stg Stage) toFileFormat() (out stageFileFormat) {
	out.Command = stg.Command
	out.WorkingDir = stg.WorkingDir

	if len(stg.Dependencies) > 0 {
		out.Dependencies = make([]*artifact.Artifact, len(stg.Dependencies))
		var i int = 0
		for _, art := range stg.Dependencies {
			out.Dependencies[i] = art
			i++
		}
	}

	if len(stg.Outputs) > 0 {
		out.Outputs = make([]*artifact.Artifact, len(stg.Outputs))
		var i int = 0
		for _, art := range stg.Outputs {
			out.Outputs[i] = art
			i++
		}
	}
	return
}

func (sff stageFileFormat) toStage() (stg Stage) {
	stg.Command = sff.Command
	stg.WorkingDir = sff.WorkingDir

	if len(sff.Dependencies) > 0 {
		stg.Dependencies = make(
			map[string]*artifact.Artifact,
			len(sff.Dependencies),
		)
		for _, art := range sff.Dependencies {
			stg.Dependencies[art.Path] = art
		}
	}

	if len(sff.Outputs) > 0 {
		stg.Outputs = make(
			map[string]*artifact.Artifact,
			len(sff.Outputs),
		)
		for _, art := range sff.Outputs {
			stg.Outputs[art.Path] = art
		}
	}
	return
}

// IsEquivalent return true if the two Stage objects are identical besides
// Artifact Checksum fields.
func (stg *Stage) IsEquivalent(other Stage) bool {
	if stg.Command != other.Command {
		return false
	}
	if stg.WorkingDir != other.WorkingDir {
		return false
	}
	if len(stg.Outputs) != len(other.Outputs) {
		return false
	}
	if len(stg.Dependencies) != len(other.Dependencies) {
		return false
	}
	for path := range stg.Outputs {
		if !stg.Outputs[path].IsEquivalent(*other.Outputs[path]) {
			return false
		}
	}
	for path := range stg.Dependencies {
		if !stg.Dependencies[path].IsEquivalent(*other.Dependencies[path]) {
			return false
		}
	}
	return true
}

// for mocking
var fromYamlFile = fsutil.FromYamlFile

// FromFile loads a Stage from a file. If a lock file exists and is equivalent
// (see stage.IsEquivalent), it loads the Stage's locked version.
var FromFile = func(stagePath string) (Stage, bool, error) {
	var (
		stg, locked Stage
		sff         stageFileFormat
	)
	if err := fromYamlFile(stagePath, &sff); err != nil {
		return stg, false, err
	}
	stg = sff.toStage()
	lockPath := FilePathForLock(stagePath)
	err := fromYamlFile(lockPath, &sff)
	if os.IsNotExist(err) {
		return stg, false, nil
	} else if err != nil {
		return locked, false, err
	}
	locked = sff.toStage()
	if locked.IsEquivalent(stg) {
		return locked, true, nil
	}
	return stg, false, nil
}

// ToFile writes a Stage to the given file path. It is important to use this
// method instead of bare fsutil.ToYamlFile because a Stage file is converted
// to a simplified format when stored on disk.
func (stg *Stage) ToFile(path string) error {
	return fsutil.ToYamlFile(path, stg.toFileFormat())
}

// FromPaths creates a Stage from one or more file paths.
// TODO: rename or delete (to differentiate from FromFile)
func FromPaths(isRecursive bool, paths ...string) (stg Stage, err error) {
	stg.Outputs = make(map[string]*artifact.Artifact, len(paths))

	for _, path := range paths {
		stg.Outputs[path], err = artifact.FromPath(path, isRecursive)
		if err != nil {
			return
		}
	}
	return
}

// FilePathForLock returns the lock file path given a Stage path.
func FilePathForLock(stagePath string) string {
	var str strings.Builder
	// TODO: check for valid YAML, or at least .y(a)ml extension?
	// TODO: check for .lock suffix already in input?
	str.WriteString(stagePath)
	str.WriteString(".lock")
	return str.String()
}

// CreateCommand return an exec.Cmd for the Stage.
func (stg Stage) CreateCommand() *exec.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	cmd := exec.Command(shell, "-c", stg.Command)
	cmd.Dir = filepath.Clean(stg.WorkingDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}