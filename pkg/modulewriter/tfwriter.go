/**
* Copyright 2022 Google LLC
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*      http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package modulewriter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"hpc-toolkit/pkg/config"
	"hpc-toolkit/pkg/modulereader"
)

const (
	tfStateFileName       = "terraform.tfstate"
	tfStateBackupFileName = "terraform.tfstate.backup"
)

// TFWriter writes terraform to the blueprint folder
type TFWriter struct{}

func writeHclFile(path string, hclFile *hclwrite.File) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error writing %q: %v", path, err)
	}
	defer f.Close()
	if _, err := f.WriteString(license); err != nil {
		return fmt.Errorf("error writing %q: %v", path, err)
	}
	if _, err := f.Write(hclwrite.Format(hclFile.Bytes())); err != nil {
		return fmt.Errorf("error writing %q: %v", path, err)
	}
	return nil
}

func writeOutputs(
	modules []config.Module,
	dst string,
) error {
	// Create hcl body
	hclFile := hclwrite.NewEmptyFile()
	hclBody := hclFile.Body()

	outputs := []string{}
	// Add all outputs from each module
	for _, mod := range modules {
		for _, output := range mod.Outputs {
			outputName := config.AutomaticOutputName(output.Name, mod.ID)
			outputs = append(outputs, outputName)

			hclBody.AppendNewline()
			hclBlock := hclBody.AppendNewBlock("output", []string{outputName})
			blockBody := hclBlock.Body()

			desc := output.Description
			if desc == "" {
				desc = fmt.Sprintf("Generated output from module '%s'", mod.ID)
			}
			blockBody.SetAttributeValue("description", cty.StringVal(desc))
			ref := config.ModuleRef(mod.ID, output.Name).AsValue()
			blockBody.SetAttributeRaw("value", config.TokensForValue(ref))
			if output.Sensitive {
				blockBody.SetAttributeValue("sensitive", cty.BoolVal(output.Sensitive))
			}
		}
	}

	if len(outputs) == 0 {
		return nil
	}
	return writeHclFile(filepath.Join(dst, "outputs.tf"), hclFile)
}

func writeTfvars(vars map[string]cty.Value, dst string) error {
	return WriteHclAttributes(vars, filepath.Join(dst, "terraform.tfvars"))
}

func relaxVarType(t cty.Type) cty.Type {
	if t.IsPrimitiveType() {
		return t
	}
	if t.IsListType() || t.IsTupleType() || t.IsSetType() {
		return cty.List(cty.DynamicPseudoType) // list of any
	}
	return cty.DynamicPseudoType // any
}

func getTypeTokens(ty cty.Type) hclwrite.Tokens {
	// TODO: don't use TokensForIdentifier
	// This is a temporary solution until we have a better way to tokenize types
	return hclwrite.TokensForIdentifier(typeexpr.TypeString(ty))
}

func writeVariables(vars map[string]cty.Value, extraVars []modulereader.VarInfo, dst string) error {
	var inputs []modulereader.VarInfo
	for k, v := range vars {
		inputs = append(inputs, modulereader.VarInfo{
			Name:        k,
			Type:        relaxVarType(v.Type()),
			Description: fmt.Sprintf("Toolkit deployment variable: %s", k),
		})
	}
	inputs = append(inputs, extraVars...)
	slices.SortFunc(inputs, func(i, j modulereader.VarInfo) int { return strings.Compare(i.Name, j.Name) })

	// Create HCL Body
	hclFile := hclwrite.NewEmptyFile()
	hclBody := hclFile.Body()

	// create variable block for each input
	for _, k := range inputs {
		hclBody.AppendNewline()
		hclBlock := hclBody.AppendNewBlock("variable", []string{k.Name})
		blockBody := hclBlock.Body()
		blockBody.SetAttributeValue("description", cty.StringVal(k.Description))
		blockBody.SetAttributeRaw("type", getTypeTokens(k.Type))
	}

	return writeHclFile(filepath.Join(dst, "variables.tf"), hclFile)
}

func writeMain(
	modules []config.Module,
	locals config.Dict,
	tfBackend config.TerraformBackend,
	dst string,
) error {
	hclFile := hclwrite.NewEmptyFile()
	hclBody := hclFile.Body()

	// Write Terraform backend if needed
	if tfBackend.Type != "" {
		hclBody.AppendNewline()
		tfBody := hclBody.AppendNewBlock("terraform", []string{}).Body()
		backendBlock := tfBody.AppendNewBlock("backend", []string{tfBackend.Type})
		backendBody := backendBlock.Body()
		vals := tfBackend.Configuration.Items()
		for _, setting := range orderKeys(vals) {
			backendBody.SetAttributeValue(setting, vals[setting])
		}
	}

	if !locals.IsZero() {
		hclBody.AppendNewline()
		b := hclBody.AppendNewBlock("locals", []string{}).Body()
		for _, setting := range orderKeys(locals.Items()) {
			value := locals.Get(setting)
			b.SetAttributeRaw(setting, config.TokensForValue(value))
		}
	}

	for _, mod := range modules {
		hclBody.AppendNewline()
		// Add block
		moduleBlock := hclBody.AppendNewBlock("module", []string{string(mod.ID)})
		moduleBody := moduleBlock.Body()

		// Add source attribute
		ds, err := DeploymentSource(mod)
		if err != nil {
			return err
		}
		moduleBody.SetAttributeValue("source", cty.StringVal(ds))

		// For each Setting
		for _, setting := range orderKeys(mod.Settings.Items()) {
			value := mod.Settings.Get(setting)
			moduleBody.SetAttributeRaw(setting, config.TokensForValue(value))
		}
	}

	return writeHclFile(filepath.Join(dst, "main.tf"), hclFile)
}

func writeProviders(providers map[string]config.TerraformProvider, dst string) error {
	hclFile := hclwrite.NewEmptyFile()
	hclBody := hclFile.Body()

	for _, k := range orderKeys(providers) {
		hclBody.AppendNewline()
		v := providers[k]
		pb := hclBody.AppendNewBlock("provider", []string{k}).Body()

		for _, s := range orderKeys(v.Configuration.Items()) {
			pb.SetAttributeRaw(s, config.TokensForValue(v.Configuration.Get(s)))
		}
	}
	return writeHclFile(filepath.Join(dst, "providers.tf"), hclFile)
}

func writeVersions(providers map[string]config.TerraformProvider, dst string) error {
	f := hclwrite.NewEmptyFile()
	body := f.Body()
	body.AppendNewline()
	tfb := body.AppendNewBlock("terraform", []string{}).Body()
	tfb.SetAttributeValue("required_version", cty.StringVal(">= 1.2"))
	tfb.AppendNewline()

	pb := tfb.AppendNewBlock("required_providers", []string{}).Body()

	for _, k := range orderKeys(providers) {
		v := providers[k]
		pb.SetAttributeValue(k, cty.ObjectVal(map[string]cty.Value{
			"source":  cty.StringVal(v.Source),
			"version": cty.StringVal(v.Version),
		}))
	}
	return writeHclFile(filepath.Join(dst, "versions.tf"), f)
}

func writeTerraformInstructions(w io.Writer, grpPath string, n config.GroupName, printExportOutputs bool, printImportInputs bool) {
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Terraform group '%s' was successfully created in directory %s\n", n, grpPath)
	fmt.Fprintln(w, "To deploy, run the following commands:")
	fmt.Fprintln(w)
	if printImportInputs {
		fmt.Fprintf(w, "gcluster import-inputs %s\n", grpPath)
	}
	fmt.Fprintf(w, "terraform -chdir=%s init\n", grpPath)
	fmt.Fprintf(w, "terraform -chdir=%s validate\n", grpPath)
	fmt.Fprintf(w, "terraform -chdir=%s apply\n", grpPath)
	if printExportOutputs {
		fmt.Fprintf(w, "gcluster export-outputs %s\n", grpPath)
	}
}

// writeGroup creates and sets up the terraform deployment group
func (w TFWriter) writeGroup(
	bp config.Blueprint,
	groupIndex int,
	groupPath string,
	instructions io.Writer,
) error {
	g := bp.Groups[groupIndex]
	deploymentVars := getUsedDeploymentVars(g, bp)

	intergroupVars := FindIntergroupVariables(g, bp)
	intergroupInputs := make(map[string]bool)
	for _, igVar := range intergroupVars {
		intergroupInputs[igVar.Name] = true
	}

	tp := g.TerraformProviders

	// Write main.tf file
	doctoredModules, err := substituteIgcReferences(g.Modules, intergroupVars)
	if err != nil {
		return fmt.Errorf("error substituting intergroup references in deployment group %s: %w", g.Name, err)
	}
	if err := writeMain(doctoredModules, g.Locals, g.TerraformBackend, groupPath); err != nil {
		return fmt.Errorf("error writing main.tf file for deployment group %s: %w", g.Name, err)
	}

	// Write variables.tf file
	if err := writeVariables(deploymentVars, maps.Values(intergroupVars), groupPath); err != nil {
		return fmt.Errorf("error writing variables.tf file for deployment group %s: %w", g.Name, err)
	}

	// Write outputs.tf file
	if err := writeOutputs(g.Modules, groupPath); err != nil {
		return fmt.Errorf("error writing outputs.tf file for deployment group %s: %w", g.Name, err)
	}

	// Write terraform.tfvars file
	if err := writeTfvars(deploymentVars, groupPath); err != nil {
		return fmt.Errorf("error writing terraform.tfvars file for deployment group %s: %w", g.Name, err)
	}

	// Write providers.tf file
	if err := writeProviders(tp, groupPath); err != nil {
		return fmt.Errorf("error writing providers.tf file for deployment group %s: %w", g.Name, err)
	}

	// Write versions.tf file
	if err := writeVersions(tp, groupPath); err != nil {
		return fmt.Errorf("error writing versions.tf file for deployment group %s: %v", g.Name, err)
	}

	multiGroupDeployment := len(bp.Groups) > 1
	printImportInputs := multiGroupDeployment && groupIndex > 0
	printExportOutputs := multiGroupDeployment && groupIndex < len(bp.Groups)-1

	writeTerraformInstructions(instructions, groupPath, g.Name, printExportOutputs, printImportInputs)

	return nil
}

// Transfers state files from previous resource groups (in .ghpc/) to a newly written blueprint
func (w TFWriter) restoreState(deploymentDir string) error {
	prevGroupPath := filepath.Join(HiddenGhpcDir(deploymentDir), prevGroupDirName)
	files, err := os.ReadDir(prevGroupPath)
	if err != nil {
		return fmt.Errorf("error trying to read previous modules in %s, %w", prevGroupPath, err)
	}

	for _, f := range files {
		var tfStateFiles = []string{tfStateFileName, tfStateBackupFileName}
		for _, stateFile := range tfStateFiles {
			src := filepath.Join(prevGroupPath, f.Name(), stateFile)
			dest := filepath.Join(deploymentDir, f.Name(), stateFile)

			if bytesRead, err := os.ReadFile(src); err == nil {
				err = os.WriteFile(dest, bytesRead, 0644)
				if err != nil {
					return fmt.Errorf("failed to write previous state file %s, %w", dest, err)
				}
			}
		}

	}
	return nil
}

func orderKeys[T any](settings map[string]T) []string {
	keys := make([]string, 0, len(settings))
	for k := range settings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func getUsedDeploymentVars(group config.Group, bp config.Blueprint) map[string]cty.Value {
	res := map[string]cty.Value{
		// labels must always be written as a variable as it is implicitly added
		"labels": bp.Vars.Get("labels"),
	}

	used := []string{}
	for _, m := range group.Modules {
		used = append(used, config.GetUsedDeploymentVars(m.Settings.AsObject())...)
	}
	for _, v := range group.TerraformProviders {
		used = append(used, config.GetUsedDeploymentVars(v.Configuration.AsObject())...)
	}
	for _, v := range used {
		res[v] = bp.Vars.Get(v)
	}
	return res
}

func substituteIgcReferences(mods []config.Module, igcRefs map[config.Reference]modulereader.VarInfo) ([]config.Module, error) {
	doctoredMods := make([]config.Module, len(mods))
	for i, mod := range mods {
		dm, err := SubstituteIgcReferencesInModule(mod, igcRefs)
		if err != nil {
			return nil, err
		}
		doctoredMods[i] = dm
	}
	return doctoredMods, nil
}

// SubstituteIgcReferencesInModule updates expressions in Module settings to use
// special IGC var name instead of the module reference
func SubstituteIgcReferencesInModule(mod config.Module, igcRefs map[config.Reference]modulereader.VarInfo) (config.Module, error) {
	v, err := cty.Transform(mod.Settings.AsObject(), func(p cty.Path, v cty.Value) (cty.Value, error) {
		e, is := config.IsExpressionValue(v)
		if !is {
			return v, nil
		}
		refs := e.References()
		for _, r := range refs {
			oi, exists := igcRefs[r]
			if !exists {
				continue
			}
			old := r.AsExpression()
			new := config.GlobalRef(oi.Name).AsExpression()
			var err error
			if e, err = config.ReplaceSubExpressions(e, old, new); err != nil {
				return cty.NilVal, err
			}
		}
		return e.AsValue(), nil
	})
	if err != nil {
		return config.Module{}, err
	}
	mod.Settings = config.NewDict(v.AsValueMap())
	return mod, nil
}

// FindIntergroupVariables returns all unique intergroup references made by
// each module settings in a group
func FindIntergroupVariables(group config.Group, bp config.Blueprint) map[config.Reference]modulereader.VarInfo {
	res := map[config.Reference]modulereader.VarInfo{}
	igcRefs := group.FindAllIntergroupReferences(bp)
	for _, r := range igcRefs {
		n := config.AutomaticOutputName(r.Name, r.Module)
		res[r] = modulereader.VarInfo{
			Name:        n,
			Type:        cty.DynamicPseudoType,
			Description: "Automatically generated input from previous groups (gcluster import-inputs --help)",
			Required:    true,
		}
	}
	return res
}

func (w TFWriter) kind() config.ModuleKind {
	return config.TerraformKind
}
