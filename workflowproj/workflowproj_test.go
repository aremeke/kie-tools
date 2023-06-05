// Copyright 2023 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workflowproj

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/kiegroup/kogito-serverless-operator/api/metadata"
)

func Test_Handler_WorkflowMinimal(t *testing.T) {
	proj, err := New("default").WithWorkflow(getWorkflowMinimal()).AsObjects()
	assert.NoError(t, err)
	assert.NotNil(t, proj)
	assert.Equal(t, "hello", proj.Workflow.Name)
}

func Test_Handler_WorkflowMinimalInvalid(t *testing.T) {
	proj, err := New("default").WithWorkflow(getWorkflowMinimalInvalid()).AsObjects()
	assert.Error(t, err)
	assert.Nil(t, proj)
}

func Test_Handler_WorkflowMinimalAndProps(t *testing.T) {
	proj, err := New("default").
		Named("minimal").
		WithWorkflow(getWorkflowMinimal()).
		WithAppProperties(getWorkflowProperties()).
		AsObjects()
	assert.NoError(t, err)
	assert.NotNil(t, proj.Workflow)
	assert.NotNil(t, proj.Properties)
	assert.Equal(t, "minimal", proj.Workflow.Name)
	assert.Equal(t, "minimal-props", proj.Properties.Name)
	assert.NotEmpty(t, proj.Properties.Data)
}

func Test_Handler_WorkflowMinimalAndPropsAndSpec(t *testing.T) {
	proj, err := New("default").
		WithWorkflow(getWorkflowMinimal()).
		WithAppProperties(getWorkflowProperties()).
		AddResource("myopenapi.json", getSpecOpenApi()).
		AsObjects()
	assert.NoError(t, err)
	assert.NotNil(t, proj.Workflow)
	assert.NotNil(t, proj.Properties)
	assert.NotEmpty(t, proj.Resources)
	assert.Equal(t, "hello", proj.Workflow.Name)
	assert.Equal(t, "hello-props", proj.Properties.Name)
	assert.NotEmpty(t, proj.Properties.Data)
	assert.Equal(t, 1, len(proj.Resources))
	assert.Equal(t, "hello-openapis", proj.Resources[0].Name)
	assert.Equal(t, proj.Workflow.Annotations[metadata.GetExtResTypeAnnotation(metadata.ExtResOpenApi)], proj.Resources[0].Name)

}

func Test_Handler_WorkflowMinimalAndPropsAndSpecAndGeneric(t *testing.T) {
	proj, err := New("default").
		WithWorkflow(getWorkflowMinimal()).
		WithAppProperties(getWorkflowProperties()).
		AddResource("myopenapi.json", getSpecOpenApi()).
		AddResource("myopenapi.json", getSpecOpenApi()).
		AddResource("myopenapi2.json", getSpecOpenApi()).
		AddResource("input.json", getSpecGeneric()).
		AsObjects()
	assert.NoError(t, err)
	assert.NotNil(t, proj.Workflow)
	assert.NotNil(t, proj.Properties)
	assert.NotEmpty(t, proj.Resources)
	assert.Equal(t, "hello", proj.Workflow.Name)
	assert.Equal(t, "hello-props", proj.Properties.Name)
	assert.NotEmpty(t, proj.Properties.Data)
	assert.Equal(t, 2, len(proj.Resources))
	assert.Equal(t, "hello-openapis", proj.Resources[0].Name)
	assert.Equal(t, "hello-genericres", proj.Resources[1].Name)
	assert.Equal(t, proj.Workflow.Annotations[metadata.GetExtResTypeAnnotation(metadata.ExtResOpenApi)], proj.Resources[0].Name)
	assert.Equal(t, proj.Workflow.Annotations[metadata.GetExtResTypeAnnotation(metadata.ExtResGeneric)], proj.Resources[1].Name)
	assert.NotEmpty(t, proj.Resources[0].Data["myopenapi.json"])
	assert.NotEmpty(t, proj.Resources[1].Data["input.json"])
}

func Test_Handler_WorklflowServiceAndPropsAndSpec_SaveAs(t *testing.T) {
	handler := New("default").
		WithWorkflow(getWorkflowService()).
		WithAppProperties(getWorkflowProperties()).
		AddResource("myopenapi.json", getSpecOpenApi()).
		AddResourceTyped("schema.json", getSpecGeneric(), metadata.ExtResGeneric)
	proj, err := handler.AsObjects()
	assert.NoError(t, err)
	assert.NotNil(t, proj.Workflow)
	assert.NotNil(t, proj.Properties)
	assert.NotEmpty(t, proj.Resources)

	tmpPath, err := os.MkdirTemp("", "*-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpPath)

	assert.NoError(t, handler.SaveAsKubernetesManifests(tmpPath))
	files, err := os.ReadDir(tmpPath)
	assert.NoError(t, err)
	assert.Len(t, files, 4)

	for _, f := range files {
		if strings.HasSuffix(f.Name(), yamlFileExt) {
			contents, err := os.ReadFile(path.Join(tmpPath, f.Name()))
			assert.NoError(t, err)
			decode := scheme.Codecs.UniversalDeserializer().Decode
			k8sObj, _, err := decode(contents, nil, nil)
			assert.NoError(t, err)
			assert.NotNil(t, k8sObj)
			assert.NotEmpty(t, k8sObj.GetObjectKind().GroupVersionKind().String())
		}
	}
}

func Test_Handler_WorkflowService_SaveAs(t *testing.T) {
	handler := New("default").
		WithWorkflow(getWorkflowService())

	proj, err := handler.AsObjects()
	assert.NoError(t, err)
	assert.NotNil(t, proj.Workflow)

	tmpPath, err := os.MkdirTemp("", "*-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpPath)

	assert.NoError(t, handler.SaveAsKubernetesManifests(tmpPath))
	files, err := os.ReadDir(tmpPath)
	assert.NoError(t, err)
	assert.Len(t, files, 1)

	for _, f := range files {
		if strings.HasSuffix(f.Name(), yamlFileExt) {
			contents, err := os.ReadFile(path.Join(tmpPath, f.Name()))
			assert.NoError(t, err)
			decode := scheme.Codecs.UniversalDeserializer().Decode
			k8sObj, _, err := decode(contents, nil, nil)
			assert.NoError(t, err)
			assert.NotNil(t, k8sObj)
			assert.NotEmpty(t, k8sObj.GetObjectKind().GroupVersionKind().String())
		}
	}
}

func getWorkflowMinimalInvalid() io.Reader {
	return mustGetFile("testdata/workflows/workflow-minimal-invalid.sw.json")
}

func getWorkflowMinimal() io.Reader {
	return mustGetFile("testdata/workflows/workflow-minimal.sw.json")
}

func getWorkflowService() io.Reader {
	return mustGetFile("testdata/workflows/workflow-service.sw.json")
}

func getWorkflowProperties() io.Reader {
	return mustGetFile("testdata/workflows/application.properties")
}

func getSpecOpenApi() io.Reader {
	return mustGetFile("testdata/workflows/specs/workflow-service-openapi.json")
}

func getSpecGeneric() io.Reader {
	return mustGetFile("testdata/workflows/specs/workflow-service-schema.json")
}

func mustGetFile(filepath string) io.Reader {
	file, err := os.OpenFile(filepath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return file
}
