// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sourcereader

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-getter"
)

// GetterSourceReader reads modules using the go-getter library
type GetterSourceReader struct{}

func copyModules(srcPath string, destPath string) error {
	client := getter.Client{
		Src:           srcPath,
		Dst:           destPath,
		Pwd:           destPath,
		Mode:          getter.ClientModeDir,
		Decompressors: map[string]getter.Decompressor{}, // none
		Ctx:           context.Background(),
	}
	err := client.Get()
	return err
}

// GetModule copies the source to a provided destination (the deployment directory)
func (r GetterSourceReader) GetModule(modPath string, copyPath string) error {
	modDir, err := ioutil.TempDir("", "module-*")
	defer os.RemoveAll(modDir)
	writeDir := filepath.Join(modDir, "mod")
	if err != nil {
		return err
	}

	if err := copyModules(modPath, writeDir); err != nil {
		return fmt.Errorf("failed to get module at %s to tmp dir %s: %v",
			modPath, writeDir, err)
	}

	return copyFromPath(writeDir, copyPath)
}
