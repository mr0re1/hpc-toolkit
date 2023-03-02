package main

import (
	"fmt"
	"hpc-toolkit/pkg/modulereader"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func getMod(src string) modulereader.ModuleInfo {
	info, err := modulereader.GetModuleInfo("./"+src, "terraform")
	if err != nil {
		panic(err)
	}
	return info
}

func getCtyType(src string) (cty.Type, error) {
	expr, diags := hclsyntax.ParseExpression([]byte(src), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return cty.Type{}, diags
	}
	typ, diags := typeexpr.TypeConstraint(expr)
	if diags.HasErrors() {
		return cty.Type{}, diags
	}
	return typ, nil
}

func printVar(mod string, t string, v modulereader.VarInfo) {
	typ := ""
	if v.Type != "" {
		ctyTyp, err := getCtyType(v.Type)
		if err != nil {
			log.Fatalf("Failed to parse %s : %v", v.Type, err)
		}
		typ = typeexpr.TypeString(ctyTyp)
	}
	fmt.Printf("%s\t%s\t%s\t%s\n", mod, t, v.Name, typ)
}

func printInfo(mod string, info modulereader.ModuleInfo) {
	for _, inp := range info.Inputs {
		printVar(mod, "I", inp)
	}
	for _, out := range info.Outputs {
		printVar(mod, "O", out)
	}
}

/*
func main() {
	typ1, err := getCtyType("object({name=string, count=number})")
	if err != nil {
		panic(err)
	}
	typ2, err := getCtyType("object({count=number, name=string})")
	if err != nil {
		panic(err)
	}
	fmt.Println(typeexpr.TypeString(typ1))
	fmt.Println(typeexpr.TypeString(typ2))

}*/

func main() {
	mods := []string{
		"modules/scripts/startup-script",
		"modules/network/pre-existing-vpc",
		"modules/network/vpc",
		"modules/file-system/filestore",
		"modules/file-system/pre-existing-network-storage",
		"modules/compute/vm-instance",
		"modules/monitoring/dashboard",
		"modules/scheduler/batch-login-node",
		"modules/scheduler/batch-job-template",

		"community/modules/scripts/htcondor-install",
		"community/modules/scripts/pbspro-install",
		"community/modules/scripts/pbspro-preinstall",
		"community/modules/scripts/spack-install",
		"community/modules/scripts/omnia-install",
		"community/modules/scripts/wait-for-startup",
		"community/modules/scripts/pbspro-qmgr",
		"community/modules/file-system/nfs-server",
		"community/modules/file-system/cloud-storage-bucket",
		"community/modules/file-system/DDN-EXAScaler",
		"community/modules/compute/htcondor-execute-point",
		"community/modules/compute/schedmd-slurm-gcp-v5-partition",
		"community/modules/compute/pbspro-execution",
		"community/modules/compute/schedmd-slurm-gcp-v5-node-group",
		"community/modules/compute/SchedMD-slurm-on-gcp-partition",
		"community/modules/remote-desktop/chrome-remote-desktop",
		"community/modules/database/slurm-cloudsql-federation",
		"community/modules/scheduler/schedmd-slurm-gcp-v5-controller",
		"community/modules/scheduler/SchedMD-slurm-on-gcp-controller",
		"community/modules/scheduler/schedmd-slurm-gcp-v5-hybrid",
		"community/modules/scheduler/SchedMD-slurm-on-gcp-login-node",
		"community/modules/scheduler/schedmd-slurm-gcp-v5-login",
		"community/modules/scheduler/pbspro-server",
		"community/modules/scheduler/htcondor-configure",
		"community/modules/scheduler/pbspro-client",
		"community/modules/project/service-enablement",
		"community/modules/project/service-account",
		"community/modules/project/new-project",
	}
	for _, mod := range mods {
		inf := getMod(mod)
		printInfo(mod, inf)
	}
}
