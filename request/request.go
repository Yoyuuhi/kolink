package request

import (
	"os"

	"github.com/juju/errors"
	"gopkg.in/yaml.v2"
)

type Defs struct {
	RepositoryName string       `yaml:"repositoryName"`
	RequestDefs    []RequestDef `yaml:"requestDefs"`
}

type RequestDef struct {
	OutDir string    `yaml:"outDir"`
	Callee CalleeDef `yaml:"callee"`
	Caller CallerDef `yaml:"caller"`

	FileParams []FileParam `yaml:"fileParams"`

	FileParamMap map[string]FileParam
}

type CalleeDef struct {
	Dir           string      `yaml:"dir"`
	FileFocus     []string    `yaml:"fileFocus"`
	FuncFocus     []string    `yaml:"funcFocus"`
	IgnoreNewFunc bool        `yaml:"ignoreNewFunc"`
	IgnoreTest    bool        `yaml:"ignoreTest"`
	Layout        string      `yaml:"layut"`
	FileParams    []FileParam `yaml:"fileParams"`

	FocusFileMap map[string]bool
	FocusFuncMap map[string]bool
}

type FileParam struct {
	FileName   string      `yaml:"fileName"`
	Split      bool        `yaml:"split"`
	Layout     string      `yaml:"layout"`
	Attributes []Attribute `yaml:"attributes"`
}

type Attribute struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type CallerDef struct {
	Dir        string   `yaml:"dir"`
	FileFocus  []string `yaml:"fileFocus"`
	IgnoreTest bool     `yaml:"ignoreTest"`

	FocusFileMap map[string]bool
}

func GetRequestDefs() (*Defs, error) {
	var requestDefs *Defs
	bytes, err := os.ReadFile("./kolink.yml")
	if err != nil {
		return requestDefs, errors.Trace(err)
	}

	if err := yaml.Unmarshal(bytes, &requestDefs); err != nil {
		return requestDefs, errors.Trace(err)
	}

	for i, requestDef := range requestDefs.RequestDefs {
		requestDefs.RequestDefs[i].Callee.FocusFileMap = map[string]bool{}
		for _, file := range requestDef.Callee.FileFocus {
			requestDefs.RequestDefs[i].Callee.FocusFileMap[file] = true
		}
		requestDefs.RequestDefs[i].Callee.FocusFuncMap = map[string]bool{}
		for _, file := range requestDef.Callee.FuncFocus {
			requestDefs.RequestDefs[i].Callee.FocusFuncMap[file] = true
		}
		requestDefs.RequestDefs[i].Caller.FocusFileMap = map[string]bool{}
		for _, file := range requestDef.Caller.FileFocus {
			requestDefs.RequestDefs[i].Caller.FocusFileMap[file] = true
		}

		requestDefs.RequestDefs[i].FileParamMap = map[string]FileParam{}
		for _, fileParam := range requestDef.FileParams {
			requestDefs.RequestDefs[i].FileParamMap[fileParam.FileName] = fileParam
		}
	}

	return requestDefs, nil
}
