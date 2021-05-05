package splitter

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	yaml "gopkg.in/laverya/yaml.v3"
)

type outputYaml struct {
	name         string
	contents     string
	overridePath string
}

type MinimalK8sYaml struct {
	Kind     string             `json:"kind" yaml:"kind" hcl:"kind"`
	Metadata MinimalK8sMetadata `json:"metadata" yaml:"metadata" hcl:"metadata"`
}

type MinimalK8sMetadata struct {
	Name      string `json:"name" yaml:"name" hcl:"name"`
	Namespace string `json:"namespace" yaml:"namespace" hcl:"namespace"`
}

type ListK8sYaml struct {
	APIVersion string        `json:"apiVersion" yaml:"apiVersion"`
	Kind       string        `json:"kind" yaml:"kind" hcl:"kind"`
	Items      []interface{} `json:"items" yaml:"items"`
}

func MaybeSplitMultidocYamlFs(localpath string) error {
	fs := afero.NewOsFs()
	return MaybeSplitMultidocYaml(afero.Afero{Fs: fs}, localpath, false)
}

func MaybeSplitCRDsFs(localpath string) error {
	fs := afero.NewOsFs()
	return MaybeSplitMultidocYaml(afero.Afero{Fs: fs}, localpath, true)
}

// this function is not perfect, and has known limitations. One of these is that it does not account for `\n---\n` in multiline strings.
func MaybeSplitMultidocYaml(fs afero.Afero, localPath string, combineNonCRDs bool) error {
	files, err := fs.ReadDir(localPath)
	if err != nil {
		return errors.Wrapf(err, "read files in %s", localPath)
	}

	allOutputFiles := []outputYaml{}
	allCrds := []string{}

	for _, file := range files {
		outputFiles := []outputYaml{}
		crds := []string{}

		if file.IsDir() {
			if err := MaybeSplitMultidocYaml(fs, filepath.Join(localPath, file.Name()), combineNonCRDs); err != nil {
				return err
			}
		}

		if filepath.Ext(file.Name()) != ".yaml" && filepath.Ext(file.Name()) != ".yml" {
			// not yaml, nothing to do
			continue
		}

		inFileBytes, err := fs.ReadFile(filepath.Join(localPath, file.Name()))
		if err != nil {
			return errors.Wrapf(err, "read %s", filepath.Join(localPath, file.Name()))
		}

		filesStrings := strings.Split(string(inFileBytes), "\n---\n")

		// generate replacement yaml files
		for idx, fileString := range filesStrings {

			newOutputFiles, newCRDs, err := generateOutputYaml(idx, fileString)
			if err != nil {
				return errors.Wrapf(err, "at path %s", file.Name())
			}

			outputFiles = append(outputFiles, newOutputFiles...)
			crds = append(crds, newCRDs...)
		}

		if len(outputFiles) == 1 && len(crds) == 0 { // don't rename files if we don't have to
			outputFiles[0].overridePath = file.Name()
		}

		// delete multidoc yaml file
		err = fs.Remove(filepath.Join(localPath, file.Name()))
		if err != nil {
			return errors.Wrapf(err, "unable to remove %s", filepath.Join(localPath, file.Name()))
		}

		allOutputFiles = append(allOutputFiles, outputFiles...)
		allCrds = append(allCrds, crds...)
	}

	if combineNonCRDs {
		allOutputStrings := []string{}
		for _, outputFile := range allOutputFiles {
			allOutputStrings = append(allOutputStrings, outputFile.contents)
		}
		nonCrdsFile := outputYaml{contents: strings.Join(allOutputStrings, "\n---\n"), name: "AllResorces"}
		allOutputFiles = []outputYaml{nonCrdsFile}
	}

	if len(allCrds) > 0 {
		crdsFile := outputYaml{contents: strings.Join(allCrds, "\n---\n"), name: "CustomResourceDefinitions"}
		allOutputFiles = append(allOutputFiles, crdsFile)
	}

	// write replacement yaml
	for _, outputFile := range allOutputFiles {
		if outputFile.overridePath != "" {
			err = fs.WriteFile(filepath.Join(localPath, outputFile.overridePath), []byte(outputFile.contents), os.FileMode(0644))
		} else {
			err = fs.WriteFile(filepath.Join(localPath, outputFile.name+".yaml"), []byte(outputFile.contents), os.FileMode(0644))
		}
		if err != nil {
			return errors.Wrapf(err, "write %s", outputFile.name)
		}
	}

	return nil
}

// this function drops files with no parsable 'kind', separates out CRD definitions, and splits list yaml into multiple files
func generateOutputYaml(idx int, fileString string) ([]outputYaml, []string, error) {

	thisOutputFile := outputYaml{contents: fileString}
	theseOutputFiles := []outputYaml{}
	crds := []string{}

	thisMetadata := MinimalK8sYaml{}
	_ = yaml.Unmarshal([]byte(fileString), &thisMetadata)

	if thisMetadata.Kind == "" {
		// ignore invalid k8s yaml
		return nil, nil, nil
	}

	if thisMetadata.Kind == "CustomResourceDefinition" {
		// collate CRDs into one file
		crds = append(crds, fileString)
		return theseOutputFiles, crds, nil
	}

	if thisMetadata.Kind == "List" {
		// split list yaml into multiple files
		thisList := ListK8sYaml{}
		_ = yaml.Unmarshal([]byte(fileString), &thisList)

		for itemIdx, item := range thisList.Items {
			itemYaml, err := MarshalIndent(2, item)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "marshal item %d from file %d", itemIdx, idx)
			}

			newOutput, newCRDs, err := generateOutputYaml(itemIdx, string(itemYaml))
			if err != nil {
				return nil, nil, errors.Wrapf(err, "at file %d", idx)
			}

			theseOutputFiles = append(theseOutputFiles, newOutput...)
			crds = append(crds, newCRDs...)
		}

		return theseOutputFiles, crds, nil
	}

	fileName := GenerateNameFromMetadata(thisMetadata, idx)
	thisOutputFile.name = fileName

	theseOutputFiles = []outputYaml{thisOutputFile}
	return theseOutputFiles, crds, nil
}

func GenerateNameFromMetadata(k8sYaml MinimalK8sYaml, idx int) string {
	fileName := fmt.Sprintf("%s-%d", k8sYaml.Kind, idx)

	if k8sYaml.Metadata.Name != "" {
		fileName = k8sYaml.Kind + "-" + k8sYaml.Metadata.Name
		if k8sYaml.Metadata.Namespace != "" && k8sYaml.Metadata.Namespace != "default" {
			fileName += "-" + k8sYaml.Metadata.Namespace
		}
	}

	fileName = regexp.MustCompile(`[/\\:]`).ReplaceAllString(fileName, "-")

	return fileName
}

func MarshalIndent(indent int, in interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indent)
	enc.SetLineLength(-1)
	err := enc.Encode(in)
	if err != nil {
		return nil, errors.Wrapf(err, "marshal with indent %d", indent)
	}

	return buf.Bytes(), nil
}
